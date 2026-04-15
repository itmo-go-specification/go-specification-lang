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

// resultInfo holds metadata about a single return value.
type resultInfo struct {
	name    string   // variable name to use in generated code
	typ     ast.Expr // Go type expression
	isNamed bool     // true if the name comes from the source (not generated)
}

// collectResultInfos builds a slice of resultInfo for all return values.
// If ALL return values are named, isNamed=true and the original names are used.
// Otherwise (unnamed or mixed), generated names __result_0, __result_1, ... are used.
func collectResultInfos(funcDecl *ast.FuncDecl) []resultInfo {
	if funcDecl.Type.Results == nil {
		return nil
	}

	// Flatten fields into individual return values.
	type raw struct {
		name string // "" if unnamed
		typ  ast.Expr
	}
	var flat []raw
	for _, field := range funcDecl.Type.Results.List {
		if len(field.Names) == 0 {
			flat = append(flat, raw{name: "", typ: field.Type})
		} else {
			for _, n := range field.Names {
				flat = append(flat, raw{name: n.Name, typ: field.Type})
			}
		}
	}

	// Check if all are named.
	allNamed := len(flat) > 0
	for _, r := range flat {
		if r.name == "" {
			allNamed = false
			break
		}
	}

	infos := make([]resultInfo, len(flat))
	for i, r := range flat {
		if allNamed {
			infos[i] = resultInfo{name: r.name, typ: r.typ, isNamed: true}
		} else {
			infos[i] = resultInfo{name: fmt.Sprintf("__result_%d", i), typ: r.typ, isNamed: false}
		}
	}
	return infos
}

// resultCount returns the total number of individual return values for a function.
func resultCount(funcDecl *ast.FuncDecl) int {
	return len(collectResultInfos(funcDecl))
}

// resultVarName returns the generated variable name for the i-th return value.
// For unnamed returns: __result_0, __result_1, ...
// For named returns: the original name from source.
// Requires collectResultInfos to have been called; use resultInfos[i].name directly when possible.
func resultVarName(i int) string {
	return fmt.Sprintf("__result_%d", i)
}

// createResultDeclarations generates variable declarations for unnamed return values only.
// For named return values, Go already declares them — no var statement is needed.
func createResultDeclarations(infos []resultInfo) []ast.Stmt {
	var stmts []ast.Stmt
	for _, info := range infos {
		if info.isNamed {
			// Named return values are already declared by Go; skip.
			continue
		}
		typ := info.typ
		if typ == nil {
			typ = &ast.Ident{Name: "interface{}"}
		}
		stmts = append(stmts, &ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{{Name: info.name}},
						Type:  typ,
					},
				},
			},
		})
	}
	return stmts
}

// replaceReturnsInBlock rewrites every "return v0, v1, ..." inside block.
//
// For unnamed returns:
//
//	return v0, v1  →  { __result_0 = v0; __result_1 = v1; return __result_0, __result_1 }
//
// For named returns:
//
//	return v0, v1  →  { name0 = v0; name1 = v1; return }
//
// Empty returns (bare "return") are left unchanged in both cases.
func replaceReturnsInBlock(block *ast.BlockStmt, infos []resultInfo) {
	for i, stmt := range block.List {
		if ret, ok := stmt.(*ast.ReturnStmt); ok && len(ret.Results) > 0 {
			block.List[i] = createResultReturnBlock(ret, infos)
		} else {
			replaceReturnsInNestedBlocks(stmt, infos)
		}
	}
}

func replaceReturnsInNestedBlocks(stmt ast.Stmt, infos []resultInfo) {
	switch s := stmt.(type) {
	case *ast.IfStmt:
		if s.Body != nil {
			replaceReturnsInBlock(s.Body, infos)
		}
		if s.Else != nil {
			switch els := s.Else.(type) {
			case *ast.BlockStmt:
				replaceReturnsInBlock(els, infos)
			case *ast.IfStmt:
				replaceReturnsInNestedBlocks(els, infos)
			}
		}
	case *ast.ForStmt:
		if s.Body != nil {
			replaceReturnsInBlock(s.Body, infos)
		}
	case *ast.RangeStmt:
		if s.Body != nil {
			replaceReturnsInBlock(s.Body, infos)
		}
	case *ast.SwitchStmt:
		for _, caseBlock := range s.Body.List {
			if cc, ok := caseBlock.(*ast.CaseClause); ok {
				replaceReturnsInBlock(&ast.BlockStmt{List: cc.Body}, infos)
			}
		}
	case *ast.TypeSwitchStmt:
		for _, caseBlock := range s.Body.List {
			if cc, ok := caseBlock.(*ast.CaseClause); ok {
				replaceReturnsInBlock(&ast.BlockStmt{List: cc.Body}, infos)
			}
		}
	case *ast.SelectStmt:
		for _, comm := range s.Body.List {
			if cc, ok := comm.(*ast.CommClause); ok {
				replaceReturnsInBlock(&ast.BlockStmt{List: cc.Body}, infos)
			}
		}
	case *ast.BlockStmt:
		replaceReturnsInBlock(s, infos)
	}
}

// createResultReturnBlock rewrites "return v0, v1, ..." as:
//
// Unnamed: { __result_0 = v0; __result_1 = v1; return __result_0, __result_1 }
// Named:   { name0 = v0; name1 = v1; return }
func createResultReturnBlock(ret *ast.ReturnStmt, infos []resultInfo) *ast.BlockStmt {
	n := len(infos)
	stmts := make([]ast.Stmt, 0, n+1)

	// Assign each return value to its variable.
	for i, val := range ret.Results {
		if i >= n {
			break
		}
		stmts = append(stmts, &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: infos[i].name}},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{val},
		})
	}

	// Build the return statement.
	allNamed := n > 0 && infos[0].isNamed
	if allNamed {
		// Named returns: bare "return" — Go returns the named variables.
		stmts = append(stmts, &ast.ReturnStmt{})
	} else {
		// Unnamed returns: explicit "return __result_0, __result_1, ..."
		retExprs := make([]ast.Expr, n)
		for i := 0; i < n; i++ {
			retExprs[i] = &ast.Ident{Name: infos[i].name}
		}
		stmts = append(stmts, &ast.ReturnStmt{Results: retExprs})
	}

	return &ast.BlockStmt{List: stmts}
}
