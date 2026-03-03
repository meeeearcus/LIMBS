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

	NamespaceUser     = "02_USER"
	NamespaceUSBDrive = "03_USB_DRIVE"
	NamespaceUnknown  = "unknown"
)

type Resolution struct {
	ResolvedPath string
	Namespace    string
	Reason       string
}

func ResolveVirtualSamplePath(samplesRoot, usbDrive, virtualPath string) Resolution {
	normalized := strings.ReplaceAll(strings.TrimSpace(virtualPath), "\\", "/")
	trimmed, ok := trimAnyPrefixFold(normalized, []string{
		legacyBrowserPrefix,
		newBrowserPrefix,
	})
	if !ok {
		return Resolution{
			Namespace: NamespaceUnknown,
			Reason:    "unrecognized library_sample_path format",
		}
	}

	trimmed = strings.TrimPrefix(trimmed, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		return Resolution{
			Namespace: NamespaceUnknown,
			Reason:    "unrecognized library_sample_path format",
		}
	}

	namespace := strings.ToUpper(parts[0])
	relativePath := strings.TrimPrefix(parts[1], "/")
	if strings.TrimSpace(relativePath) == "" {
		return Resolution{
			Namespace: NamespaceUnknown,
			Reason:    "unrecognized library_sample_path format",
		}
	}

	switch namespace {
	case NamespaceUser:
		return Resolution{
			Namespace:    NamespaceUser,
			ResolvedPath: filepath.Join(samplesRoot, filepath.FromSlash(relativePath)),
		}
	case NamespaceUSBDrive:
		if strings.TrimSpace(usbDrive) == "" {
			return Resolution{
				Namespace: NamespaceUSBDrive,
				Reason:    "usb drive path not configured (--usb-drive)",
			}
		}
		return Resolution{
			Namespace:    NamespaceUSBDrive,
			ResolvedPath: filepath.Join(usbDrive, filepath.FromSlash(relativePath)),
		}
	default:
		return Resolution{
			Namespace: NamespaceUnknown,
			Reason:    "unrecognized library_sample_path format",
		}
	}
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
