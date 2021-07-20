package lsp

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/jerloo/funny"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

func (h Handler) handleTextDocumentHover(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (result *lsp.Hover, err error) {
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

	var currentToken *funny.Token
	var lastToken funny.Token
	var fields []string
	for index, token := range parser.Tokens {
		if token.Position.Line == params.Position.Line && params.Position.Character >= token.Position.Col && params.Position.Character <= token.Position.Col+token.Position.Length {
			currentToken = &token
			if index < len(parser.Tokens)-1 {
				if parser.Tokens[index+1].Kind == funny.DOT {
					fields = append(fields, token.Data)
				}
			}
			if index > 0 {
				lastToken = parser.Tokens[index-1]
			}
			break
		}
	}
	fmt.Println(lastToken)
	bbs := collectBlocks(h.log, params.Position.Line, builtinBlock)
	pbs := collectBlocks(h.log, params.Position.Line, items)
	findedBlocks := append(bbs, pbs...)
	return findHover(h.log, findedBlocks, currentToken), nil
}

func findHover(logger *zap.Logger, blocks []*funny.Block, hoverToken *funny.Token) *lsp.Hover {
	for _, block := range blocks {
		for _, statement := range block.Statements {
			switch v := statement.(type) {
			case *funny.Function:
				if hoverToken.Data == v.Name {
					// lastPos := v.Body.Statements[len(v.Body.Statements)-1].GetPosition()
					return &lsp.Hover{
						Contents: []lsp.MarkedString{
							{
								Language: "funny",
								Value:    v.SignatureString(),
							},
						},
						Range: &lsp.Range{
							Start: lsp.Position{
								Line:      hoverToken.Position.Line,
								Character: hoverToken.Position.Col,
							},
							End: lsp.Position{
								Line:      10,
								Character: 12,
							},
						},
					}
				}
			}
		}
	}
	return nil
}
