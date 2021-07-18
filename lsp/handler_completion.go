package lsp

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	builtinParser := funny.NewParser([]byte(funny.BuiltinsDotFunny), "")
	builtinBlock := builtinParser.Parse()

	parser := funny.NewParser(contents, UriToRealPath(params.TextDocument.URI))
	items := parser.Parse()

	var currentToken *funny.Token
	for _, token := range parser.Tokens {
		if token.Position.Line == params.Position.Line && (token.Position.Col+token.Position.Length) >= params.Position.Character {
			currentToken = &token
			break
		}
	}
	if params.Context.TriggerCharacter == "." {

	} else {
		// h.log.Info("tokens", zap.Any("tokens", parser.Tokens))
		l := 0
		if currentToken != nil {
			l = len(currentToken.Data)
			h.log.Info("current", zap.Any("current", currentToken))
		}

		blocks := collectBlocks(h.log, params.Position.Line, items)
		h.log.Info("funny:completion", zap.Any("blocks", blocks))

		fds := collectCompletionItems(params, blocks, l)
		fdsBuiltins := collectCompletionItems(params, []*funny.Block{builtinBlock}, l)
		fds = append(fds, fdsBuiltins...)
		h.log.Info("funny:completion", zap.Any("fds", fds))
		cl.Items = fds
	}
	return cl, nil
}

func collectBlocks(logger *zap.Logger, line int, block *funny.Block) (results []*funny.Block) {
	if len(results) == 0 {
		results = append(results, block)
	}
	for _, statement := range block.Statements {
		if v, ok := statement.(*funny.Block); ok {
			if line >= v.GetPosition().Line && line < v.EndPosition().Line {
				results = append(results, v)
			}
		}
	}
	return
}

func collectCompletionItems(params lsp.CompletionParams, block []*funny.Block, l int) (results []lsp.CompletionItem) {
	for _, b := range block {
		var comments []*funny.Comment
		newLineCount := 0
		for _, statement := range b.Statements {
			ci := lsp.CompletionItem{}
			switch v := statement.(type) {
			case *funny.Function:
				ci.Label = v.Name
				ci.Detail = v.SignatureString()
			case *funny.Variable:
				ci.Label = v.Name
			case *funny.Assign:
				if target, ok := v.Target.(*funny.Variable); ok {
					ci.Label = target.Name
				}
			case *funny.Block:
				brs := collectCompletionItems(params, []*funny.Block{v}, l)
				results = append(results, brs...)
				newLineCount = 0
			case *funny.NewLine:
				newLineCount++
				if newLineCount > 1 {
					comments = make([]*funny.Comment, 0)
				}
			case *funny.Comment:
				comments = append(comments, v)
				newLineCount = 0
			}
			ci.Documentation = joinComments(comments)
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
				results = append(results, ci)
			}
		}
	}
	return
}

func joinComments(comments []*funny.Comment) string {
	sb := new(strings.Builder)
	for _, item := range comments {
		sb.WriteString(item.Value)
		sb.WriteString("\n")
	}
	return sb.String()
}
