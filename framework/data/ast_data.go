package data

import (
	"go/ast"
	"go/token"
	specir "gpma/framework/gosl/ir"
)

// For now gpma ir supports only function contracts and loop specs
type AdditionalInfo struct {
	Data         map[string]interface{}
	FuncContract *specir.FuncContract
	LoopSpec     *specir.LoopSpec
	Asserts      []*specir.AssertSpec
}

func NewAdditionalInfo() *AdditionalInfo {
	return &AdditionalInfo{
		Data: make(map[string]interface{}),
	}
}

type ASTData struct {
	Source   string
	File     *ast.File
	FileSet  *token.FileSet
	NodeInfo map[ast.Node]*AdditionalInfo // go ast -> Spec bindings
	Spec     *specir.SpecModule           // all gosl context
}

func NewASTData(file *ast.File, fset *token.FileSet, source string) *ASTData {
	return &ASTData{
		Source:   source,
		File:     file,
		FileSet:  fset,
		NodeInfo: make(map[ast.Node]*AdditionalInfo),
	}
}

func (d *ASTData) TryGetNodeInfo(node ast.Node) *AdditionalInfo {
	if info, exists := d.NodeInfo[node]; exists {
		return info
	}
	return nil
}

func (d *ASTData) GetNodeInfo(node ast.Node) *AdditionalInfo {
	if info, exists := d.NodeInfo[node]; exists {
		return info
	}
	info := NewAdditionalInfo()
	d.NodeInfo[node] = info
	return info
}

func (d *ASTData) SetNodeInfo(node ast.Node, info *AdditionalInfo) {
	d.NodeInfo[node] = info
}
