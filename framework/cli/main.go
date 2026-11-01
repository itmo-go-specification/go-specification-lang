package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gpma/framework"
	"gpma/framework/errors"
	_ "gpma/framework/errors/formatter_factory" // Import to trigger formatter factory registration
	"gpma/framework/plugin"
	_ "gpma/framework/plugin/dynamic_checks"
	_ "gpma/framework/plugin/examples"
	_ "gpma/framework/plugin/gobra"

	"github.com/urfave/cli/v2"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.ErrorLevel)

	app := &cli.App{
		Name:  "gpma-cli",
		Usage: "Process Go files with specification plugins",
		Commands: []*cli.Command{
			{
				Name:      "file",
				Usage:     "Process a single Go file",
				ArgsUsage: "<source-file> [output-file]",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "plugin",
						Aliases: []string{"p"},
						Usage:   "Plugin names to use (can be specified multiple times)",
					},
					&cli.StringSliceFlag{
						Name:    "plugin-config",
						Aliases: []string{"c"},
						Usage:   "Plugin configuration in format 'plugin:key:value;key:value' (can be specified multiple times)",
					},
					&cli.StringFlag{
						Name:    "error-format",
						Aliases: []string{"ef"},
						Usage:   "Error output format: text, json, compact, vscode",
						Value:   "text",
					},
					&cli.StringFlag{
						Name:    "error-output",
						Aliases: []string{"eo"},
						Usage:   "Error output destination: stdout, stderr, or file path",
						Value:   "stdout",
					},
					&cli.BoolFlag{
						Name:  "error-append",
						Usage: "Append to error output file instead of truncating",
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
					&cli.StringSliceFlag{
						Name:    "plugin-config",
						Aliases: []string{"c"},
						Usage:   "Plugin configuration in format 'plugin:key:value;key:value' (can be specified multiple times)",
					},
					&cli.StringFlag{
						Name:    "error-format",
						Aliases: []string{"ef"},
						Usage:   "Error output format: text, json, compact, vscode",
						Value:   "text",
					},
					&cli.StringFlag{
						Name:    "error-output",
						Aliases: []string{"eo"},
						Usage:   "Error output destination: stdout, stderr, or file path",
						Value:   "stdout",
					},
					&cli.BoolFlag{
						Name:  "error-append",
						Usage: "Append to error output file instead of truncating",
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

// parseSimpleConfig парсит конфигурацию в формате "plugin:key:value;key:value"
// Можно использовать несколько раз для разных плагинов
func parseSimpleConfig(configStrings []string) (map[string]map[string]interface{}, error) {
	if len(configStrings) == 0 {
		return nil, nil
	}

	configs := make(map[string]map[string]interface{})

	for _, configStr := range configStrings {
		if configStr == "" {
			continue
		}

		parts := strings.SplitN(configStr, ":", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid config format: expected 'plugin:key:value;key:value', got '%s'", configStr)
		}

		pluginName := strings.TrimSpace(parts[0])
		if pluginName == "" {
			return nil, fmt.Errorf("plugin name cannot be empty in config '%s'", configStr)
		}

		if configs[pluginName] == nil {
			configs[pluginName] = make(map[string]interface{})
		}

		paramsStr := parts[1]
		paramPairs := strings.Split(paramsStr, ";")

		for _, pair := range paramPairs {
			if pair == "" {
				continue
			}

			keyValue := strings.SplitN(pair, ":", 2)
			if len(keyValue) < 2 {
				return nil, fmt.Errorf("invalid parameter format: expected 'key:value', got '%s'", pair)
			}

			key := strings.TrimSpace(keyValue[0])
			valueStr := strings.TrimSpace(keyValue[1])

			if key == "" {
				return nil, fmt.Errorf("parameter key cannot be empty")
			}

			value := convertStringToInterface(valueStr)
			configs[pluginName][key] = value
		}
	}

	return configs, nil
}

func convertStringToInterface(s string) interface{} {
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	if intVal, err := strconv.Atoi(s); err == nil {
		return intVal
	}

	if floatVal, err := strconv.ParseFloat(s, 64); err == nil {
		return floatVal
	}

	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"' || s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}

	return s
}

func processFile(c *cli.Context) error {
	if c.NArg() < 1 {
		return cli.Exit("Usage: framework file <source-file> [output-file] --plugin <plugin-name> [--plugin <plugin-name>...]", 1)
	}

	// Create error reporter first to validate format immediately
	errorFormat := c.String("error-format")
	errorOutput := c.String("error-output")
	errorAppend := c.Bool("error-append")

	var dest errors.OutputDestination
	var filePath string

	switch errorOutput {
	case "stdout":
		dest = errors.DestinationStdout
	case "stderr":
		dest = errors.DestinationStderr
	default:
		dest = errors.DestinationFile
		filePath = errorOutput
	}

	reporterConfig := errors.ReporterConfig{
		Format:            errors.OutputFormat(errorFormat),
		OutputDestination: dest,
		FilePath:          filePath,
		Append:            errorAppend,
	}

	reporter, err := errors.NewErrorReporter(reporterConfig)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create error reporter: %v", err), 1)
	}
	defer reporter.Close()

	sourceFile := c.Args().Get(0)
	var outputFile string
	if c.NArg() >= 2 {
		outputFile = c.Args().Get(1)
	}
	pluginNames := c.StringSlice("plugin")
	configStrings := c.StringSlice("plugin-config")

	if len(pluginNames) == 0 {
		return cli.Exit("Error: at least one plugin must be specified with --plugin or -p\nNote: Flags must come before positional arguments (e.g., 'framework file --plugin <name> <source> <output>')", 1)
	}

	// Parse plugin configuration
	pluginConfigs, err := parseSimpleConfig(configStrings)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to parse plugin config: %v", err), 1)
	}

	fw := framework.New()

	// Create a list to store configured plugins
	var configuredPlugins []plugin.Plugin

	for _, pluginName := range pluginNames {
		plugin, err := plugin.Get(pluginName)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Failed to get plugin %q: %v", pluginName, err), 1)
		}

		if pluginConfigs != nil {
			if config, exists := pluginConfigs[pluginName]; exists {
				if configurable, ok := plugin.(interface {
					Configure(map[string]interface{}) error
				}); ok {
					if err := configurable.Configure(config); err != nil {
						return cli.Exit(fmt.Sprintf("Failed to configure plugin %q: %v", pluginName, err), 1)
					}
				} else {
					fmt.Printf("Warning: plugin %q does not support configuration\n", pluginName)
				}
			}
		}

		configuredPlugins = append(configuredPlugins, plugin)
	}

	fw = fw.AddPlugins(configuredPlugins...)

	result, err := fw.ProcessFile(sourceFile)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Framework error: %v", err), 1)
	}

	// Report errors using the new ErrorReporter
	if len(result.Errors) > 0 {
		if err := reporter.ReportErrors(result.Errors); err != nil {
			fmt.Printf("Failed to report errors: %v\n", err)
		}
	}

	// Handle fatal errors
	if len(result.FatalErrors) > 0 {
		fmt.Println("Plugin errors:")
		for _, err := range result.FatalErrors {
			fmt.Printf("  - %v\n", err)
		}
	}

	if result.Modified {
		if outputFile != "" {
			fmt.Println("Generated code:")
			out, err := os.Create(outputFile)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to create output file: %v", err), 1)
			}
			defer out.Close()
			fmt.Fprintln(out, result.GeneratedCode)
			fmt.Printf("Output written to: %s\n", outputFile)
		} else {
			fmt.Println("Generated code:")
			fmt.Println(result.GeneratedCode)
		}
	} else {
		if outputFile != "" {
			fmt.Println("No modifications were made, output file not created")
		} else {
			fmt.Println("No modifications were made")
		}
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
	configStrings := c.StringSlice("plugin-config")

	if len(pluginNames) == 0 {
		return cli.Exit("Error: at least one plugin must be specified with --plugin or -p\nNote: Flags must come before positional arguments (e.g., 'framework project --plugin <name> <project-dir> <output-dir>')", 1)
	}

	pluginConfigs, err := parseSimpleConfig(configStrings)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to parse plugin config: %v", err), 1)
	}

	// Get error reporting configuration
	errorFormat := c.String("error-format")
	errorOutput := c.String("error-output")
	errorAppend := c.Bool("error-append")

	return processProject(projectDir, outputDir, pluginNames, pluginConfigs, errorFormat, errorOutput, errorAppend)
}

