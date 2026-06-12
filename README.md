# file-sync-tool

A file synchronization tool with binary content change detection. Syncs files from a source directory to a destination directory using SHA-256 hashing for reliable change detection.

## Features

- **Binary-safe change detection** — detects changes in any file type using SHA-256 checksums
- **Metadata-first optimization** — checks file size + mtime first; only hashes files when metadata differs
- **Streaming hashing** — 64KB buffer, never loads entire files into memory
- **State persistence** — stores file records in `~/.file-sync-tool/state.json` between runs
- **Dry-run mode** — preview changes without copying
- **Delete propagation** — optionally remove destination files that no longer exist in source
- **Watch mode** — live sync via filesystem events (fsnotify)
- **Poll mode** — periodic sync at a configurable interval

## Prerequisites

- Go 1.21 or later

## Installation

### From source

```bash
git clone git@github.com:garenyondem/file-sync-tool.git
cd file-sync-tool
go build -o file-sync-tool .
```

Optionally, move the binary to a directory in your PATH:

```bash
mv file-sync-tool ~/.local/bin/
```

Or install directly:

```bash
go install github.com/garenyondem/file-sync-tool@latest
```

## Usage

```
file-sync-tool [flags] <source> <destination>
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--dry-run` | `false` | Show what would change without copying anything |
| `--interval` | `30s` | Poll interval between syncs. Set to `0s` to run once and exit |
| `--watch` | `false` | Watch source directory for changes and sync immediately (uses fsnotify) |
| `--delete` | `false` | Remove files in destination that no longer exist in source |
| `--state` | `~/.file-sync-tool/state.json` | Path to the state file |

### Examples

**One-shot sync:**

```bash
file-sync-tool --interval 0s ~/Documents /backup/Documents
```

**Continuous polling (every 10 seconds):**

```bash
file-sync-tool --interval 10s ~/Documents /backup/Documents
```

**Watch mode — sync on every filesystem event:**

```bash
file-sync-tool --watch ~/Projects /backup/Projects
```

**Dry-run to preview changes:**

```bash
file-sync-tool --dry-run --interval 0s ~/Documents /backup/Documents
```

**Sync with delete propagation:**

```bash
file-sync-tool --delete --interval 0s ~/Documents /backup/Documents
```

## How change detection works

1. On first run, the tool scans every file in the source directory, computes its SHA-256 checksum, and stores the path, size, mtime, and checksum in the state file.
2. On subsequent runs, it checks each file's **size and modification time** first. If both match the stored values, the file is skipped — no read or hash needed.
3. If either size or mtime differs, the file is read and hashed. If the checksum matches the stored value, only the stored mtime is updated (content is identical, no copy needed).
4. If the checksum also differs, the file is copied to the destination.

This means unchanged files are never read, and only files whose metadata changed get hashed. True binary content verification happens only when necessary.

## State file

The state file is stored at `~/.file-sync-tool/state.json` by default. It tracks every synced file:

```json
{
  "source": "/home/user/Documents",
  "dest": "/backup/Documents",
  "files": {
    "notes.txt": {
      "rel_path": "notes.txt",
      "size": 1423,
      "mod_time": 1718054400000000000,
      "checksum": "e3b0c44298fc1c149afbf4c8996fb924..."
    }
  },
  "updated_at": 1718054400000000000
}
```

You can specify a custom state path with `--state` (useful for syncing multiple directory pairs).

## Testing

Run all tests across all packages:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run tests from a specific package:

```bash
go test -v ./hash/
go test -v ./sync/
```

Run the integration test only (tests the full sync cycle end-to-end):

```bash
go test -v -run TestIntegrationSync
```

Tests use `t.TempDir()` for temporary directories and do not touch your filesystem outside temp space.

## Project structure

```
file-sync-tool/
├── main.go          # CLI entry point, flag parsing, orchestration
├── types/types.go   # Shared data types (FileRecord, SyncState, Change)
├── hash/hasher.go   # SHA-256 streaming hasher (64KB buffer)
├── scan/scanner.go  # Directory walker collecting file records
├── sync/
│   ├── state.go     # State file load/save/apply helpers
│   └── syncer.go    # Change detection + file copy/delete logic
├── watch/watcher.go # fsnotify watcher with debounce
└── go.mod
```
