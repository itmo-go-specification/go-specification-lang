package bind

import (
	"fmt"
	"go/ast"
	"gpma/framework/data"
	specir "gpma/framework/gosl/ir"
	"sort"
)

func Bind(astData *data.ASTData, spec *specir.SpecModule) (*data.ASTData, error) {
	var nodes []ast.Node

	astData.Spec = spec

	ast.Inspect(astData.File, func(n ast.Node) bool {
		if n != nil {
			nodes = append(nodes, n)
		}
		return true
	})

	sort.Slice(nodes, func(i, j int) bool {
		iPosition := astData.FileSet.Position(nodes[i].Pos())
		jPosition := astData.FileSet.Position(nodes[j].Pos())
		if iPosition.Offset == jPosition.Offset {
			iEnd := astData.FileSet.Position(nodes[i].End())
			jEnd := astData.FileSet.Position(nodes[j].End())
			return iEnd.Offset > jEnd.Offset
		}
		return iPosition.Offset < jPosition.Offset
	})

	// match function contracts
	for _, specSection := range spec.FuncContracts {
		nearestNode, _ := sort.Find(len(nodes), func(i int) int {
			iPosition := astData.FileSet.Position(nodes[i].Pos())
			return specSection.EndPos.Offset - iPosition.Offset
		})

		if nearestNode == len(nodes) {
			return nil, fmt.Errorf("Not matched node for spec at %s:%d", specSection.StartPos.File, specSection.StartPos.Line)
		}

		if funcDecl, ok := nodes[nearestNode].(*ast.FuncDecl); ok {
			node_info := data.NewAdditionalInfo()
			node_info.FuncContract = specSection
			astData.NodeInfo[funcDecl] = node_info
		} else {
			return nil, fmt.Errorf("Matched node for spec at %s:%d is not a function", specSection.StartPos.File, specSection.StartPos.Line)
		}
	}

	// match loop specs
	for _, specSection := range spec.LoopSpecs {
		nearestNode, _ := sort.Find(len(nodes), func(i int) int {
			iPosition := astData.FileSet.Position(nodes[i].Pos())
			return specSection.Span.End - iPosition.Offset
		})

		if nearestNode == len(nodes) {
			return nil, fmt.Errorf("Not matched node for spec at %s:%d", specSection.StartPos.File, specSection.StartPos.Line)
		}

		if forStmt, ok := nodes[nearestNode].(*ast.ForStmt); ok {
			node_info := data.NewAdditionalInfo()
			node_info.LoopSpec = specSection
			astData.NodeInfo[forStmt] = node_info
		} else {
			return nil, fmt.Errorf("Matched node for spec at %s:%d is not a loop", specSection.StartPos.File, specSection.StartPos.Line)
		}
	}

	for _, assertSection := range spec.Asserts {
		nearestNode, _ := sort.Find(len(nodes), func(i int) int {
			iPosition := astData.FileSet.Position(nodes[i].Pos())
			return assertSection.Span.End - iPosition.Offset
		})

		if nearestNode == len(nodes) {
			return nil, fmt.Errorf("Not matched node for spec at %s:%d", assertSection.StartPos.File, assertSection.StartPos.Line)
		}

		stmt := nodes[nearestNode]
		_, exists := astData.NodeInfo[stmt]
		if !exists {
			astData.NodeInfo[stmt] = data.NewAdditionalInfo()
		}
		astData.NodeInfo[stmt].Asserts = append(astData.NodeInfo[stmt].Asserts, assertSection)
	}

	return astData, nil
}
