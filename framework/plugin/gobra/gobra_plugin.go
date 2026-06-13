package gobra

import (
	"fmt"
	"gpma/framework/data"
	"gpma/framework/gobra_runner"
	"gpma/framework/plugin"
	"os"
)

func init() {
	plugin.Register("gobra", func() plugin.Plugin {
		return NewGobraPlugin()
	})
}

type GobraPlugin struct {
	runner *gobra_runner.Runner
	debug  bool
}

func NewGobraPlugin() *GobraPlugin {
	settings := gobra_runner.DefaultSettings()
	return &GobraPlugin{
		runner: gobra_runner.New(settings),
		debug:  false,
	}
}

func (p *GobraPlugin) Name() string {
	return "GobraPlugin"
}

func (p *GobraPlugin) Configure(params map[string]interface{}) error {
	settings := gobra_runner.DefaultSettings()

	// Apply plugin parameters to settings
	if val, ok := params["backend"]; ok {
		if s, ok := val.(string); ok {
			settings.ExecutionBackend = gobra_runner.Backend(s)
		}
	}
	if val, ok := params["java"]; ok {
		if s, ok := val.(string); ok {
			settings.JavaPath = s
		}
	}
	if val, ok := params["gobra_jar"]; ok {
		if s, ok := val.(string); ok {
			settings.GobraJAR = s
		}
	}
	if val, ok := params["z3_exe"]; ok {
		if s, ok := val.(string); ok {
			settings.Z3Exe = s
		}
	}
	if val, ok := params["docker_image"]; ok {
		if s, ok := val.(string); ok {
			settings.DockerImage = s
		}
	}
	if val, ok := params["debug"]; ok {
		if b, ok := val.(bool); ok {
			p.debug = b
		}
	}

	p.runner = gobra_runner.New(settings)
	return nil
}

func (p *GobraPlugin) Process(astData *data.ASTData) *plugin.PluginResult {
	result := plugin.NewPluginResult()

	// Generate Gobra code with hybrid mapping
	sourceFile := astData.FileSet.File(astData.File.Pos()).Name()
	generated, err := GenerateGobra(astData, sourceFile)
	if err != nil {
		result.WithError(fmt.Errorf("failed to generate gobra code: %w", err))
		fmt.Printf("Generation error: %v\n", err)
		return result
	}

	// Debug: dump generated Gobra code
	if p.debug {
		debugFile := "debug.gobra"
		if err := os.WriteFile(debugFile, []byte(generated.GobraCode), 0644); err != nil {
			fmt.Printf("[DEBUG] Warning: failed to write debug.gobra: %v\n", err)
		} else {
			fmt.Printf("[DEBUG] Generated Gobra code written to %s\n", debugFile)
		}
	}

	// Run Gobra verification with hybrid mapping
	verifResult, err := p.runner.RunWithMapping(generated.GobraCode, generated.HybridMapping)
	if err != nil {
		result.WithError(fmt.Errorf("verification failed: %w", err))
		return result
	}

	// Debug: show Gobra analyzer output
	if p.debug {
		fmt.Printf("[DEBUG] Gobra analyzer output:\n%s\n", verifResult.RawOutput)
	}

	// Process verification results
	if !verifResult.Success {
		if verifResult.TimedOut {
			fmt.Println("Verification timed out")
			result.WithError(fmt.Errorf("verification timed out"))
		} else if len(verifResult.Errors) > 0 {
			fmt.Println("Verification failed")
			for _, gobraError := range verifResult.Errors {
				if gobraError.IsMapped() {
					fmt.Printf("\tError at %s:%d:%d: %s\n",
						gobraError.File, gobraError.Line, gobraError.Column, gobraError.Message)
				} else {
					fmt.Printf("\tError in generated code at %d:%d: %s\n",
						gobraError.GenLine, gobraError.GenColumn, gobraError.Message)
				}
			}
		}
	} else {
		fmt.Println("Verification succeeded")
	}

	return result
}
