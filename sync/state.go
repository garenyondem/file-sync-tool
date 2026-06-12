package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/user/file-sync-tool/types"
)

func LoadState(path string) (*types.SyncState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s types.SyncState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func SaveState(s *types.SyncState, path string) error {
	s.UpdatedAt = time.Now().UnixNano()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func ApplyChangesToState(s *types.SyncState, records []types.FileRecord, deletedPaths []string) {
	if s.Files == nil {
		s.Files = make(map[string]types.FileRecord)
	}
	for _, r := range records {
		s.Files[r.RelPath] = r
	}
	for _, p := range deletedPaths {
		delete(s.Files, p)
	}
}
