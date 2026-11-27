package specir

type Visitor func(expr Expr) bool

func Walk(expr Expr, visitor Visitor) {
	if expr == nil || !visitor(expr) {
		return
	}

	switch e := expr.(type) {
	case *ExprSelector:
		Walk(e.Expr, visitor)
	case *ExprIndex:
		Walk(e.Expr, visitor)
		Walk(e.Index, visitor)
	case *ExprUnary:
		Walk(e.Expr, visitor)
	case *ExprBinary:
		Walk(e.Left, visitor)
		Walk(e.Right, visitor)
	case *ExprCall:
		for _, arg := range e.Args {
			Walk(arg, visitor)
		}
	case *ExprConditional:
		Walk(e.Cond, visitor)
		Walk(e.Then, visitor)
		Walk(e.Else, visitor)
	case *ExprOld:
		Walk(e.Inner, visitor)
	case *ExprQuantifier:
		Walk(e.Body, visitor)
	case *ExprResult:
	case *ExprVar:
	case *ExprLiteral:
	default:
	}
}
