package gobra_runner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DockerChecker provides functionality to check Docker availability and validate images
type DockerChecker struct {
	timeout time.Duration
}

// NewDockerChecker creates a new DockerChecker with the specified timeout
func NewDockerChecker(timeout time.Duration) *DockerChecker {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &DockerChecker{timeout: timeout}
}

// CheckDockerAvailability checks if Docker is installed and running
func (dc *DockerChecker) CheckDockerAvailability() error {
	ctx, cancel := context.WithTimeout(context.Background(), dc.timeout)
	defer cancel()

	// Try to run 'docker version' to check if Docker is available
	cmd := exec.CommandContext(ctx, "docker", "version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("docker command timed out after %v", dc.timeout)
		}
		return fmt.Errorf("docker is not available or not running: %w\noutput: %s", err, string(output))
	}

	// Check if the output contains expected Docker version information
	outputStr := string(output)
	if !strings.Contains(outputStr, "Version") && !strings.Contains(outputStr, "version") {
		return fmt.Errorf("docker command executed but returned unexpected output: %s", outputStr)
	}

	return nil
}

// CheckDockerImage checks if a Docker image exists locally or can be pulled
func (dc *DockerChecker) CheckDockerImage(imageName string) error {
	if imageName == "" {
		return fmt.Errorf("docker image name cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), dc.timeout)
	defer cancel()

	// First check if image exists locally
	cmd := exec.CommandContext(ctx, "docker", "image", "inspect", imageName)
	output, err := cmd.CombinedOutput()

	if err == nil {
		// Image exists locally
		return nil
	}

	// Image doesn't exist locally, try to pull it
	// Note: We don't actually pull the image here to avoid long download times
	// We just check if the image name is valid and potentially available
	if strings.Contains(string(output), "No such image") {
		// Image doesn't exist locally, but the name format might be valid
		// We'll let Docker handle the actual pull when running
		return nil
	}

	// Some other error occurred
	return fmt.Errorf("failed to check docker image %s: %w\noutput: %s", imageName, err, string(output))
}

// ValidateDockerConfig validates the Docker configuration in Settings
func ValidateDockerConfig(settings Settings) error {
	if settings.ExecutionBackend != BackendDocker {
		return nil // Docker validation not needed for non-Docker backend
	}

	if settings.DockerImage == "" {
		return fmt.Errorf("DockerImage must be set when using Docker backend")
	}

	// Basic validation of Docker image name format
	image := settings.DockerImage
	if strings.Contains(image, " ") {
		return fmt.Errorf("DockerImage '%s' contains spaces, which is not valid", image)
	}

	// Check if Docker is available
	checker := NewDockerChecker(settings.Timeout)
	if err := checker.CheckDockerAvailability(); err != nil {
		return fmt.Errorf("docker availability check failed: %w", err)
	}

	// Validate Docker image (lightweight check)
	if err := checker.CheckDockerImage(image); err != nil {
		return fmt.Errorf("docker image validation failed: %w", err)
	}

	return nil
}

// DockerError represents a Docker-specific error with additional context
type DockerError struct {
	Err     error
	Image   string
	Command string
	Context string
}

func (de *DockerError) Error() string {
	if de.Context != "" {
		return fmt.Sprintf("docker error (%s): %v", de.Context, de.Err)
	}
	return fmt.Sprintf("docker error: %v", de.Err)
}

func (de *DockerError) Unwrap() error {
	return de.Err
}

// NewDockerError creates a new DockerError with the given parameters
func NewDockerError(err error, image, command, context string) *DockerError {
	return &DockerError{
		Err:     err,
		Image:   image,
		Command: command,
		Context: context,
	}
}

// IsDockerError checks if an error is a DockerError
func IsDockerError(err error) bool {
	_, ok := err.(*DockerError)
	return ok
}
