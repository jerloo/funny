package lsp

import (
	"context"
	"path"
	"strings"

	"github.com/jerloo/funny"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

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
