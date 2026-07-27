# FocalSort

FocalSort analysiert JPG/JPEG-Dateien und benennt Bilder deterministisch um.

## Verhalten

- verarbeitet Dateien in stabil sortierter Reihenfolge (lexikografisch nach Pfad)
- berücksichtigt `.jpg` und `.jpeg` (case-insensitiv)
- erzeugt Dateinamen im Format `YYYYMMDD_HHMMSS_<cameraid>_<quality>_<checksum>.jpg`
- `cameraid`: kompakter Kamera-Identifier aus EXIF `Make` + `Model` (z. B. `SONA2F`)
- `quality`: kompakter Schärfe-Code `Q00..Q99` (Sobel-Kantenstärke via `bild` + log-Skalierung)
- löst Namenskollisionen deterministisch mit `_0001`, `_0002`, ...
- unterstützt Dry-Run ohne Dateiänderungen

## CLI

```bash
focalsort --import-folder ./photos
```

Wichtige Flags:

- `--import-folder, -i` (pflichtig): Quellordner
- `--recursive` (default: `true`): rekursiv durchsuchen
- `--dry-run` (default: `false`): nur geplante Umbenennungen ausgeben
- `--fallback-to-modtime` (default: `false`): `mtime` verwenden, wenn EXIF-Zeit fehlt
- `--checksum-length` (default: `16`, Bereich `1..40`): Länge des Checksum-Suffixes
- `--tui, -t`: Fullscreen Terminal-UI (Bubble Tea, Synthwave) aktivieren (`q`=quit, `c`=logs löschen)

## Beispiele

```bash
# nur planen, nichts ändern
focalsort -i ./photos --dry-run

# nur oberstes Verzeichnis auswerten
focalsort -i ./photos --recursive=false

# fehlende EXIF-Zeit mit mtime ersetzen
focalsort -i ./photos --fallback-to-modtime
```
