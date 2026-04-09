package dynamicchecks

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
	"gpma/framework/plugin"
)

func makeEmptyASTData() *data.ASTData {
	fset := token.NewFileSet()
	file := &ast.File{Name: &ast.Ident{Name: "test"}}
	return data.NewASTData(file, fset)
}

func TestCreateCheckCall_WithLabel(t *testing.T) {
	call := createCheckCall(&ast.Ident{Name: "x"}, "precondition", "bound")
	assertIdent(t, call.Fun, "checks.Check")
	if len(call.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(call.Args))
	}
	lit := call.Args[1].(*ast.BasicLit)
	if lit.Value != `"precondition: bound"` {
		t.Errorf("got %q", lit.Value)
	}
}

func TestCreateCheckCall_WithoutLabel(t *testing.T) {
	call := createCheckCall(&ast.Ident{Name: "y"}, "postcondition", "")
	lit := call.Args[1].(*ast.BasicLit)
	if lit.Value != `"postcondition"` {
		t.Errorf("got %q", lit.Value)
	}
}

func TestCreateCheckCall_AssertionPrefix(t *testing.T) {
	call := createCheckCall(&ast.Ident{Name: "z"}, "assertion", "safety")
	lit := call.Args[1].(*ast.BasicLit)
	if lit.Value != `"assertion: safety"` {
		t.Errorf("got %q", lit.Value)
	}
}

func TestCreateResultDeclaration_SingleResult(t *testing.T) {
	fd := &ast.FuncDecl{
		Name: &ast.Ident{Name: "foo"},
		Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "int"}}}},
		},
	}
	stmt := createResultDeclaration(fd)
	ds := stmt.(*ast.DeclStmt)
	gd := ds.Decl.(*ast.GenDecl)
	if gd.Tok != token.VAR {
		t.Errorf("expected VAR, got %v", gd.Tok)
	}
	vs := gd.Specs[0].(*ast.ValueSpec)
	if vs.Names[0].Name != "__result" {
		t.Errorf("expected __result, got %s", vs.Names[0].Name)
	}
	assertIdent(t, vs.Type, "int")
}

func TestCreateResultDeclaration_MultipleResults(t *testing.T) {
	fd := &ast.FuncDecl{
		Name: &ast.Ident{Name: "bar"},
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{List: []*ast.Field{
				{Type: &ast.Ident{Name: "int"}},
				{Type: &ast.Ident{Name: "error"}},
			}},
		},
	}
	stmt := createResultDeclaration(fd)
	vs := stmt.(*ast.DeclStmt).Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
	assertIdent(t, vs.Type, "interface{}")
}

func TestCreateResultDeclaration_NoResults(t *testing.T) {
	fd := &ast.FuncDecl{
		Name: &ast.Ident{Name: "baz"},
		Type: &ast.FuncType{Params: &ast.FieldList{}},
	}
	stmt := createResultDeclaration(fd)
	vs := stmt.(*ast.DeclStmt).Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
	assertIdent(t, vs.Type, "interface{}")
}

func TestReplaceReturnsInBlock_SimpleReturn(t *testing.T) {
	block := &ast.BlockStmt{List: []ast.Stmt{
		&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "x"}}},
	}}
	replaceReturnsInBlock(block)
	replaced := block.List[0].(*ast.BlockStmt)
	if len(replaced.List) != 2 {
		t.Fatalf("expected 2 stmts, got %d", len(replaced.List))
	}
	assign := replaced.List[0].(*ast.AssignStmt)
	assertIdent(t, assign.Lhs[0], "__result")
	assertIdent(t, assign.Rhs[0], "x")
	ret := replaced.List[1].(*ast.ReturnStmt)
	assertIdent(t, ret.Results[0], "__result")
}

func TestReplaceReturnsInBlock_EmptyReturn(t *testing.T) {
	block := &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{}}}
	replaceReturnsInBlock(block)
	if _, ok := block.List[0].(*ast.ReturnStmt); !ok {
		t.Fatalf("empty return should not be replaced, got %T", block.List[0])
	}
}

