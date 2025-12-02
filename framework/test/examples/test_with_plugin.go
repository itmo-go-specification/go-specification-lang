//go:build exclude

package main

import (
	"fmt"
	"log"

	"gpma/framework/codegen"
	"gpma/framework/parser"
	"gpma/framework/plugin"
	"gpma/framework/plugin/examples"
)

func main() {
	p := parser.NewParser()
	astData, err := p.ParseFile("test/sources/hello_world.go")
	if err != nil {
		log.Fatalf("Failed to parse file: %v", err)
	}

	plugins := []plugin.Plugin{
		examples.NewAddCommentPlugin(),
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
		fmt.Println("Generating modified code...")
		gen := codegen.NewGenerator(astData)
		code, err := gen.GenerateToString()
		if err != nil {
			log.Fatalf("Failed to generate code: %v", err)
		}
		fmt.Println(code)
	} else {
		fmt.Println("No modifications were made by plugins")
	}
}
