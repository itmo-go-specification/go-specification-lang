package gobra

import (
	"go/ast"
	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
	"strings"
)

type PredicateSpan struct {
	Start      int
	End        int
	SourceSpan specir.Span
}

type GobraGenerator struct {
	PredicateSpans []PredicateSpan
}

func (g *GobraGenerator) GenerateGobraFuncContract(funcContract *specir.FuncContract, sourceOffset int) string {
	var b strings.Builder

	for i, requires := range funcContract.Requires {
		if i > 0 {
			b.WriteByte('\n')
		}
		start := b.Len()
		b.WriteString("requires ")
		b.WriteString(PrintExpr(requires.Expr))
		end := b.Len()
		g.PredicateSpans = append(g.PredicateSpans, PredicateSpan{
			Start:      start + sourceOffset,
			End:        end + sourceOffset,
			SourceSpan: requires.Span,
		})
	}
	for _, ensures := range funcContract.Ensures {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		start := b.Len()
		b.WriteString("ensures ")
		b.WriteString(PrintExpr(ensures.Expr))
		end := b.Len()
		g.PredicateSpans = append(g.PredicateSpans, PredicateSpan{
			Start:      start + sourceOffset,
			End:        end + sourceOffset,
			SourceSpan: ensures.Span,
		})
	}
	return b.String()
}

func (g *GobraGenerator) GenerateGobraLoopSpec(loopSpec *specir.LoopSpec, sourceOffset int) string {
	var b strings.Builder

	for i, invariant := range loopSpec.Invariants {
		if i > 0 {
			b.WriteByte('\n')
		}
		start := b.Len()
		b.WriteString("invariant ")
		b.WriteString(PrintExpr(invariant.Expr))
		end := b.Len()
		g.PredicateSpans = append(g.PredicateSpans, PredicateSpan{
			Start:      start + sourceOffset,
			End:        end + sourceOffset,
			SourceSpan: invariant.Span,
		})
	}
	return b.String()
}

func (g *GobraGenerator) GenerateGobraAsserts(asserts *specir.AssertSpec, sourceOffset int) string {
	var b strings.Builder

	start := b.Len()
	b.WriteString("assert ")
	b.WriteString(PrintExpr(asserts.Assertion.Expr))
	b.WriteByte(';')
	end := b.Len()
	g.PredicateSpans = append(g.PredicateSpans, PredicateSpan{
		Start:      start + sourceOffset,
		End:        end + sourceOffset,
		SourceSpan: asserts.Assertion.Span,
	})

	return b.String()
}

type GobraArtifacts struct {
	gobraSource    string
	predicateSpans []PredicateSpan
}

func (g *GobraArtifacts) GetGobraSource() string {
	return g.gobraSource
}

func GenerateGobra(astData *data.ASTData) (GobraArtifacts, error) {
	type Segment struct {
		info   *data.AdditionalInfo
		offset int
	}

	acsl_segments := []Segment{}

	ast.Inspect(astData.File, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		node_info, ok := astData.NodeInfo[n]
		if ok {
			position := func() int {
				if node_info.FuncContract != nil {
					return node_info.FuncContract.Span.End
				} else if node_info.LoopSpec != nil {
					return node_info.LoopSpec.Span.End
				} else if node_info.Asserts != nil {
					return node_info.Asserts.Span.End
				}
				panic("Empty NpdeInfo")
			}()

			acsl_segments = append(acsl_segments, Segment{node_info, position})
		}

		return true
	})

	g := &GobraGenerator{
		PredicateSpans: []PredicateSpan{},
	}

	source := astData.Source

	shift := 0
	for _, segment := range acsl_segments {
		offset := shift + segment.offset

		section_text := func() string {
			if segment.info.FuncContract != nil {
				return g.GenerateGobraFuncContract(segment.info.FuncContract, offset+1)
			} else if segment.info.LoopSpec != nil {
				return g.GenerateGobraLoopSpec(segment.info.LoopSpec, offset+1)
			} else if segment.info.Asserts != nil {
				return g.GenerateGobraAsserts(segment.info.Asserts, offset+1)
			}
			panic("Empty contract")
		}()

		source = source[:offset] + "\n" + section_text + source[offset:]
		shift = shift + len(section_text) + 1
	}

	return GobraArtifacts{source, g.PredicateSpans}, nil
}