func TestReplaceReturnsInBlock_NestedInIf(t *testing.T) {
	block := &ast.BlockStmt{List: []ast.Stmt{
		&ast.IfStmt{
			Cond: &ast.Ident{Name: "true"},
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "a"}}},
			}},
			Else: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "b"}}},
			}},
		},
	}}
	replaceReturnsInBlock(block)
	ifStmt := block.List[0].(*ast.IfStmt)
	thenBlock := ifStmt.Body.List[0].(*ast.BlockStmt)
	if len(thenBlock.List) != 2 {
		t.Fatalf("expected 2 stmts in then, got %d", len(thenBlock.List))
	}
	elseBlock := ifStmt.Else.(*ast.BlockStmt)
	elseReplaced := elseBlock.List[0].(*ast.BlockStmt)
	if len(elseReplaced.List) != 2 {
		t.Fatalf("expected 2 stmts in else, got %d", len(elseReplaced.List))
	}
}

func TestReplaceReturnsInBlock_NestedInFor(t *testing.T) {
	block := &ast.BlockStmt{List: []ast.Stmt{
		&ast.ForStmt{Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "i"}}},
		}}},
	}}
	replaceReturnsInBlock(block)
	forStmt := block.List[0].(*ast.ForStmt)
	replaced := forStmt.Body.List[0].(*ast.BlockStmt)
	if len(replaced.List) != 2 {
		t.Fatalf("expected 2 stmts, got %d", len(replaced.List))
	}
}

func TestReplaceReturnsInBlock_MultipleReturns(t *testing.T) {
	block := &ast.BlockStmt{List: []ast.Stmt{
		&ast.ExprStmt{X: &ast.Ident{Name: "work"}},
		&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "first"}}},
		&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "second"}}},
	}}
	replaceReturnsInBlock(block)
	if len(block.List) != 3 {
		t.Fatalf("expected 3 stmts, got %d", len(block.List))
	}
	r1 := block.List[1].(*ast.BlockStmt).List[0].(*ast.AssignStmt)
	assertIdent(t, r1.Rhs[0], "first")
	r2 := block.List[2].(*ast.BlockStmt).List[0].(*ast.AssignStmt)
	assertIdent(t, r2.Rhs[0], "second")
}

func TestCreateResultReturnBlock(t *testing.T) {
	ret := &ast.ReturnStmt{Results: []ast.Expr{
		&ast.BinaryExpr{X: &ast.Ident{Name: "a"}, Op: token.ADD, Y: &ast.Ident{Name: "b"}},
	}}
	block := createResultReturnBlock(ret)
	if len(block.List) != 2 {
		t.Fatalf("expected 2 stmts, got %d", len(block.List))
	}
	assign := block.List[0].(*ast.AssignStmt)
	assertIdent(t, assign.Lhs[0], "__result")
	if assign.Tok != token.ASSIGN {
		t.Errorf("expected ASSIGN, got %v", assign.Tok)
	}
	newRet := block.List[1].(*ast.ReturnStmt)
	assertIdent(t, newRet.Results[0], "__result")
}

func TestProcessFunc_NilBody(t *testing.T) {
	fd := &ast.FuncDecl{
		Name: &ast.Ident{Name: "abstract"},
		Type: &ast.FuncType{Params: &ast.FieldList{}},
	}
	contract := &specir.FuncContract{
		Requires: []specir.NamedPredicate{{Label: "pos", Expr: &specir.ExprVar{Name: "x"}}},
	}
	if processFunc(fd, contract, makeEmptyASTData(), true) {
		t.Error("should return false for nil body")
	}
}

func TestProcessFunc_NoContractNoAsserts(t *testing.T) {
	fd := &ast.FuncDecl{
		Name: &ast.Ident{Name: "plain"},
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{}}},
	}
	if processFunc(fd, nil, makeEmptyASTData(), true) {
		t.Error("should return false with no contract and no asserts")
	}
}

