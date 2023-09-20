// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lsp

import (
	"github.com/hashicorp/hcl-lang/lang"
	lsp "github.com/walteh/retab/gen/gopls"
)

func Command(cmd lang.Command) (lsp.Command, error) {
	lspCmd := lsp.Command{
		Title:   cmd.Title,
		Command: cmd.ID,
	}

	for _, arg := range cmd.Arguments {
		jsonArg, err := arg.MarshalJSON()
		if err != nil {
			return lspCmd, err
		}
		lspCmd.Arguments = append(lspCmd.Arguments, jsonArg)
	}

	return lspCmd, nil
}