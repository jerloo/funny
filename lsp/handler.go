package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/jerloo/funny"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type Handler struct {
	jsonrpc2.Handler
	log              *zap.Logger
	documentContents *documentContents
}

func NewHandler(logger *zap.Logger) Handler {
	return Handler{
		log:              logger,
		documentContents: newDocumentContents(logger),
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
				DefinitionProvider:         true,
				TypeDefinitionProvider:     true,
				DocumentSymbolProvider:     true,
				HoverProvider:              true,
				ReferencesProvider:         true,
				ImplementationProvider:     true,
				DocumentFormattingProvider: true,
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

	case "textDocument/didOpen":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.DidOpenTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return nil, h.handleTextDocumentDidOpen(ctx, conn, req, params)
	case "textDocument/didChange":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.DidChangeTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		//return h.handleHover(ctx, conn, req, params)
		return nil, h.handleTextDocumentDidChange(ctx, conn, req, params)
	case "textDocument/didSave":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		//return h.handleHover(ctx, conn, req, params)
		return nil, nil
	case "textDocument/didClose":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.TextDocumentPositionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		//return h.handleHover(ctx, conn, req, params)
		return nil, nil
	case "textDocument/formatting":
		if req.Params == nil {
			return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
		}
		var params lsp.DocumentFormattingParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return h.handleTextDocumentFormating(ctx, conn, req, params)

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

func (h Handler) handleTextDocumentFormating(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.DocumentFormattingParams) (resp []lsp.TextEdit, err error) {
	_, fileName := path.Split(string(params.TextDocument.URI))
	if !strings.HasSuffix(fileName, ".funny") {
		return
	}
	// Format the current document.
	contents, _ := h.documentContents.Get(string(params.TextDocument.URI))
	formated := funny.Format(contents, UriToRealPath(params.TextDocument.URI))

	lines := len(string(contents))
	w := new(strings.Builder)
	w.WriteString(formated)

	// Replace everything.
	resp = append(resp, lsp.TextEdit{
		Range: lsp.Range{
			Start: lsp.Position{},
			End:   lsp.Position{Line: lines + 1, Character: 0},
		},
		NewText: w.String(),
	})
	return
}

func (h Handler) handleTextDocumentDidOpen(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.DidOpenTextDocumentParams) (err error) {
	_, fileName := path.Split(string(params.TextDocument.URI))
	if !strings.HasSuffix(fileName, ".funny") {
		return
	}
	// Cache the template doc.
	h.documentContents.Set(string(params.TextDocument.URI), []byte(params.TextDocument.Text))
	return
}

func (h Handler) handleTextDocumentDidChange(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.DidChangeTextDocumentParams) (err error) {
	_, fileName := path.Split(string(params.TextDocument.URI))
	if !strings.HasSuffix(fileName, ".funny") {
		return
	}
	// Apply content changes to the cached template.
	_, err = h.documentContents.Apply(string(params.TextDocument.URI), params.ContentChanges)
	if err != nil {
		return
	}
	return
}

func (h Handler) handleTextDocumentCompletion(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.CompletionParams) (*lsp.CompletionList, error) {
	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			h.log.Error("error happend", zap.Error(err.(error)))
		}
	}()
	if !IsURI(params.TextDocument.URI) {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: fmt.Sprintf("textDocument/completion not yet supported for out-of-workspace URI (%q)", params.TextDocument.URI),
		}
	}

	cl := &lsp.CompletionList{
		IsIncomplete: false,
		Items:        make([]lsp.CompletionItem, 0),
	}
	contents, ok := h.documentContents.Get(string(params.TextDocument.URI))
	if !ok {
		return cl, errors.New("document content not found")
	}
	builtinParser := funny.NewParser([]byte(funny.BuiltinsDotFunny))
	builtinBlock := builtinParser.Parse()
	builtinDescriptor := builtinBlock.Descriptor()

	parser := funny.NewParser(contents)
	parser.ContentFile = UriToRealPath(params.TextDocument.URI)
	items := parser.Parse()
	descriptor := items.Descriptor()
	var currentToken *funny.Token
	for _, token := range parser.Tokens {
		if token.Position.Line == params.Position.Line && token.Position.Col == params.Position.Character {
			currentToken = &token
			break
		}
	}
	h.log.Info("tokens", zap.Any("tokens", parser.Tokens))
	l := 0
	if currentToken != nil {
		l = len(currentToken.Data)
		h.log.Info("current", zap.Any("current", currentToken))
	}
	builtinFds := flatDescriptor(builtinDescriptor)
	fds := flatDescriptor(descriptor)
	fds = append(fds, builtinFds...)
	for _, item := range fds {
		ci := convertDescriptor(item)
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

func flatDescriptor(descriptor *funny.AstDescriptor) (items []*funny.AstDescriptor) {
	if descriptor != nil {
		items = append(items, descriptor)
		if descriptor.Children != nil {
			for _, child := range descriptor.Children {
				newItems := flatDescriptor(child)
				items = append(items, newItems...)
			}
		}
	}
	return items
}

func convertDescriptor(t *funny.AstDescriptor) lsp.CompletionItem {
	ci := lsp.CompletionItem{}
	ci.Label = t.Name
	ci.Data = t.Name
	ci.Kind = lsp.CIKVariable
	ci.InsertTextFormat = lsp.ITFPlainText
	ci.InsertText = t.Name
	switch t.Type {
	case funny.STVariable:
		ci.Kind = lsp.CIKVariable
	case funny.STFunction:
		ci.Kind = lsp.CIKFunction
	case funny.STAssign:
		ci.Kind = lsp.CIKVariable
	}
	return ci
}
