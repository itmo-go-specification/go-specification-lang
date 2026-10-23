package front

import (
	acsl "gpma/framework/gosl/front/generated/grammar/acsl"
	grammar "gpma/framework/gosl/front/generated/grammar/acsl"
	specir "gpma/framework/gosl/ir"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	log "github.com/sirupsen/logrus"
)

func (b *ACSLBuilder) VisitAcsl_comment(ctx *acsl.Acsl_commentContext) interface{} {
	switch {
	case ctx.Acsl_comment_logic_defs() != nil:
		log.Debug("logic defs\n")
		return b.Visit(ctx.Acsl_comment_logic_defs())

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

func (b *ACSLBuilder) parseNamedPred(p acsl.IPredContext) specir.NamedPredicate {
	if p == nil {
		return specir.NamedPredicate{}
	}
	log.Debugf("parseNamedPred text:%s, type:%T", p.GetText(), p)
	np := specir.NamedPredicate{Label: "", Expr: nil, Span: b.blockSpan(p)}

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
	log.Debugln("parseNamedPred: Simple")
	// Simple pred
	np.Expr = b.predToExpr(p)
	return np
}

// assigns_clause -> LocationSet
func (b *ACSLBuilder) VisitAssigns_clause(ctx *acsl.Assigns_clauseContext) interface{} {
	log.Debug("VisitAssigns_clause text:", ctx.GetText())
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
	if p == nil {
		return nil
	}
	log.Debugf("predToExpr text:%s, type:%T", p.GetText(), p)

	v := b.Visit(p)
	if e, ok := v.(specir.Expr); ok {
		return e
	}
	if np, ok := v.(specir.NamedPredicate); ok {
		// returning NamedPredicate
		return np.Expr
	}
	// fallback
	log.Debug("predToExpr fallback")
	return &specir.ExprVar{BaseExpr: specir.BaseExpr{Span: b.blockSpan(p)}, Name: p.GetText()}
}

// VisitSyntactic_naming_pred handles "label: pred" inside predicate position.
// It delegates to parseNamedPred so the label is preserved in the returned NamedPredicate.
func (b *ACSLBuilder) VisitSyntactic_naming_pred(ctx *acsl.Syntactic_naming_predContext) interface{} {
	log.Debugf("VisitSyntactic_naming_pred text:%s\n", ctx.GetText())
	return b.parseNamedPred(ctx)
}

func (b *ACSLBuilder) VisitLogical_true_pred(ctx *acsl.Logical_true_predContext) interface{} {
	log.Debug("VisitLogical_true_pred text:", ctx.GetText())
	return &specir.ExprLiteral{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Kind:     specir.LitBool,
		Bool:     true,
	}
}

func (b *ACSLBuilder) VisitLogical_false_pred(ctx *acsl.Logical_false_predContext) interface{} {
	log.Debug("VisitLogical_false_pred text:", ctx.GetText())
	return &specir.ExprLiteral{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Kind:     specir.LitBool,
		Bool:     false,
	}
}

// Comparison predicates may contain more than one comparison.
// a < b < c is represented as (a < b) && (b < c).
func (b *ACSLBuilder) VisitComparison_pred(ctx *acsl.Comparison_predContext) interface{} {
	terms := ctx.AllTerm()
	ops := ctx.AllRel_op()
	if len(terms) < 2 || len(ops) != len(terms)-1 {
		return b.rawExpr(ctx)
	}

	var result specir.Expr
	for i := 0; i < len(ops); i++ {
		op, ok := relOpToBinary(ops[i].GetText())
		if !ok {
			return b.rawExpr(ctx)
		}
		pair := specir.Expr(&specir.ExprBinary{
			BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
			Op:       op,
			Left:     b.termToExpr(terms[i]),
			Right:    b.termToExpr(terms[i+1]),
		})
		if result == nil {
			result = pair
		} else {
			result = &specir.ExprBinary{
				BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
				Op:       specir.OpAnd,
				Left:     result,
				Right:    pair,
			}
		}
	}
	return result
}

func relOpToBinary(s string) (specir.BinaryOp, bool) {
	switch s {
	case "+":
		return specir.OpAdd, true
	case "-":
		return specir.OpSub, true
	case "*":
		return specir.OpMul, true
	case "/":
		return specir.OpDiv, true
	case "%":
		return specir.OpMod, true
	case "==":
		return specir.OpEq, true
	case "!=":
		return specir.OpNeq, true
	case "<=":
		return specir.OpLe, true
	case ">=":
		return specir.OpGe, true
	case "<":
		return specir.OpLt, true
	case ">":
		return specir.OpGt, true
	case "&&":
		return specir.OpAnd, true
	case "||":
		return specir.OpOr, true
	case "^^":
		return specir.OpXor, true
	case "==>":
		return specir.OpImplies, true
	case "|":
		return specir.OpBitwiseOr, true
	case "^":
		return specir.OpBitwiseXor, true
	case "&":
		return specir.OpBitwiseAnd, true
	case "<<":
		return specir.OpShiftLeft, true
	case ">>":
		return specir.OpShiftRight, true
	default:
		return 0, false
	}
}

// && / ||
func (b *ACSLBuilder) VisitConjunction_pred(ctx *acsl.Conjunction_predContext) interface{} {
	log.Debug("VisitConjunction_pred text:", ctx.GetText())
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
	log.Debug("VisitDisjunction_pred text:", ctx.GetText())
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

// pred ==> pred
func (b *ACSLBuilder) VisitImplication_pred(ctx *acsl.Implication_predContext) interface{} {
	log.Debug("VisitImplication_pred text:", ctx.GetText())
	ps := ctx.AllPred()
	if len(ps) < 2 {
		return b.predToExpr(ctx.Pred(0))
	}
	return &specir.ExprBinary{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Op:       specir.OpImplies,
		Left:     b.predToExpr(ps[0]),
		Right:    b.predToExpr(ps[1]),
	}
}

// pred <==> pred
func (b *ACSLBuilder) VisitEquivalence_pred(ctx *acsl.Equivalence_predContext) interface{} {
	log.Debug("VisitEquivalence_pred text:", ctx.GetText())
	ps := ctx.AllPred()
	if len(ps) < 2 {
		return b.predToExpr(ctx.Pred(0))
	}
	// Represent A <==> B as (A ==> B) && (B ==> A)
	left := b.predToExpr(ps[0])
	right := b.predToExpr(ps[1])
	fwd := &specir.ExprBinary{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Op:       specir.OpImplies,
		Left:     left,
		Right:    right,
	}
	bwd := &specir.ExprBinary{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Op:       specir.OpImplies,
		Left:     right,
		Right:    left,
	}
	return &specir.ExprBinary{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Op:       specir.OpAnd,
		Left:     fwd,
		Right:    bwd,
	}
}

// pred: cond ? p1 : p2
func (b *ACSLBuilder) VisitTernary_condition_pred(ctx *acsl.Ternary_condition_predContext) interface{} {
	log.Debug("VisitTernary_condition_pred text:", ctx.GetText())
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
	log.Debug("VisitOld_pred text:", ctx.GetText())
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
	log.Debug("VisitInternal_term text:", ctx.GetText())
	// "id" / literal / ...
	return b.VisitChildren(ctx)
}

func polyIDName(poly acsl.IPoly_idContext) (string, bool) {
	if poly == nil {
		return "", false
	}
	idTok := poly.Identifier()
	if idTok == nil {
		return "", false
	}
	name := idTok.GetText()
	if name == "" {
		return "", false
	}
	return name, true
}

func (b *ACSLBuilder) VisitFunc_application_term(ctx *acsl.Func_application_termContext) interface{} {
	log.Debugln("VisitFunc_application_term text:", ctx.GetText())

	name := ""
	if n, ok := polyIDName(ctx.Poly_id()); ok {
		name = n
	}

	return b.buildCallExpr(ctx, name, ctx.AllTerm())
}

func (b *ACSLBuilder) VisitPredicate_application_pred(ctx *acsl.Predicate_application_predContext) interface{} {
	log.Debugln("VisitPredicate_application_pred text:", ctx.GetText())

	name := ""
	if ctx.Ident() != nil {
		name = ctx.Ident().GetText()
	}

	return b.buildCallExpr(ctx, name, ctx.AllTerm())
}

func (b *ACSLBuilder) buildCallExpr(spanCtx antlr.ParserRuleContext, name string, terms []acsl.ITermContext) specir.Expr {
	args := make([]specir.Expr, 0, len(terms))

	for _, t := range terms {
		args = append(args, b.termToExpr(t))
	}

	return &specir.ExprCall{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(spanCtx)},
		Name:     name,
		Args:     args,
	}
}

func (b *ACSLBuilder) VisitOld_term(ctx *acsl.Old_termContext) interface{} {
	log.Debug("VisitOld_term text:", ctx.GetText())
	inner := b.termToExpr(ctx.Term())
	return &specir.ExprOld{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Inner:    inner,
	}
}

func (b *ACSLBuilder) VisitResult_term(ctx *acsl.Result_termContext) interface{} {
	log.Debug("VisitResult_term text:", ctx.GetText())
	return &specir.ExprResult{BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)}, Index: 0}
}

