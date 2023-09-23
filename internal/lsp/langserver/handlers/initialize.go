// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"context"

	"github.com/creachadair/jrpc2"
	"github.com/walteh/retab/gen/gopls"
	lsctx "github.com/walteh/retab/internal/lsp/context"
	"github.com/walteh/retab/internal/lsp/lsp"
	"github.com/walteh/retab/internal/lsp/uri"
	"github.com/walteh/retab/internal/protocol"
)

func (svc *service) Initialize(ctx context.Context, params gopls.InitializeParams) (gopls.InitializeResult, error) {
	serverCaps := initializeResult(ctx)

	var err error

	properties := map[string]interface{}{}
	// properties["lsVersion"] = serverCaps.ServerInfo.Version

	clientCaps := params.Capabilities

	// svc.server = jrpc2.ServerFromContext(ctx)

	// setupTelemetry(expClientCaps, svc, ctx, properties)
	// defer svc.telemetry.SendEvent(ctx, "initialize", properties)

	if params.ClientInfo != nil && params.ClientInfo.Name != "" {
		err = lsp.SetClientName(ctx, params.ClientInfo.Name)
		if err != nil {
			return serverCaps, err
		}
	}

	expServerCaps := protocol.ExperimentalServerCapabilities{}

	serverCaps.Capabilities.Experimental = expServerCaps

	err = lsp.SetClientCapabilities(ctx, &clientCaps)
	if err != nil {
		return serverCaps, err
	}

	//////////////////////// configureClientCapabilities ////////////////////////

	// svc.diagsNotifier = diagnostics.NewNotifier(svc.server, svc.logger)

	// moduleHooks := []notifier.Hook{
	// 	updateDiagnostics(svc.diagsNotifier),
	// 	sendModuleTelemetry(svc.telemetry),
	// }

	// cc, err := lsp.ClientCapabilities(ctx)
	// if err == nil {
	// 	if cc.Workspace.SemanticTokens != nil && cc.Workspace.SemanticTokens.RefreshSupport {
	// 		moduleHooks = append(moduleHooks, refreshSemanticTokens(svc.server))
	// 	}
	// }

	// svc.notifier = notifier.NewNotifier(moduleHooks)
	// svc.notifier.SetLogger(svc.logger)
	// svc.notifier.Start(svc.sessCtx)

	//////////////////////// configureClientCapabilities ////////////////////////

	stCaps := clientCaps.TextDocument.SemanticTokens
	caps := lsp.SemanticTokensClientCapabilities{
		SemanticTokensClientCapabilities: clientCaps.TextDocument.SemanticTokens,
	}
	semanticTokensOpts := gopls.SemanticTokensOptions{
		Legend: gopls.SemanticTokensLegend{
			TokenTypes:     lsp.TokenTypesLegend(stCaps.TokenTypes).AsStrings(),
			TokenModifiers: lsp.TokenModifiersLegend(stCaps.TokenModifiers).AsStrings(),
		},
		Full: &gopls.Or_SemanticTokensOptions_full{
			Value: caps.FullRequest(),
		},
	}

	serverCaps.Capabilities.SemanticTokensProvider = semanticTokensOpts

	if !clientCaps.Workspace.WorkspaceFolders && len(params.WorkspaceFolders) > 0 {
		jrpc2.ServerFromContext(ctx).Notify(ctx, "window/showMessage", &gopls.ShowMessageParams{
			Type: gopls.Warning,
			Message: "Client sent workspace folders despite not declaring support. " +
				"Please report this as a bug.",
		})
	}

	if params.RootURI == "" {
		svc.singleFileMode = true
		properties["root_uri"] = "file"
		if properties["options.ignoreSingleFileWarning"] == false {
			jrpc2.ServerFromContext(ctx).Notify(ctx, "window/showMessage", &gopls.ShowMessageParams{
				Type:    gopls.Warning,
				Message: "Some capabilities may be reduced when editing a single file. We recommend opening a directory for full functionality. Use 'ignoreSingleFileWarning' to suppress this warning.",
			})
		}
	} else {
		rootURI := string(params.RootURI)

		invalidUriErr := jrpc2.Errorf(jrpc2.InvalidParams,
			"Unsupported or invalid URI: %q "+
				"This is most likely client bug, please report it.", rootURI)

		if uri.IsWSLURI(rootURI) {
			properties["root_uri"] = "invalid"
			// For WSL URIs we return additional error data
			// such that clients (e.g. VS Code) can provide better UX
			// and nudge users to open in the WSL Remote Extension instead.
			return serverCaps, invalidUriErr.WithData("INVALID_URI_WSL")
		}

		if !uri.IsURIValid(rootURI) {
			properties["root_uri"] = "invalid"

			return serverCaps, invalidUriErr
		}

	}

	return serverCaps, err
}

