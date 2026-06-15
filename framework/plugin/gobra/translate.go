package gobra

// Gobra Code Generation with Hybrid Mapping
//
// This package provides Gobra code generation with automatic tracking of code additions
// and their relationships to the original Go source code. The hybrid mapping enables
// precise error reporting by translating Gobra error positions back to original source locations.
//
// Key Features:
// - Universal automatic tracking via TrackAddition and TrackPredicate methods
// - Support for multiple predicates at the same source position
// - Non-overlapping interval validation
// - Precise point mappings for predicates
//
// Usage:
//   result, err := GenerateGobra(astData, sourceFile)
//   if err != nil {
//       return err
//   }
//
//   // Access generated code
//   gobraCode := result.GobraCode
//
//   // Access hybrid mapping for error translation
//   mapping := result.HybridMapping
//
//   // Map Gobra errors back to source
//   sourcePos, err := mapping.MapError(gobraErrorOffset)
//
// Multiple Predicates at Same Source Position:
// The system correctly handles scenarios where multiple predicates (requires, ensures)
// are generated at the same source position (e.g., end of a contract):
//
//   /*
//   @
//   requires x >= 0;      // predicateSpan.Start = 20
//   ensures res == x + 1; // predicateSpan.Start = 40
//   */
//   func AddOne(x int) (res int) { return x + 1 }
//
// Each predicate gets:
// - Separate non-overlapping interval in generated code
// - Unique predicate span pointing to its source location
// - Same general source position (end of contract)

import (
	"fmt"
	"go/ast"
	"gpma/framework/data"
	gobra_runner "gpma/framework/gobra_runner"
	specir "gpma/framework/gosl/ir"
	"strings"
)

// GobraGenerationResult contains the generated Gobra code and its hybrid mapping
type GobraGenerationResult struct {
	GobraCode     string
	HybridMapping *gobra_runner.HybridMapping
}

// GobraGenerator generates Gobra code with automatic tracking of additions
type GobraGenerator struct {
	builder       strings.Builder // For building generated code
	HybridMapping *gobra_runner.HybridMapping
	sourceOffset  int // Current position in source code
	currentOffset int // Current position in generated code
	sourceFile    string
}

// TrackAddition automatically tracks a code addition by measuring position changes
// generateFunc: function that generates the code to add
// sourcePos: position in the original source code
// predicateSpan: optional precise point mapping for predicates
func (g *GobraGenerator) TrackAddition(
	generateFunc func() string,
	sourcePos gobra_runner.SourcePosition,
	predicateSpan *gobra_runner.SourcePosition,
) (string, error) {
	// Remember position before generation
	beforeLen := g.builder.Len()

	// Generate the code and write it to builder
	generated := generateFunc()
	g.builder.WriteString(generated)

	// Remember position after generation
	afterLen := g.builder.Len()

	// Create interval automatically
	genStart := beforeLen + g.currentOffset
	genEnd := afterLen + g.currentOffset - 1 // -1 because Len() returns length, not position

	interval := gobra_runner.Interval{
		GenStart:      genStart,
		GenEnd:        genEnd,
		SourcePos:     sourcePos,
		PredicateSpan: predicateSpan,
	}

	// Add interval to mapping
	err := g.HybridMapping.AddInterval(interval)
	if err != nil {
		return "", fmt.Errorf("failed to track addition: %w", err)
	}

	return generated, nil
}

// TrackPredicate automatically tracks a predicate with point mapping
// generateFunc: function that generates the predicate code
// sourceSpan: span of the predicate in source code
// sourcePos: general position in source code (e.g., end of contract)
func (g *GobraGenerator) TrackPredicate(
	generateFunc func() string,
	sourceSpan specir.Span,
	sourcePos gobra_runner.SourcePosition,
) (string, error) {
	predicateSpan := &gobra_runner.SourcePosition{
		File:   sourceSpan.File,
		Offset: sourceSpan.Start,
	}

	return g.TrackAddition(generateFunc, sourcePos, predicateSpan)
}

// WriteString writes a string to the builder and returns its length
func (g *GobraGenerator) WriteString(s string) int {
	g.builder.WriteString(s)
	return len(s)
}

// WriteByte writes a byte to the builder and returns 1
func (g *GobraGenerator) WriteByte(b byte) int {
	g.builder.WriteByte(b)
	return 1
}

// CurrentPosition returns the current position in the generated code
func (g *GobraGenerator) CurrentPosition() int {
	return g.builder.Len() + g.currentOffset
}

