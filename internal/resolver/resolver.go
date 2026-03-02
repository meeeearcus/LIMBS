package resolver

import (
	"path/filepath"
	"strings"
)

const (
	legacyBrowserPrefix = "/browser/samples/"
	newBrowserPrefix    = "/tmp/S-4/browser/samples/"
)

type PathMode string

const (
	PathModeV11Legacy         PathMode = "v11-legacy"
	PathModeV13New            PathMode = "v13-new"
	PathModeUnknownAssumedV11 PathMode = "unknown-assumed-v11"
)

func ResolveVirtualSamplePath(samplesRoot, virtualPath string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(virtualPath), "\\", "/")
	trimmed, ok := trimAnyPrefixFold(normalized, []string{
		legacyBrowserPrefix,
		newBrowserPrefix,
	})
	if !ok {
		return ""
	}

	trimmed = strings.TrimPrefix(trimmed, "/")

	// Current S-4 exports often include this synthetic root marker.
	trimmed = strings.TrimPrefix(trimmed, "02_USER/")
	trimmed = strings.TrimPrefix(trimmed, "/")

	if trimmed == "" {
		return ""
	}
	return filepath.Join(samplesRoot, filepath.FromSlash(trimmed))
}

func BuildVirtualLimbsPath(projectName, filename string, mode PathMode) string {
	switch mode {
	case PathModeV13New:
		return "/tmp/S-4/browser/SAMPLES/02_USER/LIMBS/" + projectName + "/" + filename
	default:
		return "/browser/samples/02_USER/LIMBS/" + projectName + "/" + filename
	}
}

func trimAnyPrefixFold(s string, prefixes []string) (string, bool) {
	for _, prefix := range prefixes {
		if out, ok := trimPrefixFold(s, prefix); ok {
			return out, true
		}
	}
	return "", false
}

func trimPrefixFold(s, prefix string) (string, bool) {
	if len(s) < len(prefix) {
		return "", false
	}
	if strings.EqualFold(s[:len(prefix)], prefix) {
		return s[len(prefix):], true
	}
	return "", false
}
