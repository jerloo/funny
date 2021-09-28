package funny

func Format(data []byte, contentFile string) string {
	parser := NewParser(data, contentFile)
	parser.ContentFile = contentFile
	block, err := parser.Parse()
	if err != nil {
		panic(err)
	}
	return block.Format(true)
}
