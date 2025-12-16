package front

import (
	"fmt"
	acsl "gpma/framework/gosl/front/generated/grammar/acsl"
	specir "gpma/framework/gosl/ir"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

func (b *ACSLBuilder) VisitAcsl_comment(ctx *acsl.Acsl_commentContext) interface{} {
	switch {
	case ctx.Acsl_comment_contract() != nil:
		return b.Visit(ctx.Acsl_comment_contract())

	case ctx.Acsl_comment_loop_annot() != nil:
		return b.Visit(ctx.Acsl_comment_loop_annot())

	case ctx.Acsl_comment_assertion() != nil:
		return b.Visit(ctx.Acsl_comment_assertion())

	case ctx.Acsl_comment_statement_contract() != nil:
		return b.Visit(ctx.Acsl_comment_statement_contract())

	default:
		return nil
	}
}

func (b *ACSLBuilder) VisitAcsl_comment_contract(ctx *acsl.Acsl_comment_contractContext) interface{} {
	fc := &specir.FuncContract{
		FuncName:     "",
		RecvTypeName: "",
		Span: b.blockSpan(ctx),
	}

	b.Mod.FuncContracts = append(b.Mod.FuncContracts, fc)
	// opening function contract context
	b.currFunc = fc
	// closing function contract context at VisitAcsl_comment_contract ending
	defer func() { b.currFunc = nil }()

    fcCtx := ctx.Function_contract()
	if fcCtx != nil {
		b.Visit(fcCtx)
	}

    fc.StartPos, fc.EndPos = b.blockPositions(fcCtx)
    
	fmt.Println("[ACSLBuilder] added FuncContract. Text:", ctx.GetText())
	// fmt.Println("[ACSLBuilder] added FuncContract:", fc.FuncName, "Text:", ctx.GetText(), "comment_span:", b.blockSpan(ctx), "contract_span:", b.blockSpan(fcCtx))
    // fmt.Println("contract_start:", fc.StartPos, "contract_end:", fc.EndPos)
    // fmt.Println("block_start:", b.Block.Start, "block_end:", b.Block.End)
	return nil
}

func (b *ACSLBuilder) VisitAcsl_comment_loop_annot(ctx *acsl.Acsl_comment_loop_annotContext) interface{} {
	ls := &specir.LoopSpec{
		ParentFuncName: "",
		Span: b.blockSpan(ctx),
	}

	b.Mod.LoopSpecs = append(b.Mod.LoopSpecs, ls)

	fmt.Println("[ACSLBuilder] added LoopSpec in:", ls.ParentFuncName, "Text:", ctx.GetText())
	return nil
}

func (b *ACSLBuilder) VisitAcsl_comment_assertion(ctx *acsl.Acsl_comment_assertionContext) interface{} {
	fmt.Println("[ACSLBuilder] assertion (stub)", "Text:", ctx.GetText())
	return nil
}

func (b *ACSLBuilder) VisitAcsl_comment_statement_contract(ctx *acsl.Acsl_comment_statement_contractContext) interface{} {
	fmt.Println("[ACSLBuilder] statement contract (stub)", "Text:", ctx.GetText())
	return nil
}

// Function contract: requires/ensures/assigns
func (b *ACSLBuilder) VisitFunction_contract(ctx *acsl.Function_contractContext) interface{} {
	fmt.Println("VisitFunction_contract text:", ctx.GetText())
    if b.currFunc == nil || ctx == nil {
        return nil
    }

    // requires
    for _, rc := range ctx.AllRequires_clause() {
        np, ok := b.Visit(rc).(specir.NamedPredicate)
        if ok {
            b.currFunc.Requires = append(b.currFunc.Requires, np)
        }
    }

    // simple_clause usually contains ensures/assigns ...
    for _, sc := range ctx.AllSimple_clause() {
        b.Visit(sc)
    }

    return nil
}

func (b *ACSLBuilder) VisitSimple_clause(ctx *acsl.Simple_clauseContext) interface{} {
	fmt.Println("VisitSimple_clause text:", ctx.GetText())
    if b.currFunc == nil || ctx == nil {
        return nil
    }
    if ec := ctx.Ensures_clause(); ec != nil {
        np, ok := b.Visit(ec).(specir.NamedPredicate)
        if ok {
            b.currFunc.Ensures = append(b.currFunc.Ensures, np)
        }
        return nil
    }
    if ac := ctx.Assigns_clause(); ac != nil {
        ls, ok := b.Visit(ac).(specir.LocationSet)
        if ok {
            b.currFunc.Assigns = ls
        }
        return nil
    }
    return nil
}

// requires_clause / ensures_clause -> NamedPredicate
func (b *ACSLBuilder) VisitRequires_clause(ctx *acsl.Requires_clauseContext) interface{} {
	fmt.Println("VisitRequires_clause text:", ctx.GetText())
    if ctx == nil {
        return nil
    }
    return b.parseNamedPred(ctx.Pred())
}

func (b *ACSLBuilder) VisitEnsures_clause(ctx *acsl.Ensures_clauseContext) interface{} {
	fmt.Println("VisitEnsures_clause text:", ctx.GetText())
    if ctx == nil {
        return nil
    }
    return b.parseNamedPred(ctx.Pred())
}

func (b *ACSLBuilder) parseNamedPred(p acsl.IPredContext) specir.NamedPredicate {
	fmt.Println("parseNamedPred text:", p.GetText())
    // default without label
    np := specir.NamedPredicate{
        Label: "",
        Expr:  nil,
        Span:  b.blockSpan(p), // TODO
    }
    if p == nil {
        return np
    }

    // syntactic naming: "bound: <pred>"
    if sn, ok := p.(*acsl.Syntactic_naming_predContext); ok {
        // label can be String_ or Id(). Grammar dependent
        if sn.String_() != nil {
            np.Label = sn.String_().GetText()
        } else if sn.Id() != nil {
            np.Label = sn.Id().GetText()
        }
        np.Expr = b.predToExpr(sn.Pred())
        return np
    }

    // Simple pred
    np.Expr = b.predToExpr(p)
    return np
}


// assigns_clause -> LocationSet
func (b *ACSLBuilder) VisitAssigns_clause(ctx *acsl.Assigns_clauseContext) interface{} {
	fmt.Println("VisitAssigns_clause text:", ctx.GetText())
    if ctx == nil {
        return specir.LocationSet{Kind: specir.LocSetNothing}
    }
    txt := ctx.GetText()

    // \nothing
    if strings.Contains(txt, `\nothing`) {
        return specir.LocationSet{Kind: specir.LocSetNothing}
    }

    // collecting identifiers. TODO
    vars := collectIdentifiers(ctx)
    if len(vars) == 0 {
        // fallback
        return specir.LocationSet{Kind: specir.LocSetVars, Vars: []string{txt}}
    }
    return specir.LocationSet{Kind: specir.LocSetVars, Vars: unique(vars)}
}

// subtree walk and collecting Identifier tokens TODO
func collectIdentifiers(t antlr.Tree) []string {
    var out []string
    var walk func(n antlr.Tree)
    walk = func(n antlr.Tree) {
        if n == nil {
            return
        }
        if tn, ok := n.(antlr.TerminalNode); ok {
            sym := tn.GetSymbol()
            if sym != nil && sym.GetTokenType() == acsl.ACSLParserIdentifier {
                out = append(out, sym.GetText())
            }
        }
        for i := 0; i < n.GetChildCount(); i++ {
            if c, ok := n.GetChild(i).(antlr.Tree); ok {
                walk(c)
            }
        }
    }
    walk(t)
    return out
}

func unique(xs []string) []string {
    seen := map[string]struct{}{}
    out := make([]string, 0, len(xs))
    for _, x := range xs {
        if _, ok := seen[x]; ok {
            continue
        }
        seen[x] = struct{}{}
        out = append(out, x)
    }
    return out
}

// specir.Expr
func (b *ACSLBuilder) predToExpr(p acsl.IPredContext) specir.Expr {
	fmt.Println("predToExpr text:", p.GetText())
    if p == nil {
        return nil
    }
    v := b.Visit(p)
    if e, ok := v.(specir.Expr); ok {
        return e
    }
    if np, ok := v.(specir.NamedPredicate); ok {
        // returning NamedPredicate
        return np.Expr
    }
    // fallback
    return &specir.ExprVar{BaseExpr: specir.BaseExpr{Span: b.blockSpan(p)}, Name: p.GetText()}
}

func (b *ACSLBuilder) VisitLogical_true_pred(ctx *acsl.Logical_true_predContext) interface{} {
	fmt.Println("VisitLogical_true_pred text:", ctx.GetText())
    return &specir.ExprLiteral{
        BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
        Kind:     specir.LitBool,
        Bool:     true,
    }
}

func (b *ACSLBuilder) VisitLogical_false_pred(ctx *acsl.Logical_false_predContext) interface{} {
	fmt.Println("VisitLogical_false_pred text:", ctx.GetText())
    return &specir.ExprLiteral{
        BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
        Kind:     specir.LitBool,
        Bool:     false,
    }
}

// a <= b, a == b, ...
func (b *ACSLBuilder) VisitComparison_pred(ctx *acsl.Comparison_predContext) interface{} {
    fmt.Println("VisitComparison_pred text:", ctx.GetText())
	//Term Rel_op Term 
    terms := ctx.AllTerm()
    ops := ctx.AllRel_op()
    if len(terms) < 2 || len(ops) < 1 {
        return &specir.ExprVar{BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)}, Name: ctx.GetText()}
    }

    left := b.termToExpr(terms[0])
    right := b.termToExpr(terms[1])

    op := relOpToBinary(ops[0].GetText())
    return &specir.ExprBinary{
        BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
        Op:       op,
        Left:     left,
        Right:    right,
    }
}

