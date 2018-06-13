package main

import (
	"github.com/alecthomas/kingpin"
	"io/ioutil"
	"github.com/jeremaihloo/funny/langs"
	"os"
	"fmt"
)

var (
	app = kingpin.New("funny", "funny lang")

	script       = app.Arg("script", "script file path").String()
	optionLexer  = app.Flag("lexer", "tokenizer script").Bool()
	optionParser = app.Flag("parser", "parser AST").Bool()
	optionFormat = app.Flag("format", "format script code").Bool()
)

func main() {
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
	if *script != "" {
		run()
	}

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
		fmt.Printf("%s\n", item.String())
	}
}

func format() {
	data, _ := ioutil.ReadFile(*script)
	parser := langs.NewParser(data)
	parser.Consume("")
	for {
		item := parser.ReadStatement()
		if item == nil {
			break
		}
		fmt.Printf("%s\n", item.String())
	}
}

func run() {
	data, _ := ioutil.ReadFile(*script)
	interpreter := langs.NewInterpreter(langs.Scope{})
	parser := langs.NewParser(data)
	program := langs.Program{
		Statements: parser.Parse(),
	}
	interpreter.Run(program)
}
