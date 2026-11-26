package dynamicchecks

import (
	"go/ast"
	"go/token"

	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
)

func applyLoopPass(astData *data.ASTData) bool {
	modified := false
	for _, decl := range astData.File.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
			if instrumentLoopsInBlock(fn.Body, astData) {
				modified = true
			}
		}
	}
	return modified
}

func instrumentLoopsInBlock(block *ast.BlockStmt, astData *data.ASTData) bool {
	modified := false
	newList := make([]ast.Stmt, 0, len(block.List))
	for _, stmt := range block.List {
		if instrumentNestedLoopBlocks(stmt, astData) {
			modified = true
		}

		var loopBody *ast.BlockStmt
		var loopSpec *specir.LoopSpec
		switch loop := stmt.(type) {
		case *ast.ForStmt:
			loopBody = loop.Body
			if info := astData.TryGetNodeInfo(loop); info != nil {
				loopSpec = info.LoopSpec
			}
		case *ast.RangeStmt:
			loopBody = loop.Body
			if info := astData.TryGetNodeInfo(loop); info != nil {
				loopSpec = info.LoopSpec
			}
		}

		if loopBody != nil && loopSpec != nil {
			checks := generateLoopChecks(loopSpec, newContractConverter())
			body, bodyModified := insertLoopChecks(loopBody, checks)
			loopBody.List = body
			if len(checks.beginStmts) > 0 {
				newList = append(newList, checks.beginStmts...)
				modified = true
			}
			modified = modified || bodyModified
		}
		newList = append(newList, stmt)
	}
	block.List = newList
	return modified
}

func instrumentNestedLoopBlocks(stmt ast.Stmt, astData *data.ASTData) bool {
	modified := false
	process := func(block *ast.BlockStmt) {
		if block != nil && instrumentLoopsInBlock(block, astData) {
			modified = true
		}
	}

	switch s := stmt.(type) {
	case *ast.ForStmt:
		process(s.Body)
	case *ast.RangeStmt:
		process(s.Body)
	case *ast.BlockStmt:
		process(s)
	case *ast.IfStmt:
		process(s.Body)
		if elseBlock, ok := s.Else.(*ast.BlockStmt); ok {
			process(elseBlock)
		} else if elseIf, ok := s.Else.(*ast.IfStmt); ok {
			if instrumentNestedLoopBlocks(elseIf, astData) {
				modified = true
			}
		}
	case *ast.SwitchStmt:
		for _, item := range s.Body.List {
			if clause, ok := item.(*ast.CaseClause); ok {
				block := &ast.BlockStmt{List: clause.Body}
				process(block)
				clause.Body = block.List
			}
		}
	case *ast.TypeSwitchStmt:
		for _, item := range s.Body.List {
			if clause, ok := item.(*ast.CaseClause); ok {
				block := &ast.BlockStmt{List: clause.Body}
				process(block)
				clause.Body = block.List
			}
		}
	case *ast.SelectStmt:
		for _, item := range s.Body.List {
			if clause, ok := item.(*ast.CommClause); ok {
				block := &ast.BlockStmt{List: clause.Body}
				process(block)
				clause.Body = block.List
			}
		}
	}
	return modified
}

func applyLoopSpecToFor(forStmt *ast.ForStmt, spec *specir.LoopSpec) bool {
	if forStmt.Body == nil {
		return false
	}

	conv := newContractConverter()
	checks := generateLoopChecks(spec, conv)

	newBody, modified := insertLoopChecks(forStmt.Body, checks)
	forStmt.Body = &ast.BlockStmt{List: newBody}

	return modified
}

func applyLoopSpecToRange(rangeStmt *ast.RangeStmt, spec *specir.LoopSpec) bool {
	if rangeStmt.Body == nil {
		return false
	}

	conv := newContractConverter()
	checks := generateLoopChecks(spec, conv)

	newBody, modified := insertLoopChecks(rangeStmt.Body, checks)
	rangeStmt.Body = &ast.BlockStmt{List: newBody}

	return modified
}

type loopChecks struct {
	beginStmts []ast.Stmt
	endStmts   []ast.Stmt
}

