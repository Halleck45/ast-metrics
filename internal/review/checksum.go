package review

import (
	"crypto/sha256"
	"encoding/hex"
	"os"

	pb "github.com/ast-metrics/ast-metrics/pb"
)

// FillChecksums computes a content checksum for every parsed file that does
// not have one yet. Engines do not fill this field; the review relies on it
// to skip files whose content is identical in both versions.
func FillChecksums(files []*pb.File) {
	for _, file := range files {
		if file == nil || file.Checksum != "" || file.Path == "" {
			continue
		}
		data, err := os.ReadFile(file.Path)
		if err != nil {
			continue
		}
		sum := sha256.Sum256(data)
		file.Checksum = hex.EncodeToString(sum[:])
	}
}
