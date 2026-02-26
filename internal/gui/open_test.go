package gui

import (
	"runtime"
	"testing"
)

func TestOpenCommand(t *testing.T) {
	name, args, err := openCommand("http://127.0.0.1:9999")
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		if err == nil {
			t.Fatalf("expected error for unsupported OS")
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name == "" {
		t.Fatalf("expected command name")
	}
	if len(args) == 0 {
		t.Fatalf("expected command args")
	}
}
