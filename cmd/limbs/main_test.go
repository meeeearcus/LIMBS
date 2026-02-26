package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestParseExportFlags_USBDrive(t *testing.T) {
	cfg, err := parseExportFlags([]string{
		"--project-name", "TEST",
		"--dest-root", "/tmp/out",
		"--usb-drive", "/mnt/usb",
	})
	if err != nil {
		t.Fatalf("parseExportFlags failed: %v", err)
	}
	if cfg.USBDrive != "/mnt/usb" {
		t.Fatalf("expected usb drive to be set, got %q", cfg.USBDrive)
	}
}

func TestParseExportFlags_HelpShowsDetailedUsage(t *testing.T) {
	output := captureStderr(t, func() {
		_, _ = parseExportFlags([]string{"--help"})
	})

	assertContainsAll(t, output,
		"LIMBS - An exporter for Torso S-4 Projects",
		"Path Resolution:",
		"03_USB_DRIVE/...  resolves under --usb-drive",
		"Examples:",
	)
}

func TestParseExportFlags_InvalidFlagShowsDetailedUsage(t *testing.T) {
	output := captureStderr(t, func() {
		_, _ = parseExportFlags([]string{"--bad-flag"})
	})

	assertContainsAll(t, output,
		"LIMBS - An exporter for Torso S-4 Projects",
		"Required:",
		"Options:",
		"Examples:",
	)
}

func TestParseGUIFlags(t *testing.T) {
	cfg, err := parseGUIFlags([]string{"--port", "8080", "--no-open"})
	if err != nil {
		t.Fatalf("parseGUIFlags failed: %v", err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected port=8080, got %d", cfg.Port)
	}
	if !cfg.NoOpen {
		t.Fatalf("expected no-open=true")
	}
}

func TestParseGUIFlags_Help(t *testing.T) {
	output := captureStderr(t, func() {
		_, _ = parseGUIFlags([]string{"--help"})
	})
	assertContainsAll(t, output,
		"LIMBS GUI - Local web interface",
		"limbs gui [--port <n>] [--no-open]",
	)
}

func TestPrintRootHelpContainsGUI(t *testing.T) {
	var buf bytes.Buffer
	printRootHelp(&buf)
	assertContainsAll(t, buf.String(),
		"Commands:",
		"gui",
		"limbs gui [--port <n>] [--no-open]",
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
