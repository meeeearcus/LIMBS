package main

import (
	"flag"
	"fmt"
	"os"

	"limbs/internal/config"
	"limbs/internal/exporter"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "export" {
		printUsage()
		os.Exit(2)
	}

	cfg, err := parseExportFlags(os.Args[2:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	resolved, err := config.Resolve(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	result, err := exporter.Run(resolved)
	if err != nil {
		fmt.Fprintf(os.Stderr, "export failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Export complete\n")
	fmt.Printf("- Project: %s\n", result.ProjectName)
	fmt.Printf("- Output directory: %s\n", result.OutputDir)
	if result.ZipPath != "" {
		fmt.Printf("- Zip archive: %s\n", result.ZipPath)
	}
	fmt.Printf("- References found: %d\n", result.Stats.ReferencesFound)
	fmt.Printf("- Unique references: %d\n", result.Stats.UniqueReferences)
	fmt.Printf("- Samples copied: %d\n", result.Stats.SamplesCopied)
	fmt.Printf("- Missing samples: %d\n", len(result.Stats.MissingSamples))
	fmt.Printf("- Filename collisions: %d\n", len(result.Stats.Collisions))

	for _, c := range result.Stats.Collisions {
		fmt.Printf("  collision: %s -> %s\n", c.SourcePath, c.AssignedName)
	}
	for _, m := range result.Stats.MissingSamples {
		fmt.Printf("  warning: missing sample for ref %s (resolved: %s)\n", m.ReferencePath, m.ResolvedPath)
	}
}

func parseExportFlags(args []string) (config.Config, error) {
	var cfg config.Config
	fs := flag.NewFlagSet("export", flag.ContinueOnError)

	fs.StringVar(&cfg.SourceMount, "source-mount", "/Volumes/S-4", "Base mount path for S-4 storage")
	fs.StringVar(&cfg.ProjectsRoot, "projects-root", "", "Override projects root (defaults to <source-mount>/PROJECTS)")
	fs.StringVar(&cfg.SamplesRoot, "samples-root", "", "Override samples root (defaults to <source-mount>/SAMPLES)")
	fs.StringVar(&cfg.ProjectName, "project-name", "", "Project name without .s4project suffix")
	fs.StringVar(&cfg.ProjectFile, "project-file", "", "Path to project.json")
	fs.StringVar(&cfg.DestRoot, "dest-root", "", "Destination root directory for export (required)")
	fs.StringVar(&cfg.LimbsRoot, "limbs-root", "SAMPLES/LIMBS", "Exported limbs root path inside output")
	fs.BoolVar(&cfg.Zip, "zip", false, "Create zip archive after export")
	fs.BoolVar(&cfg.AllowMissing, "allow-missing", true, "Continue with warnings when a sample is missing")

	if err := fs.Parse(args); err != nil {
		return config.Config{}, err
	}
	return cfg, nil
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  limbs export --project-name <name> --dest-root <dir> [options]")
	fmt.Fprintln(os.Stderr, "  limbs export --project-file <path> --dest-root <dir> [options]")
}
