package funny

func Format(data []byte) string {
	parser := NewParser(data)
	block := parser.Parse()
	return block.Format(true)
}
