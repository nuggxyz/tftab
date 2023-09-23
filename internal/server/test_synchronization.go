// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/walteh/retab/gen/gopls/event"
	"github.com/walteh/retab/gen/gopls/event/tag"
	"github.com/walteh/retab/gen/gopls/jsonrpc2"
	"github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/gen/gopls/span"
	"github.com/walteh/retab/internal/source"
	// "golang.org/x/tools/gopls/internal/lsp/protocol"
	// "golang.org/x/tools/gopls/internal/lsp/source"
	// "golang.org/x/tools/gopls/internal/span"
	// "golang.org/x/tools/internal/event"
	// "golang.org/x/tools/internal/event/tag"
	// "golang.org/x/tools/internal/jsonrpc2"
)

// ModificationSource identifies the origin of a change.
type ModificationSource int

const (
	// FromDidOpen is from a didOpen notification.
	FromDidOpen = ModificationSource(iota)

	// FromDidChange is from a didChange notification.
	FromDidChange

	// FromDidChangeWatchedFiles is from didChangeWatchedFiles notification.
	FromDidChangeWatchedFiles

	// FromDidSave is from a didSave notification.
	FromDidSave

	// FromDidClose is from a didClose notification.
	FromDidClose

	// FromDidChangeConfiguration is from a didChangeConfiguration notification.
	FromDidChangeConfiguration

	// FromRegenerateCgo refers to file modifications caused by regenerating
	// the cgo sources for the workspace.
	FromRegenerateCgo

	// FromInitialWorkspaceLoad refers to the loading of all packages in the
	// workspace when the view is first created.
	FromInitialWorkspaceLoad
)

func (m ModificationSource) String() string {
	switch m {
	case FromDidOpen:
		return "opened files"
	case FromDidChange:
		return "changed files"
	case FromDidChangeWatchedFiles:
		return "files changed on disk"
	case FromDidSave:
		return "saved files"
	case FromDidClose:
		return "close files"
	case FromRegenerateCgo:
		return "regenerate cgo"
	case FromInitialWorkspaceLoad:
		return "initial workspace load"
	default:
		return "unknown file modification"
	}
}

func (s *Server) didOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	ctx, done := event.Start(ctx, "lsp.Server.didOpen", tag.URI.Of(params.TextDocument.URI))
	defer done()

	uri := params.TextDocument.URI.SpanURI()
	if !uri.IsFile() {
		return nil
	}
	// There may not be any matching view in the current session. If that's
	// the case, try creating a new view based on the opened file path.
	//
	// TODO(rstambler): This seems like it would continuously add new
	// views, but it won't because ViewOf only returns an error when there
	// are no views in the session. I don't know if that logic should go
	// here, or if we can continue to rely on that implementation detail.
	//
	// TODO(golang/go#57979): this will be generalized to a different view calculation.
	if _, err := s.session.ViewOf(uri); err != nil {
		dir := filepath.Dir(uri.Filename())
		if err := s.addFolders(ctx, []protocol.WorkspaceFolder{{
			URI:  string(protocol.URIFromPath(dir)),
			Name: filepath.Base(dir),
		}}); err != nil {
			return err
		}
	}
	return s.didModifyFiles(ctx, []source.FileModification{{
		URI:        uri,
		Action:     source.Open,
		Version:    params.TextDocument.Version,
		Text:       []byte(params.TextDocument.Text),
		LanguageID: params.TextDocument.LanguageID,
	}}, FromDidOpen)
}

func (s *Server) didChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	ctx, done := event.Start(ctx, "lsp.Server.didChange", tag.URI.Of(params.TextDocument.URI))
	defer done()

	uri := params.TextDocument.URI.SpanURI()
	if !uri.IsFile() {
		return nil
	}

	text, err := s.changedText(ctx, uri, params.ContentChanges)
	if err != nil {
		return err
	}
	c := source.FileModification{
		URI:     uri,
		Action:  source.Change,
		Version: params.TextDocument.Version,
		Text:    text,
	}
	if err := s.didModifyFiles(ctx, []source.FileModification{c}, FromDidChange); err != nil {
		return err
	}
	return s.warnAboutModifyingGeneratedFiles(ctx, uri)
}

