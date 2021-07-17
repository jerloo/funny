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
	builtinParser := funny.NewParser([]byte(funny.BuiltinsDotFunny), "")
	builtinBlock := builtinParser.Parse()

	parser := funny.NewParser(contents, UriToRealPath(params.TextDocument.URI))
	items := parser.Parse()

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

	stack := stackByPosition(params.Position.Line, items, &BlockStack{})

	fds := getNamedItems(stack.Blocks)
	fdsBuiltins := getNamedItems([]*funny.Block{builtinBlock})
	fds = append(fds, fdsBuiltins...)

	for _, item := range fds {
		ci := lsp.CompletionItem{}
		switch v := item.(type) {
		case *funny.Function:
			ci.Label = v.Name
			ci.Detail = v.SignatureString()
		case *funny.Variable:
			ci.Label = v.Name
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

func stackByPosition(line int, block *funny.Block, stack *BlockStack) *BlockStack {
	for _, statement := range block.Statements {
		if v, ok := statement.(*funny.Block); ok {
			if line >= v.GetPosition().Line && line < v.EndPosition().Line {
				stack.Push(v)
			}
		}
	}
	return stack
}

type BlockStack struct {
	Blocks []*funny.Block
}

func (bs *BlockStack) Push(block *funny.Block) {
	bs.Blocks = append(bs.Blocks, block)
}

func (bs *BlockStack) Pop() {
	bs.Blocks = bs.Blocks[:len(bs.Blocks)-1]
}

func getNamedItems(block []*funny.Block) (results []funny.Statement) {
	for _, b := range block {
		for _, statement := range b.Statements {
			switch v := statement.(type) {
			case *funny.Function:
			case *funny.Variable:
				results = append(results, v)
			}
		}
	}
	return
}
