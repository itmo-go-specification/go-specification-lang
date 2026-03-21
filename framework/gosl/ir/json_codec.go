package specir

import (
	"encoding/json"
	"fmt"
)

// This file contains JSON codecs for the IR without mixing the struct declarations
// with serialization logic.
//
// The main complication is Expr (an interface). We serialize it as a tagged object:
//   {"kind":"binary", ...}
//
// Kinds:
//   "literal", "var", "unary", "binary", "call", "cond", "result", "old"

// -----------------------------
// Expr (polymorphic) JSON codec
// -----------------------------

type exprJSON struct {
	Kind string `json:"kind"`

	// Common
	Span Span `json:"span,omitempty"`

	// Literal
	LitKind LiteralKind `json:"litKind,omitempty"`
	Bool    bool        `json:"bool,omitempty"`
	Int     int64       `json:"int,omitempty"`
	Real    float64     `json:"real,omitempty"`
	Str     string      `json:"str,omitempty"`

	// Var / Call
	Name string `json:"name,omitempty"`

	// Unary / Binary
	UnaryOp  UnaryOp  `json:"unaryOp,omitempty"`
	BinaryOp BinaryOp `json:"binaryOp,omitempty"`

	// Children
	Expr  *exprJSON   `json:"expr,omitempty"`
	Left  *exprJSON   `json:"left,omitempty"`
	Right *exprJSON   `json:"right,omitempty"`
	Args  []*exprJSON `json:"args,omitempty"`

	// Conditional
	Cond *exprJSON `json:"cond,omitempty"`
	Then *exprJSON `json:"then,omitempty"`
	Else *exprJSON `json:"else,omitempty"`

	// Old
	Inner *exprJSON `json:"inner,omitempty"`

	// Quantifier
	QuantifierKind QuantifierKind `json:"quantifierKind,omitempty"`
	Vars           []BoundVar     `json:"vars,omitempty"`
	Body           *exprJSON      `json:"body,omitempty"`
}

func marshalExpr(e Expr) (*exprJSON, error) {
	if e == nil {
		return nil, nil
	}
	switch t := e.(type) {
	case *ExprLiteral:
		return &exprJSON{
			Kind:    "literal",
			Span:    t.GetSpan(),
			LitKind: t.Kind,
			Bool:    t.Bool,
			Int:     t.Int,
			Real:    t.Real,
			Str:     t.Str,
		}, nil
	case *ExprVar:
		return &exprJSON{
			Kind: "var",
			Span: t.GetSpan(),
			Name: t.Name,
		}, nil
	case *ExprUnary:
		inner, err := marshalExpr(t.Expr)
		if err != nil {
			return nil, err
		}
		return &exprJSON{
			Kind:    "unary",
			Span:    t.GetSpan(),
			UnaryOp: t.Op,
			Expr:    inner,
		}, nil
	case *ExprBinary:
		l, err := marshalExpr(t.Left)
		if err != nil {
			return nil, err
		}
		r, err := marshalExpr(t.Right)
		if err != nil {
			return nil, err
		}
		return &exprJSON{
			Kind:     "binary",
			Span:     t.GetSpan(),
			BinaryOp: t.Op,
			Left:     l,
			Right:    r,
		}, nil
	case *ExprCall:
		args := make([]*exprJSON, 0, len(t.Args))
		for _, a := range t.Args {
			aj, err := marshalExpr(a)
			if err != nil {
				return nil, err
			}
			args = append(args, aj)
		}
		return &exprJSON{
			Kind: "call",
			Span: t.GetSpan(),
			Name: t.Name,
			Args: args,
		}, nil
	case *ExprConditional:
		c, err := marshalExpr(t.Cond)
		if err != nil {
			return nil, err
		}
		th, err := marshalExpr(t.Then)
		if err != nil {
			return nil, err
		}
		el, err := marshalExpr(t.Else)
		if err != nil {
			return nil, err
		}
		return &exprJSON{
			Kind: "cond",
			Span: t.GetSpan(),
			Cond: c,
			Then: th,
			Else: el,
		}, nil
	case *ExprResult:
		return &exprJSON{
			Kind: "result",
			Span: t.GetSpan(),
		}, nil
	case *ExprOld:
		in, err := marshalExpr(t.Inner)
		if err != nil {
			return nil, err
		}
		return &exprJSON{
			Kind:  "old",
			Span:  t.GetSpan(),
			Inner: in,
		}, nil
	case *ExprQuantifier:
		body, err := marshalExpr(t.Body)
		if err != nil {
			return nil, err
		}
		return &exprJSON{
			Kind:           "quantifier",
			Span:           t.GetSpan(),
			QuantifierKind: t.Kind,
			Vars:           t.Vars,
			Body:           body,
		}, nil
	default:
		return nil, fmt.Errorf("json: unsupported Expr dynamic type %T", e)
	}
}