// warnAboutModifyingGeneratedFiles shows a warning if a user tries to edit a
// generated file for the first time.
func (s *Server) warnAboutModifyingGeneratedFiles(ctx context.Context, uri span.URI) error {
	s.changedFilesMu.Lock()
	_, ok := s.changedFiles[uri]
	if !ok {
		s.changedFiles[uri] = struct{}{}
	}
	s.changedFilesMu.Unlock()

	// This file has already been edited before.
	if ok {
		return nil
	}

	// Ideally, we should be able to specify that a generated file should
	// be opened as read-only. Tell the user that they should not be
	// editing a generated file.
	view, err := s.session.ViewOf(uri)
	if err != nil {
		return err
	}
	snapshot, release, err := view.Snapshot()
	if err != nil {
		return err
	}
	isGenerated := source.IsGenerated(ctx, snapshot, uri)
	release()

	if !isGenerated {
		return nil
	}
	return s.client.ShowMessage(ctx, &protocol.ShowMessageParams{
		Message: fmt.Sprintf("Do not edit this file! %s is a generated file.", uri.Filename()),
		Type:    protocol.Warning,
	})
}

func (s *Server) didChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
	ctx, done := event.Start(ctx, "lsp.Server.didChangeWatchedFiles")
	defer done()

	var modifications []source.FileModification
	for _, change := range params.Changes {
		uri := change.URI.SpanURI()
		if !uri.IsFile() {
			continue
		}
		action := changeTypeToFileAction(change.Type)
		modifications = append(modifications, source.FileModification{
			URI:    uri,
			Action: action,
			OnDisk: true,
		})
	}
	return s.didModifyFiles(ctx, modifications, FromDidChangeWatchedFiles)
}

func (s *Server) didSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	ctx, done := event.Start(ctx, "lsp.Server.didSave", tag.URI.Of(params.TextDocument.URI))
	defer done()

	uri := params.TextDocument.URI.SpanURI()
	if !uri.IsFile() {
		return nil
	}
	c := source.FileModification{
		URI:    uri,
		Action: source.Save,
	}
	if params.Text != nil {
		c.Text = []byte(*params.Text)
	}
	return s.didModifyFiles(ctx, []source.FileModification{c}, FromDidSave)
}

func (s *Server) didClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	ctx, done := event.Start(ctx, "lsp.Server.didClose", tag.URI.Of(params.TextDocument.URI))
	defer done()

	uri := params.TextDocument.URI.SpanURI()
	if !uri.IsFile() {
		return nil
	}
	return s.didModifyFiles(ctx, []source.FileModification{
		{
			URI:     uri,
			Action:  source.Close,
			Version: -1,
			Text:    nil,
		},
	}, FromDidClose)
}

