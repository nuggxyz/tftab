// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"encoding/json"
	"fmt"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/walteh/retab/internal/lsp/langserver"
)

func TestSemanticTokensFull(t *testing.T) {
	tmpDir := TempDir(t)

	var testSchema tfjson.ProviderSchemas
	err := json.Unmarshal([]byte(testModuleSchemaOutput), &testSchema)
	if err != nil {
		t.Fatal(err)
	}

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {
			"textDocument": {
				"semanticTokens": {
					"tokenTypes": [
						"enumMember",
						"property",
						"string",
						"type"
					],
					"tokenModifiers": [
						"defaultLibrary",
						"deprecated"
					],
					"requests": {
						"full": true
					}
				}
			}
		},
		"rootUri": %q,
		"processId": 12345
	}`, tmpDir.URI)})

	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didOpen",
		ReqParams: fmt.Sprintf(`{
		"textDocument": {
			"version": 0,
			"languageId": "retab",
			"text": "provider \"test\" {\n\n}\n",
			"uri": "%s/main.tf"
		}
	}`, tmpDir.URI)})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "textDocument/semanticTokens/full",
		ReqParams: fmt.Sprintf(`{
			"textDocument": {
				"uri": "%s/main.tf"
			}
		}`, tmpDir.URI)}, `{
			"jsonrpc": "2.0",
			"id": 3,
			"result": {
				"data": [
					0,0,8,3,0,
					0,9,6,0,1
				]
			}
		}`)
}

func TestSemanticTokensFull_clientSupportsDelta(t *testing.T) {
	tmpDir := TempDir(t)

	var testSchema tfjson.ProviderSchemas
	err := json.Unmarshal([]byte(testModuleSchemaOutput), &testSchema)
	if err != nil {
		t.Fatal(err)
	}

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {
			"textDocument": {
				"semanticTokens": {
					"tokenTypes": [
						"enumMember",
						"property",
						"string",
						"type"
					],
					"tokenModifiers": [
						"defaultLibrary",
						"deprecated"
					],
					"requests": {
						"full": {
							"delta": true
						}
					}
				}
			}
		},
		"rootUri": %q,
		"processId": 12345
	}`, tmpDir.URI)})

	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didOpen",
		ReqParams: fmt.Sprintf(`{
		"textDocument": {
			"version": 0,
			"languageId": "retab",
			"text": "provider \"test\" {\n\n}\n",
			"uri": "%s/main.tf"
		}
	}`, tmpDir.URI)})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "textDocument/semanticTokens/full",
		ReqParams: fmt.Sprintf(`{
			"textDocument": {
				"uri": "%s/main.tf"
			}
		}`, tmpDir.URI)}, `{
			"jsonrpc": "2.0",
			"id": 3,
			"result": {
				"data": [
					0,0,8,3,0,
					0,9,6,0,1
				]
			}
		}`)
}

func TestVarsSemanticTokensFull(t *testing.T) {
	tmpDir := TempDir(t)

	var testSchema tfjson.ProviderSchemas
	err := json.Unmarshal([]byte(testModuleSchemaOutput), &testSchema)
	if err != nil {
		t.Fatal(err)
	}

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {
			"textDocument": {
				"semanticTokens": {
					"tokenTypes": [
						"type",
						"property",
						"string"
					],
					"tokenModifiers": [
						"defaultLibrary",
						"deprecated"
					],
					"requests": {
						"full": true
					}
				}
			}
		},
		"rootUri": %q,
		"processId": 12345
	}`, tmpDir.URI)})

	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didOpen",
		ReqParams: fmt.Sprintf(`{
		"textDocument": {
			"version": 0,
			"languageId": "retab",
			"text": "variable \"test\" {\n type=string\n}\n",
			"uri": "%s/variables.tf"
		}
	}`, tmpDir.URI)})
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didOpen",
		ReqParams: fmt.Sprintf(`{
			"textDocument": {
				"version": 0,
				"languageId": "terraform-vars",
				"text": "test = \"dev\"\n",
				"uri": "%s/terraform.tfvars"
			}
	}`, tmpDir.URI)})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "textDocument/semanticTokens/full",
		ReqParams: fmt.Sprintf(`{
			"textDocument": {
				"uri": "%s/terraform.tfvars"
			}
		}`, tmpDir.URI)}, `{
			"jsonrpc": "2.0",
			"id": 4,
			"result": {
				"data": [
					0,0,4,0,0,
					0,7,5,1,0
				]
			}
		}`)
}

func TestVarsSemanticTokensFull_functionToken(t *testing.T) {
	tmpDir := TempDir(t)

	var testSchema tfjson.ProviderSchemas
	err := json.Unmarshal([]byte(testModuleSchemaOutput), &testSchema)
	if err != nil {
		t.Fatal(err)
	}

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{}))
	stop := ls.Start(t)
	defer stop()

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
		"capabilities": {
			"textDocument": {
				"semanticTokens": {
					"tokenTypes": [
						"type",
						"property",
						"string",
						"function"
					],
					"tokenModifiers": [
						"defaultLibrary",
						"deprecated"
					],
					"requests": {
						"full": true
					}
				}
			}
		},
		"rootUri": %q,
		"processId": 12345
	}`, tmpDir.URI)})

	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didOpen",
		ReqParams: fmt.Sprintf(`{
		"textDocument": {
			"version": 0,
			"languageId": "retab",
			"text": "locals {\n  foo = abs(-42)\n}\n",
			"uri": "%s/locals.tf"
		}
	}`, tmpDir.URI)})

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "textDocument/semanticTokens/full",
		ReqParams: fmt.Sprintf(`{
			"textDocument": {
				"uri": "%s/locals.tf"
			}
		}`, tmpDir.URI)}, `{
			"jsonrpc": "2.0",
			"id": 3,
			"result": {
				"data": [
					0,0,6,3,0,
					1,2,3,1,0,
					0,6,3,0,0
				]
			}
		}`)
}
