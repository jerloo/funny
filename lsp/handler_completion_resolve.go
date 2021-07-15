package lsp

import (
	"context"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

func (h Handler) handleTextDocumentCompletionResolve(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.CompletionItem) (*lsp.CompletionItem, error) {
	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			h.log.Error("error happend", zap.Error(err.(error)))
		}
	}()
	return &params, nil
}
