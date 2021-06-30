package lsp

import (
	"os"

	"github.com/jeremaihloo/funny"
	"github.com/sourcegraph/go-lsp"
)

var funnyTypeCIKMap = map[string]lsp.CompletionItemKind{
	funny.STVariable: lsp.CIKVariable,
	funny.STFunction: lsp.CIKFunction,
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
	parser := funny.NewParser(file)
	parser.Consume("")
	var items funny.Block
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
		case funny.STVariable:
			t := item.(*funny.Variable)
			ci.Label = t.Name
			ci.Data = t.Name
			ci.Kind = lsp.CIKVariable
		case funny.STFunction:
			t := item.(*funny.Function)
			ci.Label = t.Name
			ci.Data = t.Name
			ci.Kind = lsp.CIKFunction
		case funny.STAssign:
			t := item.(*funny.Assign)
			switch t.Target.Type() {
			case funny.STVariable:
				tt := t.Target.(*funny.Variable)
				ci.Label = tt.Name
				ci.Data = tt.Name
				ci.Kind = lsp.CIKVariable
			}
		}
		cl.Items = append(cl.Items, ci)
	}
	return
}
