package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jeremaihloo/funny"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type Handler struct {
	jsonrpc2.Handler
	log *zap.Logger
}

func NewHandler(logger *zap.Logger) Handler {
	return Handler{
		log: logger,
	}
}

// Handle implements jsonrpc2.Handler
func (h Handler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			h.log.Error("error happend", zap.Error(err.(error)))
		}
	}()
	h.log.Info("request", zap.Any("req", req))
	resp, err := h.internal(ctx, conn, req)
	if err != nil {
		h.log.Error("response", zap.Error(err))
		return
	}
	err = conn.Reply(ctx, req.ID, resp)
	if err != nil {
		h.log.Error("handle: error sending response", zap.Error(err))
	}
	h.log.Info("response", zap.Any("resp", resp))
}

func (h Handler) internal(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	//TODO: Prevent any uncaught panics from taking the entire server down.
	switch req.Method {
	case "initialize":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		kind := lsp.TDSKIncremental
		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
					Kind: &kind,
				},
				CompletionProvider: &lsp.CompletionOptions{
					ResolveProvider:   true,
					TriggerCharacters: []string{"(", "."},
				},
				DefinitionProvider:     true,
				TypeDefinitionProvider: true,
				DocumentSymbolProvider: true,
				HoverProvider:          true,
				ReferencesProvider:     true,
				ImplementationProvider: true,
				SignatureHelpProvider: &lsp.SignatureHelpOptions{
					TriggerCharacters: []string{"(", ","},
				},
			},
		}, nil

	case "initialized":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		// A notification that the client is ready to receive requests. Ignore
		return nil, nil

	case "shutdown":
		return nil, nil

	case "exit":
		conn.Close()
		return nil, nil

	case "$/cancelRequest":
		// notification, don't send back results/errors
		if req.Params == nil {
			return nil, nil
		}
		var params lsp.CancelParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, nil
		}
		return nil, nil

	case "textDocument/hover":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		//return h.handleHover(ctx, conn, req, params)
		return nil, nil

	case "textDocument/definition":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		//return h.handleDefinition(ctx, conn, req, params)
		return nil, nil

	case "textDocument/typeDefinition":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		//return h.handleTypeDefinition(ctx, conn, req, params)
		return nil, nil

	case "textDocument/completion":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.CompletionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleTextDocumentCompletion(ctx, conn, req, params)

	case "textDocument/implementation":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		//return h.handleTextDocumentImplementation(ctx, conn, req, params)
		return nil, nil

	case "textDocument/signatureHelp":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		//return h.handleTextDocumentSignatureHelp(ctx, conn, req, params)
		return nil, nil
	}
	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

func (h Handler) handleTextDocumentCompletion(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.CompletionParams) (*lsp.CompletionList, error) {
	if !IsURI(params.TextDocument.URI) {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: fmt.Sprintf("textDocument/completion not yet supported for out-of-workspace URI (%q)", params.TextDocument.URI),
		}
	}

	filename := UriToRealPath(params.TextDocument.URI)
	cl := &lsp.CompletionList{
		IsIncomplete: false,
		Items:        make([]lsp.CompletionItem, 0),
	}
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	parser := funny.NewParser(file)
	parser.Consume("")
	var items funny.Block
	for {
		item := parser.ReadStatement()
		if item == nil {
			break
		}
		items = append(items, item)
	}
	for _, item := range items {
		ci := lsp.CompletionItem{}
		switch item.Type() {
		case funny.STVariable:
			t := item.(*funny.Variable)
			ci.Label = t.Name
			ci.Data = t.Name
			ci.Kind = lsp.CIKVariable
			ci.InsertTextFormat = lsp.ITFPlainText
			ci.InsertText = t.Name
		case funny.STFunction:
			t := item.(*funny.Function)
			ci.Label = t.Name
			ci.Data = t.Name
			ci.Kind = lsp.CIKFunction
			ci.InsertTextFormat = lsp.ITFPlainText
			ci.InsertText = t.Name
		case funny.STAssign:
			t := item.(*funny.Assign)
			switch t.Target.Type() {
			case funny.STVariable:
				tt := t.Target.(*funny.Variable)
				ci.Label = tt.Name
				ci.Data = tt.Name
				ci.Kind = lsp.CIKVariable
				ci.InsertTextFormat = lsp.ITFPlainText
				ci.InsertText = tt.Name
			}
		}
		var currentToken *funny.Token
		for _, token := range parser.Tokens {
			if token.Position.Line == params.Position.Line && token.Position.Col == params.Position.Character {
				currentToken = &token
			}
		}
		h.log.Info("tokens", zap.Any("tokens", parser.Tokens))
		l := 0
		if currentToken != nil {
			l = len(currentToken.Data)
			h.log.Info("current", zap.Any("current", currentToken))
		}
		ci.TextEdit = &lsp.TextEdit{
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      params.Position.Line,
					Character: params.Position.Character - l,
				},
				End: lsp.Position{
					Line:      params.Position.Line,
					Character: params.Position.Character,
				},
			},
			NewText: ci.Label,
		}
		if ci.Label != "" {
			cl.Items = append(cl.Items, ci)
		}
	}
	return cl, nil
}
