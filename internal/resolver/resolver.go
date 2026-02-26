package resolver

import (
	"path/filepath"
	"strings"
)

const browserAnchor = "/browser/samples/"

const (
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
	normalized := strings.ReplaceAll(virtualPath, "\\", "/")
	anchorIndex := strings.Index(strings.ToLower(normalized), browserAnchor)
	if anchorIndex == -1 {
		return Resolution{
			Namespace: NamespaceUnknown,
			Reason:    "unrecognized library_sample_path format",
		}
	}

	trimmed := normalized[anchorIndex+len(browserAnchor):]
	trimmed = strings.TrimPrefix(trimmed, "/")
	if trimmed == "" {
		return Resolution{
			Namespace: NamespaceUnknown,
			Reason:    "unrecognized library_sample_path format",
		}
	}

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

func BuildVirtualLimbsPath(projectName, filename string) string {
	return "/browser/samples/02_USER/LIMBS/" + projectName + "/" + filename
}
