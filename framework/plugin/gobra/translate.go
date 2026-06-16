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
	"go/printer"
	"go/token"
	"gpma/framework/data"
	gobra_runner "gpma/framework/gobra_runner"
	specir "gpma/framework/gosl/ir"
	"sort"
	"strings"
)

// GobraGenerationResult contains the generated Gobra code and its hybrid mapping
type GobraGenerationResult struct {
	GobraCode     string
	HybridMapping *gobra_runner.HybridMapping
}

type Replacement struct {
	SourcePos int
	SourceEnd   int
	Insertion string
	PredicateSpan *gobra_runner.SourcePosition
}

// GobraGenerator generates Gobra code with automatic tracking of additions
type GobraGenerator struct {
	builder       strings.Builder // For building generated code
	HybridMapping *gobra_runner.HybridMapping
	sourceOffset  int // Current position in source code
	currentOffset int // Current position in generated code
	varCount      int 
	fset          *token.FileSet
	sourceFile    string
	sourceCode    string
	replacements  []Replacement
}

func (g *GobraGenerator) AddReplacement(sourcePos int, sourceEnd int, insertion string, predicateSpan *gobra_runner.SourcePosition) {
	repl := Replacement{
		SourcePos: sourcePos,
		SourceEnd: sourceEnd,
		Insertion: insertion,
		PredicateSpan: predicateSpan,
	}

	g.replacements = append(g.replacements, repl)
}

func (g *GobraGenerator) Render() (string, error) {
    sort.SliceStable(g.replacements, func(i, j int) bool {
        return g.replacements[i].SourcePos < g.replacements[j].SourcePos
    })

    // Validate: no overlapping replacements.
    for i := 1; i < len(g.replacements); i++ {
        prev, curr := g.replacements[i-1], g.replacements[i]
        if prev.SourceEnd > curr.SourcePos {
            return "", fmt.Errorf("overlapping replacements at [%d,%d) and [%d,%d)",
                prev.SourcePos, prev.SourceEnd,
                curr.SourcePos, curr.SourceEnd)
        }
    }

    var sb strings.Builder
    cursor := 0

    for _, r := range g.replacements {
        if r.SourcePos > len(g.sourceCode) || r.SourceEnd > len(g.sourceCode) {
            return "", fmt.Errorf("replacement [%d,%d) out of bounds (source len %d)",
                r.SourcePos, r.SourceEnd, len(g.sourceCode))
        }
        if r.SourcePos < cursor {
            return "", fmt.Errorf("replacement at %d starts before current cursor %d",
                r.SourcePos, cursor)
        }

        // Copy source unchanged up to the replacement start.
        sb.WriteString(g.sourceCode[cursor:r.SourcePos])

		interval := gobra_runner.Interval{
			GenStart:      sb.Len(),
			GenEnd:        sb.Len() + len(r.Insertion) - 1,
			SourcePos:     gobra_runner.SourcePosition{ File: g.sourceFile, Offset: r.SourcePos, },
			SourceEnd:     gobra_runner.SourcePosition{ File: g.sourceFile, Offset: r.SourceEnd, },
			PredicateSpan: r.PredicateSpan,
		}
		err := g.HybridMapping.AddInterval(interval)
		if err != nil {
			return "", fmt.Errorf("failed to track addition: %w", err)
		}

        // Write the replacement.
        sb.WriteString(r.Insertion)
        // Advance cursor past the replaced region.
        cursor = r.SourceEnd
    }

    // Copy any remaining source after the last replacement.
    sb.WriteString(g.sourceCode[cursor:])

    return sb.String(), nil
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
		SourceEnd:     sourcePos,
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

func (g *GobraGenerator) freshResultName(fn *ast.FuncDecl) string {
    // Collect all names already used in the result list.
    used := make(map[string]bool)
    if fn.Type.Results != nil {
        for _, field := range fn.Type.Results.List {
            for _, name := range field.Names {
                used[name.Name] = true
            }
        }
    }

    // Also collect names from parameters to avoid shadowing.
    if fn.Type.Params != nil {
        for _, field := range fn.Type.Params.List {
            for _, name := range field.Names {
                used[name.Name] = true
            }
        }
    }

    // Increment until we find a name not in use.
    for {
        g.varCount++
        candidate := fmt.Sprintf("r%d", g.varCount)
        if !used[candidate] {
            return candidate
        }
    }
}

// nameAllResults mutates fn.Type.Results in-place by assigning fresh names
// to every unnamed result, then returns the result list rendered as a string.
func (g *GobraGenerator) nameAllResults(fn *ast.FuncDecl) string {
    if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
        return ""
    }

    // Mutate the AST: assign fresh names to every unnamed result.
    for _, field := range fn.Type.Results.List {
        if len(field.Names) == 0 {
            field.Names = []*ast.Ident{ast.NewIdent(g.freshResultName(fn))}
        } else {
            for _, name := range field.Names {
                if name.Name == "" || name.Name == "_" {
                    name.Name = g.freshResultName(fn)
                }
            }
        }
    }

    // Render the (now fully named) result list into a local string builder.
    var sb strings.Builder
    sb.WriteString("(")
    for i, field := range fn.Type.Results.List {
        if i > 0 {
            sb.WriteString(", ")
        }
        for j, name := range field.Names {
            if j > 0 {
                sb.WriteString(", ")
            }
            sb.WriteString(name.Name)
        }
        sb.WriteString(" ")
        // Use printer to render the type node into the local builder.
        if err := printer.Fprint(&sb, g.fset, field.Type); err != nil {
            sb.WriteString("/* error rendering type */")
        }
    }
    sb.WriteString(")")
    return sb.String()
}

