package review

import (
	"os"
	"path/filepath"
	"testing"

	pb "github.com/ast-metrics/ast-metrics/pb"
	"github.com/stretchr/testify/assert"
)

func TestFillChecksums(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.go")
	assert.NoError(t, os.WriteFile(path, []byte("package a"), 0644))

	same1 := &pb.File{Path: path}
	same2 := &pb.File{Path: path}
	missing := &pb.File{Path: filepath.Join(dir, "missing.go")}
	preset := &pb.File{Path: path, Checksum: "already"}

	FillChecksums([]*pb.File{same1, same2, missing, preset, nil})

	assert.NotEmpty(t, same1.Checksum)
	assert.Equal(t, same1.Checksum, same2.Checksum)
	assert.Empty(t, missing.Checksum)
	assert.Equal(t, "already", preset.Checksum)
}
