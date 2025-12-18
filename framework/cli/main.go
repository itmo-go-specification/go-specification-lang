package main

import (
	"fmt"
	"log"
	"os"

	"gpma/framework"
	_ "gpma/framework/plugin/dynamic_checks"
	_ "gpma/framework/plugin/examples"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: framework <source-file> <output-file> [plugin-names...]")
	}

	sourceFile := os.Args[1]
	outputFile := os.Args[2]
	pluginNames := os.Args[3:]

	if len(pluginNames) == 0 {
		fmt.Println("Available plugins:")
		for _, name := range framework.New().ListAvailablePlugins() {
			fmt.Printf("  - %s\n", name)
		}
		fmt.Println("\nUsage: framework <source-file> [plugin-names...]")
		return
	}

	fw := framework.New()
	fw, err := fw.AddPluginsByNames(pluginNames...)
	if err != nil {
		log.Fatalf("Failed to add plugins: %v", err)
	}

	result, err := fw.ProcessFile(sourceFile)
	if err != nil {
		log.Fatalf("Framework error: %v", err)
	}

	if len(result.Errors) > 0 {
		fmt.Println("Plugin errors:")
		for _, err := range result.Errors {
			fmt.Printf("  - %v\n", err)
		}
	}

	if result.Modified {
		fmt.Println("Generated code:")
		out, err := os.Create(outputFile)
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()
		fmt.Fprintln(out, result.GeneratedCode)
	} else {
		fmt.Println("No modifications were made")
	}
}
