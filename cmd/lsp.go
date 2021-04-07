/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/jeremaihloo/funny/lsp"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/sourcegraph/jsonrpc2"
)

const (
	exitCodeErr       = 1
	exitCodeInterrupt = 2
)

type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}

type rpcLogger struct {
	zapLogger *zap.Logger
}

func (l rpcLogger) Printf(format string, v ...interface{}) {
	l.zapLogger.Info(fmt.Sprintf(format, v...))
}

func run(ctx context.Context, args []string) error {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{
		"log.txt",
	}
	logger, err := cfg.Build()
	if err != nil {
		log.Printf("failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	logger.Info("Starting up...")
	handler := lsp.NewHandler(logger)

	stream := jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{})
	rpcLogger := jsonrpc2.LogMessages(rpcLogger{zapLogger: logger})
	conn := jsonrpc2.NewConn(ctx, stream, handler, rpcLogger)
	select {
	case <-ctx.Done():
		logger.Info("Signal received")
		conn.Close()
	case <-conn.DisconnectNotify():
		logger.Info("Client disconnected")
	}

	logger.Info("Stopped...")
	return nil
}

// lspCmd represents the lsp command
var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Start a funny language server.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		defer func() {
			signal.Stop(signalChan)
			cancel()
		}()
		go func() {
			select {
			case <-signalChan: // first signal, cancel context
				cancel()
			case <-ctx.Done():
			}
			<-signalChan // second signal, hard exit
			os.Exit(exitCodeInterrupt)
		}()
		if err := run(ctx, os.Args); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(exitCodeErr)
		}
	},
}

func init() {
	rootCmd.AddCommand(lspCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lspCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lspCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
