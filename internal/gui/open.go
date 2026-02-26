package gui

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenBrowser(url string) error {
	name, args, err := openCommand(url)
	if err != nil {
		return err
	}
	return exec.Command(name, args...).Start()
}

func openCommand(url string) (string, []string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "open", []string{url}, nil
	case "windows":
		return "cmd", []string{"/c", "start", "", url}, nil
	case "linux":
		return "xdg-open", []string{url}, nil
	default:
		return "", nil, fmt.Errorf("unsupported OS for auto-open: %s", runtime.GOOS)
	}
}
