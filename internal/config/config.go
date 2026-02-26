package config

import (
	"errors"
	"fmt"
	"path/filepath"
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
	if strings.TrimSpace(cfg.SourceMount) == "" {
		cfg.SourceMount = "/Volumes/S-4"
	}
	if strings.TrimSpace(cfg.ProjectsRoot) == "" {
		cfg.ProjectsRoot = filepath.Join(cfg.SourceMount, "PROJECTS")
	}
	if strings.TrimSpace(cfg.SamplesRoot) == "" {
		cfg.SamplesRoot = filepath.Join(cfg.SourceMount, "SAMPLES")
	}
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
