package resolver

import (
	"path/filepath"
	"strings"
)

const BrowserPrefix = "/browser/samples/"

func ResolveVirtualSamplePath(samplesRoot, virtualPath string) string {
	if !strings.HasPrefix(virtualPath, BrowserPrefix) {
		return ""
	}

	trimmed := strings.TrimPrefix(virtualPath, BrowserPrefix)
	trimmed = strings.TrimPrefix(trimmed, "/")

	// Current S-4 exports often include this synthetic root marker.
	trimmed = strings.TrimPrefix(trimmed, "02_USER/")
	trimmed = strings.TrimPrefix(trimmed, "/")

	if trimmed == "" {
		return ""
	}
	return filepath.Join(samplesRoot, filepath.FromSlash(trimmed))
}

func BuildVirtualLimbsPath(projectName, filename string) string {
	return "/browser/samples/02_USER/LIMBS/" + projectName + "/" + filename
}
