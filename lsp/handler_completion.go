package lsp

import (
	"context"
	"errors"
	"fmt"

	"github.com/jerloo/funny"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

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
	for i, item := range fds {
		ci := convertDescriptor(item)
		if i > 1 && fds[i-2].Type == funny.STComment {
			ci.Detail = fds[i-2].Text
			ci.Documentation = fds[i-2].Text
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

// var funnyTypeCIKMap = map[string]lsp.CompletionItemKind{
// 	funny.STVariable: lsp.CIKVariable,
// 	funny.STFunction: lsp.CIKFunction,
// }

// func convertDescriptorToCompletionItem(descriptor *funny.AstDescriptor) lsp.CompletionItem {
// 	return lsp.CompletionItem{
// 		Label:            descriptor.Name,
// 		Data:             descriptor.Text,
// 		Kind:             funnyTypeCIKMap[descriptor.Type],
// 		InsertTextFormat: lsp.ITFPlainText,
// 		InsertText:       descriptor.Text,
// 		TextEdit: &lsp.TextEdit{
// 			Range: lsp.Range{
// 				Start: lsp.Position{
// 					Line:      params.Position.Line,
// 					Character: params.Position.Character,
// 				},
// 				End: lsp.Position{
// 					Line:      params.Position.Line,
// 					Character: params.Position.Character + len(ci.Label),
// 				},
// 			},
// 			NewText: ci.Label,
// 		},
// 	}
// }

// func GetCompletionItem(filename string, params lsp.CompletionParams) (cl *lsp.CompletionList, err error) {

// 	return
// }
