package dynamicchecks

import (
	"fmt"
	"go/ast"
	"go/token"

	specir "gpma/framework/gosl/ir"
)

// createCheckCall generates a checks.CheckAt call with the source location
// (file + line from the original annotated .go file).
func createCheckCall(condition ast.Expr, prefix, label string, srcPos specir.Position) *ast.CallExpr {
	var message string
	if label != "" {
		message = fmt.Sprintf(`"%s: %s"`, prefix, label)
	} else {
		message = fmt.Sprintf(`"%s"`, prefix)
	}

	// checks.CheckAt(condition, "original.go", 42, "precondition")
	return &ast.CallExpr{
		Fun: &ast.Ident{Name: "checks.CheckAt"},
		Args: []ast.Expr{
			condition,
			&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, srcPos.File)},
			&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf(`%d`, srcPos.Line)},
			&ast.BasicLit{Kind: token.STRING, Value: message},
		},
	}
}

// resultCount returns the total number of individual return values for a function.
// A field with multiple names (e.g. "a, b int") counts as multiple values.
func resultCount(funcDecl *ast.FuncDecl) int {
	if funcDecl.Type.Results == nil {
		return 0
	}
	count := 0
	for _, field := range funcDecl.Type.Results.List {
		if len(field.Names) == 0 {
			count++ // unnamed return value
		} else {
			count += len(field.Names)
		}
	}
	return count
}

// resultTypeAt returns the Go AST type expression for the i-th return value (0-based).
// Returns nil if i is out of range.
func resultTypeAt(funcDecl *ast.FuncDecl, i int) ast.Expr {
	if funcDecl.Type.Results == nil {
		return nil
	}
	idx := 0
	for _, field := range funcDecl.Type.Results.List {
		n := len(field.Names)
		if n == 0 {
			n = 1
		}
		for range make([]struct{}, n) {
			if idx == i {
				return field.Type
			}
			idx++
		}
	}
	return nil
}

// resultVarName returns the generated variable name for the i-th return value.
// __result_0, __result_1, ...
func resultVarName(i int) string {
	return fmt.Sprintf("__result_%d", i)
}

// createResultDeclarations generates "var __result_0 T0; var __result_1 T1; ..."
// for all return values of the function.
func createResultDeclarations(funcDecl *ast.FuncDecl) []ast.Stmt {
	n := resultCount(funcDecl)
	stmts := make([]ast.Stmt, 0, n)
	for i := 0; i < n; i++ {
		typ := resultTypeAt(funcDecl, i)
		if typ == nil {
			typ = &ast.Ident{Name: "interface{}"}
		}
		stmts = append(stmts, &ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{{Name: resultVarName(i)}},
						Type:  typ,
					},
				},
			},
		})
	}
	return stmts
}

// replaceReturnsInBlock rewrites every "return v0, v1, ..." inside block as:
//
//	{ __result_0 = v0; __result_1 = v1; ...; return __result_0, __result_1, ... }
//
// n is the total number of return values.
func replaceReturnsInBlock(block *ast.BlockStmt, n int) {
	for i, stmt := range block.List {
		if ret, ok := stmt.(*ast.ReturnStmt); ok && len(ret.Results) > 0 {
			block.List[i] = createResultReturnBlock(ret, n)
		} else {
			replaceReturnsInNestedBlocks(stmt, n)
		}
	}
}

func replaceReturnsInNestedBlocks(stmt ast.Stmt, n int) {
	switch s := stmt.(type) {
	case *ast.IfStmt:
		if s.Body != nil {
			replaceReturnsInBlock(s.Body, n)
		}
		if s.Else != nil {
			switch els := s.Else.(type) {
			case *ast.BlockStmt:
				replaceReturnsInBlock(els, n)
			case *ast.IfStmt:
				replaceReturnsInNestedBlocks(els, n)
			}
		}
	case *ast.ForStmt:
		if s.Body != nil {
			replaceReturnsInBlock(s.Body, n)
		}
	case *ast.RangeStmt:
		if s.Body != nil {
			replaceReturnsInBlock(s.Body, n)
		}
	case *ast.SwitchStmt:
		for _, caseBlock := range s.Body.List {
			if cc, ok := caseBlock.(*ast.CaseClause); ok {
				replaceReturnsInBlock(&ast.BlockStmt{List: cc.Body}, n)
			}
		}
	case *ast.TypeSwitchStmt:
		for _, caseBlock := range s.Body.List {
			if cc, ok := caseBlock.(*ast.CaseClause); ok {
				replaceReturnsInBlock(&ast.BlockStmt{List: cc.Body}, n)
			}
		}
	case *ast.SelectStmt:
		for _, comm := range s.Body.List {
			if cc, ok := comm.(*ast.CommClause); ok {
				replaceReturnsInBlock(&ast.BlockStmt{List: cc.Body}, n)
			}
		}
	case *ast.BlockStmt:
		replaceReturnsInBlock(s, n)
	}
}

// createResultReturnBlock rewrites "return v0, v1, ..." as:
//
//	{ __result_0 = v0; __result_1 = v1; ...; return __result_0, __result_1, ... }
func createResultReturnBlock(ret *ast.ReturnStmt, n int) *ast.BlockStmt {
	stmts := make([]ast.Stmt, 0, n+1)

	// __result_i = ret.Results[i]  for each i
	for i, val := range ret.Results {
		stmts = append(stmts, &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: resultVarName(i)}},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{val},
		})
	}

	// return __result_0, __result_1, ...
	retExprs := make([]ast.Expr, n)
	for i := 0; i < n; i++ {
		retExprs[i] = &ast.Ident{Name: resultVarName(i)}
	}
	stmts = append(stmts, &ast.ReturnStmt{Results: retExprs})

	return &ast.BlockStmt{List: stmts}
}
