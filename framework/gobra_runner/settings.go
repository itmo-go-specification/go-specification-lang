package gobra_runner

import (
	"os"
	"time"
)

// Backend represents the execution backend for running Gobra
type Backend string

const (
	// BackendNative runs Gobra natively on the host system
	BackendNative Backend = "native"

	// BackendDocker runs Gobra inside a Docker container
	BackendDocker Backend = "docker"
)

// IsValid checks if the backend value is valid
func (b Backend) IsValid() bool {
	return b == BackendNative || b == BackendDocker
}

type Settings struct {
	// Execution backend (native or docker)
	ExecutionBackend Backend

	// Docker-specific settings
	DockerImage string

	// Native execution settings
	JavaPath string
	GobraJAR string
	GobraBin string
	Z3Exe    string

	// Gobra verifier settings
	GobraBackend string
	Timeout      time.Duration
	LogLevel     string

	// Gobra verification flags
	NoInformation             bool
	ParallelizeBranches       bool
	AssumeInjectivityOnInhale bool
	CheckOverflow             bool

	// Additional flags to pass to Gobra
	ExtraFlags []string
}

// DefaultSettings returns default settings with support for environment variables
func DefaultSettings() Settings {
	settings := Settings{
		ExecutionBackend: BackendNative,
		DockerImage:      "ghcr.io/viperproject/gobra:sha-7fa4f86",

		JavaPath: "java",
		GobraJAR: "gobra.jar",
		GobraBin: "gobra",
		Z3Exe:    "z3",

		GobraBackend: "SILICON", // Uppercase for Docker compatibility
		Timeout:      120 * time.Second,
		LogLevel:     "ERROR",

		NoInformation:       false, // Disabled - not supported by Docker image
		ParallelizeBranches: false,
	}

	// Override with environment variables if set
	if env := os.Getenv("GPMA_EXECUTION_BACKEND"); env != "" {
		if backend := Backend(env); backend.IsValid() {
			settings.ExecutionBackend = backend
		}
	}

	if env := os.Getenv("GPMA_DOCKER_IMAGE"); env != "" {
		settings.DockerImage = env
	}

	if env := os.Getenv("GPMA_JAVA_PATH"); env != "" {
		settings.JavaPath = env
	}

	if env := os.Getenv("GPMA_GOBRA_JAR"); env != "" {
		settings.GobraJAR = env
	}

	if env := os.Getenv("GPMA_GOBRA_BACKEND"); env != "" {
		settings.GobraBackend = env
	}

	if env := os.Getenv("GPMA_TIMEOUT"); env != "" {
		if duration, err := time.ParseDuration(env); err == nil {
			settings.Timeout = duration
		}
	}

	if env := os.Getenv("GPMA_LOG_LEVEL"); env != "" {
		settings.LogLevel = env
	}

	return settings
}
