package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"strings"

	"encoding/json"

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
		fmt.Println()
		return
	}
	kingpin.Usage()
}

func lexer() {
	var data []byte
	if *script != "" && strings.HasSuffix(*script, ".fun") {
		data, _ = ioutil.ReadFile(*script)
	} else {
		data = []byte(*script)
	}

	var tokens []langs.Token
	lexer := langs.NewLexer(data)
	for {
		token := lexer.Next()
		// fmt.Printf("%v\n", token.String())

		if token.Kind == langs.EOF {
			break
		}
		tokens = append(tokens, token)
	}
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}

func parser() {
	data, _ := ioutil.ReadFile(*script)
	parser := langs.NewParser(data)
	parser.Consume("")
	var items langs.Block
	for {
		item := parser.ReadStatement()
		if item == nil {
			break
		}
		items = append(items, item)
		// fmt.Printf("%s %s\n", langs.Typing(item), item.String())
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}

func format() {
	var data []byte
	if *script != "" && strings.HasSuffix(*script, ".fun") {
		data, _ = ioutil.ReadFile(*script)
	} else {
		inputReader := bufio.NewScanner(os.Stdin)
		for inputReader.Scan() {
			data = append(data, inputReader.Bytes()...)
			data = append(data, []byte("\n")...)
		}
	}

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
			flag++
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
