package exporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"limbs/internal/config"
)

func TestRun_MissingUSBDriveConfig(t *testing.T) {
	tmp := t.TempDir()
	projectFile := writeProjectFixture(t, tmp, "TEST", []string{
		"/tmp/S-4/browser/SAMPLES/03_USB_DRIVE/SAMPLES/INSTRUMENTS/DRONE_1.wav",
	})

	cfg := config.Config{
		ProjectName:  "TEST",
		ProjectFile:  projectFile,
		DestRoot:     filepath.Join(tmp, "out"),
		SamplesRoot:  filepath.Join(tmp, "samples"),
		LimbsRoot:    "SAMPLES/LIMBS",
		AllowMissing: true,
	}

	res, err := Run(cfg)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(res.Stats.MissingSamples) != 1 {
		t.Fatalf("expected 1 missing sample, got %d", len(res.Stats.MissingSamples))
	}
	got := res.Stats.MissingSamples[0]
	if got.Reason != "usb drive path not configured (--usb-drive)" {
		t.Fatalf("unexpected reason: %q", got.Reason)
	}
}

func TestRun_USBDriveResolvesAndCopies(t *testing.T) {
	tmp := t.TempDir()
	usbRoot := filepath.Join(tmp, "usb")
	if err := os.MkdirAll(filepath.Join(usbRoot, "MyFolder"), 0o755); err != nil {
		t.Fatalf("mkdir usb root: %v", err)
	}
	srcSample := filepath.Join(usbRoot, "MyFolder", "DRONE_1.wav")
	if err := os.WriteFile(srcSample, []byte("sample"), 0o644); err != nil {
		t.Fatalf("write sample: %v", err)
	}

	projectFile := writeProjectFixture(t, tmp, "TEST", []string{
		"/tmp/S-4/browser/SAMPLES/03_USB_DRIVE/MyFolder/DRONE_1.wav",
	})

	cfg := config.Config{
		ProjectName:  "TEST",
		ProjectFile:  projectFile,
		DestRoot:     filepath.Join(tmp, "out"),
		SamplesRoot:  filepath.Join(tmp, "samples"),
		USBDrive:     usbRoot,
		LimbsRoot:    "SAMPLES/LIMBS",
		AllowMissing: true,
	}

	res, err := Run(cfg)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if res.Stats.SamplesCopied != 1 {
		t.Fatalf("expected 1 copied sample, got %d", res.Stats.SamplesCopied)
	}
	if len(res.Stats.MissingSamples) != 0 {
		t.Fatalf("expected no missing samples, got %d", len(res.Stats.MissingSamples))
	}
}

func TestRun_MissingFileHasResolvedPath(t *testing.T) {
	tmp := t.TempDir()
	usbRoot := filepath.Join(tmp, "usb")
	projectFile := writeProjectFixture(t, tmp, "TEST", []string{
		"/browser/samples/03_USB_DRIVE/Synths/missing.wav",
	})

	cfg := config.Config{
		ProjectName:  "TEST",
		ProjectFile:  projectFile,
		DestRoot:     filepath.Join(tmp, "out"),
		SamplesRoot:  filepath.Join(tmp, "samples"),
		USBDrive:     usbRoot,
		LimbsRoot:    "SAMPLES/LIMBS",
		AllowMissing: true,
	}

	res, err := Run(cfg)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(res.Stats.MissingSamples) != 1 {
		t.Fatalf("expected 1 missing sample, got %d", len(res.Stats.MissingSamples))
	}
	got := res.Stats.MissingSamples[0]
	if got.Reason != "file not found" {
		t.Fatalf("unexpected reason: %q", got.Reason)
	}
	wantResolved := filepath.Join(usbRoot, "Synths", "missing.wav")
	if got.ResolvedPath != wantResolved {
		t.Fatalf("expected resolved path %q, got %q", wantResolved, got.ResolvedPath)
	}
}

func TestRun_MixedNamespaces(t *testing.T) {
	tmp := t.TempDir()

	samplesRoot := filepath.Join(tmp, "samples")
	userSample := filepath.Join(samplesRoot, "BOUNCES", "DX", "kick.wav")
	if err := os.MkdirAll(filepath.Dir(userSample), 0o755); err != nil {
		t.Fatalf("mkdir user sample: %v", err)
	}
	if err := os.WriteFile(userSample, []byte("kick"), 0o644); err != nil {
		t.Fatalf("write user sample: %v", err)
	}

	usbRoot := filepath.Join(tmp, "usb")
	usbSample := filepath.Join(usbRoot, "MyFolder", "DRONE_1.wav")
	if err := os.MkdirAll(filepath.Dir(usbSample), 0o755); err != nil {
		t.Fatalf("mkdir usb sample: %v", err)
	}
	if err := os.WriteFile(usbSample, []byte("drone"), 0o644); err != nil {
		t.Fatalf("write usb sample: %v", err)
	}

	projectFile := writeProjectFixture(t, tmp, "TEST", []string{
		"/browser/samples/02_USER/BOUNCES/DX/kick.wav",
		"/tmp/S-4/browser/SAMPLES/03_USB_DRIVE/MyFolder/DRONE_1.wav",
	})

	cfg := config.Config{
		ProjectName:  "TEST",
		ProjectFile:  projectFile,
		DestRoot:     filepath.Join(tmp, "out"),
		SamplesRoot:  samplesRoot,
		USBDrive:     usbRoot,
		LimbsRoot:    "SAMPLES/LIMBS",
		AllowMissing: true,
	}

	res, err := Run(cfg)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if res.Stats.SamplesCopied != 2 {
		t.Fatalf("expected 2 copied samples, got %d", res.Stats.SamplesCopied)
	}
	if len(res.Stats.MissingSamples) != 0 {
		t.Fatalf("expected no missing samples, got %d", len(res.Stats.MissingSamples))
	}
}

func writeProjectFixture(t *testing.T, root, projectName string, refs []string) string {
	t.Helper()

	projectDir := filepath.Join(root, "PROJECTS", projectName+".s4project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}

	items := make([]map[string]any, 0, len(refs))
	for _, ref := range refs {
		items = append(items, map[string]any{
			"library_sample_path": ref,
		})
	}

	rootJSON := map[string]any{
		"data": map[string]any{
			"tracks": items,
		},
	}

	encoded, err := json.Marshal(rootJSON)
	if err != nil {
		t.Fatalf("marshal project json: %v", err)
	}

	projectFile := filepath.Join(projectDir, "project.json")
	if err := os.WriteFile(projectFile, encoded, 0o644); err != nil {
		t.Fatalf("write project json: %v", err)
	}

	return projectFile
}
