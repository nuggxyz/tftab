// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	lsp "github.com/walteh/retab/gen/gopls"
	lsctx "github.com/walteh/retab/internal/lsp/context"
	ilsp "github.com/walteh/retab/internal/lsp/lsp"
)

func (svc *service) TextDocumentComplete(ctx context.Context, params lsp.CompletionParams) (lsp.CompletionList, error) {
	var list lsp.CompletionList

	cc, err := ilsp.ClientCapabilities(ctx)
	if err != nil {
		return list, err
	}

	dh := ilsp.HandleFromDocumentURI(params.TextDocument.URI)
	doc, err := svc.stateStore.DocumentStore.GetDocument(dh)
	if err != nil {
		return list, err
	}

	d, err := svc.decoderForDocument(ctx, doc)
	if err != nil {
		return list, err
	}

	expFeatures, err := lsctx.ExperimentalFeatures(ctx)
	if err != nil {
		return list, err
	}

	d.PrefillRequiredFields = expFeatures.PrefillRequiredFields

	pos, err := ilsp.HCLPositionFromLspPosition(params.TextDocumentPositionParams.Position, doc)
	if err != nil {
		return list, err
	}

	svc.logger.Printf("Looking for candidates at %q -> %#v", doc.Filename, pos)
	candidates, err := d.CandidatesAtPos(ctx, doc.Filename, pos)
	svc.logger.Printf("received candidates: %#v", candidates)
	return ilsp.ToCompletionList(candidates, cc.TextDocument), err
}