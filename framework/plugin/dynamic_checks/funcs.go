package dynamicchecks

import (
	"fmt"
	"go/ast"
	"go/token"

	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
	"gpma/framework/plugin"
)

func processFunctions(astData *data.ASTData, result *plugin.PluginResult, strictMode bool) {
	ast.Inspect(astData.File, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			info := astData.TryGetNodeInfo(node)
			var contract *specir.FuncContract
			if info != nil {
				contract = info.FuncContract
			}

			asserts := collectAssertSpecs(node, astData)

			if contract != nil || len(asserts) > 0 {
				if processFunc(node, contract, astData, strictMode) {
					result.WithASTModified()
				}
				node_info := astData.GetNodeInfo(node)
				node_info.Data["processed"] = true
				result.WithMapModified()
			}
		}
		return true
	})
}

func processFunc(funcDecl *ast.FuncDecl, contract *specir.FuncContract, astData *data.ASTData, strictMode bool) bool {
	if funcDecl.Body == nil {
		return false
	}

	converter := NewConverter()

	// Собираем \old выражения из постусловий (если контракт есть)
	if contract != nil {
		for _, ens := range contract.Ensures {
			converter.CollectOld(ens.Expr)
		}
	}

	// Собираем все AssertSpec из функции и \old выражения из них
	asserts := collectAssertSpecs(funcDecl, astData)
	for _, assertSpecWithNode := range asserts {
		converter.CollectOld(assertSpecWithNode.assert.Assertion.Expr)
	}

	var newStmts []ast.Stmt

	var preconditionsPassedVar *ast.Ident
	var hasRequires = false
	if contract != nil {
		hasRequires = len(contract.Requires) > 0
	}

	if !strictMode && hasRequires {
		preconditionsPassedVar = &ast.Ident{Name: "__preconditions_passed"}
		newStmts = append(newStmts, &ast.AssignStmt{
			Lhs: []ast.Expr{preconditionsPassedVar},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.Ident{Name: "true"}},
		})
	}

	if contract != nil {
		for _, req := range contract.Requires {
			condition := converter.Convert(req.Expr)

			if strictMode {
				checkCall := createCheckCall(condition, "precondition", req.Label)
				newStmts = append(newStmts, &ast.ExprStmt{X: checkCall})
			} else {
				// if !condition { __preconditions_passed = false }
				newStmts = append(newStmts, &ast.IfStmt{
					Cond: &ast.UnaryExpr{
						Op: token.NOT,
						X:  condition,
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{preconditionsPassedVar},
								Tok: token.ASSIGN,
								Rhs: []ast.Expr{&ast.Ident{Name: "false"}},
							},
						},
					},
				})
			}
		}
	}

	for _, oldInfo := range converter.GetOldVars() {
		value := converter.Convert(oldInfo.expr)
		assign := &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: oldInfo.name}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{value},
		}
		newStmts = append(newStmts, assign)
		oldInfo.isDeclared = true
	}

	hasEnsures := contract != nil && len(contract.Ensures) > 0
	has_result_var := hasEnsures && funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) == 1

	if has_result_var {
		resultDecls := createResultDeclaration(funcDecl)
		newStmts = append(newStmts, resultDecls)
	}

	if contract != nil {
		for _, ens := range contract.Ensures {
			condition := converter.Convert(ens.Expr)

			if strictMode || !hasRequires {
				checkCall := createCheckCall(condition, "postcondition", ens.Label)
				deferCheck := &ast.DeferStmt{Call: checkCall}
				newStmts = append(newStmts, deferCheck)
			} else {
				checkCall := createCheckCall(condition, "postcondition", ens.Label)

				// defer func() { if __preconditions_passed { checks.Check(...) } }()
				deferFunc := &ast.FuncLit{
					Type: &ast.FuncType{Params: &ast.FieldList{}},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.IfStmt{
								Cond: preconditionsPassedVar,
								Body: &ast.BlockStmt{
									List: []ast.Stmt{&ast.ExprStmt{X: checkCall}},
								},
							},
						},
					},
				}
				deferCheck := &ast.DeferStmt{Call: &ast.CallExpr{Fun: deferFunc}}
				newStmts = append(newStmts, deferCheck)
			}
		}
	}

	if len(asserts) > 0 {
		funcDecl.Body.List = insertAssertChecksIntoBlock(funcDecl.Body.List, asserts, converter)
	}

	if len(newStmts) > 0 {
		funcDecl.Body.List = append(newStmts, funcDecl.Body.List...)

		if has_result_var {
			replaceReturnsInBlock(funcDecl.Body)
		}

		return true
	}
	return false
}

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

