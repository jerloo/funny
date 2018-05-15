package main

import (
	"github.com/alecthomas/kingpin"
	"io/ioutil"
	"github.com/jeremaihloo/funny-lang/cores"
	"os"
	"fmt"
)

var (
	app = kingpin.New("funny", "funny lang")

	script       = app.Arg("script", "script file path").String()
	optionLexer  = app.Flag("lexer", "tokenizer script").Bool()
	optionParser = app.Flag("parser", "parser AST").Bool()
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
	if *script != "" {
		run()
	}
}

func lexer() {
	data, _ := ioutil.ReadFile(*script)
	lexer := cores.NewLexer(data)
	for {
		token := lexer.Next()
		fmt.Printf("%v\n", token.String())

		if token.Kind == cores.EOF {
			break
		}
	}
}

func parser() {
	data, _ := ioutil.ReadFile(*script)
	parser := cores.NewParser(data)
	parser.Consume("")
	for {
		item := parser.ReadStatement()
		if item == nil {
			break
		}
		fmt.Printf("%s | %s\n", cores.Typing(item), item.String())

	}
}

func run() {
	data, _ := ioutil.ReadFile(*script)
	interpreter := cores.NewInterpreter(cores.Scope{})
	parser := cores.NewParser(data)
	program := cores.Program{
		Statements: parser.Parse(),
	}
	interpreter.Run(program)
}
