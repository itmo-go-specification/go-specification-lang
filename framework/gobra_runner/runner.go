package gobra_runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Runner struct {
	settings Settings
}

func New(settings Settings) *Runner {
	return &Runner{settings: settings}
}

// RunWithMapping runs Gobra verification with hybrid mapping support
func (r *Runner) RunWithMapping(gobraCode string, mapping *HybridMapping) (*VerificationResult, error) {
	tmpFile, err := writeTempFile(gobraCode)
	if err != nil {
		return nil, fmt.Errorf("gobra_runner: write temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	var ctx context.Context
	var cancel context.CancelFunc
	if r.settings.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), r.settings.Timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	cmd, err := r.buildCommand(ctx, tmpFile)
	if err != nil {
		return nil, fmt.Errorf("gobra_runner: build command: %w", err)
	}

	rawOut, runErr := cmd.CombinedOutput()
	timedOut := ctx.Err() != nil

	output := string(rawOut)
	rawErrors := parseGobraOutput(output, tmpFile)
	mappedErrors := mapErrorsWithHybrid(rawErrors, mapping, gobraCode)

	exitOK := runErr == nil
	hasVerifErrors := len(mappedErrors) > 0

	if !exitOK && !hasVerifErrors && !timedOut {
		return nil, fmt.Errorf("gobra_runner: gobra exited with error: %w\noutput:\n%s", runErr, output)
	}

	success := exitOK && !hasVerifErrors && !timedOut

	return &VerificationResult{
		Success:   success,
		TimedOut:  timedOut,
		Errors:    mappedErrors,
		RawOutput: output,
	}, nil
}

// Run runs Gobra verification without mapping (legacy method for backward compatibility)
func (r *Runner) Run(gobraCode string) (*VerificationResult, error) {
	return r.RunWithMapping(gobraCode, nil)
}

func (r *Runner) buildCommand(ctx context.Context, inputFile string) (*exec.Cmd, error) {
	s := r.settings

	// Handle Docker backend
	if s.ExecutionBackend == BackendDocker {
		return r.buildDockerCommand(ctx, inputFile)
	}

	// Handle native backend
	return r.buildNativeCommand(ctx, inputFile)
}

// validateDockerConfig performs basic validation of Docker configuration
func (r *Runner) validateDockerConfig() error {
	s := r.settings
	if s.ExecutionBackend != BackendDocker {
		return nil
	}

	if s.DockerImage == "" {
		return fmt.Errorf("DockerImage must be set when using Docker backend")
	}

	// Basic validation of Docker image name format
	if strings.Contains(s.DockerImage, " ") {
		return fmt.Errorf("DockerImage '%s' contains spaces, which is not valid", s.DockerImage)
	}

	return nil
}

// buildGobraArgs builds Gobra command line arguments
func (r *Runner) buildGobraArgs(inputFile string) []string {
	// s := r.settings
	var args []string

	args = append(args, "-i", inputFile)

	// if s.GobraBackend != "" {
	// 	args = append(args, "--backend", s.GobraBackend)
	// }
	// if s.Z3Exe != "" {
	// 	args = append(args, "--z3Exe", s.Z3Exe)
	// }
	// if s.LogLevel != "" {
	// 	args = append(args, "--logLevel", s.LogLevel)
	// }
	// if s.NoInformation {
	// 	args = append(args, "--noInformation")
	// }
	// if s.ParallelizeBranches {
	// 	args = append(args, "--parallelizeBranches")
	// }
	// if s.AssumeInjectivityOnInhale {
	// 	args = append(args, "--assumeInjectivityOnInhale")
	// }
	// if s.CheckOverflow {
	// 	args = append(args, "--overflow")
	// }
	// args = append(args, s.ExtraFlags...)

	return args
}

// buildDockerCommand builds a Docker command for running Gobra
func (r *Runner) buildDockerCommand(ctx context.Context, inputFile string) (*exec.Cmd, error) {
	s := r.settings

	// Validate Docker configuration before building command
	if err := r.validateDockerConfig(); err != nil {
		return nil, err
	}

	// Get absolute path of input file
	absInputFile, err := filepath.Abs(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for input file: %w", err)
	}

	inputFileName := filepath.Base(absInputFile)

	containerPath := "/input/" + inputFileName

	args := []string{
		"run", "--rm",
		"-v", absInputFile + ":" + containerPath + ":ro",
		s.DockerImage,
	}

	// Add Gobra arguments (reuse the same logic)
	// Use relative path since we're setting working directory to /workspace
	gobraArgs := r.buildGobraArgs(containerPath)
	args = append(args, gobraArgs...)

	fmt.Printf("[DEBUG] Running %v\n", args)

	cmd := exec.CommandContext(ctx, "docker", args...)
	return cmd, nil
}

// buildNativeCommand builds a native command for running Gobra
func (r *Runner) buildNativeCommand(ctx context.Context, inputFile string) (*exec.Cmd, error) {
	s := r.settings

	if s.GobraBin != "" {
		args := r.buildGobraArgs(inputFile)
		cmd := exec.CommandContext(ctx, s.GobraBin, args...)
		return cmd, nil
	}

	if s.GobraJAR == "" {
		return nil, fmt.Errorf("neither GobraBin nor GobraJAR is set in Settings")
	}

	javaArgs := []string{"-jar", s.GobraJAR}
	javaArgs = append(javaArgs, r.buildGobraArgs(inputFile)...)
	cmd := exec.CommandContext(ctx, s.JavaPath, javaArgs...)
	return cmd, nil
}

func writeTempFile(content string) (string, error) {
	// Create temp file in current directory for Docker compatibility
	// macOS temp directories don't mount properly into Docker
	tmpDir := ".gobra_temp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", err
	}

	f, err := os.CreateTemp(tmpDir, "gobra_input_*.gobra")
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

// mapErrorsWithHybrid maps errors using the hybrid mapping system
func mapErrorsWithHybrid(raw []rawError, mapping *HybridMapping, genCode string) []VerificationError {
	result := make([]VerificationError, 0, len(raw))
	genContent := []byte(genCode)

	// Read original source file content for line/column conversion
	var sourceContent []byte
	if mapping != nil && mapping.SourceFile != "" {
		sourceContent, _ = os.ReadFile(mapping.SourceFile)
	}

	for _, re := range raw {
		ve := VerificationError{
			GenLine:   re.genLine,
			GenColumn: re.genColumn,
			GenFile:   re.genFile,
			Message:   re.message,
		}

		// Try to map using hybrid mapping
		if mapping != nil {
			// Convert line:col to offset for internal processing
			pos, err := mapping.MapErrorLineCol(re.genLine, re.genColumn, genContent)
			if err == nil {
				ve.File = pos.File
				ve.Offset = pos.Offset

				// Convert offset back to line:col for final error reporting
				// IMPORTANT: Use source content, not generated content!
				if len(sourceContent) > 0 {
					ve.Line, ve.Column = pos.ToLineCol(sourceContent)
				} else {
					// Fallback: if we can't read source file, use generated content
					// This won't be accurate but prevents crashes
					ve.Line, ve.Column = pos.ToLineCol(genContent)
				}
			} else {
				// Mapping failed - add metadata to help with debugging
				ve.MappingInfo = &ErrorMappingInfo{
					MappingType: MappingTypeUnmapped,
					Confidence:  0.0,
				}
			}
		}

		// Infer error kind from message
		if ve.Kind == "" {
			ve.Kind = inferKind(re.message)
		}

		result = append(result, ve)
	}

	return result
}
