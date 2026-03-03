package exporter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"limbs/internal/archive"
	"limbs/internal/config"
	"limbs/internal/resolver"
	"limbs/internal/scanner"
)

type MissingSample struct {
	ReferencePath string
	ResolvedPath  string
	Reason        string
}

type Collision struct {
	SourcePath   string
	AssignedName string
}

type Stats struct {
	ReferencesFound  int
	UniqueReferences int
	SamplesCopied    int
	MissingSamples   []MissingSample
	Collisions       []Collision
}

type Result struct {
	ProjectName              string
	ProjectVersion           string
	PathMode                 string
	MinFirmware              string
	VersionAssumptionWarning string
	OutputDir                string
	ZipPath                  string
	Stats                    Stats
}

type assignment struct {
	srcPath     string
	fileName    string
	virtualPath string
}

func Run(cfg config.Config) (Result, error) {
	projectDir := filepath.Dir(cfg.ProjectFile)
	projectDirName := filepath.Base(projectDir)
	if !strings.HasSuffix(projectDirName, ".s4project") {
		return Result{}, fmt.Errorf("project file must be inside a .s4project directory: %s", cfg.ProjectFile)
	}

	projectRaw, err := os.ReadFile(cfg.ProjectFile)
	if err != nil {
		return Result{}, err
	}

	var root any
	if err := json.Unmarshal(projectRaw, &root); err != nil {
		return Result{}, fmt.Errorf("invalid project json: %w", err)
	}

	projectVersion := readProjectVersion(root)
	pathMode, minFirmware, versionWarning := detectPathMode(projectVersion)
	firmwareToken := "fw" + minFirmware

	scan := scanner.CollectSampleRefs(root)
	stats := Stats{
		ReferencesFound:  len(scan.OrderedRefs),
		UniqueReferences: len(scan.UniqueRefs),
		MissingSamples:   make([]MissingSample, 0),
		Collisions:       make([]Collision, 0),
	}

	assignments := map[string]assignment{}
	renameMap := map[string]string{}
	nameCounts := map[string]int{}

	uniqueRefs := append([]string(nil), scan.UniqueRefs...)
	sort.Strings(uniqueRefs)

	for _, ref := range uniqueRefs {
		resolution := resolver.ResolveVirtualSamplePath(cfg.SamplesRoot, cfg.USBDrive, ref)
		if resolution.ResolvedPath == "" {
			stats.MissingSamples = append(stats.MissingSamples, MissingSample{
				ReferencePath: ref,
				ResolvedPath:  resolution.ResolvedPath,
				Reason:        resolution.Reason,
			})
			continue
		}

		if _, err := os.Stat(resolution.ResolvedPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				stats.MissingSamples = append(stats.MissingSamples, MissingSample{
					ReferencePath: ref,
					ResolvedPath:  resolution.ResolvedPath,
					Reason:        "file not found",
				})
				continue
			}
			return Result{}, err
		}

		if existing, exists := assignments[resolution.ResolvedPath]; exists {
			renameMap[ref] = existing.virtualPath
			continue
		}

		base := filepath.Base(resolution.ResolvedPath)
		count := nameCounts[base]
		nameCounts[base] = count + 1
		assigned := base
		if count > 0 {
			ext := filepath.Ext(base)
			stem := strings.TrimSuffix(base, ext)
			assigned = fmt.Sprintf("%s__%d%s", stem, count+1, ext)
			stats.Collisions = append(stats.Collisions, Collision{
				SourcePath:   resolution.ResolvedPath,
				AssignedName: assigned,
			})
		}

		virtual := resolver.BuildVirtualLimbsPath(cfg.ProjectName, assigned, pathMode)
		assignments[resolution.ResolvedPath] = assignment{
			srcPath:     resolution.ResolvedPath,
			fileName:    assigned,
			virtualPath: virtual,
		}
		renameMap[ref] = virtual
	}

	if len(stats.MissingSamples) > 0 && !cfg.AllowMissing {
		return Result{}, fmt.Errorf("missing %d sample(s); re-run with --allow-missing", len(stats.MissingSamples))
	}

	tempRoot, err := os.MkdirTemp("", "limbs-export-*")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(tempRoot)

	tempExportName := cfg.ProjectName + "_" + firmwareToken + "_export"
	tempExport := filepath.Join(tempRoot, tempExportName)
	projectOutBase := filepath.Join(tempExport, "PROJECTS")
	samplesOutBase := filepath.Join(tempExport, filepath.FromSlash(cfg.LimbsRoot))

	if err := os.MkdirAll(projectOutBase, 0o755); err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(samplesOutBase, 0o755); err != nil {
		return Result{}, err
	}

	exportProjectDirName := cfg.ProjectName + "-LIMB.s4project"
	projectOutDir := filepath.Join(projectOutBase, exportProjectDirName)
	if err := copyDir(projectDir, projectOutDir); err != nil {
		return Result{}, err
	}

	limbsProjectDir := filepath.Join(samplesOutBase, cfg.ProjectName)
	if err := os.MkdirAll(limbsProjectDir, 0o755); err != nil {
		return Result{}, err
	}

	zipFilename := cfg.ProjectName + "_" + firmwareToken + ".zip"
	importReadmePath := filepath.Join(tempExport, "IMPORT-README.md")
	if err := os.WriteFile(importReadmePath, []byte(buildImportReadme(cfg.ProjectName, projectVersion, minFirmware, versionWarning, zipFilename)), 0o644); err != nil {
		return Result{}, err
	}

	for _, a := range assignments {
		dst := filepath.Join(limbsProjectDir, a.fileName)
		if err := copyFile(a.srcPath, dst); err != nil {
			return Result{}, err
		}
		stats.SamplesCopied++
	}

	scanner.RewriteSampleRefs(root, renameMap)
	rewritten, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return Result{}, err
	}
	rewritten = append(rewritten, '\n')

	projectOutJSON := filepath.Join(projectOutDir, "project.json")
	if err := os.WriteFile(projectOutJSON, rewritten, 0o644); err != nil {
		return Result{}, err
	}

	finalOutDir := filepath.Join(cfg.DestRoot, tempExportName)
	if err := os.RemoveAll(finalOutDir); err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(cfg.DestRoot, 0o755); err != nil {
		return Result{}, err
	}
	if err := os.Rename(tempExport, finalOutDir); err != nil {
		return Result{}, err
	}

	res := Result{
		ProjectName:              cfg.ProjectName,
		ProjectVersion:           projectVersion,
		PathMode:                 string(pathMode),
		MinFirmware:              minFirmware,
		VersionAssumptionWarning: versionWarning,
		OutputDir:                finalOutDir,
		Stats:                    stats,
	}

	if cfg.Zip {
		zipPath := filepath.Join(cfg.DestRoot, zipFilename)
		if err := os.RemoveAll(zipPath); err != nil {
			return Result{}, err
		}
		if err := archive.ZipDir(finalOutDir, zipPath); err != nil {
			return Result{}, err
		}
		res.ZipPath = zipPath
	}

	return res, nil
}

