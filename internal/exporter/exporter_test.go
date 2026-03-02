package exporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"limbs/internal/config"
)

func TestRun_V11UsesLegacyRewriteAndNaming(t *testing.T) {
	tmp := t.TempDir()
	samplesRoot := filepath.Join(tmp, "SAMPLES")
	if err := os.MkdirAll(filepath.Join(samplesRoot, "BOUNCES", "DX"), 0o755); err != nil {
		t.Fatalf("mkdir samples: %v", err)
	}
	if err := os.WriteFile(filepath.Join(samplesRoot, "BOUNCES", "DX", "kick.wav"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write sample: %v", err)
	}

	projectFile := writeProject(t, tmp, "MYPROJECT", "v11", "/browser/samples/02_USER/BOUNCES/DX/kick.wav")
	res, err := Run(config.Config{
		ProjectName:  "MYPROJECT",
		ProjectFile:  projectFile,
		DestRoot:     filepath.Join(tmp, "out"),
		SamplesRoot:  samplesRoot,
		AllowMissing: true,
		Zip:          true,
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if res.PathMode != "v11-legacy" || res.MinFirmware != "2.0.4+" {
		t.Fatalf("unexpected metadata: mode=%s firmware=%s", res.PathMode, res.MinFirmware)
	}
	if !strings.Contains(filepath.Base(res.OutputDir), "MYPROJECT_fw2.0.4+_export") {
		t.Fatalf("unexpected output dir: %s", res.OutputDir)
	}
	if filepath.Base(res.ZipPath) != "MYPROJECT_fw2.0.4+.zip" {
		t.Fatalf("unexpected zip name: %s", res.ZipPath)
	}

	outJSON := filepath.Join(res.OutputDir, "PROJECTS", "MYPROJECT-LIMB.s4project", "project.json")
	data, err := os.ReadFile(outJSON)
	if err != nil {
		t.Fatalf("read output json: %v", err)
	}
	if !strings.Contains(string(data), "/browser/samples/02_USER/LIMBS/MYPROJECT/kick.wav") {
		t.Fatalf("expected legacy rewritten path, got:\n%s", string(data))
	}
	if !strings.Contains(string(data), "\"version\": \"v11\"") {
		t.Fatalf("expected version to remain v11")
	}

	readmePath := filepath.Join(res.OutputDir, "IMPORT-README.md")
	readme, _ := os.ReadFile(readmePath)
	if !strings.Contains(string(readme), "Minimum firmware: 2.0.4+") {
		t.Fatalf("README missing firmware line:\n%s", string(readme))
	}
	if !strings.Contains(string(readme), "Unzip \"MYPROJECT_fw2.0.4+.zip\".") {
		t.Fatalf("README missing zip filename:\n%s", string(readme))
	}
}

func TestRun_V13UsesNewRewriteAndNaming(t *testing.T) {
	tmp := t.TempDir()
	samplesRoot := filepath.Join(tmp, "SAMPLES")
	if err := os.MkdirAll(filepath.Join(samplesRoot, "BOUNCES", "DX"), 0o755); err != nil {
		t.Fatalf("mkdir samples: %v", err)
	}
	if err := os.WriteFile(filepath.Join(samplesRoot, "BOUNCES", "DX", "kick.wav"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write sample: %v", err)
	}

	projectFile := writeProject(t, tmp, "MYPROJECT", "v13", "/tmp/S-4/browser/SAMPLES/02_USER/BOUNCES/DX/kick.wav")
	res, err := Run(config.Config{
		ProjectName:  "MYPROJECT",
		ProjectFile:  projectFile,
		DestRoot:     filepath.Join(tmp, "out"),
		SamplesRoot:  samplesRoot,
		AllowMissing: true,
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if res.PathMode != "v13-new" || res.MinFirmware != "2.1.3+" {
		t.Fatalf("unexpected metadata: mode=%s firmware=%s", res.PathMode, res.MinFirmware)
	}
	if !strings.Contains(filepath.Base(res.OutputDir), "MYPROJECT_fw2.1.3+_export") {
		t.Fatalf("unexpected output dir: %s", res.OutputDir)
	}

	outJSON := filepath.Join(res.OutputDir, "PROJECTS", "MYPROJECT-LIMB.s4project", "project.json")
	data, err := os.ReadFile(outJSON)
	if err != nil {
		t.Fatalf("read output json: %v", err)
	}
	if !strings.Contains(string(data), "/tmp/S-4/browser/SAMPLES/02_USER/LIMBS/MYPROJECT/kick.wav") {
		t.Fatalf("expected v13 rewritten path, got:\n%s", string(data))
	}
}

func TestRun_UnknownVersionFallsBackToLegacyWithWarning(t *testing.T) {
	tmp := t.TempDir()
	samplesRoot := filepath.Join(tmp, "SAMPLES")
	if err := os.MkdirAll(filepath.Join(samplesRoot, "BOUNCES", "DX"), 0o755); err != nil {
		t.Fatalf("mkdir samples: %v", err)
	}
	if err := os.WriteFile(filepath.Join(samplesRoot, "BOUNCES", "DX", "kick.wav"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write sample: %v", err)
	}

	projectFile := writeProject(t, tmp, "MYPROJECT", "v12", "/browser/samples/02_USER/BOUNCES/DX/kick.wav")
	res, err := Run(config.Config{
		ProjectName:  "MYPROJECT",
		ProjectFile:  projectFile,
		DestRoot:     filepath.Join(tmp, "out"),
		SamplesRoot:  samplesRoot,
		AllowMissing: true,
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if res.PathMode != "unknown-assumed-v11" {
		t.Fatalf("unexpected mode: %s", res.PathMode)
	}
	if res.VersionAssumptionWarning == "" {
		t.Fatalf("expected warning for unknown version")
	}
	if res.MinFirmware != "2.0.4+" {
		t.Fatalf("unexpected firmware: %s", res.MinFirmware)
	}

	readmePath := filepath.Join(res.OutputDir, "IMPORT-README.md")
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	if !strings.Contains(string(readme), "Assumption: ") {
		t.Fatalf("expected README assumption line for unknown version:\\n%s", string(readme))
	}
}

func TestDetectPathMode(t *testing.T) {
	mode, fw, warn := detectPathMode("v11")
	if mode != "v11-legacy" || fw != "2.0.4+" || warn != "" {
		t.Fatalf("unexpected v11 mapping: %s %s %s", mode, fw, warn)
	}

	mode, fw, warn = detectPathMode("v13")
	if mode != "v13-new" || fw != "2.1.3+" || warn != "" {
		t.Fatalf("unexpected v13 mapping: %s %s %s", mode, fw, warn)
	}

	mode, fw, warn = detectPathMode("v99")
	if mode != "unknown-assumed-v11" || fw != "2.0.4+" || warn == "" {
		t.Fatalf("unexpected unknown mapping: %s %s %s", mode, fw, warn)
	}
}

func writeProject(t *testing.T, root, projectName, version, ref string) string {
	t.Helper()
	projectDir := filepath.Join(root, "PROJECTS", projectName+".s4project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}

	payload := map[string]any{
		"version": version,
		"data": map[string]any{
			"tracks": []any{
				map[string]any{"library_sample_path": ref},
			},
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal project json: %v", err)
	}
	projectFile := filepath.Join(projectDir, "project.json")
	if err := os.WriteFile(projectFile, b, 0o644); err != nil {
		t.Fatalf("write project json: %v", err)
	}
	return projectFile
}
