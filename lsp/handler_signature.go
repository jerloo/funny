package lsp

import (
	"context"
	"errors"
	"path"
	"strings"

	"github.com/jerloo/funny"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h Handler) handleTextDocumentSignatureHelp(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (result *lsp.SignatureHelp, err error) {
	_, fileName := path.Split(string(params.TextDocument.URI))
	if !strings.HasSuffix(fileName, ".funny") {
		return
	}
	contents, ok := h.documentContents.Get(string(params.TextDocument.URI))
	if !ok {
		return nil, errors.New("document content not found")
	}
	builtinParser := funny.NewParser([]byte(funny.BuiltinsDotFunny))
	builtinBlock := builtinParser.Parse()
	builtinDescriptor := builtinBlock.Descriptor()

	parser := funny.NewParser(contents)
	parser.ContentFile = UriToRealPath(params.TextDocument.URI)
	items := parser.Parse()
	descriptor := items.Descriptor()
	builtinFds := flatDescriptor(builtinDescriptor)
	fds := flatDescriptor(descriptor)
	fds = append(fds, builtinFds...)
	var signatures []*funny.AstDescriptor
	for _, item := range fds {
		if item.Type == funny.STFunction {
			signatures = append(signatures, item)
		}
	}
	for _, item := range fds {
		if item.Position.Line == params.Position.Line {
			for _, sig := range signatures {
				if sig.Name == item.Name {
					ap := len(item.Children)
					if ap > len(sig.Children) {
						ap = len(sig.Children) - 1
					}
					var infos []lsp.ParameterInformation
					for _, pas := range item.Children {
						infos = append(infos, lsp.ParameterInformation{
							Label:         pas.Name,
							Documentation: pas.Text,
						})
					}
					return &lsp.SignatureHelp{
						Signatures: []lsp.SignatureInformation{
							{
								Label:         sig.Text,
								Documentation: "",
								Parameters:    infos,
							},
						},
						ActiveSignature: 0,
						ActiveParameter: ap,
					}, nil
				}
			}
		}
	}
	return &lsp.SignatureHelp{
		Signatures:      nil,
		ActiveSignature: 0,
		ActiveParameter: 0,
	}, nil
}
