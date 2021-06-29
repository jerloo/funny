package lsp

import (
	"os"

	"github.com/jeremaihloo/funny/lang"
	"github.com/sourcegraph/go-lsp"
)

func GetCompletionItem(filename string) (cl *lsp.CompletionList, err error) {
	cl = &lsp.CompletionList{
		IsIncomplete: false,
		Items:        make([]lsp.CompletionItem, 0),
	}
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	parser := lang.NewParser(file)
	parser.Consume("")
	var items lang.Block
	for {
		item := parser.ReadStatement()
		if item == nil {
			break
		}
		items = append(items, item)
	}
	for _, item := range items {
		cl.Items = append(cl.Items, lsp.CompletionItem{
			Label: item.String(),
			Data:  item.String(),
		})
	}
	return
}
