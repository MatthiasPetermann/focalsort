# FocalSort

FocalSort analyzes JPG/JPEG files and renames them deterministically.

## Behavior

- processes files in stable, lexicographic path order
- supports `.jpg` and `.jpeg` files case-insensitively
- produces names in the format `YYYYMMDD_HHMMSS_<cameraid>_<checksum>.jpg`
- `cameraid`: compact camera identifier derived from EXIF `Make` and `Model` (for example, `SONA2F`)
- files with unparseable EXIF use `19700101_000000` and camera ID `UNK`
- resolves filename collisions deterministically with `_0001`, `_0002`, ...
- does not detect content duplicates: identical images are retained and receive a collision suffix
- supports dry runs without modifying files

## CLI

```bash
focalsort --import-folder ./photos
```

Key flags:

- `--import-folder, -i` (required): source folder
- `--recursive` (default: `true`): search folders recursively
- `--dry-run` (default: `false`): print planned renames without changing files
- `--fallback-to-modtime` (default: `false`): use the file modification time when the EXIF timestamp is missing
- `--checksum-length` (default: `16`, range `1..40`): checksum suffix length
- `--tui, -t`: enable the full-screen terminal UI (`q`/`Ctrl+C` cancels processing; `c` clears logs)

## Build for Linux

```bash
make linux
scp dist/focalsort-linux-amd64 user@server:/target/path/
```

The target builds a statically linked Linux binary without CGO dependencies. For ARM64 servers:

```bash
make linux GOARCH=arm64
```

## Examples

```bash
# Preview changes without renaming files
focalsort -i ./photos --dry-run

# Process only the top-level folder
focalsort -i ./photos --recursive=false

# Use mtime when the EXIF timestamp is missing
focalsort -i ./photos --fallback-to-modtime
```
