package lsp

import (
	"context"
	"errors"
	"path"
	"strings"

	"github.com/jerloo/funny"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h Handler) handleTextDocumentSignatureHelp(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (result *lsp.SignatureHelp, err error) {
	_, fileName := path.Split(string(params.TextDocument.URI))
	if !strings.HasSuffix(fileName, ".funny") {
		return
	}
	contents, ok := h.documentContents.Get(string(params.TextDocument.URI))
	if !ok {
		return nil, errors.New("document content not found")
	}
	builtinParser := funny.NewParser([]byte(funny.BuiltinsDotFunny), UriToRealPath(params.TextDocument.URI))
	builtinBlock := builtinParser.Parse()

	parser := funny.NewParser(contents, UriToRealPath(params.TextDocument.URI))
	parser.ContentFile = UriToRealPath(params.TextDocument.URI)
	items := parser.Parse()

	builtinFuncs := getFunctions(builtinBlock.Statements)
	parsedFuncs := getFunctions(items.Statements)
	var fns []*funny.Function
	fns = append(fns, builtinFuncs...)
	fns = append(fns, parsedFuncs...)

	signatures := getFuncCalls(items.Statements)

	for _, item := range fns {
		if item.GetPosition().Line == params.Position.Line {
			for _, sig := range signatures {
				if sig.Name == item.Name {
					ap := len(item.Parameters)
					if ap > len(sig.Parameters) {
						ap = len(sig.Parameters) - 1
					}
					var infos []lsp.ParameterInformation
					for _, pas := range item.Parameters {
						pi := lsp.ParameterInformation{}
						switch v := pas.(type) {
						case *funny.Variable:
							pi.Label = v.Name
						}
						infos = append(infos)
					}
					return &lsp.SignatureHelp{
						Signatures: []lsp.SignatureInformation{
							{
								Label:         sig.Name,
								Documentation: "",
								Parameters:    infos,
							},
						},
						ActiveSignature: 0,
						ActiveParameter: ap,
					}, nil
				}
			}
		}
	}
	return &lsp.SignatureHelp{
		Signatures:      nil,
		ActiveSignature: 0,
		ActiveParameter: 0,
	}, nil
}

func getFunctions(items []funny.Statement) (results []*funny.Function) {
	for _, item := range items {
		if v, ok := item.(*funny.Function); ok {
			results = append(results, v)
		}
	}
	return results
}

func getFuncCalls(items []funny.Statement) (results []*funny.FunctionCall) {
	for _, item := range items {
		if v, ok := item.(*funny.FunctionCall); ok {
			results = append(results, v)
		}
	}
	return results
}
