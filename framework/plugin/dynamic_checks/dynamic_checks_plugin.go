package dynamicchecks

import (
	"gpma/framework/data"
	"gpma/framework/plugin"
)

func init() {
	plugin.Register("dynamic_checks", func() plugin.Plugin {
		return NewDynamicChecksPlugin()
	})
}

type DynamicChecksPlugin struct{}

func NewDynamicChecksPlugin() *DynamicChecksPlugin {
	return &DynamicChecksPlugin{}
}

func (p *DynamicChecksPlugin) Name() string {
	return "DynamicChecksPlugin"
}

func (p *DynamicChecksPlugin) Process(astData *data.ASTData) *plugin.PluginResult {
	result := plugin.NewPluginResult()

	processFunctions(astData, result)
	processLoops(astData, result)

	return result
}
