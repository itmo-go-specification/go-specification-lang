package dynamicchecks

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"strings"
	"testing"

	specir "gpma/framework/gosl/ir"
)

func TestConvertBinary_OperatorPrecedence(t *testing.T) {
	tests := []struct {
		name string
		expr specir.Expr
		want string
	}{
		{
			name: "add_before_mul",
			expr: bin(specir.OpEq,
				bin(specir.OpAdd, intLit(1),
					bin(specir.OpMul, intLit(2), intLit(3))),
				intLit(7)),
			want: "1 + 2*3 == 7",
		},
		{
			name: "grouping_mul",
			expr: bin(specir.OpMul,
				bin(specir.OpAdd, intLit(1), intLit(2)),
				varExpr("c")),
			want: "(1 + 2) * c",
		},
		{
			name: "sub_nested",
			expr: bin(specir.OpSub,
				varExpr("a"),
				bin(specir.OpSub, varExpr("b"), varExpr("c"))),
			want: "a - (b - c)",
		},
		{
			name: "div_of_mul",
			expr: bin(specir.OpDiv,
				varExpr("a"),
				bin(specir.OpMul, varExpr("b"), varExpr("c"))),
			want: "a / (b * c)",
		},
		{
			name: "and_with_or",
			expr: bin(specir.OpAnd,
				varExpr("a"),
				bin(specir.OpOr, varExpr("b"), varExpr("c"))),
			want: "a && (b || c)",
		},
		{
			name: "or_and_grouping",
			expr: bin(specir.OpAnd,
				bin(specir.OpOr, varExpr("a"), varExpr("b")),
				varExpr("c")),
			want: "(a || b) && c",
		},
		{
			name: "implies_rhs_and",
			expr: bin(specir.OpImplies,
				bin(specir.OpAnd, varExpr("a"), varExpr("b")),
				varExpr("c")),
			want: "!(a && b) || c",
		},
		{
			name: "implies_lhs_simple",
			expr: bin(specir.OpImplies,
				varExpr("a"),
				bin(specir.OpAnd, varExpr("b"), varExpr("c"))),
			want: "!a || b && c",
		},
	}

	conv := &baseConverter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeGo(renderExpr(conv.Convert(tt.expr)))
			want := normalizeGo(tt.want)

			if !strings.Contains(got, want) {
				t.Fatalf("rendered = %q, want substring = %q",
					got, want)
			}
		})
	}
}

func TestExtractBounds_VarOnEitherSide(t *testing.T) {
	tests := []struct {
		name   string
		guard  specir.Expr
		wantLo int64
		wantHi int64
	}{
		{
			name: "le_left_and_lt_right",
			guard: bin(specir.OpAnd,
				bin(specir.OpLe, intLit(0), varExpr("i")),
				bin(specir.OpLt, varExpr("i"), intLit(3))),
			wantLo: 0,
			wantHi: 3,
		},
		{
			name: "ge_left_and_lt_right",
			guard: bin(specir.OpAnd,
				bin(specir.OpGe, varExpr("i"), intLit(3)),
				bin(specir.OpLt, varExpr("i"), intLit(10))),
			wantLo: 3,
			wantHi: 10,
		},
		{
			name: "gt_left_and_lt_right",
			guard: bin(specir.OpAnd,
				bin(specir.OpGt, varExpr("i"), intLit(0)),
				bin(specir.OpLt, varExpr("i"), intLit(3))),
			wantLo: 1,
			wantHi: 3,
		},
		{
			name: "le_left_and_ge_right",
			guard: bin(specir.OpAnd,
				bin(specir.OpLe, varExpr("i"), intLit(3)),
				bin(specir.OpLe, intLit(0), varExpr("i"))),
			wantLo: 0,
			wantHi: 4,
		},
		{
			name: "lower_bound_on_right",
			guard: bin(specir.OpAnd,
				bin(specir.OpLe, intLit(2), varExpr("i")),
				bin(specir.OpLt, varExpr("i"), intLit(8))),
			wantLo: 2,
			wantHi: 8,
		},
		{
			name: "upper_bound_on_left",
			guard: bin(specir.OpAnd,
				bin(specir.OpLe, varExpr("i"), intLit(5)),
				bin(specir.OpLe, intLit(1), varExpr("i"))),
			wantLo: 1,
			wantHi: 6,
		},
		{
			name: "lower_bound_defaults_to_zero",
			guard: bin(specir.OpLt,
				varExpr("i"), intLit(5)),
			wantLo: 0,
			wantHi: 5,
		},
		{
			name: "inclusive_upper_bound",
			guard: bin(specir.OpLe,
				varExpr("i"), intLit(5)),
			wantLo: 0,
			wantHi: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lo, hi := extractBounds(tt.guard, "i")

			if lo == nil || hi == nil {
				t.Fatalf("extractBounds returned nil: lo=%v hi=%v",
					lo, hi)
			}

			if got := evalIntExpr(t, lo); got != tt.wantLo {
				t.Errorf("lo=%d, want %d; expression=%s",
					got, tt.wantLo, renderIRExpr(lo))
			}

			if got := evalIntExpr(t, hi); got != tt.wantHi {
				t.Errorf("hi=%d, want %d; expression=%s",
					got, tt.wantHi, renderIRExpr(hi))
			}
		})
	}
}

