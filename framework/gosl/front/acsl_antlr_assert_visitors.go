package front

import (
	acsl "gpma/framework/gosl/front/generated/grammar/acsl"
	specir "gpma/framework/gosl/ir"

	log "github.com/sirupsen/logrus"
)

func (b *ACSLBuilder) VisitAcsl_comment_assertion(ctx *acsl.Acsl_comment_assertionContext) interface{} {
	log.Debugf("VisitAcsl_comment_assertion Text:%s, type:%T\n", ctx.GetText(), ctx.Assertion())

	a := &specir.AssertSpec{
		Span: b.blockSpan(ctx),
	}
	a.StartPos, a.EndPos = b.blockPositions(ctx)

	// only one assertion
	np, ok := b.Visit(ctx.Assertion()).(specir.NamedPredicate)
	if ok {
		a.Assertion = np
		b.Mod.Asserts = append(b.Mod.Asserts, a)
	}

	return nil
}

func (b *ACSLBuilder) VisitAssertion(ctx *acsl.AssertionContext) interface{} {
	log.Debugf("VisitAssertion Text:%s, type:%T, pred_type:%T\n", ctx.GetText(), ctx, ctx.Pred())
	return b.parseNamedPred(ctx.Pred())
}

func (b *ACSLBuilder) VisitAcsl_comment_statement_contract(ctx *acsl.Acsl_comment_statement_contractContext) interface{} {
	log.Debugln("[ACSLBuilder] statement contract (stub)", "Text:", ctx.GetText())
	return nil
}

// Function contract: requires/ensures/assigns
func (b *ACSLBuilder) VisitFunction_contract(ctx *acsl.Function_contractContext) interface{} {
	log.Debugln("VisitFunction_contract text:", ctx.GetText())
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