func relOpToBinary(s string) specir.BinaryOp {
    switch s {
    case "==":
        return specir.OpEq
    case "!=":
        return specir.OpNeq
    case "<=":
        return specir.OpLe
    case ">=":
        return specir.OpGe
    case "<":
        return specir.OpLt
    case ">":
        return specir.OpGt
    default:
        return specir.OpEq
    }
}

// && / ||
func (b *ACSLBuilder) VisitConjunction_pred(ctx *acsl.Conjunction_predContext) interface{} {
    fmt.Println("VisitConjunction_pred text:", ctx.GetText())
	ps := ctx.AllPred()
    if len(ps) < 2 {
        return b.predToExpr(ctx.Pred(0))
    }
    return &specir.ExprBinary{
        BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
        Op:       specir.OpAnd,
        Left:     b.predToExpr(ps[0]),
        Right:    b.predToExpr(ps[1]),
    }
}

func (b *ACSLBuilder) VisitDisjunction_pred(ctx *acsl.Disjunction_predContext) interface{} {
	fmt.Println("VisitDisjunction_pred text:", ctx.GetText())
    ps := ctx.AllPred()
    if len(ps) < 2 {
        return b.predToExpr(ctx.Pred(0))
    }
    return &specir.ExprBinary{
        BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
        Op:       specir.OpOr,
        Left:     b.predToExpr(ps[0]),
        Right:    b.predToExpr(ps[1]),
    }
}