func (s *Server) didModifyFiles(ctx context.Context, modifications []source.FileModification, cause ModificationSource) error {
	// wg guards two conditions:
	//  1. didModifyFiles is complete
	//  2. the goroutine diagnosing changes on behalf of didModifyFiles is
	//     complete, if it was started
	//
	// Both conditions must be satisfied for the purpose of testing: we don't
	// want to observe the completion of change processing until we have received
	// all diagnostics as well as all server->client notifications done on behalf
	// of this function.
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Done()

	if s.Options().VerboseWorkDoneProgress {
		work := s.progress.Start(ctx, DiagnosticWorkTitle(cause), "Calculating file diagnostics...", nil, nil)
		go func() {
			wg.Wait()
			work.End(ctx, "Done.")
		}()
	}

	onDisk := cause == FromDidChangeWatchedFiles

	s.stateMu.Lock()
	if s.state >= serverShutDown {
		// This state check does not prevent races below, and exists only to
		// produce a better error message. The actual race to the cache should be
		// guarded by Session.viewMu.
		s.stateMu.Unlock()
		return errors.New("server is shut down")
	}
	s.stateMu.Unlock()

	// If the set of changes included directories, expand those directories
	// to their files.
	modifications = s.session.ExpandModificationsToDirectories(ctx, modifications)

	// Build a lookup map for file modifications, so that we can later join
	// with the snapshot file associations.
	modMap := make(map[span.URI]source.FileModification)
	for _, mod := range modifications {
		modMap[mod.URI] = mod
	}

	snapshots, release, err := s.session.DidModifyFiles(ctx, modifications)
	if err != nil {
		return err
	}

	// golang/go#50267: diagnostics should be re-sent after an open or close. For
	// some clients, it may be helpful to re-send after each change.
	for snapshot, uris := range snapshots {
		for _, uri := range uris {
			mod := modMap[uri]
			if snapshot.Options().ChattyDiagnostics || mod.Action == source.Open || mod.Action == source.Close {
				s.mustPublishDiagnostics(uri)
			}
		}
	}

	wg.Add(1)
	go func() {
		s.diagnoseSnapshots(snapshots, onDisk)
		release()
		wg.Done()
	}()

	// After any file modifications, we need to update our watched files,
	// in case something changed. Compute the new set of directories to watch,
	// and if it differs from the current set, send updated registrations.
	return s.updateWatchedDirectories(ctx)
}

// DiagnosticWorkTitle returns the title of the diagnostic work resulting from a
// file change originating from the given cause.
func DiagnosticWorkTitle(cause ModificationSource) string {
	return fmt.Sprintf("diagnosing %v", cause)
}

func (s *Server) changedText(ctx context.Context, uri span.URI, changes []protocol.TextDocumentContentChangeEvent) ([]byte, error) {
	if len(changes) == 0 {
		return nil, fmt.Errorf("%w: no content changes provided", jsonrpc2.ErrInternal)
	}

	// Check if the client sent the full content of the file.
	// We accept a full content change even if the server expected incremental changes.
	if len(changes) == 1 && changes[0].Range == nil && changes[0].RangeLength == 0 {
		return []byte(changes[0].Text), nil
	}
	return s.applyIncrementalChanges(ctx, uri, changes)
}

func (s *Server) applyIncrementalChanges(ctx context.Context, uri span.URI, changes []protocol.TextDocumentContentChangeEvent) ([]byte, error) {
	fh, err := s.session.ReadFile(ctx, uri)
	if err != nil {
		return nil, err
	}
	content, err := fh.Content()
	if err != nil {
		return nil, fmt.Errorf("%w: file not found (%v)", jsonrpc2.ErrInternal, err)
	}
	for _, change := range changes {
		// TODO(adonovan): refactor to use diff.Apply, which is robust w.r.t.
		// out-of-order or overlapping changes---and much more efficient.

		// Make sure to update mapper along with the content.
		m := protocol.NewMapper(uri, content)
		if change.Range == nil {
			return nil, fmt.Errorf("%w: unexpected nil range for change", jsonrpc2.ErrInternal)
		}
		spn, err := m.RangeSpan(*change.Range)
		if err != nil {
			return nil, err
		}
		start, end := spn.Start().Offset(), spn.End().Offset()
		if end < start {
			return nil, fmt.Errorf("%w: invalid range for content change", jsonrpc2.ErrInternal)
		}
		var buf bytes.Buffer
		buf.Write(content[:start])
		buf.WriteString(change.Text)
		buf.Write(content[end:])
		content = buf.Bytes()
	}
	return content, nil
}

func changeTypeToFileAction(ct protocol.FileChangeType) source.FileAction {
	switch ct {
	case protocol.Changed:
		return source.Change
	case protocol.Created:
		return source.Create
	case protocol.Deleted:
		return source.Delete
	}
	return source.UnknownFileAction
}
