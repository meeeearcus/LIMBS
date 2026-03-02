package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"limbs/internal/config"
	"limbs/internal/exporter"
)

func main() {
	if len(os.Args) < 2 {
		printRootHelp(os.Stdout)
		return
	}

	switch os.Args[1] {
	case "-h", "--help", "help":
		printRootHelp(os.Stdout)
		return
	case "export":
		// continue
	default:
		fmt.Fprintf(os.Stderr, "error: unknown subcommand: %s\n", os.Args[1])
		printRootHelp(os.Stderr)
		os.Exit(2)
	}

	cfg, err := parseExportFlags(os.Args[2:])
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	resolved, err := config.Resolve(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		printExportHelp(os.Stderr)
		os.Exit(2)
	}

	result, err := exporter.Run(resolved)
	if err != nil {
		fmt.Fprintf(os.Stderr, "export failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Export complete\n")
	fmt.Printf("- Project: %s\n", result.ProjectName)
	fmt.Printf("- Project version: %s\n", result.ProjectVersion)
	fmt.Printf("- Path mode: %s\n", result.PathMode)
	fmt.Printf("- Minimum firmware: %s\n", result.MinFirmware)
	if result.VersionAssumptionWarning != "" {
		fmt.Printf("- Version assumption: %s\n", result.VersionAssumptionWarning)
	}
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
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		printExportHelp(os.Stderr)
	}

	fs.StringVar(&cfg.SourceMount, "source-mount", "", "Base mount path for S-4 storage (OS-aware default when omitted)")
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

func printRootHelp(w io.Writer) {
	fmt.Fprintln(w, "LIMBS - An exporter for Torso S-4 Projects")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Command:")
	fmt.Fprintln(w, "  export   Export a project and rewrite sample paths")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  limbs export [options]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Run `limbs export --help` for full flag details and examples.")
}

func printExportHelp(w io.Writer) {
	fmt.Fprintln(w, "LIMBS - An exporter for Torso S-4 Projects")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  limbs export --project-name <name> --dest-root <dir> [options]")
	fmt.Fprintln(w, "  limbs export --project-file <path> --dest-root <dir> [options]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Required:")
	fmt.Fprintln(w, "  --dest-root <dir>")
	fmt.Fprintln(w, "  Exactly one of:")
	fmt.Fprintln(w, "    --project-name <name>")
	fmt.Fprintln(w, "    --project-file <path>")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Basic Options:")
	fmt.Fprintln(w, "  --project-name <name>   Project name without .s4project suffix")
	fmt.Fprintln(w, "  --project-file <path>   Path to project.json")
	fmt.Fprintln(w, "  --dest-root <dir>       Destination root directory for export")
	fmt.Fprintln(w, "  --zip                   Create zip archive after export")
	fmt.Fprintln(w, "  --allow-missing         Continue with warnings when samples are missing (default: true)")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Advanced Options:")
	fmt.Fprintln(w, "  --source-mount <path>   Base mount path (OS-aware default when omitted)")
	fmt.Fprintln(w, "  --projects-root <path>  Override projects root (default: <source-mount>/PROJECTS)")
	fmt.Fprintln(w, "  --samples-root <path>   Override samples root (default: <source-mount>/SAMPLES)")
	fmt.Fprintln(w, "  --limbs-root <path>     Exported limbs sample path (default: SAMPLES/LIMBS)")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Firmware-Aware Rewrite:")
	fmt.Fprintln(w, "  v11 -> /browser/samples/02_USER/LIMBS/...  (minimum firmware: 2.0.4+)")
	fmt.Fprintln(w, "  v13 -> /tmp/S-4/browser/SAMPLES/02_USER/LIMBS/...  (minimum firmware: 2.1.3+)")
	fmt.Fprintln(w, "  Unknown version -> warning + v11-compatible rewrite mode")
	fmt.Fprintln(w, "  Output names include firmware suffix (e.g. MYPROJECT_fw2.1.3+_export)")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Platform Mount Defaults:")
	fmt.Fprintln(w, "  macOS:   /Volumes/S-4")
	fmt.Fprintln(w, "  Linux:   /media/$USER/S-4, fallback /run/media/$USER/S-4")
	fmt.Fprintln(w, "  Windows: no fixed default; set --source-mount (e.g. E:\\)")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  limbs export --project-name MYPROJECT --dest-root \"$HOME/limbs-exports\" --zip")
	fmt.Fprintln(w, "  limbs export --project-file /Volumes/S-4/PROJECTS/MYPROJECT.s4project/project.json --dest-root \"$HOME/limbs-exports\"")
	fmt.Fprintln(w, "  limbs export --project-name MYPROJECT --source-mount \"/media/$USER/S-4\" --dest-root \"$HOME/limbs-exports\" --zip")
	fmt.Fprintln(w, "  limbs export --project-name MYPROJECT --source-mount E:\\ --dest-root \"$HOME\\Desktop\\limbs-exports\" --zip")
}
