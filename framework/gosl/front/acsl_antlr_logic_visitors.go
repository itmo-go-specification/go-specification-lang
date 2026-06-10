package front

import (
	acsl "gpma/framework/gosl/front/generated/grammar/acsl"
	specir "gpma/framework/gosl/ir"

	log "github.com/sirupsen/logrus"
)

func (b *ACSLBuilder) VisitAcsl_comment_logic_defs(ctx *acsl.Acsl_comment_logic_defsContext) interface{} {
	log.Debugln("Visit_Acsl_comment_logic_defs text:", ctx.GetText())
	if ctx == nil {
		return nil
	}

	for _, def := range ctx.AllLogic_def() {
		b.Visit(def)
	}

	return nil
}

func (b *ACSLBuilder) VisitLogic_def(ctx *acsl.Logic_defContext) interface{} {
	log.Debugln("VisitLogic_def text:", ctx.GetText())

	switch {
	case ctx.Logic_predicate_def() != nil:
		return b.Visit(ctx.Logic_predicate_def())

	case ctx.Logic_function_def() != nil:
		return b.Visit(ctx.Logic_function_def())

	case ctx.Lemma_def() != nil:
		return b.Visit(ctx.Lemma_def())

	default:
		return nil
	}
}

func (b *ACSLBuilder) VisitLogic_predicate_def(ctx *acsl.Logic_predicate_defContext) interface{} {
	name, _ := polyIDName(ctx.Poly_id())

	decl := &specir.LogicDecl{
		Name:       name,
		Kind:       specir.LogicPredicate,
		Params:     b.logicParams(ctx.Parameters()),
		ResultType: specir.TypeExpr{Kind: specir.TypeBoolean},
		Body:       b.predToExpr(ctx.Pred()),
		Span:       b.blockSpan(ctx),
	}

	b.Mod.LogicDecls[name] = decl
	return decl
}

func (b *ACSLBuilder) VisitLogic_function_def(ctx *acsl.Logic_function_defContext) interface{} {
	name, _ := polyIDName(ctx.Poly_id())

	decl := &specir.LogicDecl{
		Name:       name,
		Kind:       specir.LogicFunction,
		Params:     b.logicParams(ctx.Parameters()),
		ResultType: b.typeExprFromContext(ctx.Type_expr()),
		Body:       b.termToExpr(ctx.Term()),
		Span:       b.blockSpan(ctx),
	}

	b.Mod.LogicDecls[name] = decl
	return decl
}

func (b *ACSLBuilder) VisitLemma_def(ctx *acsl.Lemma_defContext) interface{} {
	log.Debugln("VisitLemma_def text:", ctx.GetText())
	if ctx == nil {
		return nil
	}

	name, ok := polyIDName(ctx.Poly_id())

	if !ok {
		return ctx.Poly_id().GetText()
	}

	decl := &specir.LogicDecl{
		Name:       name,
		Kind:       specir.LogicLemma,
		ResultType: specir.TypeExpr{Kind: specir.TypeBoolean},
		Body:       b.predToExpr(ctx.Pred()),
		ClauseKind: clauseKindFromContext(ctx.Clause_kind()),
		Span:       b.blockSpan(ctx),
	}

	if b.Mod.LogicDecls == nil {
		b.Mod.LogicDecls = make(map[string]*specir.LogicDecl)
	}

	log.Debugln("Add logic decl")
	b.Mod.LogicDecls[name] = decl

	return decl
}

func clauseKindFromContext(ctx acsl.IClause_kindContext) specir.ClauseKind {
	if ctx == nil {
		return specir.ClauseDefault
	}

	switch ctx.GetText() {
	case "check":
		return specir.ClauseCheck
	case "admit":
		return specir.ClauseAdmit
	default:
		return specir.ClauseDefault
	}
}

func (b *ACSLBuilder) logicParams(ctx acsl.IParametersContext) []specir.LogicParam {
	if ctx == nil {
		return nil
	}

	var out []specir.LogicParam

	for _, p := range ctx.AllParameter() {
		ty := b.typeExprFromContext(p.Type_expr())

		for _, v := range p.AllVariable_ident() {
			out = append(out, specir.LogicParam{
				Name: b.variableIdentName(v),
				Type: ty,
			})
		}
	}

	return out
}
