/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/jeremaihloo/funny/langs"
	"github.com/spf13/cobra"
)

// lexerCmd represents the lexer command
var lexerCmd = &cobra.Command{
	Use:   "lexer",
	Short: "Lexer dumps json for tokenizer a funny script file or funny script text.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			filename := args[0]
			if _, err := os.Stat(filename); err != nil {
				fmt.Printf("file not found %s\n", filename)
				return
			}
			var data []byte
			if filename != "" && strings.HasSuffix(filename, ".fun") {
				data, _ = ioutil.ReadFile(filename)
			} else {
				data = []byte(filename)
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
	},
}

func init() {
	rootCmd.AddCommand(lexerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lexerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lexerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