// VisitResult_n_term handles \result_N — indexed result for multi-return functions.
// \result_0 is equivalent to \result (first return value).
func (b *ACSLBuilder) VisitResult_n_term(ctx *acsl.Result_n_termContext) interface{} {
	log.Debug("VisitResult_n_term text:", ctx.GetText())
	idx := 0
	if tok := ctx.Constant(); tok != nil {
		if n, err := strconv.Atoi(tok.GetText()); err == nil {
			idx = n
		}
	}
	return &specir.ExprResult{BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)}, Index: idx}
}

func (b *ACSLBuilder) VisitParentheses_pred(ctx *acsl.Parentheses_predContext) interface{} {
	log.Debug("VisitParentheses_pred text:", ctx.GetText())
	return b.predToExpr(ctx.Pred())
}

func (b *ACSLBuilder) VisitLiteral_term(ctx *grammar.Literal_termContext) interface{} {
	if ctx == nil {
		return nil
	}

	lit := ctx.Literal()

	log.Debugf("VisitLiteral_term text:%s, type:%T\n", ctx.GetText(), lit)
	if lit == nil {
		// fallback
		return &specir.ExprVar{
			BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
			Name:     ctx.GetText(),
		}
	}

	switch x := lit.(type) {

	case *grammar.True_constantContext:
		return &specir.ExprLiteral{
			BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
			Kind:     specir.LitBool,
			Bool:     true,
		}

	case *grammar.False_constantContext:
		return &specir.ExprLiteral{
			BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
			Kind:     specir.LitBool,
			Bool:     false,
		}

	case *grammar.String_constantContext:
		raw := ""
		if s := x.String_(); s != nil {
			raw = s.GetText()
		} else {
			raw = x.GetText()
		}

		// removing ".."
		unq, err := strconv.Unquote(raw)
		if err != nil {
			unq = raw // fallback
		}

		return &specir.ExprLiteral{
			BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
			Kind:     specir.LitString,
			Str:      unq,
		}

	case *grammar.Trivial_constantContext:
		// Constant()
		raw := ""
		if c := x.Constant(); c != nil {
			raw = c.GetText()
		} else {
			raw = x.GetText()
		}

		// int (base=0)
		if i, err := strconv.ParseInt(raw, 0, 64); err == nil {
			return &specir.ExprLiteral{
				BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
				Kind:     specir.LitInteger,
				Int:      i,
			}
		}

		// float
		if f, err := strconv.ParseFloat(raw, 64); err == nil {
			return &specir.ExprLiteral{
				BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
				Kind:     specir.LitFloat,
				Float:    f,
			}
		}

		// fallback
		return &specir.ExprVar{
			BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
			Name:     raw,
		}

	default:
		// TODO: extra literals
		return &specir.ExprVar{
			BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
			Name:     lit.(antlr.ParserRuleContext).GetText(),
		}
	}
}

