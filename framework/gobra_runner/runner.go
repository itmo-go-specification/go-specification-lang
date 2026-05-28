package gobra_runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type Runner struct {
	settings Settings
}

func New(settings Settings) *Runner {
	return &Runner{settings: settings}
}

func (r *Runner) Run(gobraCode string, mapping *CoordMapping) (*VerificationResult, error) {
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
	mappedErrors := mapErrors(rawErrors, mapping)

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

func (r *Runner) buildCommand(ctx context.Context, inputFile string) (*exec.Cmd, error) {
	s := r.settings

	if s.GobraBin != "" {
		args := r.buildArgs(inputFile)
		cmd := exec.CommandContext(ctx, s.GobraBin, args...)
		return cmd, nil
	}

	if s.GobraJAR == "" {
		return nil, fmt.Errorf("neither GobraBin nor GobraJAR is set in Settings")
	}

	javaArgs := []string{"-jar", s.GobraJAR}
	javaArgs = append(javaArgs, r.buildArgs(inputFile)...)
	cmd := exec.CommandContext(ctx, s.JavaPath, javaArgs...)
	return cmd, nil
}

func (r *Runner) buildArgs(inputFile string) []string {
	s := r.settings
	var args []string

	args = append(args, "--input", inputFile)

	if s.Backend != "" {
		args = append(args, "--backend", s.Backend)
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

func writeTempFile(content string) (string, error) {
	f, err := os.CreateTemp("", "gobra_input_*.go")
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

func mapErrors(raw []rawError, mapping *CoordMapping) []VerificationError {
	result := make([]VerificationError, 0, len(raw))

	for _, re := range raw {
		ve := VerificationError{
			GenLine:   re.genLine,
			GenColumn: re.genColumn,
			Message:   re.message,
		}

		if mapping != nil {
			if entry, ok := mapping.Lookup(re.genLine, re.genColumn); ok {
				ve.File = entry.Orig.File
				ve.Line = entry.Orig.Line
				ve.Column = entry.Orig.Column
				ve.Kind = entry.Kind
			}
		}

		if ve.Kind == "" {
			ve.Kind = inferKind(re.message)
		}

		result = append(result, ve)
	}

	return result
}
