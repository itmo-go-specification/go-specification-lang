package front

import (
	"go/ast"
	"go/token"
	specir "gpma/framework/gosl/ir"
)

func CollectRawBlocks(fset *token.FileSet, file *ast.File, filename string) []RawSpecBlock {
	var blocks []RawSpecBlock

	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if !isACSLComment(c.Text) {
				continue
			}

			start := fset.Position(c.Pos()).Offset
			end := fset.Position(c.End()).Offset

			blocks = append(blocks, RawSpecBlock{
				Text: trimACSLMarkers(c.Text), // removing /*@ и */
				Span: specir.Span{
					File:  filename,
					Start: start,
					End:   end,
				},
			})
		}
	}
	return blocks
}

func isACSLComment(text string) bool {
	// strings.HasPrefix("/*@") && strings.HasSuffix("*/")
	return true
}

func trimACSLMarkers(text string) string {
	// Cut "/*@" && "*/"
	return text
}