func TestExtractBounds_InvalidPatterns(t *testing.T) {
	tests := []struct {
		name  string
		guard specir.Expr
	}{
		{
			name:  "no_upper_bound",
			guard: bin(specir.OpLe, intLit(0), varExpr("i")),
		},
		{
			name: "duplicate_lower_bounds",
			guard: bin(specir.OpAnd,
				bin(specir.OpLe, intLit(0), varExpr("i")),
				bin(specir.OpLe, intLit(1), varExpr("i"))),
		},
		{
			name: "duplicate_upper_bounds",
			guard: bin(specir.OpAnd,
				bin(specir.OpLt, varExpr("i"), intLit(5)),
				bin(specir.OpLe, varExpr("i"), intLit(10))),
		},
		{
			name:  "unrelated_comparison",
			guard: bin(specir.OpLt, varExpr("x"), intLit(5)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lo, hi := extractBounds(tt.guard, "i")
			if lo != nil || hi != nil {
				t.Fatalf("expected (nil, nil), got lo=%v hi=%v",
					lo, hi)
			}
		})
	}
}

func bin(op specir.BinaryOp, left, right specir.Expr) specir.Expr {
	return &specir.ExprBinary{
		Op:    op,
		Left:  left,
		Right: right,
	}
}

func intLit(v int64) specir.Expr {
	return &specir.ExprLiteral{
		Kind: specir.LitInteger,
		Int:  v,
	}
}

func varExpr(name string) specir.Expr {
	return &specir.ExprVar{Name: name}
}

func normalizeGo(s string) string {
	return strings.Join(strings.Fields(s), "")
}

func renderExpr(expr ast.Expr) string {
	var b bytes.Buffer
	if err := format.Node(&b, token.NewFileSet(), expr); err != nil {
		panic(err)
	}
	return strings.TrimSpace(b.String())
}

func renderIRExpr(expr specir.Expr) string {
	if expr == nil {
		return "<nil>"
	}
	return renderExpr((&baseConverter{}).Convert(expr))
}

// evalIntExpr evaluates the integer expressions used by these tests.
// It intentionally supports only integer literals and addition.
func evalIntExpr(t *testing.T, expr specir.Expr) int64 {
	t.Helper()

	switch e := expr.(type) {
	case *specir.ExprLiteral:
		if e.Kind != specir.LitInteger {
			t.Fatalf("expected integer literal, got kind %v", e.Kind)
		}
		return e.Int

	case *specir.ExprBinary:
		if e.Op != specir.OpAdd {
			t.Fatalf("unsupported expression in test: operator %v", e.Op)
		}
		return evalIntExpr(t, e.Left) + evalIntExpr(t, e.Right)

	default:
		t.Fatalf("unsupported expression in test: %T", expr)
		return 0
	}
}

func TestConvertForall_RecognizesVarOnEitherSide(t *testing.T) {
	q := &specir.ExprQuantifier{
		Kind: specir.QuantifierForall,
		Vars: []specir.BoundVar{
			{
				Name: "i",
				Type: specir.TypeExpr{Kind: specir.TypeInteger},
			},
		},
		Body: &specir.ExprBinary{
			Op: specir.OpImplies,
			Left: &specir.ExprBinary{
				Op: specir.OpAnd,
				Left: &specir.ExprBinary{
					Op:    specir.OpGe,
					Left:  &specir.ExprVar{Name: "i"},
					Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 3},
				},
				Right: &specir.ExprBinary{
					Op:    specir.OpLt,
					Left:  &specir.ExprVar{Name: "i"},
					Right: &specir.ExprLiteral{Kind: specir.LitInteger, Int: 10},
				},
			},
			Right: &specir.ExprVar{Name: "i"},
		},
	}

	got := convertQuantifier(&baseConverter{}, q)
	if got == nil {
		t.Fatal("convertQuantifier returned nil")
	}
}
