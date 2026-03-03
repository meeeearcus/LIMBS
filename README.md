# limbs - A cross-platform CLI exporter for Torso S-4 Projects

Limbs is a flexible CLI tool that collects referenced samples, rewrites `project.json` sample paths, and optionally creates a zip archive for easy sharing.

## WARNING: This application comes with no license or warrenty and is not affiliated with Torso Electronics. The author makes no claim of the safety of your data.
TL;DR: Works for me. Use at your own risk. Backup your projects before using this tool.

## Features

- Supports export from an S4 attached via USB Mass Storage Mode or a backup of the S4's filesystem.
- Copies full project directory contents into `PROJECTS/<project>-LIMB.s4project` in the destination folder
- Copies referenced samples and flattens them into one directory: `SAMPLES/LIMBS/<project>/` in the destination folder
- Rewrites every `library_sample_path` to:
  - Firmware-aware root based on top-level JSON `version`:
  - `v11`: `/browser/samples/02_USER/LIMBS/<project>/<filename>`
  - `v13`: `/tmp/S-4/browser/SAMPLES/02_USER/LIMBS/<project>/<filename>`
- Handles flattened sample filename collisions with suffixing:
  - `kick.wav`, `kick__2.wav`, `kick__3.wav`, ...
- Warns on missing samples by default (`--allow-missing=true`)
- With the `--zip` flag it creates `<project>_fw<min-firmware>.zip` containing `PROJECTS/` and `SAMPLES/` at archive root
- Adds `IMPORT-README.md` to each export with end-user import instructions

## Firmware Compatibility

LIMBS uses top-level `project.json` `version` to decide rewritten `library_sample_path` roots and minimum firmware guidance:

- `v11` -> rewrites to `/browser/samples/02_USER/LIMBS/...` -> minimum firmware `2.0.4+`
- `v13` -> rewrites to `/tmp/S-4/browser/SAMPLES/02_USER/LIMBS/...` -> minimum firmware `2.1.3+`
- Unknown versions -> warning + assume v11-compatible rewrite mode (`2.0.4+`)\
- Note: "v11" project files are automatically rewritten to the "v13" format on the S4 when loaded on 2.1.3 firmware.

Export names include assumed firmware it was exported from:

- Folder: `<project>_fw<min-firmware>_export`
- Zip: `<project>_fw<min-firmware>.zip`

## Sample Resolution Namespaces

Project file `library_sample_path` is resolved from the `/browser/samples/` anchor (legacy or new firmware path shapes are both accepted for input):

- `02_USER/...` -> resolved under `--samples-root` (or `<source-mount>/SAMPLES` by default)
- `03_USB_DRIVE/...` -> resolved under `--usb-drive`

For `03_USB_DRIVE`, LIMBS treats the tail after `03_USB_DRIVE/` as authoritative and joins it directly to `--usb-drive`. In short, use the external storage you use with your S4 as the `--usb-drive` value, or a directory that mirrors that structure.

Example source resolution:

- JSON ref: `/tmp/S-4/browser/SAMPLES/03_USB_DRIVE/SAMPLES/INSTRUMENTS/EXAMPLE.wav`
- `--usb-drive /Volumes/MY_USB`
- Resolved source path: `/Volumes/MY_USB/SAMPLES/INSTRUMENTS/EXAMPLE.wav`

## Install

### Option 1: Download a release binary (recommended)

1. Open the GitHub Releases page for this repo.
2. Download the archive for your OS:
   - macOS: `limbs_<version>_darwin_<arch>.tar.gz`
   - Linux: `limbs_<version>_linux_<arch>.tar.gz`
   - Windows: `limbs_<version>_windows_amd64.zip`
3. Extract it and run:
   - macOS/Linux: `./limbs`
   - Windows (PowerShell): `.\limbs.exe`

### Option 2: Build from source

Prerequisite: Go 1.22+