func collectAssertSpecs(funcDecl *ast.FuncDecl, astData *data.ASTData) []*assertSpecWithNode {
	var asserts []*assertSpecWithNode

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if stmt, ok := n.(ast.Stmt); ok {
			if info := astData.TryGetNodeInfo(stmt); info != nil && info.Asserts != nil {
				asserts = append(asserts, &assertSpecWithNode{
					stmt:   stmt,
					assert: info.Asserts,
				})
			}
		}
		return true
	})

	return asserts
}

type assertSpecWithNode struct {
	stmt   ast.Stmt
	assert *specir.AssertSpec
}

func insertAssertChecksIntoBlock(stmts []ast.Stmt, asserts []*assertSpecWithNode, converter *Converter) []ast.Stmt {
	assertMap := make(map[ast.Stmt]*specir.AssertSpec)
	processedStmts := make(map[ast.Stmt]bool)

	for _, a := range asserts {
		assertMap[a.stmt] = a.assert
	}

	var result []ast.Stmt

	for _, stmt := range stmts {
		if assert, ok := assertMap[stmt]; ok {
			// Собираем \old выражения прямо перед преобразованием
			converter.CollectOld(assert.Assertion.Expr)
			condition := converter.Convert(assert.Assertion.Expr)
			checkCall := createCheckCall(condition, "assertion", assert.Assertion.Label)
			checkStmt := &ast.ExprStmt{X: checkCall}

			// Вставляем Check ПЕРЕД statement'ом
			result = append(result, checkStmt)
			processedStmts[stmt] = true
		}

		result = append(result, stmt)
		processAssertInNestedBlocks(stmt, assertMap, converter)
	}

	// Обрабатываем оставшиеся AssertSpec которые не были привязаны к конкретным statement'ам
	for stmt, assert := range assertMap {
		if !processedStmts[stmt] {
			// Проверяем что этот statement не из другого блока
			found := false
			for _, s := range stmts {
				if s == stmt {
					found = true
					break
				}
			}
			if found {
				// Собираем \old выражения перед преобразованием
				converter.CollectOld(assert.Assertion.Expr)
				condition := converter.Convert(assert.Assertion.Expr)
				checkCall := createCheckCall(condition, "assertion", assert.Assertion.Label)
				checkStmt := &ast.ExprStmt{X: checkCall}
				result = append(result, checkStmt)
			}
		}
	}

	return result
}

func processAssertInNestedBlocks(stmt ast.Stmt, assertMap map[ast.Stmt]*specir.AssertSpec, converter *Converter) {
	switch s := stmt.(type) {
	case *ast.IfStmt:
		if s.Body != nil {
			s.Body.List = insertAssertChecksIntoBlock(s.Body.List, convertMapToSlice(assertMap), converter)
		}
		if s.Else != nil {
			if elseBlock, ok := s.Else.(*ast.BlockStmt); ok {
				elseBlock.List = insertAssertChecksIntoBlock(elseBlock.List, convertMapToSlice(assertMap), converter)
			} else if elseIf, ok := s.Else.(*ast.IfStmt); ok {
				processAssertInNestedBlocks(elseIf, assertMap, converter)
			}
		}
	case *ast.ForStmt:
		if s.Body != nil {
			s.Body.List = insertAssertChecksIntoBlock(s.Body.List, convertMapToSlice(assertMap), converter)
		}
	case *ast.RangeStmt:
		if s.Body != nil {
			s.Body.List = insertAssertChecksIntoBlock(s.Body.List, convertMapToSlice(assertMap), converter)
		}
	case *ast.SwitchStmt:
		for _, caseBlock := range s.Body.List {
			if caseClause, ok := caseBlock.(*ast.CaseClause); ok {
				caseClause.Body = insertAssertChecksIntoBlock(caseClause.Body, convertMapToSlice(assertMap), converter)
			}
		}
	case *ast.TypeSwitchStmt:
		for _, caseBlock := range s.Body.List {
			if caseClause, ok := caseBlock.(*ast.CaseClause); ok {
				caseClause.Body = insertAssertChecksIntoBlock(caseClause.Body, convertMapToSlice(assertMap), converter)
			}
		}
	case *ast.SelectStmt:
		for _, comm := range s.Body.List {
			if commClause, ok := comm.(*ast.CommClause); ok {
				commClause.Body = insertAssertChecksIntoBlock(commClause.Body, convertMapToSlice(assertMap), converter)
			}
		}
	case *ast.BlockStmt:
		s.List = insertAssertChecksIntoBlock(s.List, convertMapToSlice(assertMap), converter)
	}
}

