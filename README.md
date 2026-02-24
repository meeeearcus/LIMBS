# limbs - A cross-platform CLI exporter for Torso S-4 Projects

Limbs is a CLI tool that collects referenced samples, rewrites `project.json` sample paths, and creates a zip archive for easy sharing.

## Known Issues

- Issue uncovered with different directory structure across firmware versions (e.g., 2.0.4 vs 2.1.*) causing issues with finding samples and rewriting project.json paths.
- Does not handle sample content from USB drives

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


## Features

- Copies full project directory contents into `PROJECTS/<project>-LIMB.s4project` in the destination folder
- Copies referenced samples into a flat directory: `SAMPLES/LIMBS/<project>/` in the destination folder
- Rewrites every `library_sample_path` to:
  - `/browser/samples/02_USER/LIMBS/<project>/<filename>`
- Handles flattened filename collisions with suffixing:
  - `kick.wav`, `kick__2.wav`, `kick__3.wav`, ...
- Warns on missing samples by default (`--allow-missing=true`)
- With the `--zip` flag it creates `<project>.zip` containing `PROJECTS/` and `SAMPLES/` at archive root
- Adds `IMPORT-README.md` to each export with end-user import instructions

## Defaults

- `--source-mount`: `/Volumes/S-4`
- `--projects-root`: `<source-mount>/PROJECTS`
- `--samples-root`: `<source-mount>/SAMPLES`
- `--limbs-root`: `SAMPLES/LIMBS`

## Usage (macOS assumed)

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

## Cross-platform usage

`--source-mount` is the base path that contains both `PROJECTS` and `SAMPLES` on your mounted Torso S-4 drive.

### macOS (default mount)

Default source mount:
- `/Volumes/S-4`

Example:

```bash
./limbs export \
  --project-name ONEBEATWK2 \
  --dest-root "$HOME/Desktop/limbs-exports" \
  --zip
```

### Linux (common mount points)

Common source mounts:
- `/media/$USER/S-4`
- `/run/media/$USER/S-4`

Example:

```bash
./limbs export \
  --project-name ONEBEATWK2 \
  --source-mount "/media/$USER/S-4" \
  --dest-root "$HOME/limbs-exports" \
  --zip
```

### Windows (USB drive letter)

Common source mounts:
- `E:\`
- `F:\`

Example (PowerShell):

```powershell
.\limbs.exe export `
  --project-name ONEBEATWK2 `
  --source-mount E:\ `
  --dest-root "$HOME\Desktop\limbs-exports" `
  --zip
```
