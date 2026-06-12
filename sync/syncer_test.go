package sync

import (
	"testing"

	"github.com/user/file-sync-tool/types"
)

func TestDetectChanges(t *testing.T) {
	t.Run("new files", func(t *testing.T) {
		current := []types.FileRecord{
			{RelPath: "a.txt", Size: 10, ModTime: 100, Checksum: "aaa"},
			{RelPath: "b.txt", Size: 20, ModTime: 200, Checksum: "bbb"},
		}
		prev := map[string]types.FileRecord{}

		changes := DetectChanges(current, prev, false)

		if len(changes) != 2 {
			t.Fatalf("expected 2 changes, got %d", len(changes))
		}
		for _, c := range changes {
			if c.Type != types.New {
				t.Errorf("expected New change for %s, got %v", c.RelPath, c.Type)
			}
		}
	})

	t.Run("no changes", func(t *testing.T) {
		current := []types.FileRecord{
			{RelPath: "a.txt", Size: 10, ModTime: 100, Checksum: "aaa"},
		}
		prev := map[string]types.FileRecord{
			"a.txt": {RelPath: "a.txt", Size: 10, ModTime: 100, Checksum: "aaa"},
		}

		changes := DetectChanges(current, prev, false)

		if len(changes) != 0 {
			t.Errorf("expected 0 changes, got %d", len(changes))
		}
	})

	t.Run("modified content", func(t *testing.T) {
		current := []types.FileRecord{
			{RelPath: "a.txt", Size: 11, ModTime: 100, Checksum: "bbb"},
		}
		prev := map[string]types.FileRecord{
			"a.txt": {RelPath: "a.txt", Size: 10, ModTime: 100, Checksum: "aaa"},
		}

		changes := DetectChanges(current, prev, false)

		if len(changes) != 1 {
			t.Fatalf("expected 1 change, got %d", len(changes))
		}
		if changes[0].Type != types.Modified {
			t.Errorf("expected Modified, got %v", changes[0].Type)
		}
	})

	t.Run("same content different mtime", func(t *testing.T) {
		current := []types.FileRecord{
			{RelPath: "a.txt", Size: 10, ModTime: 999, Checksum: "aaa"},
		}
		prev := map[string]types.FileRecord{
			"a.txt": {RelPath: "a.txt", Size: 10, ModTime: 100, Checksum: "aaa"},
		}

		changes := DetectChanges(current, prev, false)

		if len(changes) != 0 {
			t.Errorf("expected 0 changes (same checksum), got %d", len(changes))
		}
	})

	t.Run("deleted with delete flag", func(t *testing.T) {
		current := []types.FileRecord{
			{RelPath: "a.txt", Size: 10, ModTime: 100, Checksum: "aaa"},
		}
		prev := map[string]types.FileRecord{
			"a.txt": {RelPath: "a.txt", Size: 10, ModTime: 100, Checksum: "aaa"},
			"b.txt": {RelPath: "b.txt", Size: 20, ModTime: 200, Checksum: "bbb"},
		}

		changes := DetectChanges(current, prev, true)

		if len(changes) != 1 {
			t.Fatalf("expected 1 change (delete), got %d", len(changes))
		}
		if changes[0].Type != types.Deleted || changes[0].RelPath != "b.txt" {
			t.Errorf("expected Deleted b.txt, got %v %s", changes[0].Type, changes[0].RelPath)
		}
	})

	t.Run("deleted without delete flag", func(t *testing.T) {
		current := []types.FileRecord{}
		prev := map[string]types.FileRecord{
			"a.txt": {RelPath: "a.txt", Size: 10, ModTime: 100, Checksum: "aaa"},
		}

		changes := DetectChanges(current, prev, false)

		if len(changes) != 0 {
			t.Errorf("expected 0 changes (delete not set), got %d", len(changes))
		}
	})
}