func (b *ACSLBuilder) VisitNegation_pred(ctx *acsl.Negation_predContext) interface{} {
	return &specir.ExprUnary{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Op:       specir.UnaryNot,
		Expr:     b.predToExpr(ctx.Pred()),
	}
}

func (b *ACSLBuilder) VisitUniversal_quantification_pred(ctx *acsl.Universal_quantification_predContext) interface{} {
	return &specir.ExprQuantifier{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Kind:     specir.QuantifierForall,
		Vars:     b.bindersToVars(ctx.Binders()),
		Body:     b.predToExpr(ctx.Pred()),
	}
}

func (b *ACSLBuilder) VisitExistential_quantification_pred(ctx *acsl.Existential_quantification_predContext) interface{} {
	log.Debug("VisitExistential_quantification_pred")

	return &specir.ExprQuantifier{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Kind:     specir.QuantifierExists,
		Vars:     b.bindersToVars(ctx.Binders()),
		Body:     b.predToExpr(ctx.Pred()),
	}
}

func (b *ACSLBuilder) bindersToVars(ctx acsl.IBindersContext) []specir.BoundVar {

	if ctx == nil {
		return nil
	}

	var out []specir.BoundVar

	for _, binderCtx := range ctx.AllBinder() {
		ty := b.typeExprFromContext(binderCtx.Type_expr())
		for _, vctx := range binderCtx.AllVariable_ident() {
			out = append(out, specir.BoundVar{
				Name: b.variableIdentName(vctx),
				Type: ty,
			})
		}
	}

	return out
}

