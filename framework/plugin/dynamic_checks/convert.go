package dynamicchecks

import (
	"fmt"
	"go/ast"
	"go/token"
	specir "gpma/framework/gosl/ir"
	"strconv"
)

type exprConverter interface {
	Convert(specir.Expr) ast.Expr
}

// baseConverter translates "pure" IR expressions: literals, variables,
// arithmetic, function calls, and the ternary operator.
// Use for assert and loop invariant contexts where \old and \result are not allowed.
type baseConverter struct{}

func (c *baseConverter) Convert(expr specir.Expr) ast.Expr {
	return convertExpr(c, expr)
}

// contractConverter translates IR expressions in the context of a function contract.
// Extends baseConverter with support for \old and \result.
// Use for requires and ensures clauses.
type contractConverter struct {
	oldVars  map[specir.Expr]*oldVarInfo
	oldOrder []specir.Expr
	counter  int
}

type oldVarInfo struct {
	name       string
	expr       specir.Expr
	isDeclared bool
}

func newContractConverter() *contractConverter {
	return &contractConverter{
		oldVars:  make(map[specir.Expr]*oldVarInfo),
		oldOrder: make([]specir.Expr, 0),
	}
}

func (c *contractConverter) Convert(expr specir.Expr) ast.Expr {
	switch e := expr.(type) {
	case *specir.ExprOld:
		return c.convertOld(e)
	case *specir.ExprResult:
		return &ast.Ident{Name: "__result"}
	default:
		return convertExpr(c, expr)
	}
}

func (c *contractConverter) CollectOld(expr specir.Expr) {
	specir.Walk(expr, func(e specir.Expr) bool {
		if old, ok := e.(*specir.ExprOld); ok {
			if _, exists := c.oldVars[old.Inner]; !exists {
				c.counter++
				c.oldVars[old.Inner] = &oldVarInfo{
					name: fmt.Sprintf("__old_%d", c.counter),
					expr: old.Inner,
				}
				c.oldOrder = append(c.oldOrder, old.Inner)
			}
		}
		return true
	})
}

func (c *contractConverter) GetOldVars() []*oldVarInfo {
	result := make([]*oldVarInfo, 0, len(c.oldOrder))
	for _, expr := range c.oldOrder {
		result = append(result, c.oldVars[expr])
	}
	return result
}

func (c *contractConverter) newVar() string {
	c.counter++
	return strconv.Itoa(c.counter)
}

func (c *contractConverter) convertOld(old *specir.ExprOld) ast.Expr {
	if info, exists := c.oldVars[old.Inner]; exists && info.isDeclared {
		return &ast.Ident{Name: info.name}
	}
	return c.Convert(old.Inner)
}

func convertExpr(conv exprConverter, expr specir.Expr) ast.Expr {
	switch e := expr.(type) {
	case *specir.ExprLiteral:
		return convertLiteral(e)
	case *specir.ExprVar:
		return &ast.Ident{Name: e.Name}
	case *specir.ExprUnary:
		return &ast.UnaryExpr{
			Op: convertUnaryOp(e.Op),
			X:  conv.Convert(e.Expr),
		}
	case *specir.ExprBinary:
		return &ast.BinaryExpr{
			X:  conv.Convert(e.Left),
			Op: convertBinaryOp(e.Op),
			Y:  conv.Convert(e.Right),
		}
	case *specir.ExprCall:
		return convertCall(conv, e)
	case *specir.ExprConditional:
		return convertConditional(conv, e)
	case *specir.ExprQuantifier:
		return convertQuantifier(conv, e)
	case *specir.ExprOld:
		panic(`\old is not valid in this context (only allowed in ensures)`)
	case *specir.ExprResult:
		panic(`\result is not valid in this context (only allowed in ensures)`)
	default:
		panic(fmt.Sprintf("unsupported expression type: %T", expr))
	}
}

func convertLiteral(lit *specir.ExprLiteral) ast.Expr {
	switch lit.Kind {
	case specir.LitBool:
		if lit.Bool {
			return &ast.Ident{Name: "true"}
		}
		return &ast.Ident{Name: "false"}
	case specir.LitInteger:
		return &ast.BasicLit{
			Kind:  token.INT,
			Value: fmt.Sprintf("%d", lit.Int),
		}
	case specir.LitReal:
		return &ast.BasicLit{
			Kind:  token.FLOAT,
			Value: fmt.Sprintf("%f", lit.Real),
		}
	case specir.LitFloat:
		return &ast.BasicLit{
			Kind:  token.FLOAT,
			Value: fmt.Sprintf("%f", lit.Float),
		}
	case specir.LitString:
		return &ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf(`"%s"`, lit.Str),
		}
	default:
		panic(fmt.Sprintf("unsupported literal kind: %v", lit.Kind))
	}
}

func convertUnaryOp(op specir.UnaryOp) token.Token {
	switch op {
	case specir.UnaryPlus:
		return token.ADD
	case specir.UnaryMinus:
		return token.SUB
	case specir.UnaryNot:
		return token.NOT
	default:
		return token.ADD
	}
}

func convertBinaryOp(op specir.BinaryOp) token.Token {
	switch op {
	case specir.OpAdd:
		return token.ADD
	case specir.OpSub:
		return token.SUB
	case specir.OpMul:
		return token.MUL
	case specir.OpDiv:
		return token.QUO
	case specir.OpMod:
		return token.REM
	case specir.OpEq:
		return token.EQL
	case specir.OpNeq:
		return token.NEQ
	case specir.OpLe:
		return token.LEQ
	case specir.OpGe:
		return token.GEQ
	case specir.OpLt:
		return token.LSS
	case specir.OpGt:
		return token.GTR
	case specir.OpAnd:
		return token.LAND
	case specir.OpOr:
		return token.LOR
	case specir.OpXor:
		return token.XOR
	default:
		return token.EQL
	}
}

func convertCall(conv exprConverter, call *specir.ExprCall) ast.Expr {
	args := make([]ast.Expr, len(call.Args))
	for i, arg := range call.Args {
		args[i] = conv.Convert(arg)
	}
	return &ast.CallExpr{
		Fun:  &ast.Ident{Name: call.Name},
		Args: args,
	}
}

func convertConditional(conv exprConverter, cond *specir.ExprConditional) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
				Results: &ast.FieldList{
					List: []*ast.Field{
						{Type: &ast.Ident{Name: "bool"}},
					},
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.IfStmt{
						Cond: conv.Convert(cond.Cond),
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{conv.Convert(cond.Then)},
								},
							},
						},
						Else: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{conv.Convert(cond.Else)},
								},
							},
						},
					},
				},
			},
		},
	}
}
