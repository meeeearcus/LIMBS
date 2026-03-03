package resolver

import (
	"path/filepath"
	"testing"
)

func TestResolveVirtualSamplePath_UserPaths(t *testing.T) {
	samplesRoot := filepath.Join("root", "SAMPLES")

	tests := []struct {
		name string
		ref  string
	}{
		{
			name: "legacy root",
			ref:  "/browser/samples/02_USER/BOUNCES/DX/kick.wav",
		},
		{
			name: "new root",
			ref:  "/tmp/S-4/browser/SAMPLES/02_USER/BOUNCES/DX/kick.wav",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveVirtualSamplePath(samplesRoot, "", tc.ref)
			want := filepath.Join(samplesRoot, "BOUNCES", "DX", "kick.wav")
			if got.ResolvedPath != want {
				t.Fatalf("expected %q, got %q", want, got.ResolvedPath)
			}
			if got.Reason != "" {
				t.Fatalf("expected empty reason, got %q", got.Reason)
			}
		})
	}
}

func TestResolveVirtualSamplePath_USBDrive(t *testing.T) {
	usbRoot := filepath.Join("mnt", "usb")
	got := ResolveVirtualSamplePath("", usbRoot, "/tmp/S-4/browser/SAMPLES/03_USB_DRIVE/MyFolder/DRONE_1.wav")
	want := filepath.Join(usbRoot, "MyFolder", "DRONE_1.wav")
	if got.ResolvedPath != want {
		t.Fatalf("expected %q, got %q", want, got.ResolvedPath)
	}
	if got.Namespace != NamespaceUSBDrive {
		t.Fatalf("unexpected namespace: %q", got.Namespace)
	}
}

func TestResolveVirtualSamplePath_USBMissingRoot(t *testing.T) {
	got := ResolveVirtualSamplePath("", "", "/browser/samples/03_USB_DRIVE/SAMPLES/INSTRUMENTS/DRONE_1.wav")
	if got.ResolvedPath != "" {
		t.Fatalf("expected empty resolved path, got %q", got.ResolvedPath)
	}
	if got.Reason != "usb drive path not configured (--usb-drive)" {
		t.Fatalf("unexpected reason: %q", got.Reason)
	}
}

func TestResolveVirtualSamplePath_Unrecognized(t *testing.T) {
	got := ResolveVirtualSamplePath("", "", "/tmp/other/path.wav")
	if got.ResolvedPath != "" {
		t.Fatalf("expected empty resolved path, got %q", got.ResolvedPath)
	}
	if got.Reason != "unrecognized library_sample_path format" {
		t.Fatalf("unexpected reason: %q", got.Reason)
	}
}

func TestBuildVirtualLimbsPath_UsesMode(t *testing.T) {
	legacy := BuildVirtualLimbsPath("MYPROJECT", "sample.wav", PathModeV11Legacy)
	if legacy != "/browser/samples/02_USER/LIMBS/MYPROJECT/sample.wav" {
		t.Fatalf("unexpected legacy path: %s", legacy)
	}

	newPath := BuildVirtualLimbsPath("MYPROJECT", "sample.wav", PathModeV13New)
	if newPath != "/tmp/S-4/browser/SAMPLES/02_USER/LIMBS/MYPROJECT/sample.wav" {
		t.Fatalf("unexpected new path: %s", newPath)
	}

	unknown := BuildVirtualLimbsPath("MYPROJECT", "sample.wav", PathModeUnknownAssumedV11)
	if unknown != "/browser/samples/02_USER/LIMBS/MYPROJECT/sample.wav" {
		t.Fatalf("unexpected unknown fallback path: %s", unknown)
	}
}
