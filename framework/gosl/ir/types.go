package specir

// TypeKind - kind of logic type
type TypeKind int

const (
	TypeBoolean TypeKind = iota
	TypeInteger
	TypeReal
	TypeCustom // named logic type. Ex: rounding_mode, value_type
	// TODO: add TypeGoRef, TypeArray, TypeStruct ...
)

// TypeExpr - logic type
type TypeExpr struct {
	Kind TypeKind

	// For TypeCustom. Ex: Name == "rounding_mode"
	// For not TypeCustom. Ex: Name can be empty
	Name string
}
