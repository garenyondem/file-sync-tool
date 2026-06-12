package scan

import (
	"os"
	"path/filepath"

	"github.com/user/file-sync-tool/hash"
	"github.com/user/file-sync-tool/types"
)

func Tree(root string) ([]types.FileRecord, error) {
	var records []types.FileRecord

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		checksum, err := hash.FileChecksum(path)
		if err != nil {
			return err
		}

		records = append(records, types.FileRecord{
			RelPath:  rel,
			Size:     info.Size(),
			ModTime:  info.ModTime().UnixNano(),
			Checksum: checksum,
		})
		return nil
	})

	return records, err
}