func TestProcessFunc_StrictMode_RequiresOnly(t *testing.T) {
	fd := &ast.FuncDecl{
		Name: &ast.Ident{Name: "foo"},
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: &ast.Ident{Name: "work"}}}},
	}
	contract := &specir.FuncContract{
		Requires: []specir.NamedPredicate{{Label: "pos", Expr: &specir.ExprBinary{
			Op: specir.OpGt, Left: &specir.ExprVar{Name: "x"}, Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 0},
		}}},
	}
	if !processFunc(fd, contract, makeEmptyASTData(), true) {
		t.Fatal("should return true")
	}
	if len(fd.Body.List) < 2 {
		t.Fatalf("expected >= 2 stmts, got %d", len(fd.Body.List))
	}
	es := fd.Body.List[0].(*ast.ExprStmt)
	call := es.X.(*ast.CallExpr)
	assertIdent(t, call.Fun, "checks.Check")
}

func TestProcessFunc_StrictMode_EnsuresWithResult(t *testing.T) {
	fd := &ast.FuncDecl{
		Name: &ast.Ident{Name: "abs"},
		Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "int"}}}},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "x"}}},
		}},
	}
	contract := &specir.FuncContract{
		Ensures: []specir.NamedPredicate{{Expr: &specir.ExprBinary{
			Op: specir.OpGe, Left: &specir.ExprResult{}, Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 0},
		}}},
	}
	if !processFunc(fd, contract, makeEmptyASTData(), true) {
		t.Fatal("should return true")
	}
	var hasVar, hasDefer, hasResultRet bool
	for _, stmt := range fd.Body.List {
		if ds, ok := stmt.(*ast.DeclStmt); ok {
			if gd, ok := ds.Decl.(*ast.GenDecl); ok && gd.Tok == token.VAR {
				if gd.Specs[0].(*ast.ValueSpec).Names[0].Name == "__result" {
					hasVar = true
				}
			}
		}
		if _, ok := stmt.(*ast.DeferStmt); ok {
			hasDefer = true
		}
		if bs, ok := stmt.(*ast.BlockStmt); ok {
			for _, s := range bs.List {
				if r, ok := s.(*ast.ReturnStmt); ok && len(r.Results) == 1 {
					if id, ok := r.Results[0].(*ast.Ident); ok && id.Name == "__result" {
						hasResultRet = true
					}
				}
			}
		}
	}
	if !hasVar {
		t.Error("missing var __result")
	}
	if !hasDefer {
		t.Error("missing defer postcondition")
	}
	if !hasResultRet {
		t.Error("missing return __result")
	}
}

func TestProcessFunc_LenientMode(t *testing.T) {
	fd := &ast.FuncDecl{
		Name: &ast.Ident{Name: "foo"},
		Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "int"}}}},
		},
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "x"}}},
		}},
	}
	contract := &specir.FuncContract{
		Requires: []specir.NamedPredicate{{Expr: &specir.ExprBinary{
			Op: specir.OpGt, Left: &specir.ExprVar{Name: "x"}, Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 0},
		}}},
		Ensures: []specir.NamedPredicate{{Expr: &specir.ExprBinary{
			Op: specir.OpGe, Left: &specir.ExprResult{}, Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 0},
		}}},
	}
	if !processFunc(fd, contract, makeEmptyASTData(), false) {
		t.Fatal("should return true")
	}
	var hasPrecondVar, hasIfNot, hasDeferGuard bool
	for _, stmt := range fd.Body.List {
		if a, ok := stmt.(*ast.AssignStmt); ok && a.Tok == token.DEFINE {
			if id, ok := a.Lhs[0].(*ast.Ident); ok && id.Name == "__preconditions_passed" {
				hasPrecondVar = true
			}
		}
		if is, ok := stmt.(*ast.IfStmt); ok {
			if u, ok := is.Cond.(*ast.UnaryExpr); ok && u.Op == token.NOT {
				hasIfNot = true
			}
		}
		if ds, ok := stmt.(*ast.DeferStmt); ok {
			if _, ok := ds.Call.Fun.(*ast.FuncLit); ok {
				hasDeferGuard = true
			}
		}
	}
	if !hasPrecondVar {
		t.Error("missing __preconditions_passed decl")
	}
	if !hasIfNot {
		t.Error("missing if !condition guard")
	}
	if !hasDeferGuard {
		t.Error("missing defer with func literal guard")
	}
}

