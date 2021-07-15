package funny

func Format(data []byte, contentFile string) string {
	parser := NewParser(data)
	parser.ContentFile = contentFile
	block := parser.Parse()
	return block.Format(true)
}
