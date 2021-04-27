package lsp

import "github.com/sourcegraph/go-lsp"

func GetCompletionItem(filename string) (*lsp.CompletionList, error) {
	return &lsp.CompletionList{
		IsIncomplete: false,
		Items:        nil,
	}, nil
}
