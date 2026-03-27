package dynamicchecks

import (
	"fmt"
	"go/ast"
	"go/token"
)

func createCheckCall(condition ast.Expr, prefix, label string) *ast.CallExpr {
	var message string
	if label != "" {
		message = fmt.Sprintf(`"%s: %s"`, prefix, label)
	} else {
		message = fmt.Sprintf(`"%s"`, prefix)
	}

	return &ast.CallExpr{
		Fun: &ast.Ident{Name: "checks.Check"},
		Args: []ast.Expr{
			condition,
			&ast.BasicLit{
				Kind:  token.STRING,
				Value: message,
			},
		},
	}
}

func createResultDeclaration(funcDecl *ast.FuncDecl) ast.Stmt {
	var resultType ast.Expr
	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
		if len(funcDecl.Type.Results.List) == 1 {
			resultType = funcDecl.Type.Results.List[0].Type
		} else {
			resultType = &ast.Ident{Name: "interface{}"}
		}
	} else {
		resultType = &ast.Ident{Name: "interface{}"}
	}

	return &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{{Name: "__result"}},
					Type:  resultType,
				},
			},
		},
	}
}

func replaceReturnsInBlock(block *ast.BlockStmt) {
	for i, stmt := range block.List {
		if ret, ok := stmt.(*ast.ReturnStmt); ok && len(ret.Results) > 0 {
			block.List[i] = createResultReturnBlock(ret)
		} else {
			replaceReturnsInNestedBlocks(stmt)
		}
	}
}

func replaceReturnsInNestedBlocks(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.IfStmt:
		if s.Body != nil {
			replaceReturnsInBlock(s.Body)
		}
		if s.Else != nil {
			switch els := s.Else.(type) {
			case *ast.BlockStmt:
				replaceReturnsInBlock(els)
			case *ast.IfStmt:
				replaceReturnsInNestedBlocks(els)
			}
		}
	case *ast.ForStmt:
		if s.Body != nil {
			replaceReturnsInBlock(s.Body)
		}
	case *ast.RangeStmt:
		if s.Body != nil {
			replaceReturnsInBlock(s.Body)
		}
	case *ast.SwitchStmt:
		for _, caseBlock := range s.Body.List {
			if cc, ok := caseBlock.(*ast.CaseClause); ok {
				replaceReturnsInBlock(&ast.BlockStmt{List: cc.Body})
			}
		}
	case *ast.TypeSwitchStmt:
		for _, caseBlock := range s.Body.List {
			if cc, ok := caseBlock.(*ast.CaseClause); ok {
				replaceReturnsInBlock(&ast.BlockStmt{List: cc.Body})
			}
		}
	case *ast.SelectStmt:
		for _, comm := range s.Body.List {
			if cc, ok := comm.(*ast.CommClause); ok {
				replaceReturnsInBlock(&ast.BlockStmt{List: cc.Body})
			}
		}
	case *ast.BlockStmt:
		replaceReturnsInBlock(s)
	}
}

// return <expr> --> { __result = <expr>; return __result }.
func createResultReturnBlock(ret *ast.ReturnStmt) *ast.BlockStmt {
	return &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "__result"}},
				Tok: token.ASSIGN,
				Rhs: ret.Results,
			},
			&ast.ReturnStmt{
				Results: []ast.Expr{&ast.Ident{Name: "__result"}},
			},
		},
	}
}
