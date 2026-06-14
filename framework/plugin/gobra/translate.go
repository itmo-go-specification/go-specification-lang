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

func (g *GobraGenerator) GenerateGobraFuncContract(funcContract *specir.FuncContract, sourceOffset int) string {
	g.builder.Reset()
	g.currentOffset = sourceOffset // Update offset for this generation

	sourcePos := gobra_runner.SourcePosition{
		File:   g.sourceFile,
		Offset: funcContract.Span.End,
	}

	// Generate requires with automatic tracking
	for i, requires := range funcContract.Requires {
		if i > 0 {
			g.builder.WriteByte('\n')
		}

		_, err := g.TrackPredicate(
			func() string {
				return "requires " + PrintExpr(requires.Expr)
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
				return "ensures " + PrintExpr(ensures.Expr)
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

// GenerateGobra generates Gobra code from AST data with automatic tracking
// Returns GobraGenerationResult containing both code and hybrid mapping
func GenerateGobra(astData *data.ASTData, sourceFile string) (*GobraGenerationResult, error) {
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
		HybridMapping: gobra_runner.NewHybridMapping(sourceFile),
		currentOffset: 0,
		sourceFile:    sourceFile,
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

	return &GobraGenerationResult{
		GobraCode:     source,
		HybridMapping: g.HybridMapping,
	}, nil
}
