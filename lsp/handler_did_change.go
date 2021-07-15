package lsp

import (
	"context"
	"path"
	"strings"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

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