func TestProcessFunc_WithOld(t *testing.T) {
	fd := &ast.FuncDecl{
		Name: &ast.Ident{Name: "swap"},
		Type: &ast.FuncType{Params: &ast.FieldList{List: []*ast.Field{
			{Names: []*ast.Ident{{Name: "p"}}, Type: &ast.StarExpr{X: &ast.Ident{Name: "int"}}},
		}}},
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: &ast.Ident{Name: "doSwap"}}}},
	}
	inner := &specir.ExprVar{Name: "*p"}
	contract := &specir.FuncContract{
		Ensures: []specir.NamedPredicate{{Expr: &specir.ExprBinary{
			Op: specir.OpEq, Left: &specir.ExprVar{Name: "*p"}, Right: &specir.ExprOld{Inner: inner},
		}}},
	}
	if !processFunc(fd, contract, makeEmptyASTData(), true) {
		t.Fatal("should return true")
	}
	var hasOld bool
	for _, stmt := range fd.Body.List {
		if a, ok := stmt.(*ast.AssignStmt); ok && a.Tok == token.DEFINE {
			if id, ok := a.Lhs[0].(*ast.Ident); ok && id.Name == "__old_1" {
				hasOld = true
			}
		}
	}
	if !hasOld {
		t.Error("missing __old_1 assignment")
	}
}

func TestProcessFunc_DirectDefer_NoRequires(t *testing.T) {
	fd := &ast.FuncDecl{
		Name: &ast.Ident{Name: "foo"},
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: &ast.Ident{Name: "work"}}}},
	}
	contract := &specir.FuncContract{
		Ensures: []specir.NamedPredicate{{Label: "safe", Expr: &specir.ExprLiteral{Kind: specir.LitBool, Bool: true}}},
	}
	processFunc(fd, contract, makeEmptyASTData(), true)
	var found bool
	for _, stmt := range fd.Body.List {
		if ds, ok := stmt.(*ast.DeferStmt); ok {
			if id, ok := ds.Call.Fun.(*ast.Ident); ok && id.Name == "checks.Check" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected direct defer checks.Check()")
	}
}

func TestProcessFunc_MultipleRequires(t *testing.T) {
	fd := &ast.FuncDecl{
		Name: &ast.Ident{Name: "foo"},
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: &ast.Ident{Name: "work"}}}},
	}
	contract := &specir.FuncContract{
		Requires: []specir.NamedPredicate{
			{Label: "r1", Expr: &specir.ExprVar{Name: "a"}},
			{Label: "r2", Expr: &specir.ExprVar{Name: "b"}},
			{Label: "r3", Expr: &specir.ExprVar{Name: "c"}},
		},
	}
	processFunc(fd, contract, makeEmptyASTData(), true)
	count := 0
	for _, stmt := range fd.Body.List {
		if es, ok := stmt.(*ast.ExprStmt); ok {
			if call, ok := es.X.(*ast.CallExpr); ok {
				if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "checks.Check" {
					count++
				}
			}
		}
	}
	if count != 3 {
		t.Errorf("expected 3 checks, got %d", count)
	}
}

func TestProcessFunc_MultipleEnsures(t *testing.T) {
	fd := &ast.FuncDecl{
		Name: &ast.Ident{Name: "foo"},
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: &ast.Ident{Name: "work"}}}},
	}
	contract := &specir.FuncContract{
		Ensures: []specir.NamedPredicate{
			{Expr: &specir.ExprVar{Name: "a"}},
			{Expr: &specir.ExprVar{Name: "b"}},
		},
	}
	processFunc(fd, contract, makeEmptyASTData(), true)
	count := 0
	for _, stmt := range fd.Body.List {
		if _, ok := stmt.(*ast.DeferStmt); ok {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 defers, got %d", count)
	}
}