func setupTelemetry(expClientCaps protocol.ExpClientCapabilities, svc *service, ctx context.Context, properties map[string]interface{}) {
	if tv, ok := expClientCaps.TelemetryVersion(); ok {
		svc.logger.Printf("enabling telemetry (version: %d)", tv)
		err := svc.setupTelemetry(tv, svc.server)
		if err != nil {
			svc.logger.Printf("failed to setup telemetry: %s", err)
		}
		svc.logger.Printf("telemetry enabled (version: %d)", tv)
	}
}

// func getTelemetryProperties(out *settings.DecodedOptions) map[string]interface{} {
// 	properties := map[string]interface{}{
// 		"experimentalCapabilities.referenceCountCodeLens": false,
// 		"options.ignoreSingleFileWarning":                 false,
// 		"options.rootModulePaths":                         false,
// 		"options.excludeModulePaths":                      false,
// 		"options.commandPrefix":                           false,
// 		"options.indexing.ignoreDirectoryNames":           false,
// 		"options.indexing.ignorePaths":                    false,
// 		"options.experimentalFeatures.validateOnSave":     false,
// 		"options.terraform.path":                          false,
// 		"options.terraform.timeout":                       "",
// 		"options.terraform.logFilePath":                   false,
// 		"root_uri":                                        "dir",
// 		"lsVersion":                                       "",
// 	}

// properties["options.commandPrefix"] = len(out.Options.CommandPrefix) > 0
// properties["options.indexing.ignoreDirectoryNames"] = len(out.Options.Indexing.IgnoreDirectoryNames) > 0
// properties["options.indexing.ignorePaths"] = len(out.Options.Indexing.IgnorePaths) > 0
// properties["options.experimentalFeatures.prefillRequiredFields"] = out.Options.ExperimentalFeatures.PrefillRequiredFields
// properties["options.experimentalFeatures.validateOnSave"] = out.Options.ExperimentalFeatures.ValidateOnSave
// properties["options.ignoreSingleFileWarning"] = out.Options.IgnoreSingleFileWarning

// 	return properties
// }

func initializeResult(ctx context.Context) gopls.InitializeResult {
	serverCaps := gopls.InitializeResult{
		ServerInfo: &gopls.PServerInfoMsg_initialize{},
		Capabilities: gopls.ServerCapabilities{
			TextDocumentSync: gopls.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    gopls.Incremental,
			},
			CompletionProvider: &gopls.CompletionOptions{
				ResolveProvider:   true,
				TriggerCharacters: []string{".", "["},
			},
			CodeActionProvider: gopls.CodeActionOptions{
				CodeActionKinds: lsp.SupportedCodeActions.AsSlice(),
				ResolveProvider: false,
			},
			DeclarationProvider:        &gopls.Or_ServerCapabilities_declarationProvider{Value: true},
			DefinitionProvider:         &gopls.Or_ServerCapabilities_definitionProvider{Value: true},
			CodeLensProvider:           &gopls.CodeLensOptions{},
			ReferencesProvider:         &gopls.Or_ServerCapabilities_referencesProvider{Value: true},
			HoverProvider:              &gopls.Or_ServerCapabilities_hoverProvider{Value: true},
			DocumentFormattingProvider: &gopls.Or_ServerCapabilities_documentFormattingProvider{Value: true},
			DocumentSymbolProvider:     &gopls.Or_ServerCapabilities_documentSymbolProvider{Value: true},
			WorkspaceSymbolProvider:    &gopls.Or_ServerCapabilities_workspaceSymbolProvider{Value: true},
			Workspace: &gopls.Workspace6Gn{
				WorkspaceFolders: &gopls.WorkspaceFolders5Gn{
					Supported:           true,
					ChangeNotifications: "workspace/didChangeWorkspaceFolders",
				},
			},
			SignatureHelpProvider: &gopls.SignatureHelpOptions{
				TriggerCharacters: []string{"(", ","},
			},
		},
	}

	serverCaps.ServerInfo.Name = "terraform-ls"
	version, ok := lsctx.LanguageServerVersion(ctx)
	if ok {
		serverCaps.ServerInfo.Version = version
	}

	return serverCaps
}
