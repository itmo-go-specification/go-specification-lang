package gobra_runner

import (
	"os"
	"testing"
	"time"
)

func TestBackendIsValid(t *testing.T) {
	tests := []struct {
		name     string
		backend  Backend
		expected bool
	}{
		{"Native backend", BackendNative, true},
		{"Docker backend", BackendDocker, true},
		{"Invalid backend", Backend("invalid"), false},
		{"Empty backend", Backend(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.backend.IsValid(); got != tt.expected {
				t.Errorf("Backend.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultSettings(t *testing.T) {
	// Clear environment variables before test
	os.Unsetenv("GPMA_EXECUTION_BACKEND")
	os.Unsetenv("GPMA_DOCKER_IMAGE")
	os.Unsetenv("GPMA_JAVA_PATH")
	os.Unsetenv("GPMA_GOBRA_JAR")
	os.Unsetenv("GPMA_GOBRA_BACKEND")
	os.Unsetenv("GPMA_TIMEOUT")
	os.Unsetenv("GPMA_LOG_LEVEL")

	settings := DefaultSettings()

	// Check default values
	if settings.ExecutionBackend != BackendNative {
		t.Errorf("DefaultSettings() ExecutionBackend = %v, want %v", settings.ExecutionBackend, BackendNative)
	}

	if settings.DockerImage != "ghcr.io/viperproject/gobra:sha-7fa4f86" {
		t.Errorf("DefaultSettings() DockerImage = %v, want %v", settings.DockerImage, "ghcr.io/viperproject/gobra:sha-7fa4f86")
	}

	if settings.JavaPath != "java" {
		t.Errorf("DefaultSettings() JavaPath = %v, want %v", settings.JavaPath, "java")
	}

	if settings.GobraJAR != "gobra.jar" {
		t.Errorf("DefaultSettings() GobraJAR = %v, want %v", settings.GobraJAR, "gobra.jar")
	}

	if settings.GobraBackend != "silicon" {
		t.Errorf("DefaultSettings() GobraBackend = %v, want %v", settings.GobraBackend, "silicon")
	}

	if settings.Timeout != 120*time.Second {
		t.Errorf("DefaultSettings() Timeout = %v, want %v", settings.Timeout, 120*time.Second)
	}

	if settings.LogLevel != "ERROR" {
		t.Errorf("DefaultSettings() LogLevel = %v, want %v", settings.LogLevel, "ERROR")
	}

	if !settings.NoInformation {
		t.Errorf("DefaultSettings() NoInformation = %v, want %v", settings.NoInformation, true)
	}

	if settings.ParallelizeBranches {
		t.Errorf("DefaultSettings() ParallelizeBranches = %v, want %v", settings.ParallelizeBranches, false)
	}
}

func TestDefaultSettingsWithEnvVars(t *testing.T) {
	tests := []struct {
		name          string
		envVar        string
		envValue      string
		checkFunc     func(*testing.T, Settings)
		cleanupNeeded bool
	}{
		{
			name:     "GPMA_EXECUTION_BACKEND",
			envVar:   "GPMA_EXECUTION_BACKEND",
			envValue: "docker",
			checkFunc: func(t *testing.T, s Settings) {
				if s.ExecutionBackend != BackendDocker {
					t.Errorf("ExecutionBackend = %v, want %v", s.ExecutionBackend, BackendDocker)
				}
			},
			cleanupNeeded: true,
		},
		{
			name:     "GPMA_DOCKER_IMAGE",
			envVar:   "GPMA_DOCKER_IMAGE",
			envValue: "custom/gobra:latest",
			checkFunc: func(t *testing.T, s Settings) {
				if s.DockerImage != "custom/gobra:latest" {
					t.Errorf("DockerImage = %v, want %v", s.DockerImage, "custom/gobra:latest")
				}
			},
			cleanupNeeded: true,
		},
		{
			name:     "GPMA_JAVA_PATH",
			envVar:   "GPMA_JAVA_PATH",
			envValue: "/usr/bin/java",
			checkFunc: func(t *testing.T, s Settings) {
				if s.JavaPath != "/usr/bin/java" {
					t.Errorf("JavaPath = %v, want %v", s.JavaPath, "/usr/bin/java")
				}
			},
			cleanupNeeded: true,
		},
		{
			name:     "GPMA_GOBRA_JAR",
			envVar:   "GPMA_GOBRA_JAR",
			envValue: "/path/to/gobra.jar",
			checkFunc: func(t *testing.T, s Settings) {
				if s.GobraJAR != "/path/to/gobra.jar" {
					t.Errorf("GobraJAR = %v, want %v", s.GobraJAR, "/path/to/gobra.jar")
				}
			},
			cleanupNeeded: true,
		},
		{
			name:     "GPMA_GOBRA_BACKEND",
			envVar:   "GPMA_GOBRA_BACKEND",
			envValue: "carbon",
			checkFunc: func(t *testing.T, s Settings) {
				if s.GobraBackend != "carbon" {
					t.Errorf("GobraBackend = %v, want %v", s.GobraBackend, "carbon")
				}
			},
			cleanupNeeded: true,
		},
		{
			name:     "GPMA_TIMEOUT",
			envVar:   "GPMA_TIMEOUT",
			envValue: "5m",
			checkFunc: func(t *testing.T, s Settings) {
				expected := 5 * time.Minute
				if s.Timeout != expected {
					t.Errorf("Timeout = %v, want %v", s.Timeout, expected)
				}
			},
			cleanupNeeded: true,
		},
		{
			name:     "GPMA_LOG_LEVEL",
			envVar:   "GPMA_LOG_LEVEL",
			envValue: "INFO",
			checkFunc: func(t *testing.T, s Settings) {
				if s.LogLevel != "INFO" {
					t.Errorf("LogLevel = %v, want %v", s.LogLevel, "INFO")
				}
			},
			cleanupNeeded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			os.Unsetenv("GPMA_EXECUTION_BACKEND")
			os.Unsetenv("GPMA_DOCKER_IMAGE")
			os.Unsetenv("GPMA_JAVA_PATH")
			os.Unsetenv("GPMA_GOBRA_JAR")
			os.Unsetenv("GPMA_GOBRA_BACKEND")
			os.Unsetenv("GPMA_TIMEOUT")
			os.Unsetenv("GPMA_LOG_LEVEL")

			// Set the specific env var for this test
			os.Setenv(tt.envVar, tt.envValue)

			// Get settings
			settings := DefaultSettings()

			// Check the result
			tt.checkFunc(t, settings)

			// Cleanup
			if tt.cleanupNeeded {
				os.Unsetenv(tt.envVar)
			}
		})
	}
}

func TestDefaultSettingsInvalidBackend(t *testing.T) {
	// Clear all env vars first
	os.Unsetenv("GPMA_EXECUTION_BACKEND")
	os.Unsetenv("GPMA_DOCKER_IMAGE")
	os.Unsetenv("GPMA_JAVA_PATH")
	os.Unsetenv("GPMA_GOBRA_JAR")
	os.Unsetenv("GPMA_GOBRA_BACKEND")
	os.Unsetenv("GPMA_TIMEOUT")
	os.Unsetenv("GPMA_LOG_LEVEL")

	// Set invalid backend
	os.Setenv("GPMA_EXECUTION_BACKEND", "invalid")
	defer os.Unsetenv("GPMA_EXECUTION_BACKEND")

	settings := DefaultSettings()

	// Should fall back to default when invalid
	if settings.ExecutionBackend != BackendNative {
		t.Errorf("ExecutionBackend = %v, want %v (default)", settings.ExecutionBackend, BackendNative)
	}
}

func TestDefaultSettingsInvalidTimeout(t *testing.T) {
	// Clear all env vars first
	os.Unsetenv("GPMA_EXECUTION_BACKEND")
	os.Unsetenv("GPMA_DOCKER_IMAGE")
	os.Unsetenv("GPMA_JAVA_PATH")
	os.Unsetenv("GPMA_GOBRA_JAR")
	os.Unsetenv("GPMA_GOBRA_BACKEND")
	os.Unsetenv("GPMA_TIMEOUT")
	os.Unsetenv("GPMA_LOG_LEVEL")

	// Set invalid timeout
	os.Setenv("GPMA_TIMEOUT", "invalid")
	defer os.Unsetenv("GPMA_TIMEOUT")

	settings := DefaultSettings()

	// Should fall back to default when invalid
	if settings.Timeout != 120*time.Second {
		t.Errorf("Timeout = %v, want %v (default)", settings.Timeout, 120*time.Second)
	}
}

func TestSettingsStructure(t *testing.T) {
	settings := Settings{
		ExecutionBackend:          BackendDocker,
		DockerImage:               "custom/gobra:v1.0",
		JavaPath:                  "/usr/bin/java",
		GobraJAR:                  "/opt/gobra/gobra.jar",
		GobraBin:                  "/opt/gobra/gobra",
		Z3Exe:                     "/opt/z3/bin/z3",
		GobraBackend:              "carbon",
		Timeout:                   5 * time.Minute,
		LogLevel:                  "DEBUG",
		NoInformation:             false,
		ParallelizeBranches:       true,
		AssumeInjectivityOnInhale: true,
		CheckOverflow:             true,
		ExtraFlags:                []string{"--flag1", "--flag2"},
	}

	// Verify all fields are set correctly
	if settings.ExecutionBackend != BackendDocker {
		t.Errorf("ExecutionBackend not set correctly")
	}

	if settings.DockerImage != "custom/gobra:v1.0" {
		t.Errorf("DockerImage not set correctly")
	}

	if len(settings.ExtraFlags) != 2 {
		t.Errorf("ExtraFlags length = %d, want 2", len(settings.ExtraFlags))
	}
}