func (g *GobraGenerator) GenerateGobraFuncContract(node *ast.Node, funcContract *specir.FuncContract) {
	funcDecl, ok := (*node).(*ast.FuncDecl)
	if !ok {
		panic("Received not function decl node to contract generation")
	}

	p := &Printer{}
	if funcDecl.Type.Results != nil && funcDecl.Type.Results.List != nil {
		for _, field := range funcDecl.Type.Results.List {
			if len(field.Names) == 0 {
				panic("unnamed results")
			}
			p.resultNames = append(p.resultNames, field.Names[0].Name)
		}
	}

	// Generate requires with automatic tracking
	sourcePos := funcContract.Span.End
	for _, requires := range funcContract.Requires {
		g.AddReplacement(sourcePos, sourcePos, "\n", nil)
		g.AddReplacement(sourcePos, sourcePos, "requires ", nil)

		predicateSpan := &gobra_runner.SourcePosition{
			File:   requires.Span.File,
			Offset: requires.Span.Start,
		}

		p.buf.Reset()
		p.PrintExpr(requires.Expr)
		g.AddReplacement(sourcePos, sourcePos, p.buf.String(), predicateSpan)
	}

	// Generate ensures with automatic tracking
	for _, ensures := range funcContract.Ensures {
		g.AddReplacement(sourcePos, sourcePos, "\n", nil)
		g.AddReplacement(sourcePos, sourcePos, "ensures ", nil)

		predicateSpan := &gobra_runner.SourcePosition{
			File:   ensures.Span.File,
			Offset: ensures.Span.Start,
		}

		p.buf.Reset()
		p.PrintExpr(ensures.Expr)
		g.AddReplacement(sourcePos, sourcePos, p.buf.String(), predicateSpan)
	}
}

func (g *GobraGenerator) GenerateGobraLoopSpec(loopSpec *specir.LoopSpec) {
	sourcePos := loopSpec.Span.End
	for _, invariant := range loopSpec.Invariants {
		g.AddReplacement(sourcePos, sourcePos, "\n", nil)
		g.AddReplacement(sourcePos, sourcePos, "invariant ", nil)

		predicateSpan := &gobra_runner.SourcePosition{
			File:   invariant.Span.File,
			Offset: invariant.Span.Start,
		}

		g.AddReplacement(sourcePos, sourcePos, PrintExpr(invariant.Expr), predicateSpan)
	}
}