func (g *GobraGenerator) GenerateGobraFuncContract(node *ast.Node, funcContract *specir.FuncContract, sourceOffset int) string {
	g.builder.Reset()
	g.currentOffset = sourceOffset // Update offset for this generation

	sourcePos := gobra_runner.SourcePosition{
		File:   g.sourceFile,
		Offset: funcContract.Span.End,
	}

	funcDecl, ok := (*node).(*ast.FuncDecl)
	if !ok {
		panic("Received not function decl node to contract generation")
	}

	p := &Printer{}
	if funcDecl.Type.Results != nil && funcDecl.Type.Results.List != nil {
		for _, field := range funcDecl.Type.Results.List {
			for i, name := range field.Names {
				if name == nil {
					p.resultNames = append(p.resultNames, fmt.Sprintf("__result_%d", i))
				} else {
					p.resultNames = append(p.resultNames, name.Name)
				}
			}
		}
	}

	// Generate requires with automatic tracking
	for i, requires := range funcContract.Requires {
		if i > 0 {
			g.builder.WriteByte('\n')
		}

		_, err := g.TrackPredicate(
			func() string {
				p.buf.Reset()
				p.PrintExpr(requires.Expr)
				return "requires " + p.buf.String()
			},
			requires.Span,
			sourcePos,
		)
		if err != nil {
			fmt.Printf("Warning: failed to track requires: %v\n", err)
		}
	}

	// Generate ensures with automatic tracking
	for _, ensures := range funcContract.Ensures {
		if g.builder.Len() > 0 {
			g.builder.WriteByte('\n')
		}

		_, err := g.TrackPredicate(
			func() string {
				p.buf.Reset()
				p.PrintExpr(ensures.Expr)
				return "ensures " + p.buf.String()
			},
			ensures.Span,
			sourcePos,
		)
		if err != nil {
			fmt.Printf("Warning: failed to track ensures: %v\n", err)
		}
	}

	return g.builder.String()
}

func (g *GobraGenerator) GenerateGobraLoopSpec(loopSpec *specir.LoopSpec, sourceOffset int) string {
	g.builder.Reset()
	g.currentOffset = sourceOffset // Update offset for this generation

	sourcePos := gobra_runner.SourcePosition{
		File:   g.sourceFile,
		Offset: loopSpec.Span.End,
	}

	for i, invariant := range loopSpec.Invariants {
		if i > 0 {
			g.builder.WriteByte('\n')
		}

		_, err := g.TrackPredicate(
			func() string {
				return "invariant " + PrintExpr(invariant.Expr)
			},
			invariant.Span,
			sourcePos,
		)
		if err != nil {
			fmt.Printf("Warning: failed to track invariant: %v\n", err)
		}
	}

	return g.builder.String()
}

func (g *GobraGenerator) GenerateGobraAsserts(asserts *specir.AssertSpec, sourceOffset int) string {
	g.builder.Reset()
	g.currentOffset = sourceOffset // Update offset for this generation

	sourcePos := gobra_runner.SourcePosition{
		File:   g.sourceFile,
		Offset: asserts.Span.End,
	}

	_, err := g.TrackPredicate(
		func() string {
			return "assert " + PrintExpr(asserts.Assertion.Expr) + ";"
		},
		asserts.Assertion.Span,
		sourcePos,
	)
	if err != nil {
		fmt.Printf("Warning: failed to track assert: %v\n", err)
	}

	return g.builder.String()
}

func (g *GobraGenerator) GenerateGobraPredicate(predicate *specir.LogicDecl, sourceOffset int) string {
	g.builder.Reset()
	g.currentOffset = sourceOffset

	g.builder.WriteString("ghost\ndecreases\n")
	g.builder.WriteString(fmt.Sprintf("pure func %s(", predicate.Name))
	for i, param := range predicate.Params {
		if i > 0 {
			g.builder.WriteString(", ")
		}

		g.WriteString(param.Name)
		g.WriteString(" ")
		g.WriteString(typeExprStr(param.Type))
	}
	g.builder.WriteString(") bool {\n\t")
	g.builder.WriteString("return " + PrintExpr(predicate.Body))
	g.builder.WriteString("\n}\n")

	return g.builder.String()
}

// GenerateGobra generates Gobra code from AST data with automatic tracking
// Returns GobraGenerationResult containing both code and hybrid mapping
func GenerateGobra(astData *data.ASTData, sourceFile string) (*GobraGenerationResult, error) {
	type Segment struct {
		node   *ast.Node
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

			acsl_segments = append(acsl_segments, Segment{&n, node_info, position})
		}

		return true
	})

	g := &GobraGenerator{
		HybridMapping: gobra_runner.NewHybridMapping(sourceFile),
		sourceOffset: 0,
		currentOffset: 0,
		sourceFile:    sourceFile,
	}

	builder := strings.Builder{}

	for _, segment := range acsl_segments {
		if g.sourceOffset < segment.offset {
			builder.WriteString(astData.Source[g.sourceOffset:segment.offset])
			builder.WriteString("\n")
			g.sourceOffset = segment.offset
		}

		section_text := func() string {
			if segment.info.FuncContract != nil {
				return g.GenerateGobraFuncContract(segment.node, segment.info.FuncContract, builder.Len())
			} else if segment.info.LoopSpec != nil {
				return g.GenerateGobraLoopSpec(segment.info.LoopSpec, builder.Len())
			} else if segment.info.Asserts != nil {
				return g.GenerateGobraAsserts(segment.info.Asserts, builder.Len())
			}
			panic("Empty contract")
		}()

		builder.WriteString(section_text)
	}

	if g.sourceOffset < len(astData.Source) {
		builder.WriteString(astData.Source[g.sourceOffset:])
		builder.WriteString("\n")
		g.sourceOffset = len(astData.Source)
	}

	for _, specDecl := range astData.Spec.LogicDecls {
		if specDecl.Kind == specir.LogicPredicate {
			builder.WriteString(
				g.GenerateGobraPredicate(specDecl, builder.Len()),
			)
		}
	}

	return &GobraGenerationResult{
		GobraCode:     builder.String(),
		HybridMapping: g.HybridMapping,
	}, nil
}
