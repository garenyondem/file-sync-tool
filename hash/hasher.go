package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

const CopyBufSize = 64 * 1024

func FileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.CopyBuffer(h, f, make([]byte, CopyBufSize)); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
