package examples

import (
	"go/ast"

	"gpma/framework/data"
	"gpma/framework/plugin"
)

func init() {
	plugin.Register("example", func() plugin.Plugin {
		return NewExamplePlugin()
	})
}

type ExamplePlugin struct{}

func NewExamplePlugin() *ExamplePlugin {
	return &ExamplePlugin{}
}

func (p *ExamplePlugin) Name() string {
	return "ExamplePlugin"
}

func (p *ExamplePlugin) Process(astData *data.ASTData) *plugin.PluginResult {
	result := plugin.NewPluginResult()

	ast.Inspect(astData.File, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			info := astData.GetNodeInfo(funcDecl)
			info.Data["processed"] = true
			result.WithMapModified()
		}
		return true
	})

	return result
}
