package gobra

import (
	"strings"
	"testing"
)

func TestDefaultConfigFromEnv(t *testing.T) {
	t.Setenv("GOBRA_BACKEND", "docker")
	t.Setenv("GOBRA_JAR", "/custom/gobra.jar")
	t.Setenv("Z3_EXE", "/custom/z3")

	cfg := DefaultConfig()
	if cfg.Backend != BackendDocker {
		t.Fatalf("expected docker backend, got %q", cfg.Backend)
	}
	if cfg.Jar != "/custom/gobra.jar" {
		t.Fatalf("expected custom jar, got %q", cfg.Jar)
	}
	if cfg.Z3Exe != "/custom/z3" {
		t.Fatalf("expected custom z3, got %q", cfg.Z3Exe)
	}
}

func TestApplyPluginParams(t *testing.T) {
	cfg := DefaultConfig()
	err := cfg.ApplyPluginParams(map[string]interface{}{
		"backend":    "native",
		"gobra_jar":  "/opt/gobra.jar",
		"z3_exe":     "/opt/z3",
		"boogie_exe": "/opt/Boogie",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Backend != BackendNative {
		t.Fatalf("expected native backend, got %q", cfg.Backend)
	}
	if cfg.Jar != "/opt/gobra.jar" {
		t.Fatalf("unexpected jar: %q", cfg.Jar)
	}
}

func TestGobraProcessEnv(t *testing.T) {
	t.Setenv("Z3_EXE", "old")

	env := gobraProcessEnv(GobraConfig{
		Z3Exe:     "/new/z3",
		BoogieExe: "/new/Boogie",
	})
	foundZ3 := false
	foundLD := false
	for _, e := range env {
		if e == "Z3_EXE=/new/z3" {
			foundZ3 = true
		}
		if e == "Z3_EXE=old" {
			t.Fatal("old Z3_EXE should be replaced")
		}
		if strings.HasPrefix(e, "LD_LIBRARY_PATH=") && strings.Contains(e, defaultLibPath) {
			foundLD = true
		}
	}
	if !foundZ3 {
		t.Fatal("Z3_EXE not set in environment")
	}
	if !foundLD {
		t.Fatal("LD_LIBRARY_PATH not set in environment")
	}
}
