package utils

import (
	"bytes"
	"os"
)

// IsBinaryFile determines if a file is binary by checking for null bytes
func IsBinaryFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return false
	}
	buf = buf[:n]

	return bytes.Contains(buf, []byte{0})
}