```bash
git clone https://github.com/meeeearcus/LIMBS.git
cd Limbs
go build -o limbs ./cmd/limbs
```

Windows (PowerShell):

```powershell
go build -o limbs.exe .\cmd\limbs
```

## Options
Required:
  `--dest-root <dir>`
  Exactly one of:
    `--project-name <name>`
    `--project-file <path>`

Basic Options:
  `--project-name <name>`   Project name without .s4project suffix
  `--project-file <path>`   Path to project.json
  `--dest-root <dir>`       Destination root directory for export
  `--zip`                   Create zip archive after export
  `--allow-missing`         Continue export with warnings when samples are missing (default: true)

Advanced Options:
  `--source-mount <path>`   Base S-4 mount path; default may apply on macOS/Linux
                           On Windows set explicitly (e.g. E:\) unless both roots are set
  `--projects-root <path>`  Full PROJECTS path override
  `--samples-root <path>`   Full SAMPLES path override 
  `--usb-drive <path>`      Host root for S-4 03_USB_DRIVE namespace
  `--limbs-root <path>`     Export-relative sample path (default: `SAMPLES/LIMBS`)
                           Effective location: `<dest-root>/<project>_..._export/<limbs-root>/<project>/`

## Usage (macOS/Linux)
First attach your S-4 in USB mass storage mode to your computer.

```bash
limbs export \
  --project-name PROJECT_1 \
  --dest-root /path/to/exports \
  --zip
```

Or:

```bash
limbs export \
  --project-file /Volumes/S-4/PROJECTS/PROJECT_1.s4project/project.json \
  --dest-root /path/to/exports
```

## Cross-platform usage for attaching the S-4 data sources

### macOS (default mount)

Default source mount:
- `/Volumes/S-4`

### Linux (common mount points)

Supported common source mounts:
- `/media/$USER/S-4`
- `/run/media/$USER/S-4`

### Windows (USB drive letter)

No default, set `--source-mount` manually.

### Setting --source-mount

`--source-mount` is the base path that contains both `PROJECTS` and `SAMPLES` on your mounted Torso S-4 drive, or even a copy on your computer.
On Windows, provide it explicitly unless you pass both `--projects-root` and `--samples-root`.

`--projects-root` and `--samples-root` are full path overrides. LIMBS does not append `PROJECTS` or `SAMPLES` to values you provide.

## USB sample references

If your project contains references to `03_USB_DRIVE`, provide the mounted root of that USB device with `--usb-drive`.

Example:

```bash
./limbs export \
  --project-file /Volumes/S-4/PROJECTS/MYPROJECT.s4project/project.json \
  --usb-drive /Volumes/MY_USB \
  --dest-root "$HOME/limbs-exports"
```

## More Examples
Assume default mounts, provide a destination and create a zip:
  `limbs export --project-name MYPROJECT --dest-root "$HOME/limbs-exports" --zip`
Provide a specific project file location and destination:
  `limbs export --project-file /Volumes/S-4/PROJECTS/MYPROJECT.s4project/project.json --dest-root "$HOME/limbs-exports"`
Provide a project name, project and sample source locations by setting projects-root and samples-root, and destination:
  `limbs export --project-name MYPROJECT --projects-root /path/to/S-4/PROJECTS --samples-root /path/to/S-4/SAMPLES --dest-root "$HOME/limbs-exports"`
Provide a project file, USB sample drive and destination root:
  `limbs export --project-file /Volumes/S-4/PROJECTS/MYPROJECT.s4project/project.json --usb-drive /Volumes/MY_USB --dest-root "$HOME/limbs-exports"`
Provide project name, its a source location, destination , and create a zip:
  `limbs export --project-name MYPROJECT --source-mount "/media/$USER/S-4" --dest-root "$HOME/limbs-exports" --zip`
Provide project name, its a source location, destination , and create a zip (Windows):
  `limbs export --project-name MYPROJECT --source-mount E:\ --dest-root "$HOME\Desktop\limbs-exports" --zip`