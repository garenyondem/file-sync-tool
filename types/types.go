package types

type FileRecord struct {
	RelPath  string `json:"rel_path"`
	Size     int64  `json:"size"`
	ModTime  int64  `json:"mod_time"`
	Checksum string `json:"checksum"`
}

type SyncState struct {
	Source    string                 `json:"source"`
	Dest      string                 `json:"dest"`
	Files     map[string]FileRecord  `json:"files"`
	UpdatedAt int64                  `json:"updated_at"`
}

type ChangeType uint8

const (
	New ChangeType = iota
	Modified
	Deleted
)

type Change struct {
	Type    ChangeType
	RelPath string
	Old     *FileRecord
	New     *FileRecord
}
