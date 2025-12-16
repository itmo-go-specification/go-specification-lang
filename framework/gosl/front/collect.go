package front

import (
	"go/ast"
	"go/token"
	specir "gpma/framework/gosl/ir"
	"strings"
)

func CollectRawBlocks(fset *token.FileSet, file *ast.File, filename string) []RawSpecBlock {
	var blocks []RawSpecBlock

	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if !isACSLComment(c.Text) {
				continue
			}
			
			blocks = append(blocks, RawSpecBlock{
				Text: c.Text,
				Start: specir.Position{
					File:  filename,
					Offset: fset.Position(c.Pos()).Offset,
					Line:   fset.Position(c.Pos()).Line,
					Column: fset.Position(c.Pos()).Column,
				},
				End: specir.Position{
					File:  filename,
					Offset: fset.Position(c.End()).Offset,
					Line:   fset.Position(c.End()).Line,
					Column: fset.Position(c.End()).Column,
				},
			})
		}
	}
	return blocks
}

// isACSLComment returns true if comment text is an ACSL block/line comment.
// Supports:
//   /*@ ... */
//   /* @ ... */
//   /*\n@\n ... */   (what gofmt may produce)
//   //@ ...
//   // @ ...
func isACSLComment(text string) bool {
	if strings.HasPrefix(text, "//") {
		i := 2
		for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
			i++
		}
		return i < len(text) && text[i] == '@'
	}

	if strings.HasPrefix(text, "/*") {
		// Go block comments produced by parser always include the closing "*/" in Text,
		// but we don't strictly require it for recognition.
		i := 2
		for i < len(text) {
			ch := text[i]
			if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
				i++
				continue
			}
			return ch == '@'
		}
	}

	return false
}