func (g *GobraGenerator) GenerateGobraAsserts(asserts *specir.AssertSpec) {
	sourcePos := asserts.Span.End

	g.AddReplacement(sourcePos, sourcePos, "\nassert ", nil)
	predicateSpan := &gobra_runner.SourcePosition{
		File:   asserts.Span.File,
		Offset: asserts.Span.Start,
	}
	g.AddReplacement(sourcePos, sourcePos, PrintExpr(asserts.Assertion.Expr), predicateSpan)
}

func (g *GobraGenerator) GenerateGobraPredicate(predicate *specir.LogicDecl) {
	sourcePos := len(g.sourceCode)

	g.AddReplacement(sourcePos, sourcePos, "\nghost\ndecreases\n", nil)
	g.AddReplacement(sourcePos, sourcePos, fmt.Sprintf("pure func %s(", predicate.Name), nil)
	for i, param := range predicate.Params {
		if i > 0 {
			g.AddReplacement(sourcePos, sourcePos, ", ", nil)
		}

		g.AddReplacement(sourcePos, sourcePos, param.Name, nil)
		g.AddReplacement(sourcePos, sourcePos, " ", nil)
		g.AddReplacement(sourcePos, sourcePos, typeExprStr(param.Type), nil)
	}
	g.AddReplacement(sourcePos, sourcePos, ") bool {\n\treturn ", nil)

	predicateSpan := &gobra_runner.SourcePosition{
		File:   predicate.Span.File,
		Offset: predicate.Span.Start,
	}
	g.AddReplacement(sourcePos, sourcePos, PrintExpr(predicate.Body), predicateSpan)

	g.AddReplacement(sourcePos, sourcePos, "\n}\n", nil)
}

// GenerateGobra generates Gobra code from AST data with automatic tracking
// Returns GobraGenerationResult containing both code and hybrid mapping
func GenerateGobra(astData *data.ASTData, sourceFile string) (*GobraGenerationResult, error) {
	g := &GobraGenerator{
		HybridMapping: gobra_runner.NewHybridMapping(sourceFile),
		sourceOffset: 0,
		currentOffset: 0,
		sourceFile:    sourceFile,
		sourceCode: astData.Source,
		fset: astData.FileSet,
	}

	for _, decl := range astData.File.Decls {
		switch decl.(type) {
		case *ast.FuncDecl: {
			sourceResultSection := decl.(*ast.FuncDecl).Type.Results
			if sourceResultSection == nil {
				continue
			}

			sourcePos := astData.FileSet.Position(sourceResultSection.Pos()).Offset
			sourceEnd := astData.FileSet.Position(sourceResultSection.End()).Offset

			namedResultsSection := g.nameAllResults(decl.(*ast.FuncDecl))
			
			g.AddReplacement(sourcePos, sourceEnd, namedResultsSection, nil)
		}
		}
	}

	ast.Inspect(astData.File, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		node_info, ok := astData.NodeInfo[n]
		if ok {
			if node_info.FuncContract != nil {
				g.GenerateGobraFuncContract(&n, node_info.FuncContract)
			} else if node_info.LoopSpec != nil {
				g.GenerateGobraLoopSpec(node_info.LoopSpec)
			} else if node_info.Asserts != nil {
				g.GenerateGobraAsserts(node_info.Asserts)
			} else {
				panic("Empty NodeInfo")
			}
		}

		return true
	})

	for _, specDecl := range astData.Spec.LogicDecls {
		if specDecl.Kind == specir.LogicPredicate {
			g.GenerateGobraPredicate(specDecl)
		}
	}

	generated, err := g.Render()
	if err != nil {
		return nil, err
	}

	return &GobraGenerationResult{
		GobraCode:     generated,
		HybridMapping: g.HybridMapping,
	}, nil
}
