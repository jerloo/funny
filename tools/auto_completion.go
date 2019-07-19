package tools

import (
	"github.com/jeremaihloo/funny/langs"
)

func parseGoodSymbols(data []byte) langs.Block {
	parser := langs.NewParser(data)
	block := langs.Block{}
	parser.Consume("")
	for {
		if parser.Current.Kind == langs.EOF {
			break
		}
		element := parser.ReadStatement()
		if element == nil {
			break
		}
		block = append(block, element)
	}
}
