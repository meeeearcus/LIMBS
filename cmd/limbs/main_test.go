package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestParseExportFlags_HelpShowsRichUsage(t *testing.T) {
	output := captureStderr(t, func() {
		_, _ = parseExportFlags([]string{"--help"})
	})

	assertContainsAll(t, output,
		"Basic Options:",
		"Advanced Options (optional):",
		"OS-Specific Mount Defaults:",
		"USB example mapping:",
	)
}

func TestParseExportFlags_InvalidFlagShowsRichUsage(t *testing.T) {
	output := captureStderr(t, func() {
		_, _ = parseExportFlags([]string{"--bad-flag"})
	})

	assertContainsAll(t, output,
		"LIMBS - An exporter for Torso S-4 Projects",
		"Basic Usage (macOS/Linux):",
		"Basic Usage (Windows):",
		"Basic Options:",
	)
}

func TestParseExportFlags_USBDriveParses(t *testing.T) {
	cfg, err := parseExportFlags([]string{"--project-name", "MYPROJECT", "--dest-root", "/tmp/out", "--usb-drive", "/Volumes/MY_USB"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if cfg.USBDrive != "/Volumes/MY_USB" {
		t.Fatalf("unexpected usb drive path: %q", cfg.USBDrive)
	}
}

func TestPrintRootHelpExportOnly(t *testing.T) {
	var buf bytes.Buffer
	printRootHelp(&buf)
	out := buf.String()

	assertContainsAll(t, out,
		"Command:",
		"export",
		"limbs export [options]",
	)

	if strings.Contains(out, "gui") {
		t.Fatalf("root help should not mention gui command")
	}
}

func TestPrintExportHelp_IncludesUSBGuidance(t *testing.T) {
	var buf bytes.Buffer
	printExportHelp(&buf)
	out := buf.String()

	assertContainsAll(t, out,
		"--usb-drive <path>",
		"03_USB_DRIVE",
		"EXAMPLE.wav",
		"On Windows set explicitly",
		"Full PROJECTS path override",
		"Full SAMPLES path override",
		"Effective location: <dest-root>/<project-name>_..._export/<limbs-root>/<project-name>/",
		"OS-Specific Mount Defaults:",
		"Windows: no fixed default; set --source-mount (e.g. E:\\)",
	)
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}
	os.Stderr = w
	defer func() {
		os.Stderr = orig
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}
	return buf.String()
}

func assertContainsAll(t *testing.T, got string, want ...string) {
	t.Helper()
	for _, s := range want {
		if !strings.Contains(got, s) {
			t.Fatalf("expected output to contain %q, got:\n%s", s, got)
		}
	}
}
