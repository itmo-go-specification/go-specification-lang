package front

import (
	specir "gpma/framework/gosl/ir"
)

// BuildSpecModule - can use different parser to parse GOSL language for building SpecModule
func BuildSpecModule(blocks []RawSpecBlock) (*specir.SpecModule, error) {
	spec := &specir.SpecModule{
		LogicDecls: make(map[string]*specir.LogicDecl),
	}

	spec, err := buildSpecModuleACSL_ANTLR(spec, blocks)

	return spec, err
}
