package storage

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"os"

	pb "github.com/halleck45/ast-metrics/pb"
	"google.golang.org/protobuf/proto"
)

// Provides the hash of a file, in order to avoid to parse it twice
func GetFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
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
