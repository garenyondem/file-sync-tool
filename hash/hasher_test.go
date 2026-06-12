package hash

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileChecksum(t *testing.T) {
	dir := t.TempDir()

	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	t.Run("same content same hash", func(t *testing.T) {
		h1, _ := FileChecksum(write("a.txt", "hello"))
		h2, _ := FileChecksum(write("b.txt", "hello"))
		if h1 != h2 {
			t.Errorf("expected same hash, got %s vs %s", h1, h2)
		}
	})

	t.Run("different content different hash", func(t *testing.T) {
		h1, _ := FileChecksum(write("c.txt", "hello"))
		h2, _ := FileChecksum(write("d.txt", "world"))
		if h1 == h2 {
			t.Errorf("expected different hashes, got %s", h1)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		h, err := FileChecksum(write("empty.txt", ""))
		if err != nil {
			t.Fatal(err)
		}
		if h != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
			t.Errorf("unexpected empty file hash: %s", h)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := FileChecksum("/nonexistent/foo.bar")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})
}
