package gobra_runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Runner struct {
	settings Settings
}

func New(settings Settings) *Runner {
	return &Runner{settings: settings}
}

// Verify executes Gobra verification with the provided request
func (r *Runner) Verify(ctx context.Context, req VerificationRequest) (*VerificationResult, error) {
	tmpFile, err := writeTempFile(req.GeneratedCode)
	if err != nil {
		return nil, fmt.Errorf("gobra_runner: write temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Use provided context or create default one
	if ctx == nil {
		var cancel context.CancelFunc
		if r.settings.Timeout > 0 {
			ctx, cancel = context.WithTimeout(context.Background(), r.settings.Timeout)
		} else {
			ctx, cancel = context.WithCancel(context.Background())
		}
		defer cancel()
	}

	cmd, err := r.buildCommand(ctx, tmpFile)
	if err != nil {
		return nil, fmt.Errorf("gobra_runner: build command: %w", err)
	}

	rawOut, runErr := cmd.CombinedOutput()
	timedOut := ctx.Err() != nil

	output := string(rawOut)
	rawErrors := parseGobraOutput(output, tmpFile)

	// Use ErrorMapper for mapping errors
	mapper := NewErrorMapper(req.Mapping, req.GeneratedCode)
	mappedErrors := make([]VerificationError, 0, len(rawErrors))
	for _, re := range rawErrors {
		mappedErrors = append(mappedErrors, mapper.MapError(re))
	}

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

// VerifyWithoutMapping is a convenience method for simple cases without mapping
func (r *Runner) VerifyWithoutMapping(ctx context.Context, gobraCode string) (*VerificationResult, error) {
	return r.Verify(ctx, VerificationRequest{
		GeneratedCode: gobraCode,
		Mapping:       nil,
	})
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
func (r *Runner) buildGobraArgs(inputFile string, s Settings) []string {
	// s := r.settings
	var args []string

	args = append(args, "-i", inputFile)

	if s.GobraBackend != "" {
		args = append(args, "--backend", s.GobraBackend)
	}
	if s.Z3Exe != "" {
		args = append(args, "--z3Exe", s.Z3Exe)
	}
	if s.LogLevel != "" {
		args = append(args, "--logLevel", s.LogLevel)
	}
	if s.NoInformation {
		args = append(args, "--noInformation")
	}
	if s.ParallelizeBranches {
		args = append(args, "--parallelizeBranches")
	}
	if s.AssumeInjectivityOnInhale {
		args = append(args, "--assumeInjectivityOnInhale")
	}
	if s.CheckOverflow {
		args = append(args, "--overflow")
	}
	args = append(args, s.ExtraFlags...)

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

	// Generate a unique container name based on input file
	containerName := "gobra-" + strings.TrimSuffix(inputFileName, ".gobra") + "-" + fmt.Sprintf("%d", time.Now().UnixNano())

	args := []string{
		"run", "--rm", "--name", containerName,
		"-v", absInputFile + ":" + containerPath + ":ro",
		s.DockerImage,
	}

	// Add Gobra arguments (reuse the same logic)
	// Use relative path since we're setting working directory to /workspace
	gobraArgs := r.buildGobraArgs(containerPath, s)
	args = append(args, gobraArgs...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	
	// Start a goroutine to stop the container when context is cancelled
	go func() {
		<-ctx.Done()
		// Stop the container when context is cancelled
		stopCmd := exec.Command("docker", "stop", containerName)
		stopCmd.Run() // Ignore errors, container might not exist
	}()
	
	return cmd, nil
}

// buildNativeCommand builds a native command for running Gobra
func (r *Runner) buildNativeCommand(ctx context.Context, inputFile string) (*exec.Cmd, error) {
	s := r.settings

	if s.GobraBin != "" {
		args := r.buildGobraArgs(inputFile, s)
		cmd := exec.CommandContext(ctx, s.GobraBin, args...)
		return cmd, nil
	}

	if s.GobraJAR == "" {
		return nil, fmt.Errorf("neither GobraBin nor GobraJAR is set in Settings")
	}

	javaArgs := []string{"-jar", s.GobraJAR}
	javaArgs = append(javaArgs, r.buildGobraArgs(inputFile, s)...)
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
// Deprecated: Use ErrorMapper instead
func mapErrorsWithHybrid(raw []rawError, mapping *HybridMapping, genCode string) []VerificationError {
	// This function is kept for backward compatibility during transition
	// Use ErrorMapper for new code
	result := make([]VerificationError, 0, len(raw))
	genContent := []byte(genCode)

	for _, re := range raw {
		ve := VerificationError{
			Message: re.message,
		}

		// Try to map using hybrid mapping
		if mapping != nil {
			// Convert line:col to offset for internal processing
			pos, err := mapping.MapErrorLineCol(re.genLine, re.genColumn, genContent)
			if err == nil {
				ve.Position = pos
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
