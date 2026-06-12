package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/file-sync-tool/types"
)

func TestIntegrationSync(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	statePath := filepath.Join(t.TempDir(), "state.json")

	state := &types.SyncState{
		Source: src,
		Dest:   dst,
		Files:  make(map[string]types.FileRecord),
	}

	write := func(dir, name, content string) {
		path := filepath.Join(dir, name)
		os.MkdirAll(filepath.Dir(path), 0755)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	write(src, "hello.txt", "hello world")
	write(src, "sub/nested.txt", "nested file")

	if err := syncOnce(src, dst, statePath, false, false, state); err != nil {
		t.Fatal(err)
	}

	check := func(name, content string) {
		data, err := os.ReadFile(filepath.Join(dst, name))
		if err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
		if string(data) != content {
			t.Errorf("%s: got %q, want %q", name, string(data), content)
		}
	}

	check("hello.txt", "hello world")
	check("sub/nested.txt", "nested file")

	t.Run("second run no changes", func(t *testing.T) {
		if err := syncOnce(src, dst, statePath, false, false, state); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("detects file modification", func(t *testing.T) {
		write(src, "hello.txt", "modified content")

		if err := syncOnce(src, dst, statePath, false, false, state); err != nil {
			t.Fatal(err)
		}
		check("hello.txt", "modified content")
	})

	t.Run("dry run does not copy", func(t *testing.T) {
		write(src, "dry-test.txt", "should not appear")

		dryState := &types.SyncState{
			Source: src,
			Dest:   dst,
			Files:  make(map[string]types.FileRecord),
		}

		if err := syncOnce(src, dst, filepath.Join(t.TempDir(), "dry.json"), true, false, dryState); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(filepath.Join(dst, "dry-test.txt")); !os.IsNotExist(err) {
			t.Error("dry run should not create files")
		}
	})
}
