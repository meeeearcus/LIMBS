package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Config struct {
	SourceMount  string
	ProjectsRoot string
	SamplesRoot  string
	USBDrive     string
	ProjectName  string
	ProjectFile  string
	DestRoot     string
	LimbsRoot    string
	Zip          bool
	AllowMissing bool
}

func Resolve(cfg Config) (Config, error) {
	if strings.TrimSpace(cfg.DestRoot) == "" {
		return Config{}, errors.New("--dest-root is required")
	}
	sourceMount := strings.TrimSpace(cfg.SourceMount)
	projectsRoot := strings.TrimSpace(cfg.ProjectsRoot)
	samplesRoot := strings.TrimSpace(cfg.SamplesRoot)

	if sourceMount == "" && projectsRoot == "" && samplesRoot == "" {
		sourceMount = defaultSourceMount()
	}
	cfg.SourceMount = sourceMount

	if projectsRoot == "" && sourceMount != "" {
		projectsRoot = filepath.Join(sourceMount, "PROJECTS")
	}
	if samplesRoot == "" && sourceMount != "" {
		samplesRoot = filepath.Join(sourceMount, "SAMPLES")
	}
	cfg.ProjectsRoot = projectsRoot
	cfg.SamplesRoot = samplesRoot

	if strings.TrimSpace(cfg.LimbsRoot) == "" {
		cfg.LimbsRoot = "SAMPLES/LIMBS"
	}

	projectFile := strings.TrimSpace(cfg.ProjectFile)
	projectName := strings.TrimSpace(cfg.ProjectName)
	if projectFile == "" && projectName == "" {
		return Config{}, errors.New("either --project-name or --project-file must be provided")
	}
	if projectFile != "" && projectName != "" {
		return Config{}, errors.New("provide only one of --project-name or --project-file")
	}

	if projectFile == "" {
		if strings.TrimSpace(cfg.ProjectsRoot) == "" {
			return Config{}, errors.New("--projects-root could not be derived; provide --source-mount or --projects-root")
		}
		cfg.ProjectFile = filepath.Join(cfg.ProjectsRoot, projectName+".s4project", "project.json")
	} else {
		cfg.ProjectFile = projectFile
	}

	if projectName == "" {
		baseDir := filepath.Base(filepath.Dir(cfg.ProjectFile))
		projectName = strings.TrimSuffix(baseDir, ".s4project")
		if strings.TrimSpace(projectName) == "" || projectName == baseDir {
			return Config{}, fmt.Errorf("could not infer project name from path: %s", cfg.ProjectFile)
		}
	}
	cfg.ProjectName = projectName
	return cfg, nil
}

func defaultSourceMount() string {
	switch runtime.GOOS {
	case "darwin":
		return "/Volumes/S-4"
	case "linux":
		user := strings.TrimSpace(os.Getenv("USER"))
		if user == "" {
			return "/media/S-4"
		}
		primary := filepath.Join("/media", user, "S-4")
		if pathExists(primary) {
			return primary
		}
		fallback := filepath.Join("/run/media", user, "S-4")
		if pathExists(fallback) {
			return fallback
		}
		return primary
	default:
		return ""
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
