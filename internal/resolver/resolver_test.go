package resolver

import (
	"path/filepath"
	"testing"
)

func TestResolveVirtualSamplePath_LegacyAndNewRoots(t *testing.T) {
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
			got := ResolveVirtualSamplePath(samplesRoot, tc.ref)
			want := filepath.Join(samplesRoot, "BOUNCES", "DX", "kick.wav")
			if got != want {
				t.Fatalf("expected %q, got %q", want, got)
			}
		})
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
