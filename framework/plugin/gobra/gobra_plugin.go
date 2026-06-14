package gobra

import (
	"context"
	"fmt"
	"gpma/framework/data"
	"gpma/framework/gobra_runner"
	"gpma/framework/plugin"
	"os"
	"strings"
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

	// Create verification request with all needed data
	req := gobra_runner.VerificationRequest{
		GeneratedCode: generated.GobraCode,
		Mapping:       generated.HybridMapping,
	}

	// Run Gobra verification
	verifResult, err := p.runner.Verify(context.Background(), req)
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
			result.WithError(fmt.Errorf("verification timed out"))
		} else if len(verifResult.Errors) > 0 {
			var errorMsgs []string
			for _, gobraError := range verifResult.Errors {
				var errMsg string
				if gobraError.Position.File != "" {
					// Convert offset to line:column for display
					line, col, _ := gobra_runner.OffsetToLineCol([]byte(astData.Source), gobraError.Position.Offset)
					errMsg = fmt.Sprintf("%s:%d:%d: %s", gobraError.Position.File, line, col, gobraError.Message)
				} else {
					// No position information available
					errMsg = fmt.Sprintf("%s: %s", sourceFile, gobraError.Message)
				}
				errorMsgs = append(errorMsgs, errMsg)
			}
			result.WithError(fmt.Errorf("verification failed: %s", strings.Join(errorMsgs, "\n")))
		}
	}

	return result
}