func TestCollectAssertSpecs_NoAsserts(t *testing.T) {
	fd := &ast.FuncDecl{
		Name: &ast.Ident{Name: "foo"},
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{X: &ast.Ident{Name: "work"}}}},
	}
	asserts := collectAssertSpecs(fd, makeEmptyASTData())
	if len(asserts) != 0 {
		t.Errorf("expected 0, got %d", len(asserts))
	}
}

func TestCollectAssertSpecs_WithAsserts(t *testing.T) {
	s1 := &ast.ExprStmt{X: &ast.Ident{Name: "a"}}
	s2 := &ast.ExprStmt{X: &ast.Ident{Name: "b"}}
	fd := &ast.FuncDecl{
		Name: &ast.Ident{Name: "foo"},
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{List: []ast.Stmt{s1, s2}},
	}
	ad := makeEmptyASTData()
	info := data.NewAdditionalInfo()
	info.Asserts = &specir.AssertSpec{Assertion: specir.NamedPredicate{Expr: &specir.ExprVar{Name: "x"}}}
	ad.NodeInfo[s1] = info
	asserts := collectAssertSpecs(fd, ad)
	if len(asserts) != 1 {
		t.Fatalf("expected 1, got %d", len(asserts))
	}
	if asserts[0].stmt != s1 {
		t.Error("wrong stmt bound")
	}
}

func TestInsertAssertChecks_BeforeStatement(t *testing.T) {
	s1 := &ast.ExprStmt{X: &ast.Ident{Name: "a"}}
	s2 := &ast.ExprStmt{X: &ast.Ident{Name: "b"}}
	asserts := []*assertSpecWithNode{{
		stmt:   s2,
		assert: &specir.AssertSpec{Assertion: specir.NamedPredicate{Label: "mid", Expr: &specir.ExprVar{Name: "x"}}},
	}}
	result := insertAssertChecksIntoBlock([]ast.Stmt{s1, s2}, asserts, NewConverter())
	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}
	cs := result[1].(*ast.ExprStmt)
	call := cs.X.(*ast.CallExpr)
	assertIdent(t, call.Fun, "checks.Check")
	if result[0] != s1 {
		t.Error("s1 should be at 0")
	}
	if result[2] != s2 {
		t.Error("s2 should be at 2")
	}
}

func TestProcessFunctions_Integration(t *testing.T) {
	src := `package test
func abs(x int) int {
	if x < 0 { return -x }
	return x
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	ad := data.NewASTData(file, fset)
	var fd *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if f, ok := n.(*ast.FuncDecl); ok {
			fd = f
		}
		return true
	})
	info := data.NewAdditionalInfo()
	info.FuncContract = &specir.FuncContract{
		Ensures: []specir.NamedPredicate{{Expr: &specir.ExprBinary{
			Op: specir.OpGe, Left: &specir.ExprResult{}, Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 0},
		}}},
	}
	ad.NodeInfo[fd] = info
	result := plugin.NewPluginResult()
	processFunctions(ad, result, true)
	if !result.ASTModified {
		t.Error("expected ASTModified")
	}
	if !result.MapModified {
		t.Error("expected MapModified")
	}
}

func TestConvertMapToSlice(t *testing.T) {
	s1 := &ast.ExprStmt{X: &ast.Ident{Name: "a"}}
	a1 := &specir.AssertSpec{Assertion: specir.NamedPredicate{Expr: &specir.ExprVar{Name: "x"}}}
	m := map[ast.Stmt]*specir.AssertSpec{s1: a1}
	sl := convertMapToSlice(m)
	if len(sl) != 1 {
		t.Fatalf("expected 1, got %d", len(sl))
	}
	if sl[0].stmt != s1 || sl[0].assert != a1 {
		t.Error("wrong mapping")
	}
}
