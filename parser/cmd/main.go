package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

func printTree(node ast.Node, level int, out *os.File) {
	if node == nil {
		return
	}
	indent := strings.Repeat(" ", level*2)

	switch n := node.(type) {
	case *ast.File:
		fmt.Fprintf(out, "%s*ast.File: Name=%s, Imports=%d, Decls=%d\n", indent, n.Name, len(n.Imports), len(n.Decls))
	case *ast.GenDecl:
		fmt.Fprintf(out, "%s*ast.GenDecl: Tok=%s, Specs=%d\n", indent, n.Tok, len(n.Specs))
	case *ast.FuncDecl:
		name := ""
		if n.Name != nil {
			name = n.Name.Name
		}
		recv := ""
		if n.Recv != nil && len(n.Recv.List) > 0 {
			recv = fmt.Sprintf(" (receiver: %s)", exprToString(n.Recv.List[0].Type))
		}
		fmt.Fprintf(out, "%s*ast.FuncDecl: Name=%s%s\n", indent, name, recv)
	case *ast.ImportSpec:
		value := ""
		if n.Path != nil {
			value = n.Path.Value
		}
		fmt.Fprintf(out, "%s*ast.ImportSpec: Path=%s\n", indent, value)
	case *ast.ValueSpec:
		fmt.Fprintf(out, "%s*ast.ValueSpec:\n", indent)
		for _, name := range n.Names {
			fmt.Fprintf(out, "%s  Name=%s\n", indent, name.Name)
		}
	case *ast.TypeSpec:
		fmt.Fprintf(out, "%s*ast.TypeSpec: Name=%s\n", indent, n.Name.Name)
	case *ast.Ident:
		fmt.Fprintf(out, "%s*ast.Ident: Name=%s\n", indent, n.Name)
	case *ast.BasicLit:
		fmt.Fprintf(out, "%s*ast.BasicLit: Kind=%s, Value=%s\n", indent, n.Kind, n.Value)
	default:
		fmt.Fprintf(out, "%s%T\n", indent, node)
	}

	ast.Inspect(node, func(n ast.Node) bool {
		if n == node {
			return true
		}
		printTree(n, level+1, out)
		return false
	})
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	default:
		return fmt.Sprintf("%T", expr)
	}
}

func main() {
	files, err := os.ReadDir("./cmd/sources")
	if err != nil {
		fmt.Printf("Error while reading sources dir")
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
		}
		printTree(f, 0, out)
	}
}
