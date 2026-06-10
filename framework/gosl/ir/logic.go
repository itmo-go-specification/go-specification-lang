package specir

// LogicDeclKind - logic function, predicate, axiom
// Currently we will be using only LogicFunction. Ex: Count(...)
type LogicDeclKind int

const (
	LogicFunction LogicDeclKind = iota
	LogicPredicate
	LogicAxiom
	LogicLemma
)

// Clause kind for describing lemmas annotation
type ClauseKind int

const (
	ClauseDefault ClauseKind = iota
	ClauseCheck
	ClauseAdmit
)

// LogicDecl - describe logic function, predicate, axiom
type LogicDecl struct {
	Name   string
	Kind   LogicDeclKind
	Params []LogicParam

	// Result type for functions: return type: TypeExprInteger, TypeExprBoolean, ...
	// Resulr for pred: can be Boolean/Prop
	ResultType TypeExpr

	// Body. Can be nil
	Body Expr

	// ClauseKind is meaningful for lemmas: default/check/admit.
	// for olf logic declarations it ramains ClauseDefault.
	ClauseKind ClauseKind

	Span Span
}

// LogicParam - parameter of logic function or predicate
type LogicParam struct {
	Name string
	Type TypeExpr
}

// AxiomaticBlock - "axiomatic Name { ... }" from ACSL.
type AxiomaticBlock struct {
	Name  string
	Span  Span
	Decls []*LogicDecl
}
