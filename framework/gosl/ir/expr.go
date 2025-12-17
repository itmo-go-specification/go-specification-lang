package specir

type Expr interface {
	exprNode()
	GetSpan() Span
}

type BaseExpr struct {
	Span Span
}

func (b BaseExpr) GetSpan() Span { return b.Span }

type LiteralKind int

const (
	LitBool LiteralKind = iota
	LitInteger
	LitReal
	LitString
	LitFloat
)

// ExprLiteral - \true, \false, numbers, strings
type ExprLiteral struct {
	BaseExpr
	Kind LiteralKind
	Bool bool
	Int  int64
	Real float64
	Str  string
	Float float64
}

func (*ExprLiteral) exprNode() {}

// ExprVar — variable, identifier, function paramentrs, ...
type ExprVar struct {
	BaseExpr
	Name string
}

func (*ExprVar) exprNode() {}

// UnaryOp - unary-op ::= + | - | !
type UnaryOp int

const (
	UnaryPlus UnaryOp = iota
	UnaryMinus
	UnaryNot
)

type ExprUnary struct {
	BaseExpr
	Op   UnaryOp
	Expr Expr
}

func (*ExprUnary) exprNode() {}

// BinaryOp — ...
type BinaryOp int

const (
	OpAdd BinaryOp = iota // +
	OpSub                 // -
	OpMul                 // *
	OpDiv                 // /
	OpMod                 // %

	OpEq  // ==
	OpNeq // !=
	OpLe  // <=
	OpGe  // >=
	OpLt  // <
	OpGt  // >

	OpAnd // &&
	OpOr  // ||
	OpXor // ^^
	// TODO: bitwise operation for go ???
)

type ExprBinary struct {
	BaseExpr
	Op    BinaryOp
	Left  Expr
	Right Expr
}

func (*ExprBinary) exprNode() {}

// ExprCall - Ex: call (Count(a, n, v)).
// term ::= ident ( term (, term)* )   (* function application *)
// pred ::= ident ( term (, term)* )   (* predicate application *)
type ExprCall struct {
	BaseExpr
	Name string
	Args []Expr
}

func (*ExprCall) exprNode() {}

// ExprConditional - ternary operator (cond ? thenExpr : elseExpr).
// term ::= term ? term : term
// pred ::= term ? pred : pred | pred ? pred : pred
type ExprConditional struct {
	BaseExpr
	Cond Expr
	Then Expr
	Else Expr
}

func (*ExprConditional) exprNode() {}

// ExprResult - \result.
// term ::= \result
type ExprResult struct {
	BaseExpr
}

func (*ExprResult) exprNode() {}

// ExprOld — \old(expr)
// term ::= \old ( term )
// pred ::= \old ( pred )
type ExprOld struct {
	BaseExpr
	Inner Expr
}

func (*ExprOld) exprNode() {}
