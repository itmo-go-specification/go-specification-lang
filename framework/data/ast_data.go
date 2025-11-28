package data

import (
	"go/ast"
	"go/token"
)

type AdditionalInfo struct {
	Data map[string]interface{}
}

func NewAdditionalInfo() *AdditionalInfo {
	return &AdditionalInfo{
		Data: make(map[string]interface{}),
	}
}

type ASTData struct {
	File    *ast.File
	FileSet *token.FileSet
	NodeInfo map[ast.Node]*AdditionalInfo
}

func NewASTData(file *ast.File, fset *token.FileSet) *ASTData {
	return &ASTData{
		File:     file,
		FileSet:  fset,
		NodeInfo: make(map[ast.Node]*AdditionalInfo),
	}
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