package gobra

import (
	"context"
	"fmt"
	"gpma/framework/data"
	"gpma/framework/errors"
	"gpma/framework/gobra_runner"
	"gpma/framework/plugin"
	"os"
	"strings"
	"time"
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
	return "gobra"
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
	if val, ok := params["gobra_bin"]; ok {
		if s, ok := val.(string); ok {
			settings.GobraBin = s
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
	if val, ok := params["timeout"]; ok {
		if s, ok := val.(string); ok {
			if duration, err := time.ParseDuration(s); err == nil {
				settings.Timeout = duration
			}
		}
	}
	if val, ok := params["log_level"]; ok {
		if s, ok := val.(string); ok {
			settings.LogLevel = s
		}
	}
	if val, ok := params["no_information"]; ok {
		if b, ok := val.(bool); ok {
			settings.NoInformation = b
		}
	}
	if val, ok := params["parallelize_branches"]; ok {
		if b, ok := val.(bool); ok {
			settings.ParallelizeBranches = b
		}
	}
	if val, ok := params["assume_injectivity_on_inhale"]; ok {
		if b, ok := val.(bool); ok {
			settings.AssumeInjectivityOnInhale = b
		}
	}
	if val, ok := params["check_overflow"]; ok {
		if b, ok := val.(bool); ok {
			settings.CheckOverflow = b
		}
	}
	if val, ok := params["extra_flags"]; ok {
		if flags, ok := val.([]string); ok {
			settings.ExtraFlags = flags
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
	generated, err := p.GenerateGobraCode(astData, sourceFile)
	if err != nil {
		return result.WithFatalError(fmt.Errorf("failed to generate gobra code: %w", err))
	}

	// Run verification
	verifResult, err := p.RunVerification(generated)
	if err != nil {
		return result.WithFatalError(fmt.Errorf("verification failed: %w", err))
	}

	// Convert verification errors to unified format
	verificationErrors := p.convertVerificationErrors(verifResult, astData, sourceFile)

	return result.WithErrors(verificationErrors)
}

// generateGobraCode generates Gobra code from AST
func (p *GobraPlugin) GenerateGobraCode(astData *data.ASTData, sourceFile string) (*GobraGenerationResult, error) {
	generated, err := GenerateGobra(astData, sourceFile)
	if err != nil {
		return nil, err
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

	return generated, nil
}

// runVerification executes Gobra verification
func (p *GobraPlugin) RunVerification(generated *GobraGenerationResult) (*gobra_runner.VerificationResult, error) {
	req := gobra_runner.VerificationRequest{
		GeneratedCode: generated.GobraCode,
		Mapping:       generated.HybridMapping,
	}

	verifResult, err := p.runner.Verify(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// Debug: show Gobra analyzer output
	if p.debug {
		fmt.Printf("[DEBUG] Gobra analyzer output:\n%s\n", verifResult.RawOutput)
	}

	return verifResult, nil
}

// convertVerificationErrors converts Gobra errors to unified format
func (p *GobraPlugin) convertVerificationErrors(
	verifResult *gobra_runner.VerificationResult,
	astData *data.ASTData,
	sourceFile string,
) []*errors.Error {
	if verifResult.Success || len(verifResult.Errors) == 0 {
		return nil
	}

	if verifResult.TimedOut {
		return []*errors.Error{
			{
				Position: errors.Position{
					File:   sourceFile,
					Line:   1,
					Column: 1,
				},
				Message:  "verification timed out",
				Severity: errors.SeverityCritical,
				Source:   p.Name(),
			},
		}
	}

	unifiedErrors := make([]*errors.Error, 0, len(verifResult.Errors))
	for i := range verifResult.Errors {
		unifiedError := p.convertSingleError(&verifResult.Errors[i], astData, sourceFile)
		unifiedErrors = append(unifiedErrors, unifiedError)
	}

	return unifiedErrors
}

// convertSingleError converts a single Gobra error to unified format
func (p *GobraPlugin) convertSingleError(
	gobraError *gobra_runner.VerificationError,
	astData *data.ASTData,
	sourceFile string,
) *errors.Error {
	var position errors.Position

	if gobraError.Position.File != "" {
		line, col, _ := gobra_runner.OffsetToLineCol([]byte(astData.Source), gobraError.Position.Offset)
		position = errors.Position{
			File:   gobraError.Position.File,
			Line:   line,
			Column: col,
		}
	} else {
		position = errors.Position{
			File:   sourceFile,
			Line:   1,
			Column: 1,
		}
	}

	severity := errors.SeverityCritical
	if strings.Contains(strings.ToLower(gobraError.Message), "warning") {
		severity = errors.SeverityWarning
	}

	return &errors.Error{
		Position: position,
		Message:  gobraError.Message,
		Severity: severity,
		Source:   p.Name(),
	}
}
