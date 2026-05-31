package gobra

import (
	"fmt"
	"os"
	"strings"
)

const (
	defaultJava       = "java"
	defaultJavaOpts   = "-Xss128m"
	defaultDockerImg  = "ghcr.io/viperproject/gobra:sha-7fa4f86"
	defaultGobraJar   = "/usr/lib/gpma/gobra/gobra.jar"
	defaultZ3Exe      = "/usr/lib/gpma/bin/z3"
	defaultBoogieExe  = "/usr/lib/gpma/bin/Boogie"
	defaultLibPath    = "/usr/lib/gpma/lib"
)

// Backend selects how Gobra is invoked.
type Backend string

const (
	BackendNative Backend = "native"
	BackendDocker Backend = "docker"
)

// GobraConfig holds runtime settings for invoking the Gobra verifier.
type GobraConfig struct {
	Backend     Backend
	Java        string
	JavaOpts    string
	Jar         string
	Z3Exe       string
	BoogieExe   string
	DockerImage string
}

// DefaultConfig returns configuration resolved from environment variables with
// packaged-install defaults.
func DefaultConfig() GobraConfig {
	cfg := GobraConfig{
		Backend:     BackendNative,
		Java:        defaultJava,
		JavaOpts:    defaultJavaOpts,
		Jar:         defaultGobraJar,
		Z3Exe:       defaultZ3Exe,
		BoogieExe:   defaultBoogieExe,
		DockerImage: defaultDockerImg,
	}
	cfg.applyEnv()
	return cfg
}

func (c *GobraConfig) applyEnv() {
	if v := os.Getenv("GOBRA_BACKEND"); v != "" {
		c.Backend = Backend(strings.ToLower(strings.TrimSpace(v)))
	}
	if v := os.Getenv("GOBRA_JAVA"); v != "" {
		c.Java = v
	}
	if v := os.Getenv("GOBRA_JAVA_OPTS"); v != "" {
		c.JavaOpts = v
	}
	if v := os.Getenv("GOBRA_JAR"); v != "" {
		c.Jar = v
	}
	if v := os.Getenv("Z3_EXE"); v != "" {
		c.Z3Exe = v
	}
	if v := os.Getenv("BOOGIE_EXE"); v != "" {
		c.BoogieExe = v
	}
	if v := os.Getenv("GOBRA_DOCKER_IMAGE"); v != "" {
		c.DockerImage = v
	}
}

// ApplyPluginParams merges plugin configuration keys into the config.
func (c *GobraConfig) ApplyPluginParams(params map[string]interface{}) error {
	if val, ok := params["backend"]; ok {
		switch v := val.(type) {
		case string:
			c.Backend = Backend(strings.ToLower(strings.TrimSpace(v)))
		default:
			return fmt.Errorf("backend: must be a string")
		}
	}
	if val, ok := params["java"]; ok {
		if s, ok := val.(string); ok {
			c.Java = s
		} else {
			return fmt.Errorf("java: must be a string")
		}
	}
	if val, ok := params["java_opts"]; ok {
		if s, ok := val.(string); ok {
			c.JavaOpts = s
		} else {
			return fmt.Errorf("java_opts: must be a string")
		}
	}
	if val, ok := params["gobra_jar"]; ok {
		if s, ok := val.(string); ok {
			c.Jar = s
		} else {
			return fmt.Errorf("gobra_jar: must be a string")
		}
	}
	if val, ok := params["z3_exe"]; ok {
		if s, ok := val.(string); ok {
			c.Z3Exe = s
		} else {
			return fmt.Errorf("z3_exe: must be a string")
		}
	}
	if val, ok := params["boogie_exe"]; ok {
		if s, ok := val.(string); ok {
			c.BoogieExe = s
		} else {
			return fmt.Errorf("boogie_exe: must be a string")
		}
	}
	if val, ok := params["docker_image"]; ok {
		if s, ok := val.(string); ok {
			c.DockerImage = s
		} else {
			return fmt.Errorf("docker_image: must be a string")
		}
	}
	return nil
}

func (c *GobraConfig) javaArgs() []string {
	if c.JavaOpts == "" {
		return nil
	}
	return strings.Fields(c.JavaOpts)
}
