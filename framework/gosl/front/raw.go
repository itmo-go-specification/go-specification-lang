package front

import specir "gpma/framework/gosl/ir"

// Raw text block for parsing GOSL from it.
// This block is enreached with hints for more simple future bindig FunctionContracts and LoopSpec to go ast Func and Loop nodes
type RawSpecBlock struct {
	Text string
	Span specir.Span

	// For function contracts
	FuncName     string
	RecvTypeName string

	// For loop annotation
	ParentFuncName string
}
