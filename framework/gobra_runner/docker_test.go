package gobra_runner

import (
	"fmt"
	"testing"
	"time"
)

func TestNewDockerChecker(t *testing.T) {
	checker := NewDockerChecker(5 * time.Second)

	if checker == nil {
		t.Fatal("NewDockerChecker() returned nil")
	}

	if checker.timeout != 5*time.Second {
		t.Errorf("NewDockerChecker() timeout = %v, want %v", checker.timeout, 5*time.Second)
	}
}

func TestNewDockerChecker_DefaultTimeout(t *testing.T) {
	checker := NewDockerChecker(0)

	if checker.timeout != 10*time.Second {
		t.Errorf("NewDockerChecker(0) timeout = %v, want %v", checker.timeout, 10*time.Second)
	}
}

func TestDockerChecker_CheckDockerAvailability(t *testing.T) {
	checker := NewDockerChecker(10 * time.Second)

	err := checker.CheckDockerAvailability()

	// This test may fail if Docker is not installed or not running
	// We just check that the error message is informative if it fails
	if err != nil {
		t.Logf("Docker not available (this is expected if Docker is not installed): %v", err)
		// Check that error message is informative
		if err.Error() == "" {
			t.Error("CheckDockerAvailability() returned empty error message")
		}
	}
}

func TestDockerChecker_CheckDockerImage_EmptyImage(t *testing.T) {
	checker := NewDockerChecker(10 * time.Second)

	err := checker.CheckDockerImage("")

	if err == nil {
		t.Error("CheckDockerImage() expected error for empty image name, got nil")
	}

	expectedMsg := "docker image name cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("CheckDockerImage() error = %v, want %v", err.Error(), expectedMsg)
	}
}

func TestDockerChecker_CheckDockerImage_ValidFormat(t *testing.T) {
	checker := NewDockerChecker(10 * time.Second)

	testImages := []string{
		"ghcr.io/viperproject/gobra:sha-7fa4f86",
		"gobra:latest",
		"my-registry.com/gobra:v1.0",
		"localhost:5000/gobra:dev",
	}

	for _, image := range testImages {
		t.Run(image, func(t *testing.T) {
			err := checker.CheckDockerImage(image)
			// We don't expect this to fail for format validation
			// It may fail if the image doesn't exist, but that's OK
			if err != nil {
				t.Logf("Image check failed for %s: %v (this may be expected)", image, err)
			}
		})
	}
}

func TestValidateDockerConfig_NativeBackend(t *testing.T) {
	settings := Settings{
		ExecutionBackend: BackendNative,
		GobraJAR:         "gobra.jar",
	}

	err := ValidateDockerConfig(settings)

	if err != nil {
		t.Errorf("ValidateDockerConfig() unexpected error for native backend: %v", err)
	}
}

func TestValidateDockerConfig_DockerBackend_NoImage(t *testing.T) {
	settings := Settings{
		ExecutionBackend: BackendDocker,
	}

	err := ValidateDockerConfig(settings)

	if err == nil {
		t.Error("ValidateDockerConfig() expected error for missing DockerImage, got nil")
	}

	expectedMsg := "DockerImage must be set when using Docker backend"
	if err.Error() != expectedMsg {
		t.Errorf("ValidateDockerConfig() error = %v, want %v", err.Error(), expectedMsg)
	}
}

func TestValidateDockerConfig_DockerBackend_InvalidImageName(t *testing.T) {
	settings := Settings{
		ExecutionBackend: BackendDocker,
		DockerImage:      "invalid image name with spaces",
	}

	err := ValidateDockerConfig(settings)

	if err == nil {
		t.Error("ValidateDockerConfig() expected error for invalid image name, got nil")
	}

	expectedMsg := "contains spaces"
	if err.Error() == "" || !containsSubstring(err.Error(), expectedMsg) {
		t.Errorf("ValidateDockerConfig() error = %v, want to contain %v", err.Error(), expectedMsg)
	}
}

func TestValidateDockerConfig_DockerBackend_ValidImage(t *testing.T) {
	settings := Settings{
		ExecutionBackend: BackendDocker,
		DockerImage:      "ghcr.io/viperproject/gobra:sha-7fa4f86",
		Timeout:          10 * time.Second,
	}

	err := ValidateDockerConfig(settings)

	// This may fail if Docker is not available, but that's OK for testing
	if err != nil {
		t.Logf("Docker validation failed (expected if Docker not available): %v", err)
	}
}

func TestDockerError(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	dockerErr := NewDockerError(baseErr, "test-image", "test-command", "test-context")

	if dockerErr.Err != baseErr {
		t.Errorf("DockerError.Err = %v, want %v", dockerErr.Err, baseErr)
	}

	if dockerErr.Image != "test-image" {
		t.Errorf("DockerError.Image = %v, want %v", dockerErr.Image, "test-image")
	}

	if dockerErr.Command != "test-command" {
		t.Errorf("DockerError.Command = %v, want %v", dockerErr.Command, "test-command")
	}

	if dockerErr.Context != "test-context" {
		t.Errorf("DockerError.Context = %v, want %v", dockerErr.Context, "test-context")
	}
}

func TestDockerError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *DockerError
		expected string
	}{
		{
			name:     "With context",
			err:      NewDockerError(fmt.Errorf("test error"), "image", "cmd", "context"),
			expected: "docker error (context): test error",
		},
		{
			name:     "Without context",
			err:      NewDockerError(fmt.Errorf("test error"), "image", "cmd", ""),
			expected: "docker error: test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("DockerError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDockerError_Unwrap(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	dockerErr := NewDockerError(baseErr, "image", "cmd", "context")

	unwrapped := dockerErr.Unwrap()

	if unwrapped != baseErr {
		t.Errorf("DockerError.Unwrap() = %v, want %v", unwrapped, baseErr)
	}
}

func TestIsDockerError(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	dockerErr := NewDockerError(baseErr, "image", "cmd", "context")

	if !IsDockerError(dockerErr) {
		t.Error("IsDockerError() returned false for DockerError")
	}

	if IsDockerError(baseErr) {
		t.Error("IsDockerError() returned true for non-DockerError")
	}
}

// Helper function
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
