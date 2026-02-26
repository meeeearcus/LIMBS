package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"limbs/internal/config"
	"limbs/internal/exporter"
	"limbs/internal/gui"
)

type guiConfig struct {
	Port   int
	NoOpen bool
}

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
		runExportCommand(os.Args[2:])
	case "gui":
		runGUICommand(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "error: unknown subcommand: %s\n", os.Args[1])
		printRootHelp(os.Stderr)
		os.Exit(2)
	}
}

func runExportCommand(args []string) {
	cfg, err := parseExportFlags(args)
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
		fmt.Printf("  warning: missing sample for ref %s (reason: %s, resolved: %s)\n", m.ReferencePath, m.Reason, m.ResolvedPath)
	}
}

func runGUICommand(args []string) {
	cfg, err := parseGUIFlags(args)
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	srv, err := gui.Start(gui.Options{Host: "127.0.0.1", Port: cfg.Port})
	if err != nil {
		fmt.Fprintf(os.Stderr, "gui failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("GUI listening at %s\n", srv.URL())
	if !cfg.NoOpen {
		if err := gui.OpenBrowser(srv.URL()); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not open browser automatically: %v\n", err)
			fmt.Fprintf(os.Stderr, "Open this URL manually: %s\n", srv.URL())
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func parseExportFlags(args []string) (config.Config, error) {
	var cfg config.Config
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		printExportHelp(os.Stderr)
	}

	fs.StringVar(&cfg.SourceMount, "source-mount", "/Volumes/S-4", "Base mount path for S-4 storage")
	fs.StringVar(&cfg.ProjectsRoot, "projects-root", "", "Override projects root (defaults to <source-mount>/PROJECTS)")
	fs.StringVar(&cfg.SamplesRoot, "samples-root", "", "Override samples root (defaults to <source-mount>/SAMPLES)")
	fs.StringVar(&cfg.USBDrive, "usb-drive", "", "Host path for S-4 03_USB_DRIVE sample namespace")
	fs.StringVar(&cfg.ProjectName, "project-name", "", "Project name without .s4project suffix")
	fs.StringVar(&cfg.ProjectFile, "project-file", "", "Path to project.json")
	fs.StringVar(&cfg.DestRoot, "dest-root", "", "Destination root directory for export (required)")
	fs.StringVar(&cfg.LimbsRoot, "limbs-root", "SAMPLES/LIMBS", "Exported limbs sample path (default: SAMPLES/LIMBS)")
	fs.BoolVar(&cfg.Zip, "zip", false, "Create zip archive after export")
	fs.BoolVar(&cfg.AllowMissing, "allow-missing", true, "Continue with warnings when a sample is missing")

	if err := fs.Parse(args); err != nil {
		return config.Config{}, err
	}
	return cfg, nil
}

func parseGUIFlags(args []string) (guiConfig, error) {
	cfg := guiConfig{}
	fs := flag.NewFlagSet("gui", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		printGUIHelp(os.Stderr)
	}

	fs.IntVar(&cfg.Port, "port", 0, "Port to bind (default: 0 for auto)")
	fs.BoolVar(&cfg.NoOpen, "no-open", false, "Do not auto-open browser")

	if err := fs.Parse(args); err != nil {
		return guiConfig{}, err
	}
	return cfg, nil
}

func printRootHelp(w io.Writer) {
	fmt.Fprintln(w, "LIMBS - An exporter for Torso S-4 Projects")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  export   Run project export from CLI")
	fmt.Fprintln(w, "  gui      Launch local web UI")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  limbs export [options]")
	fmt.Fprintln(w, "  limbs gui [--port <n>] [--no-open]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Run `limbs export --help` or `limbs gui --help` for command details.")
}

func printGUIHelp(w io.Writer) {
	fmt.Fprintln(w, "LIMBS GUI - Local web interface")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  limbs gui [--port <n>] [--no-open]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --port <n>    Port to bind (default: 0 for auto-assigned free port)")
	fmt.Fprintln(w, "  --no-open     Do not auto-open browser")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Notes:")
	fmt.Fprintln(w, "  GUI binds to localhost (127.0.0.1) only.")
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
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --source-mount <path>   Base mount path for S-4 storage (default: /Volumes/S-4)")
	fmt.Fprintln(w, "  --projects-root <path>  Override projects root (default: <source-mount>/PROJECTS)")
	fmt.Fprintln(w, "  --samples-root <path>   Override samples root (default: <source-mount>/SAMPLES)")
	fmt.Fprintln(w, "  --usb-drive <path>      Host path for S-4 03_USB_DRIVE sample namespace")
	fmt.Fprintln(w, "  --limbs-root <path>     Exported limbs sample path (default: SAMPLES/LIMBS)")
	fmt.Fprintln(w, "  --zip                   Create zip archive after export")
	fmt.Fprintln(w, "  --allow-missing         Continue with warnings when a sample is missing (default: true)")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Path Resolution:")
	fmt.Fprintln(w, "  library_sample_path is anchored on /browser/samples/ (case-insensitive)")
	fmt.Fprintln(w, "  02_USER/...       resolves under --samples-root")
	fmt.Fprintln(w, "  03_USB_DRIVE/...  resolves under --usb-drive")
	fmt.Fprintln(w, "  If USB refs exist and --usb-drive is not set, warnings are emitted;")
	fmt.Fprintln(w, "  use --allow-missing=false to fail instead.")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  limbs export --project-name MYPROJECT --dest-root \"$HOME/limbs-exports\" --zip")
	fmt.Fprintln(w, "  limbs export --project-file /Volumes/S-4/PROJECTS/MYPROJECT.s4project/project.json --dest-root \"$HOME/limbs-exports\"")
	fmt.Fprintln(w, "  limbs export --project-file /path/to/project.json --usb-drive /media/$USER/S4_USB --dest-root \"$HOME/limbs-exports\"")
}
