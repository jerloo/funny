package funny

import (
	"strings"
)

func Format(data []byte) string {
	sb := new(strings.Builder)
	parser := NewParser(data)
	parser.Consume("")
	flag := 0
	for {
		item := parser.ReadStatement()
		if item == nil {
			break
		}
		switch item.(type) {
		case *NewLine:
			flag++
			if flag < 1 {
				continue
			}
		}
		sb.WriteString(item.String())
	}
	return sb.String()
}