func processProject(projectDir, outputDir string, pluginNames []string, pluginConfigs map[string]map[string]interface{}, errorFormat, errorOutput string, errorAppend bool) error {
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

	var configuredPlugins []plugin.Plugin

	for _, pluginName := range pluginNames {
		plugin, err := plugin.Get(pluginName)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Failed to get plugin %q: %v", pluginName, err), 1)
		}

		if pluginConfigs != nil {
			if config, exists := pluginConfigs[pluginName]; exists {
				if configurable, ok := plugin.(interface {
					Configure(map[string]interface{}) error
				}); ok {
					if err := configurable.Configure(config); err != nil {
						return cli.Exit(fmt.Sprintf("Failed to configure plugin %q: %v", pluginName, err), 1)
					}
				} else {
					fmt.Printf("Warning: plugin %q does not support configuration\n", pluginName)
				}
			}
		}

		configuredPlugins = append(configuredPlugins, plugin)
	}

	fw = fw.AddPlugins(configuredPlugins...)

	// Create error reporter from CLI flags
	var dest errors.OutputDestination
	var filePath string

	switch errorOutput {
	case "stdout":
		dest = errors.DestinationStdout
	case "stderr":
		dest = errors.DestinationStderr
	default:
		dest = errors.DestinationFile
		filePath = errorOutput
	}

	reporterConfig := errors.ReporterConfig{
		Format:            errors.OutputFormat(errorFormat),
		OutputDestination: dest,
		FilePath:          filePath,
		Append:            errorAppend,
	}

	reporter, err := errors.NewErrorReporter(reporterConfig)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to create error reporter: %v", err), 1)
	}
	defer reporter.Close()

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
			if err := reporter.ReportErrors(result.Errors); err != nil {
				fmt.Printf("Failed to report errors for %s: %v\n", relPath, err)
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

	overlayPath := filepath.Join(absOutputDir, "gpma_overlay.json")
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