func (b *ACSLBuilder) typeExprFromContext(ctx acsl.IType_exprContext) specir.TypeExpr {
	if ctx == nil {
		return specir.TypeExpr{}
	}

	switch ctx.GetText() {
	case "boolean":
		return specir.TypeExpr{Kind: specir.TypeBoolean}
	case "integer":
		return specir.TypeExpr{Kind: specir.TypeInteger}
	case "real":
		return specir.TypeExpr{Kind: specir.TypeReal}
	default:
		return specir.TypeExpr{Kind: specir.TypeCustom, Name: ctx.GetText()}
	}
}

func (b *ACSLBuilder) variableIdentName(ctx acsl.IVariable_identContext) string {
	if ctx == nil {
		return ""
	}
	ids := collectIdentifiers(ctx)
	if len(ids) > 0 {
		return ids[0]
	}
	return ctx.GetText()
}

// -----------------------------------------------------------------------------
// Expression visitors for the precedence-based term grammar.
//
// Each binary precedence rule is a flat sequence:
// operand (operator operand)*.
// Folding the children from left to right preserves the grammar's associativity
// while the separate visitor methods preserve precedence.
// -----------------------------------------------------------------------------

func (b *ACSLBuilder) rawExpr(ctx antlr.ParserRuleContext) specir.Expr {
	if ctx == nil {
		return nil
	}
	return &specir.ExprVar{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Name:     ctx.GetText(),
	}
}

