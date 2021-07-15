package lsp

import (
	"context"
	"path"
	"strings"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h Handler) handleTextDocumentDidOpen(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.DidOpenTextDocumentParams) (err error) {
	_, fileName := path.Split(string(params.TextDocument.URI))
	if !strings.HasSuffix(fileName, ".funny") {
		return
	}
	// Cache the template doc.
	h.documentContents.Set(string(params.TextDocument.URI), []byte(params.TextDocument.Text))
	return
}
