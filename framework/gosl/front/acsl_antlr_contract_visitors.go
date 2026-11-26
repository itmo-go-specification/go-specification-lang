package front

import (
	acsl "gpma/framework/gosl/front/generated/grammar/acsl"
	specir "gpma/framework/gosl/ir"

	log "github.com/sirupsen/logrus"
)

func (b *ACSLBuilder) VisitAcsl_comment_contract(ctx *acsl.Acsl_comment_contractContext) interface{} {
	fc := &specir.FuncContract{
		FuncName:     "",
		RecvTypeName: "",
		Span:         b.blockSpan(ctx),
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

	log.Debugln("[ACSLBuilder] added FuncContract. Text:", ctx.GetText())
	// log.Debugln("[ACSLBuilder] added FuncContract:", fc.FuncName, "Text:", ctx.GetText(), "comment_span:", b.blockSpan(ctx), "contract_span:", b.blockSpan(fcCtx))
	// log.Debugln("contract_start:", fc.StartPos, "contract_end:", fc.EndPos)
	// log.Debugln("block_start:", b.Block.Start, "block_end:", b.Block.End)
	return nil
}

func (b *ACSLBuilder) VisitSimple_clause(ctx *acsl.Simple_clauseContext) interface{} {
	log.Debugln("VisitSimple_clause text:", ctx.GetText())
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
	log.Debugln("VisitRequires_clause text:", ctx.GetText())
	if ctx == nil {
		return nil
	}
	np := b.parseNamedPred(ctx.Pred())
	// StartPos/EndPos span the entire "requires <pred>;" clause.
	np.StartPos, np.EndPos = b.blockPositions(ctx)
	return np
}

func (b *ACSLBuilder) VisitEnsures_clause(ctx *acsl.Ensures_clauseContext) interface{} {
	log.Debugln("VisitEnsures_clause text:", ctx.GetText())
	if ctx == nil {
		return nil
	}
	np := b.parseNamedPred(ctx.Pred())
	// StartPos/EndPos span the entire "ensures <pred>;" clause.
	np.StartPos, np.EndPos = b.blockPositions(ctx)
	return np
}

func (b *ACSLBuilder) VisitAssumes_clause(ctx *acsl.Assumes_clauseContext) interface{} {
	if ctx == nil {
		return nil
	}
	np := b.parseNamedPred(ctx.Pred())
	np.StartPos, np.EndPos = b.blockPositions(ctx)
	return np
}

func (b *ACSLBuilder) VisitNamed_behavior(ctx *acsl.Named_behaviorContext) interface{} {
	if ctx == nil {
		return nil
	}

	behavior := specir.Behavior{
		Name: ctx.Id().GetText(),
		Span: b.blockSpan(ctx),
	}
	behavior.StartPos, behavior.EndPos = b.blockPositions(ctx)

	body := ctx.Behavior_body()
	if body == nil {
		return behavior
	}

	for _, assumes := range body.AllAssumes_clause() {
		np, ok := b.Visit(assumes).(specir.NamedPredicate)
		if ok {
			behavior.Assumes = append(behavior.Assumes, np)
		}
	}
	for _, requires := range body.AllRequires_clause() {
		np, ok := b.Visit(requires).(specir.NamedPredicate)
		if ok {
			behavior.Requires = append(behavior.Requires, np)
		}
	}
	for _, clause := range body.AllSimple_clause() {
		if ensures := clause.Ensures_clause(); ensures != nil {
			np, ok := b.Visit(ensures).(specir.NamedPredicate)
			if ok {
				behavior.Ensures = append(behavior.Ensures, np)
			}
		}
		if assigns := clause.Assigns_clause(); assigns != nil {
			locations, ok := b.Visit(assigns).(specir.LocationSet)
			if ok {
				behavior.Assigns = locations
			}
		}
	}

	return behavior
}
