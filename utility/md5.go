package utility

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
)

func GetFileMD5(reader io.Reader) (string, error) {
	// Create a new MD5 hash object
	hash := md5.New()

	// Copy the content from the reader into the hash
	if _, err := io.Copy(hash, reader); err != nil {
		return "", fmt.Errorf("failed to compute MD5: %w", err)
	}

	// Get the MD5 checksum in hexadecimal format
	return hex.EncodeToString(hash.Sum(nil)), nil
}
