package main

import (
	"fmt"
	"go/ast"
	"os"
	"strings"
)

type ASTPrinter struct {
	out *os.File
}


func NewASTPrinter(out *os.File) *ASTPrinter {
	return &ASTPrinter{
		out: out,
	}
}

func (p *ASTPrinter) PrintTree(node ast.Node, level int) {
	if node == nil {
		return
	}
	indent := strings.Repeat(" ", level*2)

	switch n := node.(type) {
	case *ast.File:
		fmt.Fprintf(p.out, "%s*ast.File: Name=%s, Imports=%d, Decls=%d\n", indent, n.Name, len(n.Imports), len(n.Decls))
	case *ast.GenDecl:
		fmt.Fprintf(p.out, "%s*ast.GenDecl: Tok=%s, Specs=%d\n", indent, n.Tok, len(n.Specs))
	case *ast.FuncDecl:
		name := ""
		if n.Name != nil {
			name = n.Name.Name
		}
		recv := ""
		if n.Recv != nil && len(n.Recv.List) > 0 {
			recv = fmt.Sprintf(" (receiver: %s)", p.exprToString(n.Recv.List[0].Type))
		}
		fmt.Fprintf(p.out, "%s*ast.FuncDecl: Name=%s%s\n", indent, name, recv)
	case *ast.ImportSpec:
		value := ""
		if n.Path != nil {
			value = n.Path.Value
		}
		fmt.Fprintf(p.out, "%s*ast.ImportSpec: Path=%s\n", indent, value)
	case *ast.ValueSpec:
		fmt.Fprintf(p.out, "%s*ast.ValueSpec:\n", indent)
		for _, name := range n.Names {
			fmt.Fprintf(p.out, "%s  Name=%s\n", indent, name.Name)
		}
	case *ast.TypeSpec:
		fmt.Fprintf(p.out, "%s*ast.TypeSpec: Name=%s\n", indent, n.Name.Name)
	case *ast.Ident:
		fmt.Fprintf(p.out, "%s*ast.Ident: Name=%s\n", indent, n.Name)
	case *ast.BasicLit:
		fmt.Fprintf(p.out, "%s*ast.BasicLit: Kind=%s, Value=%s\n", indent, n.Kind, n.Value)
	default:
		fmt.Fprintf(p.out, "%s%T\n", indent, node)
	}

	ast.Inspect(node, func(n ast.Node) bool {
		if n == node {
			return true
		}
		p.PrintTree(n, level+1)
		return false
	})
}

func (p *ASTPrinter) exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + p.exprToString(e.X)
	case *ast.SelectorExpr:
		return p.exprToString(e.X) + "." + e.Sel.Name
	default:
		return fmt.Sprintf("%T", expr)
	}
}