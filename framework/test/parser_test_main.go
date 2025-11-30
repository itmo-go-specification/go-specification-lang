//go:build exclude

package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
)

func main() {
	files, err := os.ReadDir("./cmd/sources")
	if err != nil {
		fmt.Printf("Error while reading sources dir: %v\n", err)
		return
	}

	for _, file := range files {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, fmt.Sprintf("cmd/sources/%s", file.Name()), nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			return
		}

		out, err := os.Create(fmt.Sprintf("cmd/asts/%s.tree", file.Name()))
		if err != nil {
			fmt.Printf("Error while creating file: %v\n", err)
			continue
		}

		printer := NewASTPrinter(out)
		printer.PrintTree(f, 0)
		out.Close()
	}
}
