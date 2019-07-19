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
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/jeremaihloo/funny/langs"
	"github.com/spf13/cobra"
)

// formatCmd represents the format command
var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Format a funny script file or funny script text.",
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
	},
}

func init() {
	rootCmd.AddCommand(formatCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// formatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// formatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
