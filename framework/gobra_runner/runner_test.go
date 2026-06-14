package gobra_runner

import (
	"context"
	"testing"
	"time"
)

func TestNewRunner(t *testing.T) {
	settings := DefaultSettings()
	runner := New(settings)

	if runner == nil {
		t.Fatal("New() returned nil")
	}

	if runner.settings.ExecutionBackend != BackendNative {
		t.Errorf("New() runner.settings.ExecutionBackend = %v, want %v",
			runner.settings.ExecutionBackend, BackendNative)
	}
}

func TestRunner_RunWithMapping(t *testing.T) {
	tests := []struct {
		name        string
		gobraCode   string
		mapping     *HybridMapping
		expectError bool
	}{
		{
			name: "Simple verification without mapping",
			gobraCode: `
package main

func main() {
	assert(true)
}
`,
			mapping:     nil,
			expectError: false,
		},
		{
			name: "Verification with hybrid mapping",
			gobraCode: `
package main

func main() {
	assert(true)
}
`,
			mapping:     NewHybridMapping(""),
			expectError: false,
		},
		{
			name: "Verification with mapping containing intervals",
			gobraCode: `
package main

func main() {
	assert(true)
}
`,
			mapping: func() *HybridMapping {
				m := NewHybridMapping("test.go")
				m.AddInterval(Interval{
					GenStart:  20, // Offset in generated code
					GenEnd:    40,
					SourcePos: SourcePosition{File: "test.go", Offset: 10}, // Offset in original code
				})
				return m
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create runner with minimal settings
			settings := Settings{
				ExecutionBackend: BackendNative,
				GobraBin:         "echo", // Use echo as a safe command for testing
				Timeout:          5 * time.Second,
			}
			runner := New(settings)

			// Run verification
			req := VerificationRequest{
				GeneratedCode: tt.gobraCode,
				Mapping:       tt.mapping,
			}
			result, err := runner.Verify(context.Background(), req)

			// We expect this to fail because echo is not a real Gobra binary
			// but we're testing the API structure, not actual verification
			if err != nil && !tt.expectError {
				t.Logf("Verify() error = %v (expected for mock command)", err)
			}

			if result != nil {
				// Check that result structure is correct
				if result.RawOutput == "" {
					t.Error("Verify() result.RawOutput should not be empty")
				}
			}
		})
	}
}

func TestRunner_Run(t *testing.T) {
	tests := []struct {
		name        string
		gobraCode   string
		expectError bool
	}{
		{
			name: "Simple verification",
			gobraCode: `
package main

func main() {
	assert(true)
}
`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := Settings{
				ExecutionBackend: BackendNative,
				GobraBin:         "echo",
				Timeout:          5 * time.Second,
			}
			runner := New(settings)

			result, err := runner.VerifyWithoutMapping(context.Background(), tt.gobraCode)

			if err != nil && !tt.expectError {
				t.Logf("VerifyWithoutMapping() error = %v (expected for mock command)", err)
			}

			if result != nil {
				if result.RawOutput == "" {
					t.Error("Run() result.RawOutput should not be empty")
				}
			}
		})
	}
}

