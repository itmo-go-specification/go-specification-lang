package gobra

import (
	"fmt"
	specir "gpma/framework/gosl/ir"
	"strings"
)

// PrintExpr returns a human-readable, single-line string for any specir.Expr.
func PrintExpr(expr specir.Expr) string {
	p := &Printer{}
	p.PrintExpr(expr)
	return p.buf.String()
}

type Printer struct {
	buf strings.Builder
	resultNames []string
}

// ── Flat (single-line) Printer ────────────────────────────────────────────────

func (p *Printer) PrintExpr(expr specir.Expr) {
	if expr == nil {
		p.buf.WriteString("<nil>")
		return
	}
	switch e := expr.(type) {
	case *specir.ExprLiteral:
		p.printLiteral(e)
	case *specir.ExprVar:
		p.buf.WriteString(e.Name)
	case *specir.ExprUnary:
		p.printUnary(e)
	case *specir.ExprBinary:
		p.printBinary(e)
	case *specir.ExprCall:
		p.printCall(e)
	case *specir.ExprConditional:
		p.printConditional(e)
	case *specir.ExprResult:
		p.printResult(e)
	case *specir.ExprOld:
		p.buf.WriteString(`old(`)
		p.PrintExpr(e.Inner)
		p.buf.WriteByte(')')
	case *specir.ExprQuantifier:
		p.printQuantifier(e)
	default:
		fmt.Fprintf(&p.buf, "<unknown:%T>", expr)
	}
}

func (p *Printer) printLiteral(e *specir.ExprLiteral) {
	switch e.Kind {
	case specir.LitBool:
		if e.Bool {
			p.buf.WriteString(`true`)
		} else {
			p.buf.WriteString(`false`)
		}
	case specir.LitInteger:
		fmt.Fprintf(&p.buf, "%d", e.Int)
	case specir.LitReal:
		fmt.Fprintf(&p.buf, "%g", e.Real)
	case specir.LitFloat:
		fmt.Fprintf(&p.buf, "%g", e.Float)
	case specir.LitString:
		fmt.Fprintf(&p.buf, "%q", e.Str)
	default:
		fmt.Fprintf(&p.buf, "<literal:%d>", e.Kind)
	}
}

func (p *Printer) printUnary(e *specir.ExprUnary) {
	p.buf.WriteString(unaryOpStr(e.Op))
	// Parenthesise the operand when it is itself a binary/conditional/quantifier
	// to avoid ambiguity (unary binds tighter).
	needParen := requiresParens(e.Expr)
	if needParen {
		p.buf.WriteByte('(')
	}
	p.PrintExpr(e.Expr)
	if needParen {
		p.buf.WriteByte(')')
	}
}

func (p *Printer) printBinary(e *specir.ExprBinary) {
	if needsLeftParen(e.Op, e.Left) {
		p.buf.WriteByte('(')
		p.PrintExpr(e.Left)
		p.buf.WriteByte(')')
	} else {
		p.PrintExpr(e.Left)
	}
	fmt.Fprintf(&p.buf, " %s ", binaryOpStr(e.Op))
	if needsRightParen(e.Op, e.Right) {
		p.buf.WriteByte('(')
		p.PrintExpr(e.Right)
		p.buf.WriteByte(')')
	} else {
		p.PrintExpr(e.Right)
	}
}

func (p *Printer) printCall(e *specir.ExprCall) {
	p.buf.WriteString(e.Name)
	p.buf.WriteByte('(')
	for i, arg := range e.Args {
		if i > 0 {
			p.buf.WriteString(", ")
		}
		p.PrintExpr(arg)
	}
	p.buf.WriteByte(')')
}

func (p *Printer) printConditional(e *specir.ExprConditional) {
	p.buf.WriteByte('(')
	p.PrintExpr(e.Cond)
	p.buf.WriteString(" ? ")
	p.PrintExpr(e.Then)
	p.buf.WriteString(" : ")
	p.PrintExpr(e.Else)
	p.buf.WriteByte(')')
}