func unmarshalExpr(j *exprJSON) (Expr, error) {
	if j == nil {
		return nil, nil
	}
	switch j.Kind {
	case "literal":
		return &ExprLiteral{
			BaseExpr: BaseExpr{Span: j.Span},
			Kind:     j.LitKind,
			Bool:     j.Bool,
			Int:      j.Int,
			Real:     j.Real,
			Str:      j.Str,
		}, nil
	case "var":
		return &ExprVar{
			BaseExpr: BaseExpr{Span: j.Span},
			Name:     j.Name,
		}, nil
	case "unary":
		inner, err := unmarshalExpr(j.Expr)
		if err != nil {
			return nil, err
		}
		return &ExprUnary{
			BaseExpr: BaseExpr{Span: j.Span},
			Op:       j.UnaryOp,
			Expr:     inner,
		}, nil
	case "binary":
		l, err := unmarshalExpr(j.Left)
		if err != nil {
			return nil, err
		}
		r, err := unmarshalExpr(j.Right)
		if err != nil {
			return nil, err
		}
		return &ExprBinary{
			BaseExpr: BaseExpr{Span: j.Span},
			Op:       j.BinaryOp,
			Left:     l,
			Right:    r,
		}, nil
	case "call":
		args := make([]Expr, 0, len(j.Args))
		for _, aj := range j.Args {
			a, err := unmarshalExpr(aj)
			if err != nil {
				return nil, err
			}
			args = append(args, a)
		}
		return &ExprCall{
			BaseExpr: BaseExpr{Span: j.Span},
			Name:     j.Name,
			Args:     args,
		}, nil
	case "cond":
		c, err := unmarshalExpr(j.Cond)
		if err != nil {
			return nil, err
		}
		th, err := unmarshalExpr(j.Then)
		if err != nil {
			return nil, err
		}
		el, err := unmarshalExpr(j.Else)
		if err != nil {
			return nil, err
		}
		return &ExprConditional{
			BaseExpr: BaseExpr{Span: j.Span},
			Cond:     c,
			Then:     th,
			Else:     el,
		}, nil
	case "result":
		return &ExprResult{
			BaseExpr: BaseExpr{Span: j.Span},
		}, nil
	case "old":
		in, err := unmarshalExpr(j.Inner)
		if err != nil {
			return nil, err
		}
		return &ExprOld{
			BaseExpr: BaseExpr{Span: j.Span},
			Inner:    in,
		}, nil
	case "quantifier":
		body, err := unmarshalExpr(j.Body)
		if err != nil {
			return nil, err
		}
		return &ExprQuantifier{
			BaseExpr: BaseExpr{Span: j.Span},
			Kind:     j.QuantifierKind,
			Vars:     j.Vars,
			Body:     body,
		}, nil
	default:
		return nil, fmt.Errorf("json: unknown Expr kind %q", j.Kind)
	}
}

// -----------------------------
// Containers that include Expr
// -----------------------------

// NamedPredicate has Expr (interface) -> custom JSON.
func (p NamedPredicate) MarshalJSON() ([]byte, error) {
	ej, err := marshalExpr(p.Expr)
	if err != nil {
		return nil, err
	}
	type alias struct {
		Label string    `json:"label"`
		Expr  *exprJSON `json:"expr,omitempty"`
		Span  Span      `json:"span,omitempty"`
	}
	return json.Marshal(alias{
		Label: p.Label,
		Expr:  ej,
		Span:  p.Span,
	})
}

func (p *NamedPredicate) UnmarshalJSON(data []byte) error {
	type alias struct {
		Label string    `json:"label"`
		Expr  *exprJSON `json:"expr,omitempty"`
		Span  Span      `json:"span,omitempty"`
	}
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	e, err := unmarshalExpr(a.Expr)
	if err != nil {
		return err
	}
	p.Label = a.Label
	p.Expr = e
	p.Span = a.Span
	return nil
}

