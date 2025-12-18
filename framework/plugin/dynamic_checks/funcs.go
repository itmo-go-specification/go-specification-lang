package dynamicchecks

import (
	"fmt"
	"go/ast"
	"go/token"

	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
	"gpma/framework/plugin"
)

func processFunctions(astData *data.ASTData, result *plugin.PluginResult) {
	ast.Inspect(astData.File, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			info := astData.TryGetNodeInfo(node)
			if info != nil && info.FuncContract != nil {
				if processFunc(node, info.FuncContract) {
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

func processFunc(funcDecl *ast.FuncDecl, contract *specir.FuncContract) bool {
	if funcDecl.Body == nil {
		return false
	}

	converter := NewConverter()

	// Собираем \old выражения из постусловий
	for _, ens := range contract.Ensures {
		converter.CollectOld(ens.Expr)
	}

	var newStmts []ast.Stmt

	// Добавляем проверки requires
	for _, req := range contract.Requires {
		condition := converter.Convert(req.Expr)
		checkCall := createCheckCall(condition, "requires", req.Label)
		newStmts = append(newStmts, &ast.ExprStmt{X: checkCall})
	}

	// Объявление переменных для \old
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

	has_result_var := funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) == 1
	// Добавляем объявление __result если есть ensures
	if has_result_var {
		resultDecls := createResultDeclaration(funcDecl)
		newStmts = append(newStmts, resultDecls)
	}

	// Добавляем проверки ensures (defer)
	for _, ens := range contract.Ensures {
		condition := converter.Convert(ens.Expr)
		checkCall := createCheckCall(condition, "ensures", ens.Label)
		deferCheck := &ast.DeferStmt{Call: checkCall}
		newStmts = append(newStmts, deferCheck)
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
