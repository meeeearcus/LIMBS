package resolver

import (
	"path/filepath"
	"testing"
)

func TestResolveVirtualSamplePath_UserNamespace(t *testing.T) {
	root := filepath.Join("disk", "SAMPLES")

	tests := []struct {
		name string
		ref  string
	}{
		{
			name: "v2.0.4 format",
			ref:  "/browser/samples/02_USER/BOUNCES/DX/kick.wav",
		},
		{
			name: "v2.1 format",
			ref:  "/tmp/S-4/browser/SAMPLES/02_USER/BOUNCES/DX/kick.wav",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveVirtualSamplePath(root, "", tc.ref)
			want := filepath.Join(root, filepath.Join("BOUNCES", "DX", "kick.wav"))

			if got.ResolvedPath != want {
				t.Fatalf("expected resolved path %q, got %q", want, got.ResolvedPath)
			}
			if got.Namespace != NamespaceUser {
				t.Fatalf("expected namespace %q, got %q", NamespaceUser, got.Namespace)
			}
			if got.Reason != "" {
				t.Fatalf("expected empty reason, got %q", got.Reason)
			}
		})
	}
}

func TestResolveVirtualSamplePath_USBNamespace(t *testing.T) {
	usbRoot := filepath.Join("mnt", "usb")
	ref := "/tmp/S-4/browser/SAMPLES/03_USB_DRIVE/MyFolder/Sub/DRONE_1.wav"

	got := ResolveVirtualSamplePath("", usbRoot, ref)
	want := filepath.Join(usbRoot, filepath.Join("MyFolder", "Sub", "DRONE_1.wav"))
	if got.ResolvedPath != want {
		t.Fatalf("expected resolved path %q, got %q", want, got.ResolvedPath)
	}
	if got.Namespace != NamespaceUSBDrive {
		t.Fatalf("expected namespace %q, got %q", NamespaceUSBDrive, got.Namespace)
	}
	if got.Reason != "" {
		t.Fatalf("expected empty reason, got %q", got.Reason)
	}
}

func TestResolveVirtualSamplePath_USBRequiresFlag(t *testing.T) {
	ref := "/browser/samples/03_USB_DRIVE/INSTRUMENTS/DRONE_1.wav"
	got := ResolveVirtualSamplePath("", "", ref)

	if got.ResolvedPath != "" {
		t.Fatalf("expected empty resolved path, got %q", got.ResolvedPath)
	}
	if got.Namespace != NamespaceUSBDrive {
		t.Fatalf("expected namespace %q, got %q", NamespaceUSBDrive, got.Namespace)
	}
	if got.Reason != "usb drive path not configured (--usb-drive)" {
		t.Fatalf("unexpected reason: %q", got.Reason)
	}
}

func TestResolveVirtualSamplePath_Unrecognized(t *testing.T) {
	got := ResolveVirtualSamplePath("", "", "/tmp/not-s4/path.wav")

	if got.ResolvedPath != "" {
		t.Fatalf("expected empty resolved path, got %q", got.ResolvedPath)
	}
	if got.Namespace != NamespaceUnknown {
		t.Fatalf("expected namespace %q, got %q", NamespaceUnknown, got.Namespace)
	}
	if got.Reason != "unrecognized library_sample_path format" {
		t.Fatalf("unexpected reason: %q", got.Reason)
	}
}
