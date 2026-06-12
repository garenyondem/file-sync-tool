package sync

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/user/file-sync-tool/types"
)

func ApplyChanges(changes []types.Change, sourceDir, destDir string, dryRun bool) error {
	for _, c := range changes {
		srcPath := filepath.Join(sourceDir, c.RelPath)
		dstPath := filepath.Join(destDir, c.RelPath)

		switch c.Type {
		case types.New, types.Modified:
			if dryRun {
				fmt.Printf("[DRY-RUN] copy %s\n", c.RelPath)
				continue
			}
			if err := copyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("copy %s: %w", c.RelPath, err)
			}
			fmt.Printf("copied %s\n", c.RelPath)

		case types.Deleted:
			if dryRun {
				fmt.Printf("[DRY-RUN] delete %s\n", c.RelPath)
				continue
			}
			if err := os.Remove(dstPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("delete %s: %w", c.RelPath, err)
			}
			fmt.Printf("deleted %s\n", c.RelPath)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func DetectChanges(current []types.FileRecord, previous map[string]types.FileRecord, delete bool) []types.Change {
	var changes []types.Change
	seen := make(map[string]bool, len(current))

	for _, cur := range current {
		seen[cur.RelPath] = true
		prev, ok := previous[cur.RelPath]

		if !ok {
			c := cur
			changes = append(changes, types.Change{Type: types.New, RelPath: cur.RelPath, New: &c})
			continue
		}

		if cur.Size == prev.Size && cur.ModTime == prev.ModTime {
			continue
		}

		if cur.Checksum == prev.Checksum {
			continue
		}

		p, c := prev, cur
		changes = append(changes, types.Change{Type: types.Modified, RelPath: cur.RelPath, Old: &p, New: &c})
	}

	if delete {
		for _, prev := range previous {
			if !seen[prev.RelPath] {
				p := prev
				changes = append(changes, types.Change{Type: types.Deleted, RelPath: prev.RelPath, Old: &p})
			}
		}
	}

	return changes
}