func (b *ACSLBuilder) visitExprChild(tree antlr.Tree) specir.Expr {
	if tree == nil {
		return nil
	}
	value, ok := tree.(antlr.ParseTree)
	if !ok {
		return nil
	}
	v := b.Visit(value)
	if e, ok := v.(specir.Expr); ok {
		return e
	}
	if np, ok := v.(specir.NamedPredicate); ok {
		return np.Expr
	}
	return nil
}
func (b *ACSLBuilder) foldTermPrecedence(ctx antlr.ParserRuleContext) interface{} {
	if ctx == nil || ctx.GetChildCount() == 0 {
		return nil
	}

	var result specir.Expr
	var pendingOp string

	for i := 0; i < ctx.GetChildCount(); i++ {
		child := ctx.GetChild(i)

		if terminal, ok := child.(antlr.TerminalNode); ok {
			pendingOp = terminal.GetText()
			continue
		}

		pt, ok := child.(antlr.ParseTree)
		if !ok || pt == nil {
			continue
		}

		rhs := b.visitExprChild(child)

		if rhs == nil {
			continue
		}

		if result == nil {
			result = rhs
			continue
		}

		op, ok := relOpToBinary(pendingOp)
		if !ok {
			log.Warnf(
				"Unsupported operator %q in %q",
				pendingOp, ctx.GetText(),
			)
			return b.rawExpr(ctx)
		}

		result = &specir.ExprBinary{
			BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
			Op:       op,
			Left:     result,
			Right:    rhs,
		}

		pendingOp = ""
	}

	if result == nil {
		return b.rawExpr(ctx)
	}

	return result
}

func (b *ACSLBuilder) VisitTerm(ctx *acsl.TermContext) interface{} {
	if ctx == nil || ctx.Conditional_term() == nil {
		return b.rawExpr(ctx)
	}
	return b.Visit(ctx.Conditional_term())
}

func (b *ACSLBuilder) VisitConditional_term(ctx *acsl.Conditional_termContext) interface{} {
	if ctx == nil {
		return nil
	}
	if ctx.GetChildCount() < 5 {
		return b.visitExprChild(ctx.GetChild(0).(antlr.Tree))
	}

	cond := b.visitExprChild(ctx.GetChild(0).(antlr.Tree))
	thenExpr := b.visitExprChild(ctx.GetChild(2).(antlr.Tree))
	elseExpr := b.visitExprChild(ctx.GetChild(4).(antlr.Tree))
	if cond == nil || thenExpr == nil || elseExpr == nil {
		return b.rawExpr(ctx)
	}
	return &specir.ExprConditional{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Cond:     cond,
		Then:     thenExpr,
		Else:     elseExpr,
	}
}

func (b *ACSLBuilder) VisitLogical_or_term(ctx *acsl.Logical_or_termContext) interface{} {
	return b.foldTermPrecedence(ctx)
}
func (b *ACSLBuilder) VisitLogical_xor_term(ctx *acsl.Logical_xor_termContext) interface{} {
	return b.foldTermPrecedence(ctx)
}
func (b *ACSLBuilder) VisitLogical_and_term(ctx *acsl.Logical_and_termContext) interface{} {
	return b.foldTermPrecedence(ctx)
}
func (b *ACSLBuilder) VisitBitwise_or_term(ctx *acsl.Bitwise_or_termContext) interface{} {
	return b.foldTermPrecedence(ctx)
}
func (b *ACSLBuilder) VisitBitwise_xor_term(ctx *acsl.Bitwise_xor_termContext) interface{} {
	return b.foldTermPrecedence(ctx)
}
func (b *ACSLBuilder) VisitBitwise_and_term(ctx *acsl.Bitwise_and_termContext) interface{} {
	return b.foldTermPrecedence(ctx)
}
func (b *ACSLBuilder) VisitEquality_term(ctx *acsl.Equality_termContext) interface{} {
	return b.foldTermPrecedence(ctx)
}
func (b *ACSLBuilder) VisitRelational_term(ctx *acsl.Relational_termContext) interface{} {
	return b.foldTermPrecedence(ctx)
}
func (b *ACSLBuilder) VisitShift_term(ctx *acsl.Shift_termContext) interface{} {
	return b.foldTermPrecedence(ctx)
}
func (b *ACSLBuilder) VisitAdditive_term(ctx *acsl.Additive_termContext) interface{} {
	return b.foldTermPrecedence(ctx)
}
func (b *ACSLBuilder) VisitMultiplicative_term(ctx *acsl.Multiplicative_termContext) interface{} {
	return b.foldTermPrecedence(ctx)
}

