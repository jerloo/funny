package lsp

// var funnyTypeCIKMap = map[string]lsp.CompletionItemKind{
// 	funny.STVariable: lsp.CIKVariable,
// 	funny.STFunction: lsp.CIKFunction,
// }

// func convertDescriptorToCompletionItem(descriptor *funny.AstDescriptor) lsp.CompletionItem {
// 	return lsp.CompletionItem{
// 		Label:            descriptor.Name,
// 		Data:             descriptor.Text,
// 		Kind:             funnyTypeCIKMap[descriptor.Type],
// 		InsertTextFormat: lsp.ITFPlainText,
// 		InsertText:       descriptor.Text,
// 		TextEdit: &lsp.TextEdit{
// 			Range: lsp.Range{
// 				Start: lsp.Position{
// 					Line:      params.Position.Line,
// 					Character: params.Position.Character,
// 				},
// 				End: lsp.Position{
// 					Line:      params.Position.Line,
// 					Character: params.Position.Character + len(ci.Label),
// 				},
// 			},
// 			NewText: ci.Label,
// 		},
// 	}
// }

// func GetCompletionItem(filename string, params lsp.CompletionParams) (cl *lsp.CompletionList, err error) {

// 	return
// }
