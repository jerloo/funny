package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/jeremaihloo/funny/langs"
)

var (
	app = kingpin.New("funny", "funny lang")

	script        = app.Arg("script", "script file path").String()
	optionLexer   = app.Flag("lexer", "tokenizer script").Bool()
	optionParser  = app.Flag("parser", "parser AST").Bool()
	optionFormat  = app.Flag("format", "format script code").Bool()
	optionVersion = app.Flag("version", "Show version.").Bool()
)

func main() {
	kingpin.Version(langs.VERSION)
	kingpin.MustParse(app.Parse(os.Args[1:]))
	if *optionLexer {
		lexer()
		return
	}
	if *optionParser {
		parser()
		return
	}
	if *optionFormat {
		format()
		return
	}
	if *optionVersion {
		fmt.Printf("Version: %s\n", langs.VERSION)
		return
	}
	if *script != "" {
		run()
		return
	}
	kingpin.Usage()
}

func lexer() {
	data, _ := ioutil.ReadFile(*script)
	lexer := langs.NewLexer(data)
	for {
		token := lexer.Next()
		fmt.Printf("%v\n", token.String())

		if token.Kind == langs.EOF {
			break
		}
	}
}

func parser() {
	data, _ := ioutil.ReadFile(*script)
	parser := langs.NewParser(data)
	parser.Consume("")
	for {
		item := parser.ReadStatement()
		if item == nil {
			break
		}
		fmt.Printf("%s %s\n", langs.Typing(item), item.String())
	}
}

func format() {
	data, _ := ioutil.ReadFile(*script)
	parser := langs.NewParser(data)
	parser.Consume("")
	flag := 0
	for {
		item := parser.ReadStatement()
		if item == nil {
			break
		}
		switch item.(type) {
		case *langs.NewLine:
			flag += 1
			if flag < 1 {
				continue
			}
			break
		}
		fmt.Printf("%s", item.String())
	}
}

func run() {
	data, err := ioutil.ReadFile(*script)
	if err != nil {
		fmt.Printf("open file error : %s", err)
	}
	interpreter := langs.NewInterpreterWithScope(langs.Scope{})
	interpreter.Run(data)
}
