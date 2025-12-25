package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"gpma/framework"
	_ "gpma/framework/plugin/dynamic_checks"
	_ "gpma/framework/plugin/examples"
)

func main() {
	app := &cli.App{
		Name:  "framework",
		Usage: "Process Go files with specification plugins",
		Commands: []*cli.Command{
			{
				Name:      "file",
				Usage:     "Process a single Go file",
				ArgsUsage: "<source-file> <output-file>",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "plugin",
						Aliases: []string{"p"},
						Usage:   "Plugin names to use (can be specified multiple times)",
					},
				},
				Action: processFile,
			},
			{
				Name:      "project",
				Usage:     "Process an entire Go project",
				ArgsUsage: "<project-dir> <output-dir>",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "plugin",
						Aliases: []string{"p"},
						Usage:   "Plugin names to use (can be specified multiple times)",
					},
				},
				Action: processProjectCmd,
			},
			{
				Name:  "plugins",
				Usage: "List available plugins",
				Action: func(c *cli.Context) error {
					fmt.Println("Available plugins:")
					for _, name := range framework.New().ListAvailablePlugins() {
						fmt.Printf("  - %s\n", name)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func processFile(c *cli.Context) error {
	if c.NArg() < 2 {
		return cli.Exit("Usage: framework file <source-file> <output-file> --plugin <plugin-name> [--plugin <plugin-name>...]", 1)
	}

	sourceFile := c.Args().Get(0)
	outputFile := c.Args().Get(1)
	pluginNames := c.StringSlice("plugin")

	if len(pluginNames) == 0 {
		return cli.Exit("Error: at least one plugin must be specified with --plugin or -p\nNote: Flags must come before positional arguments (e.g., 'framework file --plugin <name> <source> <output>')", 1)
	}

	fw := framework.New()
	fw, err := fw.AddPluginsByNames(pluginNames...)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to add plugins: %v", err), 1)
	}

	result, err := fw.ProcessFile(sourceFile)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Framework error: %v", err), 1)
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
			return cli.Exit(fmt.Sprintf("Failed to create output file: %v", err), 1)
		}
		defer out.Close()
		fmt.Fprintln(out, result.GeneratedCode)
		fmt.Printf("Output written to: %s\n", outputFile)
	} else {
		content, err := os.ReadFile(sourceFile)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Failed to read source file: %v", err), 1)
		}
		if err := os.WriteFile(outputFile, content, 0644); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to create output file: %v", err), 1)
		}
		fmt.Println("No modifications were made, output file created with original content")
		fmt.Printf("Output written to: %s\n", outputFile)
	}

	return nil
}

func processProjectCmd(c *cli.Context) error {
	if c.NArg() < 2 {
		return cli.Exit("Usage: framework project <project-dir> <output-dir> --plugin <plugin-name> [--plugin <plugin-name>...]", 1)
	}

	projectDir := c.Args().Get(0)
	outputDir := c.Args().Get(1)
	pluginNames := c.StringSlice("plugin")

	if len(pluginNames) == 0 {
		return cli.Exit("Error: at least one plugin must be specified with --plugin or -p\nNote: Flags must come before positional arguments (e.g., 'framework project --plugin <name> <project-dir> <output-dir>')", 1)
	}

	return processProject(projectDir, outputDir, pluginNames)
}

func processProject(projectDir, outputDir string, pluginNames []string) error {
	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get absolute path for project directory: %v", err), 1)
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get absolute path for output directory: %v", err), 1)
	}

	if err := os.MkdirAll(absOutputDir, 0755); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create output directory: %v", err), 1)
	}

	fw := framework.New()
	fw, err = fw.AddPluginsByNames(pluginNames...)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to add plugins: %v", err), 1)
	}

	overlay := make(map[string]string)

	var processedFiles int
	var errorCount int

	err = filepath.Walk(absProjectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", path, err)
			return nil
		}

		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" || name == ".idea" || name == ".vscode" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		relPath, err := filepath.Rel(absProjectDir, path)
		if err != nil {
			fmt.Printf("Error calculating relative path for %q: %v\n", path, err)
			return nil
		}

		outputPath := filepath.Join(absOutputDir, relPath)
		outputDirPath := filepath.Dir(outputPath)

		if err := os.MkdirAll(outputDirPath, 0755); err != nil {
			fmt.Printf("Error creating output directory %q: %v\n", outputDirPath, err)
			return nil
		}

		fmt.Printf("Processing: %s\n", relPath)
		result, err := fw.ProcessFile(path)
		if err != nil {
			fmt.Printf("Error processing %q: %v\n", path, err)
			errorCount++
			return nil
		}

		if result.Modified {
			if err := os.WriteFile(outputPath, []byte(result.GeneratedCode), 0644); err != nil {
				fmt.Printf("Error writing output file %q: %v\n", outputPath, err)
				errorCount++
				return nil
			}

			overlay[path] = outputPath
			processedFiles++
		} else {
			content, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading original file %q: %v\n", path, err)
				errorCount++
				return nil
			}
			if err := os.WriteFile(outputPath, content, 0644); err != nil {
				fmt.Printf("Error copying file %q: %v\n", outputPath, err)
				errorCount++
				return nil
			}
		}

		if len(result.Errors) > 0 {
			fmt.Printf("  Plugin errors for %s:\n", relPath)
			for _, err := range result.Errors {
				fmt.Printf("    - %v\n", err)
			}
		}

		return nil
	})

	if err != nil {
		return cli.Exit(fmt.Sprintf("Error walking project directory: %v", err), 1)
	}

	overlayJSON := map[string]map[string]string{
		"Replace": overlay,
	}

	overlayPath := filepath.Join(absOutputDir, "gmpa_overlay.json")
	overlayData, err := json.MarshalIndent(overlayJSON, "", "  ")
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to marshal overlay JSON: %v", err), 1)
	}

	if err := os.WriteFile(overlayPath, overlayData, 0644); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to write overlay file: %v", err), 1)
	}

	fmt.Printf("\nProcessing complete!\n")
	fmt.Printf("  Processed files: %d\n", processedFiles)
	fmt.Printf("  Errors: %d\n", errorCount)
	fmt.Printf("  Output directory: %s\n", absOutputDir)
	fmt.Printf("  Overlay file: %s\n", overlayPath)

	return nil
}
