package dynamicchecks

import (
	"go/ast"
	"go/token"
	"sort"

	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
)

func applyLogicPass(astData *data.ASTData) bool {
	if astData.Spec == nil || len(astData.Spec.LogicDecls) == 0 {
		return false
	}

	existing := make(map[string]struct{})
	for _, decl := range astData.File.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			existing[fn.Name.Name] = struct{}{}
		}
	}

	names := make([]string, 0, len(astData.Spec.LogicDecls))
	for name := range astData.Spec.LogicDecls {
		names = append(names, name)
	}
	sort.Strings(names)

	var generated []ast.Decl
	for _, name := range names {
		decl := astData.Spec.LogicDecls[name]
		if decl == nil || decl.Body == nil {
			continue
		}
		if decl.Kind != specir.LogicFunction && decl.Kind != specir.LogicPredicate {
			continue
		}
		if _, ok := existing[decl.Name]; ok {
			continue
		}

		generated = append(generated, logicFuncDecl(decl))
		existing[decl.Name] = struct{}{}
	}

	if len(generated) == 0 {
		return false
	}

	insertAt := 0
	for insertAt < len(astData.File.Decls) {
		genDecl, ok := astData.File.Decls[insertAt].(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			break
		}
		insertAt++
	}
	decls := make([]ast.Decl, 0, len(astData.File.Decls)+len(generated))
	decls = append(decls, astData.File.Decls[:insertAt]...)
	decls = append(decls, generated...)
	decls = append(decls, astData.File.Decls[insertAt:]...)
	astData.File.Decls = decls
	return true
}

func logicFuncDecl(decl *specir.LogicDecl) *ast.FuncDecl {
	params := make([]*ast.Field, 0, len(decl.Params))
	for _, param := range decl.Params {
		params = append(params, &ast.Field{
			Names: []*ast.Ident{{Name: param.Name}},
			Type:  logicGoType(param.Type),
		})
	}

	resultType := decl.ResultType
	if decl.Kind == specir.LogicPredicate {
		resultType = specir.TypeExpr{Kind: specir.TypeBoolean}
	}

	conv := &baseConverter{}
	return &ast.FuncDecl{
		Name: &ast.Ident{Name: decl.Name},
		Type: &ast.FuncType{
			Params:  &ast.FieldList{List: params},
			Results: &ast.FieldList{List: []*ast.Field{{Type: logicGoType(resultType)}}},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{
			Results: []ast.Expr{conv.Convert(decl.Body)},
		}}},
	}
}

func logicGoType(typ specir.TypeExpr) ast.Expr {
	switch typ.Kind {
	case specir.TypeBoolean:
		return &ast.Ident{Name: "bool"}
	case specir.TypeInteger:
		return &ast.Ident{Name: "int"}
	case specir.TypeReal:
		return &ast.Ident{Name: "float64"}
	case specir.TypeCustom:
		if typ.Name != "" {
			return &ast.Ident{Name: typ.Name}
		}
	}
	return &ast.Ident{Name: "any"}
}
