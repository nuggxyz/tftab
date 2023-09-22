// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"fmt"
	"testing"

	"github.com/walteh/retab/internal/lsp/langserver"
	"github.com/walteh/retab/internal/lsp/langserver/session"
)

func TestCompletionResolve_withoutInitialization(t *testing.T) {
	ls := langserver.NewLangServerMock(t, NewMockSession(nil))
	stop := ls.Start(t)
	defer stop()

	ls.CallAndExpectError(t, &langserver.CallRequest{
		Method:    "completionItem/resolve",
		ReqParams: "{}"}, session.SessionNotInitialized.Err())
}

func TestCompletionResolve_withoutHook(t *testing.T) {
	tmpDir := TempDir(t)
	InitPluginCache(t, tmpDir.Path())

	ls := langserver.NewLangServerMock(t, NewMockSession(nil))
	stop := ls.Start(t)
	defer stop()

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {},
		"rootUri": %q,
		"processId": 12345
	}`, tmpDir.URI)})

	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})

	ls.CallAndExpectResponse(t,
		&langserver.CallRequest{
			Method: "completionItem/resolve",
			ReqParams: `{
			"label": "\"test\"",
			"kind": 1,
			"data": {
				"resolve_hook": "test",
				"path": "` + TempDir(t).URI + `/main.tf"
			}}`,
		}, `{
			"jsonrpc": "2.0",
			"id": 2,
			"result": {
				"label": "\"test\"",
				"kind": 1,
				"data": {
					"resolve_hook": "test",
					"path": "`+TempDir(t).URI+`/main.tf"
				}
			}
	}`)
}