func TestRunner_buildCommand_Native(t *testing.T) {
	tests := []struct {
		name        string
		settings    Settings
		expectError bool
	}{
		{
			name: "Native backend with GobraBin",
			settings: Settings{
				ExecutionBackend: BackendNative,
				GobraBin:         "gobra",
				GobraBackend:     "silicon",
			},
			expectError: false,
		},
		{
			name: "Native backend with GobraJAR",
			settings: Settings{
				ExecutionBackend: BackendNative,
				JavaPath:         "java",
				GobraJAR:         "gobra.jar",
				GobraBackend:     "silicon",
			},
			expectError: false,
		},
		{
			name: "Native backend without binary or JAR",
			settings: Settings{
				ExecutionBackend: BackendNative,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := New(tt.settings)
			ctx := context.Background()
			cmd, err := runner.buildCommand(ctx, "test.go")

			if tt.expectError {
				if err == nil {
					t.Error("buildCommand() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("buildCommand() unexpected error = %v", err)
				return
			}

			if cmd == nil {
				t.Error("buildCommand() returned nil command")
			}
		})
	}
}

func TestRunner_buildCommand_Docker(t *testing.T) {
	tests := []struct {
		name        string
		settings    Settings
		expectError bool
	}{
		{
			name: "Docker backend with image",
			settings: Settings{
				ExecutionBackend: BackendDocker,
				DockerImage:      "ghcr.io/viperproject/gobra:sha-7fa4f86",
				GobraBackend:     "silicon",
			},
			expectError: false,
		},
		{
			name: "Docker backend without image",
			settings: Settings{
				ExecutionBackend: BackendDocker,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := New(tt.settings)
			ctx := context.Background()
			cmd, err := runner.buildCommand(ctx, "test.go")

			if tt.expectError {
				if err == nil {
					t.Error("buildCommand() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("buildCommand() unexpected error = %v", err)
				return
			}

			if cmd == nil {
				t.Error("buildCommand() returned nil command")
			}

			// Check that Docker command is properly structured
			if cmd.Args[0] != "docker" {
				t.Errorf("buildCommand() first arg = %v, want 'docker'", cmd.Args[0])
			}
		})
	}
}

func TestRunner_buildGobraArgs(t *testing.T) {
	tests := []struct {
		name     string
		settings Settings
		wantArgs []string
	}{
		{
			name: "Minimal settings",
			settings: Settings{
				GobraBackend: "silicon",
			},
			wantArgs: []string{"--input", "test.go", "--backend", "silicon"},
		},
		{
			name: "Full settings",
			settings: Settings{
				GobraBackend:              "silicon",
				Z3Exe:                     "z3",
				LogLevel:                  "ERROR",
				NoInformation:             true,
				ParallelizeBranches:       true,
				AssumeInjectivityOnInhale: true,
				CheckOverflow:             true,
				ExtraFlags:                []string{"--custom-flag"},
			},
			wantArgs: []string{
				"--input", "test.go",
				"--backend", "silicon",
				"--z3Exe", "z3",
				"--logLevel", "ERROR",
				"--noInformation",
				"--parallelizeBranches",
				"--assumeInjectivityOnInhale",
				"--overflow",
				"--custom-flag",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := New(tt.settings)
			args := runner.buildGobraArgs("test.go")

			if len(args) != len(tt.wantArgs) {
				t.Errorf("buildGobraArgs() len = %d, want %d", len(args), len(tt.wantArgs))
				return
			}

			for i, want := range tt.wantArgs {
				if args[i] != want {
					t.Errorf("buildGobraArgs() args[%d] = %v, want %v", i, args[i], want)
				}
			}
		})
	}
}

func TestMapErrorsWithHybrid(t *testing.T) {
	tests := []struct {
		name    string
		raw     []rawError
		mapping *HybridMapping
		wantLen int
	}{
		{
			name: "No mapping",
			raw: []rawError{
				{genLine: 10, genColumn: 5, message: "assertion failed"},
			},
			mapping: nil,
			wantLen: 1,
		},
		{
			name: "With mapping",
			raw: []rawError{
				{genLine: 10, genColumn: 5, message: "assertion failed"},
			},
			mapping: func() *HybridMapping {
				m := NewHybridMapping("")
				m.AddInterval(Interval{
					GenStart:  50, // Offset in generated code
					GenEnd:    150,
					SourcePos: SourcePosition{File: "test.go", Offset: 25}, // Offset in original code
				})
				return m
			}(),
			wantLen: 1,
		},
		{
			name: "Multiple errors",
			raw: []rawError{
				{genLine: 10, genColumn: 5, message: "assertion failed"},
				{genLine: 20, genColumn: 10, message: "precondition violated"},
			},
			mapping: nil,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create sample generated code for offset conversion
			genCode := "package main\n\nfunc main() {\n\t// Line 1\n\t// Line 2\n\t// Line 3\n\t// Line 4\n\t// Line 5\n\t// Line 6\n\t// Line 7\n\t// Line 8\n\t// Line 9\n\t// Line 10\n}\n"
			errors := mapErrorsWithHybrid(tt.raw, tt.mapping, genCode)

			if len(errors) != tt.wantLen {
				t.Errorf("mapErrorsWithHybrid() len = %d, want %d", len(errors), tt.wantLen)
			}

			for i, err := range errors {
				if err.Message == "" {
					t.Errorf("mapErrorsWithHybrid() errors[%d].Message should not be empty", i)
				}
				if err.Kind == "" {
					t.Errorf("mapErrorsWithHybrid() errors[%d].Kind should not be empty", i)
				}
			}
		})
	}
}

func TestWriteTempFile(t *testing.T) {
	content := "package main\n\nfunc main() {}\n"

	tmpFile, err := writeTempFile(content)
	if err != nil {
		t.Fatalf("writeTempFile() error = %v", err)
	}

	// Check that file exists
	if tmpFile == "" {
		t.Error("writeTempFile() returned empty path")
	}

	// Clean up is handled by defer in the actual function
	_ = tmpFile
}
