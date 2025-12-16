package front

import (
	acsl "gpma/framework/gosl/front/generated/grammar"
	specir "gpma/framework/gosl/ir"

	"github.com/antlr4-go/antlr/v4"
)

type ACSLBuilder struct {
	*acsl.BaseACSLVisitor

	Mod   *specir.SpecModule
	Block RawSpecBlock

	currFunc *specir.FuncContract
	currLoop *specir.LoopSpec
}

func newACSLBuilder(mod *specir.SpecModule, block RawSpecBlock) *ACSLBuilder {
	return &ACSLBuilder{
		BaseACSLVisitor: &acsl.BaseACSLVisitor{},
		Mod:             mod,
		Block:           block,
		currFunc:        nil,
		currLoop:        nil,
	}
}

func (b *ACSLBuilder) blockSpan(ctx antlr.ParserRuleContext) specir.Span {
	base := b.Block.Start.Offset

	// fallback value == RawSpecBlock.Span
	fallback := specir.Span{
		File:  b.Block.Start.File,
		Start: b.Block.Start.Offset,
		End:   b.Block.End.Offset,
	}

	startTok := ctx.GetStart()
	stopTok := ctx.GetStop()
	if startTok == nil || stopTok == nil {
		return fallback
	}

	st, ok1 := startTok.(*antlr.CommonToken)
	en, ok2 := stopTok.(*antlr.CommonToken)
	if !ok1 || !ok2 {
		return fallback
	}

	// CommonToken.GetStart/GetStop - offsets inside of spec block 
	localStart := st.GetStart()
	localStop := en.GetStop()

	if localStart < 0 || localStop < localStart {
		return fallback
	}

	return specir.Span{
		File:  b.Block.Start.File,
		Start: base + localStart,
		End:   base + localStop + 1, // [Start, End)
	}
}

func (b *ACSLBuilder) blockPositions(ctx antlr.ParserRuleContext) (specir.Position, specir.Position) {
	base := b.Block.Start

	fallbackStart := base
	fallbackEnd := b.Block.End

	startTok := ctx.GetStart()
	stopTok := ctx.GetStop()
	if startTok == nil || stopTok == nil {
		return fallbackStart, fallbackEnd
	}

	st, ok1 := startTok.(*antlr.CommonToken)
	en, ok2 := stopTok.(*antlr.CommonToken)
	if !ok1 || !ok2 {
		return fallbackStart, fallbackEnd
	}

	localStart := st.GetStart()
	localStop := en.GetStop()
	if localStart < 0 || localStop < localStart {
		return fallbackStart, fallbackEnd
	}

	// start: startTok
	startPos := specir.Position{
		File:   base.File,
		Offset: base.Offset + localStart,
		Line:   addLine(base.Line, st.GetLine()),
		Column: addColumn(base.Line, base.Column, st.GetLine(), st.GetColumn()),
	}

	// end: stopTok (end-exclusive)
	endPos := specir.Position{
		File:   base.File,
		Offset: base.Offset + localStop + 1,
		Line:   addLine(base.Line, en.GetLine()),
		Column: addColumn(base.Line, base.Column, en.GetLine(), en.GetColumn()) + tokenTextLenApprox(en),
	}

	return startPos, endPos
}

// addLine ANTLR line is 1-based.
func addLine(baseLine int, localLine int) int {
	if localLine <= 1 {
		return baseLine
	}
	return baseLine + (localLine - 1)
}

// addColumn ANTLR column is 0-based
func addColumn(baseLine int, baseColumn int, localLine int, localColumn int) int {
    localColumn += 1 // bacause ANTLR column 0-based.
	if localLine <= 1 {
		return baseColumn + localColumn
	}
	return localColumn
}

func tokenTextLenApprox(tok *antlr.CommonToken) int {
	txt := tok.GetText()
	if txt == "" {
		return 0
	}
	return len([]rune(txt))
}



func (b *ACSLBuilder) BuildFromAcslComment(tree acsl.IAcsl_commentContext) error {
	// if _, ok := any(b).(acsl.ACSLVisitor); !ok {
	// 	fmt.Println("ACSLBuilder DOES NOT implement acsl.ACSLVisitor")
	// } else {
	// 	fmt.Println("ACSLBuilder implements acsl.ACSLVisitor")
	// }
	_ = b.Visit(tree.(antlr.ParseTree))
	return nil
}

func (b *ACSLBuilder) Visit(t antlr.ParseTree) interface{} {
	if t == nil {
		return nil
	}
	return t.Accept(b)
}
