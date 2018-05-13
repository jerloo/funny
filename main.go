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
	}
	if *optionParser {
		parser()
	}

}

func lexer() {
	data, _ := ioutil.ReadFile("funny.fl")
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
	data, _ := ioutil.ReadFile("funny.fl")
	parser := cores.NewParser(data)
	parser.Consume("")
	for {
		item := parser.ReadStatement()
		if item == nil {
			break
		}
		fmt.Printf("%v\n", item.String())

	}
}