func generateLoopChecks(spec *specir.LoopSpec, conv *contractConverter) *loopChecks {
	result := &loopChecks{}

	for _, inv := range spec.Invariants {
		condition := conv.Convert(inv.Expr)
		checkCall := createCheckCall(condition, "loop invariant", inv.Label, spec.StartPos)
		result.beginStmts = append(result.beginStmts, &ast.ExprStmt{X: checkCall})
	}

	for _, inv := range spec.Invariants {
		condition := conv.Convert(inv.Expr)
		checkCall := createCheckCall(condition, "loop invariant", inv.Label, spec.StartPos)
		result.endStmts = append(result.endStmts, &ast.ExprStmt{X: checkCall})
	}

	if spec.Variant != nil {
		variantExpr := conv.Convert(spec.Variant)

		beforeVarName := "__variant_before_" + conv.newVar()
		result.beginStmts = append(result.beginStmts, &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: beforeVarName}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{variantExpr},
		})

		afterVarName := "__variant_after_" + conv.newVar()
		result.endStmts = append(result.endStmts,
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: afterVarName}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{conv.Convert(spec.Variant)},
			},
			&ast.ExprStmt{
				X: createCheckCall(
					&ast.BinaryExpr{
						X:  &ast.Ident{Name: afterVarName},
						Op: token.LSS,
						Y:  &ast.Ident{Name: beforeVarName},
					},
					"loop variant",
					"",
					spec.StartPos,
				),
			},
			&ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: beforeVarName}},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{&ast.Ident{Name: afterVarName}},
			},
		)
	}

	return result
}

func insertLoopChecks(body *ast.BlockStmt, checks *loopChecks) (newList []ast.Stmt, modified bool) {
	if body == nil || len(checks.endStmts) == 0 {
		if body == nil {
			return nil, false
		}
		return body.List, false
	}
	newList = insertChecksBeforeContinue(body.List, checks.endStmts)
	newList = append(newList, checks.endStmts...)
	return newList, true
}

func insertChecksBeforeContinue(stmts []ast.Stmt, checks []ast.Stmt) []ast.Stmt {
	result := make([]ast.Stmt, 0, len(stmts)+len(checks))
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.BranchStmt:
			if s.Tok == token.CONTINUE && s.Label == nil {
				result = append(result, checks...)
			}
		case *ast.BlockStmt:
			s.List = insertChecksBeforeContinue(s.List, checks)
		case *ast.IfStmt:
			s.Body.List = insertChecksBeforeContinue(s.Body.List, checks)
			if elseBlock, ok := s.Else.(*ast.BlockStmt); ok {
				elseBlock.List = insertChecksBeforeContinue(elseBlock.List, checks)
			} else if elseIf, ok := s.Else.(*ast.IfStmt); ok {
				insertChecksInIf(elseIf, checks)
			}
		case *ast.SwitchStmt:
			insertChecksInCaseClauses(s.Body.List, checks)
		case *ast.TypeSwitchStmt:
			insertChecksInCaseClauses(s.Body.List, checks)
		case *ast.SelectStmt:
			for _, item := range s.Body.List {
				if clause, ok := item.(*ast.CommClause); ok {
					clause.Body = insertChecksBeforeContinue(clause.Body, checks)
				}
			}
		}
		result = append(result, stmt)
	}
	return result
}

func insertChecksInIf(stmt *ast.IfStmt, checks []ast.Stmt) {
	stmt.Body.List = insertChecksBeforeContinue(stmt.Body.List, checks)
	if elseBlock, ok := stmt.Else.(*ast.BlockStmt); ok {
		elseBlock.List = insertChecksBeforeContinue(elseBlock.List, checks)
	} else if elseIf, ok := stmt.Else.(*ast.IfStmt); ok {
		insertChecksInIf(elseIf, checks)
	}
}

func insertChecksInCaseClauses(items []ast.Stmt, checks []ast.Stmt) {
	for _, item := range items {
		if clause, ok := item.(*ast.CaseClause); ok {
			clause.Body = insertChecksBeforeContinue(clause.Body, checks)
		}
	}
}
