package dynamicchecks

import (
	"fmt"
	"go/ast"
	"go/token"
	specir "gpma/framework/gosl/ir"
	"strconv"
)

type Converter struct {
	oldVars  map[specir.Expr]*oldVarInfo
	oldOrder []specir.Expr
	counter  int
}

type oldVarInfo struct {
	name       string
	expr       specir.Expr
	isDeclared bool
}

func NewConverter() *Converter {
	return &Converter{
		oldVars:  make(map[specir.Expr]*oldVarInfo),
		oldOrder: make([]specir.Expr, 0),
		counter:  0,
	}
}

// CollectOld собирает все \old выражения
func (c *Converter) CollectOld(expr specir.Expr) {
	specir.Walk(expr, func(e specir.Expr) bool {
		if old, ok := e.(*specir.ExprOld); ok {
			if _, exists := c.oldVars[old.Inner]; !exists {
				c.counter++
				c.oldVars[old.Inner] = &oldVarInfo{
					name:       fmt.Sprintf("__old_%d", c.counter),
					expr:       old.Inner,
					isDeclared: false,
				}
				c.oldOrder = append(c.oldOrder, old.Inner)
			}
		}
		return true
	})
}

func (c *Converter) GetOldVars() []*oldVarInfo {
	result := make([]*oldVarInfo, 0, len(c.oldOrder))
	for _, expr := range c.oldOrder {
		result = append(result, c.oldVars[expr])
	}
	return result
}

func (c *Converter) Convert(expr specir.Expr) ast.Expr {
	switch e := expr.(type) {
	case *specir.ExprLiteral:
		return c.convertLiteral(e)
	case *specir.ExprVar:
		return &ast.Ident{Name: e.Name}
	case *specir.ExprUnary:
		return c.convertUnary(e)
	case *specir.ExprBinary:
		return c.convertBinary(e)
	case *specir.ExprCall:
		return c.convertCall(e)
	case *specir.ExprConditional:
		return c.convertConditional(e)
	case *specir.ExprResult:
		return &ast.Ident{Name: "__result"}
	case *specir.ExprOld:
		return c.convertOld(e)
	default:
		return &ast.Ident{Name: "true"}
	}
}

func (c *Converter) convertLiteral(lit *specir.ExprLiteral) ast.Expr {
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
	case specir.LitString:
		return &ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf(`"%s"`, lit.Str),
		}
	default:
		return &ast.Ident{Name: "true"}
	}
}

func (c *Converter) convertUnary(unary *specir.ExprUnary) ast.Expr {
	op := token.ADD
	switch unary.Op {
	case specir.UnaryPlus:
		op = token.ADD
	case specir.UnaryMinus:
		op = token.SUB
	case specir.UnaryNot:
		op = token.NOT
	}

	return &ast.UnaryExpr{
		Op: op,
		X:  c.Convert(unary.Expr),
	}
}

func (c *Converter) convertBinary(binary *specir.ExprBinary) ast.Expr {
	op := c.convertBinaryOp(binary.Op)
	return &ast.BinaryExpr{
		X:  c.Convert(binary.Left),
		Op: op,
		Y:  c.Convert(binary.Right),
	}
}

func (c *Converter) convertBinaryOp(op specir.BinaryOp) token.Token {
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
		return token.NEQ // Для булевых значений XOR через !=
	default:
		return token.EQL
	}
}

func (c *Converter) convertCall(call *specir.ExprCall) ast.Expr {
	args := make([]ast.Expr, len(call.Args))
	for i, arg := range call.Args {
		args[i] = c.Convert(arg)
	}
	return &ast.CallExpr{
		Fun:  &ast.Ident{Name: call.Name},
		Args: args,
	}
}

func (c *Converter) convertConditional(cond *specir.ExprConditional) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
				Results: &ast.FieldList{
					List: []*ast.Field{
						{Type: &ast.Ident{Name: "bool"}}, // предполагаем bool
					},
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.IfStmt{
						Cond: c.Convert(cond.Cond),
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{c.Convert(cond.Then)},
								},
							},
						},
						Else: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{c.Convert(cond.Else)},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (c *Converter) convertOld(old *specir.ExprOld) ast.Expr {
	if info, exists := c.oldVars[old.Inner]; exists && info.isDeclared {
		return &ast.Ident{Name: info.name}
	}
	return c.Convert(old.Inner)
}

func (c *Converter) NewVar() string {
	c.counter++
	return strconv.Itoa(c.counter)
}