func buildImportReadme(projectName, projectVersion, minFirmware, versionWarning, zipFilename string) string {
	assumption := ""
	if strings.TrimSpace(versionWarning) != "" {
		assumption = fmt.Sprintf("- Assumption: %s\n", versionWarning)
	}

	return fmt.Sprintf(`# IMPORT README

This export was generated by limbs for project "%[1]s".

## Compatibility

- Source project version: %[2]s
- Minimum firmware: %[3]s
%[4]s

## Import Steps (Torso S-4)

1. Unzip "%[5]s". (optional depending on whether --zip was used)
2. Put your Torso S-4 in USB Mass Storage mode and mount it on your computer.
3. Copy the exported project folder from:
   - PROJECTS/%[1]s-LIMB.s4project
   into the S-4:
   - PROJECTS/
4. Copy the exported sample folder from:
   - SAMPLES/LIMBS/%[1]s
   into the S-4 in the folder:
   - SAMPLES/LIMBS/ (you may need to create this directory, ensure it is uppercase "LIMBS").
5. Safely eject the S-4 and load "%[1]s-LIMB" on the device.
`, projectName, projectVersion, minFirmware, assumption, zipFilename)
}

func readProjectVersion(root any) string {
	obj, ok := root.(map[string]any)
	if !ok {
		return ""
	}
	if v, ok := obj["version"].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func detectPathMode(version string) (resolver.PathMode, string, string) {
	switch strings.ToLower(strings.TrimSpace(version)) {
	case "v11":
		return resolver.PathModeV11Legacy, "2.0.4+", ""
	case "v13":
		return resolver.PathModeV13New, "2.1.3+", ""
	default:
		warning := fmt.Sprintf("unrecognized project version %q; assuming v11-compatible path mode", version)
		return resolver.PathModeUnknownAssumedV11, "2.0.4+", warning
	}
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