// LogicDecl has Body Expr (interface) -> custom JSON.
func (d LogicDecl) MarshalJSON() ([]byte, error) {
	body, err := marshalExpr(d.Body)
	if err != nil {
		return nil, err
	}
	type alias struct {
		Name       string        `json:"name"`
		Kind       LogicDeclKind `json:"kind"`
		Params     []LogicParam  `json:"params,omitempty"`
		ResultType TypeExpr      `json:"resultType"`
		Body       *exprJSON     `json:"body,omitempty"`
		Span       Span          `json:"span,omitempty"`
	}
	return json.Marshal(alias{
		Name:       d.Name,
		Kind:       d.Kind,
		Params:     d.Params,
		ResultType: d.ResultType,
		Body:       body,
		Span:       d.Span,
	})
}

func (d *LogicDecl) UnmarshalJSON(data []byte) error {
	type alias struct {
		Name       string        `json:"name"`
		Kind       LogicDeclKind `json:"kind"`
		Params     []LogicParam  `json:"params,omitempty"`
		ResultType TypeExpr      `json:"resultType"`
		Body       *exprJSON     `json:"body,omitempty"`
		Span       Span          `json:"span,omitempty"`
	}
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	body, err := unmarshalExpr(a.Body)
	if err != nil {
		return err
	}
	d.Name = a.Name
	d.Kind = a.Kind
	d.Params = a.Params
	d.ResultType = a.ResultType
	d.Body = body
	d.Span = a.Span
	return nil
}

// FuncContract includes Expr through Requires/Ensures (NamedPredicate already handled).
func (c FuncContract) MarshalJSON() ([]byte, error) {
	type alias struct {
		FuncName     string           `json:"funcName"`
		RecvTypeName string           `json:"recvTypeName,omitempty"`
		Span         Span             `json:"span,omitempty"`
		Requires     []NamedPredicate `json:"requires,omitempty"`
		Ensures      []NamedPredicate `json:"ensures,omitempty"`
		Assigns      LocationSet      `json:"assigns"`
	}
	return json.Marshal(alias{
		FuncName:     c.FuncName,
		RecvTypeName: c.RecvTypeName,
		Span:         c.Span,
		Requires:     c.Requires,
		Ensures:      c.Ensures,
		Assigns:      c.Assigns,
	})
}

func (c *FuncContract) UnmarshalJSON(data []byte) error {
	type alias struct {
		FuncName     string           `json:"funcName"`
		RecvTypeName string           `json:"recvTypeName,omitempty"`
		Span         Span             `json:"span,omitempty"`
		Requires     []NamedPredicate `json:"requires,omitempty"`
		Ensures      []NamedPredicate `json:"ensures,omitempty"`
		Assigns      LocationSet      `json:"assigns"`
	}
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	c.FuncName = a.FuncName
	c.RecvTypeName = a.RecvTypeName
	c.Span = a.Span
	c.Requires = a.Requires
	c.Ensures = a.Ensures
	c.Assigns = a.Assigns
	return nil
}

// LoopSpec includes Variant Expr.
func (s LoopSpec) MarshalJSON() ([]byte, error) {
	v, err := marshalExpr(s.Variant)
	if err != nil {
		return nil, err
	}
	type alias struct {
		Span           Span             `json:"span,omitempty"`
		ParentFuncName string           `json:"parentFuncName,omitempty"`
		Invariants     []NamedPredicate `json:"invariants,omitempty"`
		Assigns        LocationSet      `json:"assigns"`
		Variant        *exprJSON        `json:"variant,omitempty"`
	}
	return json.Marshal(alias{
		Span:           s.Span,
		ParentFuncName: s.ParentFuncName,
		Invariants:     s.Invariants,
		Assigns:        s.Assigns,
		Variant:        v,
	})
}

func (s *LoopSpec) UnmarshalJSON(data []byte) error {
	type alias struct {
		Span           Span             `json:"span,omitempty"`
		ParentFuncName string           `json:"parentFuncName,omitempty"`
		Invariants     []NamedPredicate `json:"invariants,omitempty"`
		Assigns        LocationSet      `json:"assigns"`
		Variant        *exprJSON        `json:"variant,omitempty"`
	}
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	v, err := unmarshalExpr(a.Variant)
	if err != nil {
		return err
	}
	s.Span = a.Span
	s.ParentFuncName = a.ParentFuncName
	s.Invariants = a.Invariants
	s.Assigns = a.Assigns
	s.Variant = v
	return nil
}
