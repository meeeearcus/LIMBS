package config

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveExplicitSourceMountOverridesDefaults(t *testing.T) {
	cfg, err := Resolve(Config{
		SourceMount: "/custom/mount",
		ProjectName: "MYPROJECT",
		DestRoot:    "/tmp/out",
	})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if cfg.SourceMount != "/custom/mount" {
		t.Fatalf("expected source mount override, got %q", cfg.SourceMount)
	}
	if cfg.ProjectsRoot != filepath.Join("/custom/mount", "PROJECTS") {
		t.Fatalf("unexpected projects root: %q", cfg.ProjectsRoot)
	}
}

func TestResolveAllowsExplicitRootsWithoutSourceMount(t *testing.T) {
	cfg, err := Resolve(Config{
		ProjectsRoot: "/x/PROJECTS",
		SamplesRoot:  "/x/SAMPLES",
		ProjectName:  "MYPROJECT",
		DestRoot:     "/tmp/out",
	})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if cfg.SourceMount != "" {
		t.Fatalf("expected empty source mount, got %q", cfg.SourceMount)
	}
	if cfg.ProjectFile != filepath.Join("/x/PROJECTS", "MYPROJECT.s4project", "project.json") {
		t.Fatalf("unexpected project file: %q", cfg.ProjectFile)
	}
}

func TestDefaultSourceMountByOS(t *testing.T) {
	got := defaultSourceMount()
	switch runtime.GOOS {
	case "darwin":
		if got != "/Volumes/S-4" {
			t.Fatalf("expected macOS default, got %q", got)
		}
	case "linux":
		if got == "" {
			t.Fatalf("expected linux default path")
		}
	case "windows":
		if got != "" {
			t.Fatalf("expected empty windows default, got %q", got)
		}
	default:
		if got != "" {
			t.Fatalf("expected empty default for unsupported OS, got %q", got)
		}
	}
}

func TestResolveErrorsWhenProjectsRootCannotBeDerived(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-specific behavior")
	}

	_, err := Resolve(Config{
		ProjectName: "MYPROJECT",
		DestRoot:    "C:\\tmp\\out",
	})
	if err == nil {
		t.Fatalf("expected error when projects root cannot be derived")
	}
}
