package storage

import (
	"errors"
	"fmt"
	"os"

	pb "github.com/halleck45/ast-metrics/pb"
	"google.golang.org/protobuf/proto"
)

// GetFileHash returns a fast cache key based on file mtime and size,
// avoiding the cost of reading the entire file to compute MD5.
func GetFileHash(filePath string) (string, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d_%d", stat.ModTime().UnixNano(), stat.Size()), nil
}

func UnmarshalProtobuf(file string) (*pb.File, error) {
	pbFile := &pb.File{}

	// load AST via ProtoBuf (using NodeType package)
	in, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// if file is empty, return
	if len(in) == 0 {
		return nil, errors.New("File is empty: " + file)
	}

	if err := proto.Unmarshal(in, pbFile); err != nil {
		return nil, err
	}

	return pbFile, nil
}
