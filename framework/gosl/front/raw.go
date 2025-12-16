package front

import specir "gpma/framework/gosl/ir"

// Raw text block for parsing GOSL from it.
// This block is enreached with hints for more simple future bindig FunctionContracts and LoopSpec to go ast Func and Loop nodes
type RawSpecBlock struct {
	Text  string
	Start specir.Position
	End   specir.Position
}
