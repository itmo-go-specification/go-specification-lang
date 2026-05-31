package gobra

import (
	"fmt"
	"gpma/framework/data"
	"gpma/framework/plugin"
)

func init() {
	plugin.Register("gobra", func() plugin.Plugin {
		return NewGobraPlugin()
	})
}

type GobraPlugin struct {
	config GobraConfig
}

func NewGobraPlugin() *GobraPlugin {
	return &GobraPlugin{
		config: DefaultConfig(),
	}
}

func (p *GobraPlugin) Name() string {
	return "GobraPlugin"
}

func (p *GobraPlugin) Configure(params map[string]interface{}) error {
	return p.config.ApplyPluginParams(params)
}

func (p *GobraPlugin) Process(astData *data.ASTData) *plugin.PluginResult {
	result := plugin.NewPluginResult()

	failed_predicates, err := RunGobraVerifier(astData, p.config)
	if err != nil {
		result.WithError(err)
	}

	if len(failed_predicates) > 0 {
		fmt.Println("Verification failed")
		for _, predicate := range failed_predicates {
			line, col, err := OffsetToLineCol(astData.Source, predicate.Span.Start)
			if err != nil {
				result.WithError(fmt.Errorf("failed to get error coords: %v", err))
				continue
			}

			fmt.Printf("\tError: condition at %d:%d may fail\n", line, col)
			fmt.Printf("\t\t%s\n", PrintExpr(predicate.Expr))
		}
	} else {
		fmt.Println("Verification succeed")
	}

	return result
}
