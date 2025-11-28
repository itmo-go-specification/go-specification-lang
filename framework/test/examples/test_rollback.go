package main

import (
	"errors"
	"fmt"
	"log"

	"gpma/framework/codegen"
	"gpma/framework/data"
	"gpma/framework/parser"
	"gpma/framework/plugin"
	"gpma/framework/plugin/examples"
)

type FailingPlugin struct{}

func (p *FailingPlugin) Name() string {
	return "FailingPlugin"
}

func (p *FailingPlugin) Process(astData *data.ASTData) *plugin.PluginResult {
	result := plugin.NewPluginResult()
	result.WithASTModified()
	return result.WithError(errors.New("intentional error"))
}

func main() {
	p := parser.NewParser()
	astData, err := p.ParseFile("test/sources/hello_world.go")
	if err != nil {
		log.Fatalf("Failed to parse file: %v", err)
	}

	plugins := []plugin.Plugin{
		examples.NewAddCommentPlugin(),
		&FailingPlugin{},
		examples.NewExamplePlugin(),
	}

	runner := plugin.NewPluginRunner(plugins)
	result, err := runner.Run(astData)
	if err != nil {
		log.Fatalf("Plugin runner failed: %v", err)
	}

	if len(result.Errors) > 0 {
		fmt.Println("Plugin errors:")
		for _, err := range result.Errors {
			fmt.Printf("  - %v\n", err)
		}
	}

	if result.HasModifications {
		fmt.Println("\nGenerating modified code...")
		gen := codegen.NewGenerator(astData)
		code, err := gen.GenerateToString()
		if err != nil {
			log.Fatalf("Failed to generate code: %v", err)
		}
		fmt.Println(code)
	} else {
		fmt.Println("\nNo modifications were made by plugins")
	}
}