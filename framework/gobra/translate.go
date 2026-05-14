package gobra

import (
	"go/ast"
	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
	"strings"
)

func generateGobraFuncContract(funcContract *specir.FuncContract) string {
	clauses := []string{}

	for _, requires := range funcContract.Requires {
		clauses = append(clauses, "requires " + PrintExpr(requires.Expr))
	}

	for _, ensures := range funcContract.Ensures {
		clauses = append(clauses, "ensures " + PrintExpr(ensures.Expr))
	}

	return strings.Join(clauses, "\n")
}

func generateGobraLoopSpec(loopSpec *specir.LoopSpec) string {
	clauses := []string{}

	for _, invariant := range loopSpec.Invariants {
		clauses = append(clauses, "invariant " + PrintExpr(invariant.Expr))
	}

	return strings.Join(clauses, "\n")
}

func generateGobraAsserts(asserts *specir.AssertSpec) string {
	return "assert " + PrintExpr(asserts.Assertion.Expr) + ";"
}

type gobraSection struct {
	offset int
	len int
}

type GobraArtifacts struct {
	gobraSource string
	gobraSections []gobraSection
}

func (g *GobraArtifacts) GetGobraSource() string {
	return g.gobraSource
}

func (g *GobraArtifacts) ToAstOffset(offset int) int {
	for _, section := range g.gobraSections {
		if offset > section.offset {
			offset -= min(section.len, offset - section.offset)
		}
	}

	return offset
}

func (g *GobraArtifacts) ToGobraOffset(offset int) int {
	shift := 0
	for _, section := range g.gobraSections {
		if offset > section.offset {
			shift += section.len
		}
	}
	return offset + shift
}

func GenerateGobra(source string, astData *data.ASTData) (GobraArtifacts, error) {
	type segment struct {
		info *data.AdditionalInfo
		offset int
	}
	
	acsl_segments := []segment{}
	gobra_sections := []gobraSection{}

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

				panic("")
			}()

			acsl_segments = append(acsl_segments, segment{node_info, position})
		}

		return true
	})

	shift := 0
	for _, segment := range acsl_segments {
		section_text := func() string {
			if segment.info.FuncContract != nil {
				return generateGobraFuncContract(segment.info.FuncContract)
			} else if segment.info.LoopSpec != nil {
				return generateGobraLoopSpec(segment.info.LoopSpec)
			} else if segment.info.Asserts != nil {
				return generateGobraAsserts(segment.info.Asserts)
			}
			panic("")
		}()

		offset := shift + segment.offset
		source = source[:offset] + "\n" + section_text + source[offset:]
		shift = shift + len(section_text) + 1

		gobra_sections = append(gobra_sections, gobraSection{segment.offset, len(section_text) + 1})
	}

	return GobraArtifacts{source, gobra_sections}, nil
}
