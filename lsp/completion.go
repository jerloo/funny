package lsp

import (
	"os"

	"github.com/jeremaihloo/funny/lang"
	"github.com/sourcegraph/go-lsp"
)

var funnyTypeCIKMap = map[string]lsp.CompletionItemKind{
	lang.STVariable: lsp.CIKVariable,
	lang.STFunction: lsp.CIKFunction,
}

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
		ci := lsp.CompletionItem{}
		switch item.Type() {
		case lang.STVariable:
			t := item.(*lang.Variable)
			ci.Label = t.Name
			ci.Data = t.Name
			ci.Kind = lsp.CIKVariable
		case lang.STFunction:
			t := item.(*lang.Function)
			ci.Label = t.Name
			ci.Data = t.Name
			ci.Kind = lsp.CIKFunction
		case lang.STAssign:
			t := item.(*lang.Assign)
			switch t.Target.Type() {
			case lang.STVariable:
				tt := t.Target.(*lang.Variable)
				ci.Label = tt.Name
				ci.Data = tt.Name
				ci.Kind = lsp.CIKVariable
			}
		}
		cl.Items = append(cl.Items, ci)
	}
	return
}