// pred: cond ? p1 : p2
func (b *ACSLBuilder) VisitTernary_condition_pred(ctx *acsl.Ternary_condition_predContext) interface{} {
	fmt.Println("VisitTernary_condition_pred text:", ctx.GetText())
    ps := ctx.AllPred()
    if len(ps) != 3 {
        return &specir.ExprVar{BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)}, Name: ctx.GetText()}
    }

    cond := b.predToExpr(ps[0])
    thenE := b.predToExpr(ps[1])
    elseE := b.predToExpr(ps[2])
    return &specir.ExprConditional{
        BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
        Cond:     cond,
        Then:     thenE,
        Else:     elseE,
    }
}

// \old(pred)
func (b *ACSLBuilder) VisitOld_pred(ctx *acsl.Old_predContext) interface{} {
	fmt.Println("VisitOld_pred text:", ctx.GetText())
    inner := b.predToExpr(ctx.Pred())
    return &specir.ExprOld{
        BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
        Inner:    inner,
    }
}

// terms
func (b *ACSLBuilder) termToExpr(t acsl.ITermContext) specir.Expr {
    if t == nil {
        return nil
    }
    v := b.Visit(t)
    if e, ok := v.(specir.Expr); ok {
        return e
    }
    return &specir.ExprVar{BaseExpr: specir.BaseExpr{Span: b.blockSpan(t)}, Name: t.GetText()}
}

func (b *ACSLBuilder) VisitInternal_term(ctx *acsl.Internal_termContext) interface{} {
	fmt.Println("VisitInternal_term text:", ctx.GetText())
    // "id" / literal / ...
    return b.VisitChildren(ctx)
}

func (b *ACSLBuilder) VisitFunc_application_term(ctx *acsl.Func_application_termContext) interface{} {
	fmt.Println("VisitFunc_application_term text:", ctx.GetText())
    // term ::= ident '(' term (',' term)* ')'
    name := ""
    // if ctx.Id() != nil {
    //     name = ctx.Id().GetText()
    // }
    argsCtx := ctx.AllTerm()
    args := make([]specir.Expr, 0, len(argsCtx))
    for _, a := range argsCtx {
        args = append(args, b.termToExpr(a))
    }
    return &specir.ExprCall{
        BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
        Name:     name,
        Args:     args,
    }
}

func (b *ACSLBuilder) VisitOld_term(ctx *acsl.Old_termContext) interface{} {
	fmt.Println("VisitOld_term text:", ctx.GetText())
    inner := b.termToExpr(ctx.Term())
    return &specir.ExprOld{
        BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
        Inner:    inner,
    }
}

func (b *ACSLBuilder) VisitResult_term(ctx *acsl.Result_termContext) interface{} {
	fmt.Println("VisitResult_term text:", ctx.GetText())
    return &specir.ExprResult{BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)}}
}

func (b *ACSLBuilder) VisitParentheses_term(ctx *acsl.Parentheses_termContext) interface{} {
	fmt.Println("VisitParentheses_term text:", ctx.GetText())
    return b.termToExpr(ctx.Term())
}
