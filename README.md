# limbs - A cross-platform CLI exporter for Torso S-4 Projects

Limbs is a CLI tool that collects referenced samples, rewrites `project.json` sample paths, and creates a zip archive for easy sharing.

## Features

- Copies full project directory contents into `PROJECTS/<project>-LIMB.s4project`
- Copies referenced samples into a flat directory: `SAMPLES/LIMBS/<project>/`
- Rewrites every `library_sample_path` to:
  - `/browser/samples/02_USER/LIMBS/<project>/<filename>`
- Handles flattened filename collisions with suffixing:
  - `kick.wav`, `kick__2.wav`, `kick__3.wav`, ...
- Warns on missing samples by default (`--allow-missing=true`)
- Creates `<project>.zip` containing `PROJECTS/` and `SAMPLES/` at archive root
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
