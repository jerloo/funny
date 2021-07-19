package lsp

import (
	"context"
	"errors"
	"path"
	"strings"

	"github.com/jerloo/funny"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
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
	var funcDefines []*funny.Function
	funcDefines = append(funcDefines, builtinFuncs...)
	funcDefines = append(funcDefines, parsedFuncs...)
	h.log.Info("signatures", zap.Any("fns", funcDefines))

	calls := getFuncCalls(items.Statements)
	h.log.Info("signatures", zap.Any("signatures", calls))

	for _, call := range calls {
		if call.GetPosition().Line == params.Position.Line {
			for _, fnDefine := range funcDefines {
				if fnDefine.Name == call.Name {
					activeParam := len(call.Parameters)
					if activeParam > len(fnDefine.Parameters) {
						activeParam = len(fnDefine.Parameters) - 1
					}
					var infos []lsp.ParameterInformation
					var argNames []string
					for _, pas := range fnDefine.Parameters {
						pi := lsp.ParameterInformation{}
						switch v := pas.(type) {
						case *funny.Variable:
							pi.Label = v.Name
							argNames = append(argNames, v.Name)
						}
						infos = append(infos, pi)
					}
					return &lsp.SignatureHelp{
						Signatures: []lsp.SignatureInformation{
							{
								Label:         strings.Join(argNames, ","),
								Documentation: "",
								Parameters:    infos,
							},
						},
						ActiveSignature: 0,
						ActiveParameter: activeParam,
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
