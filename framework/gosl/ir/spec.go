package specir

// SpecModule - specification for all GOSL annotations in one .go file
type SpecModule struct {
	LogicDecls    map[string]*LogicDecl // GOSL logic functions: Count(...)
	FuncContracts []*FuncContract       // function contracts: ...
	LoopSpecs     []*LoopSpec           // loop specification: ...
	// AxiomaticBlocks map[string]*AxiomaticBlock // axiomatic blocks. For now it is excessive for GOSL
}
