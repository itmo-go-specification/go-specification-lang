package data

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
)

type ASTSnapshot struct {
	source   []byte
	nodeInfo map[ast.Node]*AdditionalInfo
}

func (d *ASTData) CreateSnapshot() (*ASTSnapshot, error) {
	var buf bytes.Buffer
	err := printer.Fprint(&buf, d.FileSet, d.File)
	if err != nil {
		return nil, err
	}

	nodeInfoCopy := make(map[ast.Node]*AdditionalInfo)
	for node, info := range d.NodeInfo {
		infoCopy := &AdditionalInfo{
			Data: make(map[string]interface{}),
		}
		for k, v := range info.Data {
			infoCopy.Data[k] = v
		}
		nodeInfoCopy[node] = infoCopy
	}

	return &ASTSnapshot{
		source:   buf.Bytes(),
		nodeInfo: nodeInfoCopy,
	}, nil
}

func (d *ASTData) RestoreFromSnapshot(snapshot *ASTSnapshot) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", snapshot.source, parser.ParseComments)
	if err != nil {
		return err
	}

	d.File = file
	d.FileSet = fset
	d.NodeInfo = make(map[ast.Node]*AdditionalInfo)
	for _, info := range snapshot.nodeInfo {
		infoCopy := &AdditionalInfo{
			Data: make(map[string]interface{}),
		}
		for k, v := range info.Data {
			infoCopy.Data[k] = v
		}
	}

	return nil
}