func convertMapToSlice(assertMap map[ast.Stmt]*specir.AssertSpec) []*assertSpecWithNode {
	var result []*assertSpecWithNode
	for stmt, assert := range assertMap {
		result = append(result, &assertSpecWithNode{
			stmt:   stmt,
			assert: assert,
		})
	}
	return result
}

func createResultDeclaration(funcDecl *ast.FuncDecl) ast.Stmt {
	var resultType ast.Expr
	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
		if len(funcDecl.Type.Results.List) == 1 {
			resultType = funcDecl.Type.Results.List[0].Type
		} else {
			// Такой ensures не поддерживается (если возвращается кортеж)
			resultType = &ast.Ident{Name: "interface{}"} // временное решение
		}
	} else {
		resultType = &ast.Ident{Name: "interface{}"}
	}

	decl := &ast.DeclStmt{
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
	return decl
}

// replaceReturnsInBlock заменяет все return <expr> в блоке на блоки с присваиванием __result
func replaceReturnsInBlock(block *ast.BlockStmt) {
	for i, stmt := range block.List {
		if ret, ok := stmt.(*ast.ReturnStmt); ok && len(ret.Results) > 0 {
			block.List[i] = createResultReturnBlock(ret)
		} else {
			processNestedBlocks(stmt)
		}
	}
}

func processNestedBlocks(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.IfStmt:
		if s.Body != nil {
			replaceReturnsInBlock(s.Body)
		}
		if s.Else != nil {
			if elseBlock, ok := s.Else.(*ast.BlockStmt); ok {
				replaceReturnsInBlock(elseBlock)
			} else if elseIf, ok := s.Else.(*ast.IfStmt); ok {
				processNestedBlocks(elseIf)
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
			if caseClause, ok := caseBlock.(*ast.CaseClause); ok {
				// caseClause.Body - это []ast.Stmt, преобразуем в BlockStmt для обработки
				replaceReturnsInBlock(&ast.BlockStmt{List: caseClause.Body})
			}
		}
	case *ast.TypeSwitchStmt:
		for _, caseBlock := range s.Body.List {
			if caseClause, ok := caseBlock.(*ast.CaseClause); ok {
				replaceReturnsInBlock(&ast.BlockStmt{List: caseClause.Body})
			}
		}
	case *ast.SelectStmt:
		for _, comm := range s.Body.List {
			if commClause, ok := comm.(*ast.CommClause); ok {
				replaceReturnsInBlock(&ast.BlockStmt{List: commClause.Body})
			}
		}
	case *ast.BlockStmt:
		replaceReturnsInBlock(s)
	}
}

func createResultReturnBlock(ret *ast.ReturnStmt) *ast.BlockStmt {
	assign := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "__result"}},
		Tok: token.ASSIGN,
		Rhs: ret.Results,
	}

	newReturn := &ast.ReturnStmt{
		Results: []ast.Expr{&ast.Ident{Name: "__result"}},
	}

	return &ast.BlockStmt{
		List: []ast.Stmt{assign, newReturn},
	}
}
