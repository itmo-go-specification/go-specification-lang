package front

import (
	acsl "gpma/framework/gosl/front/generated/grammar/acsl"
	specir "gpma/framework/gosl/ir"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	log "github.com/sirupsen/logrus"
)

func (b *ACSLBuilder) VisitAcsl_comment_loop_annot(ctx *acsl.Acsl_comment_loop_annotContext) interface{} {
	ls := &specir.LoopSpec{
		ParentFuncName: "",
		Span:           b.blockSpan(ctx),
	}

	b.currLoop = ls
	defer func() { b.currLoop = nil }()

	la := ctx.Loop_annot()
	if la != nil {
		b.Visit(la)
		ls.StartPos, ls.EndPos = b.blockPositions(la.(antlr.ParserRuleContext))
	}

	b.Mod.LoopSpecs = append(b.Mod.LoopSpecs, ls)
	log.Debugln("[ACSLBuilder] added LoopSpec. Text:", ctx.GetText())
	return nil
}

func (b *ACSLBuilder) VisitLoop_annot(ctx *acsl.Loop_annotContext) interface{} {
	if b.currLoop == nil || ctx == nil {
		return nil
	}

	for _, lc := range ctx.AllLoop_clause() {
		b.Visit(lc)
	}

	if v := ctx.Loop_variant(); v != nil {
		b.Visit(v)
	}

	return nil
}

func (b *ACSLBuilder) VisitLoop_clause(ctx *acsl.Loop_clauseContext) interface{} {
	if b.currLoop == nil || ctx == nil {
		return nil
	}
	if li := ctx.Loop_invariant(); li != nil {
		return b.Visit(li)
	}
	if la := ctx.Loop_assigns(); la != nil {
		return b.Visit(la)
	}

	if al := ctx.Loop_allocation(); al != nil {
		_ = al
	}
	return nil
}

func (b *ACSLBuilder) VisitLoop_invariant(ctx *acsl.Loop_invariantContext) interface{} {
	if b.currLoop == nil || ctx == nil {
		return nil
	}
	np := b.parseNamedPred(ctx.Pred())
	b.currLoop.Invariants = append(b.currLoop.Invariants, np)
	return nil
}

func (b *ACSLBuilder) VisitLoop_assigns(ctx *acsl.Loop_assignsContext) interface{} {
	if b.currLoop == nil || ctx == nil {
		return nil
	}

	// \nothing
	txt := ctx.GetText()
	if strings.Contains(txt, `\nothing`) {
		b.currLoop.Assigns = specir.LocationSet{Kind: specir.LocSetNothing}
		return nil
	}

	locs := ctx.Locations()
	vars := collectIdentifiers(locs)
	if len(vars) == 0 {
		b.currLoop.Assigns = specir.LocationSet{Kind: specir.LocSetVars, Vars: []string{txt}}
		return nil
	}

	b.currLoop.Assigns = specir.LocationSet{Kind: specir.LocSetVars, Vars: unique(vars)}
	return nil
}

func (b *ACSLBuilder) VisitLoop_variant(ctx *acsl.Loop_variantContext) interface{} {
	if b.currLoop == nil || ctx == nil {
		return nil
	}

	t := ctx.Term()
	b.currLoop.Variant = b.termToExpr(t)

	return nil
}
