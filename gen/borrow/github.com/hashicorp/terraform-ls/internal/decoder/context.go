// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package decoder

import (
	"context"

	ilsp "github.com/walteh/retab/gen/borrow/github.com/hashicorp/terraform-ls/internal/lsp"
)

type languageIdCtxKey struct{}

func WithLanguageId(ctx context.Context, langId ilsp.LanguageID) context.Context {
	return context.WithValue(ctx, languageIdCtxKey{}, langId)
}

func LanguageId(ctx context.Context) (ilsp.LanguageID, bool) {
	id, ok := ctx.Value(languageIdCtxKey{}).(ilsp.LanguageID)
	if !ok {
		return "", false
	}
	return id, true
}