func (p *Printer) printResult(e *specir.ExprResult) {
	if p.resultNames == nil || len(p.resultNames) <= e.Index {
		panic("invalid result index")
	} else {
		p.buf.WriteString(p.resultNames[e.Index])
	}
}

func (p *Printer) printQuantifier(e *specir.ExprQuantifier) {
	switch e.Kind {
	case specir.QuantifierForall:
		p.buf.WriteString(`forall`)
	case specir.QuantifierExists:
		p.buf.WriteString(`exists`)
	default:
		fmt.Fprintf(&p.buf, `\quantifier%d`, e.Kind)
	}
	p.buf.WriteByte(' ')
	for i, v := range e.Vars {
		if i > 0 {
			p.buf.WriteString(", ")
		}
		fmt.Fprintf(&p.buf, "%s %s", v.Name, typeExprStr(v.Type))
	}
	p.buf.WriteString(" :: ")
	p.PrintExpr(e.Body)
}

// ── Operator helpers ──────────────────────────────────────────────────────────

func unaryOpStr(op specir.UnaryOp) string {
	switch op {
	case specir.UnaryPlus:
		return "+"
	case specir.UnaryMinus:
		return "-"
	case specir.UnaryNot:
		return "!"
	default:
		return fmt.Sprintf("<unary:%d>", op)
	}
}

func binaryOpStr(op specir.BinaryOp) string {
	switch op {
	case specir.OpAdd:
		return "+"
	case specir.OpSub:
		return "-"
	case specir.OpMul:
		return "*"
	case specir.OpDiv:
		return "/"
	case specir.OpMod:
		return "%"
	case specir.OpEq:
		return "=="
	case specir.OpNeq:
		return "!="
	case specir.OpLe:
		return "<="
	case specir.OpGe:
		return ">="
	case specir.OpLt:
		return "<"
	case specir.OpGt:
		return ">"
	case specir.OpAnd:
		return "&&"
	case specir.OpOr:
		return "||"
	case specir.OpXor:
		return "^^"
	case specir.OpImplies:
		return "==>"
	default:
		return fmt.Sprintf("<op:%d>", op)
	}
}

// precedence returns a numeric precedence for binary operators (higher = tighter).
func precedence(op specir.BinaryOp) int {
	switch op {
	case specir.OpImplies:
		return 1
	case specir.OpOr, specir.OpXor:
		return 2
	case specir.OpAnd:
		return 3
	case specir.OpEq, specir.OpNeq, specir.OpLe, specir.OpGe, specir.OpLt, specir.OpGt:
		return 4
	case specir.OpAdd, specir.OpSub:
		return 5
	case specir.OpMul, specir.OpDiv, specir.OpMod:
		return 6
	default:
		return 0
	}
}

// requiresParens reports whether expr should be wrapped in parens when used as
// the operand of a unary operator.
func requiresParens(e specir.Expr) bool {
	switch e.(type) {
	case *specir.ExprBinary, *specir.ExprConditional, *specir.ExprQuantifier:
		return true
	}
	return false
}

func needsLeftParen(op specir.BinaryOp, left specir.Expr) bool {
	b, ok := left.(*specir.ExprBinary)
	if !ok {
		return false
	}
	return precedence(b.Op) < precedence(op)
}

func needsRightParen(op specir.BinaryOp, right specir.Expr) bool {
	b, ok := right.(*specir.ExprBinary)
	if !ok {
		return false
	}
	// Left-associative: wrap the right child only when it has strictly lower-or-
	// equal precedence, so that a-(b-c) keeps its parens but a+b+c does not.
	return precedence(b.Op) <= precedence(op)
}

// typeExprStr renders a specir.TypeExpr as a string.
// If TypeExpr implements fmt.Stringer that is used; otherwise %v is the fallback.
func typeExprStr(t specir.TypeExpr) string {
	switch t.Kind {
	case specir.TypeBoolean:
		return "bool"
	case specir.TypeInteger:
		return "int"
	case specir.TypeReal:
		return "real"
	case specir.TypeCustom:
		return t.Name
	default:
		return fmt.Sprintf("<type:%d>", t.Kind)
	}
}