func (b *ACSLBuilder) VisitUnary_term(ctx *acsl.Unary_termContext) interface{} {
	if ctx == nil || ctx.GetChildCount() == 0 {
		return b.rawExpr(ctx)
	}
	if ctx.GetChildCount() == 1 {
		return b.visitExprChild(ctx.GetChild(0).(antlr.Tree))
	}

	opText := ctx.GetChild(0).(antlr.Tree).(antlr.ParseTree).GetText()
	inner := b.visitExprChild(ctx.GetChild(1).(antlr.Tree))
	var op specir.UnaryOp
	switch opText {
	case "!":
		op = specir.UnaryNot
	case "-":
		op = specir.UnaryMinus
	case "+":
		op = specir.UnaryPlus
	default:
		// The current IR has no dereference/address-of unary node.
		return b.rawExpr(ctx)
	}
	return &specir.ExprUnary{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Op:       op,
		Expr:     inner,
	}
}

func (b *ACSLBuilder) VisitCast_term(ctx *acsl.Cast_termContext) interface{} {
	if ctx == nil || ctx.GetChildCount() == 0 {
		return b.rawExpr(ctx)
	}
	// ExprCast is not present in the supplied IR. Do not silently erase a cast:
	// preserve it as a raw expression until the IR has an explicit cast node.
	if ctx.GetChildCount() > 1 {
		return b.rawExpr(ctx)
	}
	return b.visitExprChild(ctx.GetChild(0).(antlr.Tree))
}

func (b *ACSLBuilder) VisitPostfix_term(ctx *acsl.Postfix_termContext) interface{} {
	if ctx == nil || ctx.GetChildCount() == 0 {
		return b.rawExpr(ctx)
	}
	// The current IR has no array-index/member-access expression nodes.
	// Keep the complete source expression rather than dropping postfix suffixes.
	if ctx.GetChildCount() > 1 {
		return b.rawExpr(ctx)
	}
	return b.visitExprChild(ctx.GetChild(0).(antlr.Tree))
}

func (b *ACSLBuilder) VisitVariable_term(ctx *acsl.Variable_termContext) interface{} {
	if ctx == nil {
		return nil
	}
	name := ctx.GetText()
	if ctx.Poly_id() != nil {
		if n, ok := polyIDName(ctx.Poly_id()); ok {
			name = n
		}
	}
	return &specir.ExprVar{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Name:     name,
	}
}

func (b *ACSLBuilder) VisitParentheses_term(ctx *acsl.Parentheses_termContext) interface{} {
	if ctx == nil || ctx.Term() == nil {
		return b.rawExpr(ctx)
	}
	return b.termToExpr(ctx.Term())
}

// Predicates not covered by the original builder.
func (b *ACSLBuilder) VisitExclusive_or_pred(ctx *acsl.Exclusive_or_predContext) interface{} {
	ps := ctx.AllPred()
	if len(ps) != 2 {
		return b.rawExpr(ctx)
	}
	return &specir.ExprBinary{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Op:       specir.OpXor,
		Left:     b.predToExpr(ps[0]),
		Right:    b.predToExpr(ps[1]),
	}
}

func (b *ACSLBuilder) VisitTernary_condition_term_pred(ctx *acsl.Ternary_condition_term_predContext) interface{} {
	if ctx == nil || ctx.Term() == nil || len(ctx.AllPred()) != 2 {
		return b.rawExpr(ctx)
	}
	return &specir.ExprConditional{
		BaseExpr: specir.BaseExpr{Span: b.blockSpan(ctx)},
		Cond:     b.termToExpr(ctx.Term()),
		Then:     b.predToExpr(ctx.Pred(0)),
		Else:     b.predToExpr(ctx.Pred(1)),
	}
}
