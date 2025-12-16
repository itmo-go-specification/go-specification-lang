package specir

// NamedPredicate — предикат с синтаксическим именем (bound:, count:).
// ensures bound: 0 <= \result <= n;
// Label: "bound"
// Expr: "0 <= \result <= n"

type NamedPredicate struct {
	Label string
	Expr  Expr
	Span  Span
}

// LocationSetKind — location
type LocationSetKind int

const (
	LocSetNothing LocationSetKind = iota // \nothing
	LocSetVars                           // assigns x, y;
	// TODO: assignes a[0...n-1], etc.
)

// LocationSet - assigns / loop assigns
type LocationSet struct {
	Kind LocationSetKind

	// if Kind == LocSetNothing then Vars is empty. Ex: \nothinf
	// Vars используется, если Kind == LocSetVars
	Vars []string // if Kind == LocSetVars. Ex: assigns i, counted;
}

// ===== GOSL (ACSL) Contract type =====
// requires valid: \valid_read(a + (0..n-1));
// ensures bound: 0 <= \result <= n;
// ensures count: \result == Count(a, n, v);
type FuncContract struct {
	// Func name from source go ast
	FuncName string

	// funtion receiver. RecvTypeName = "" if func f(...)
	RecvTypeName string

	// Positions
	Span     Span
	StartPos Position
	EndPos   Position

	Requires []NamedPredicate
	Ensures  []NamedPredicate

	// assigns locations. Ex: assigns \nothing;
	Assigns LocationSet
}

// ===== GOSL (ACSL) Loop invariant type =====
// loop invariant / loop assigns / loop variant
// loop invariant bound: 0 <= i <= n;
// loop invariant count: counted == Count(a, i, v);
type LoopSpec struct {
	Span     Span
	StartPos Position
	EndPos   Position

	// Optional: parent function name
	ParentFuncName string

	Invariants []NamedPredicate
	// loop assigns ...;
	Assigns LocationSet
	// loop variant term;
	Variant Expr
}
