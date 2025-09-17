package storage

import (
	"os"
	"testing"
)

func TestGetFileHash_ValidFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := "test content"
	tmpFile.WriteString(content)
	tmpFile.Close()

	hash, err := GetFileHash(tmpFile.Name())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if len(hash) != 32 { // MD5 hash length
		t.Errorf("expected hash length 32, got %d", len(hash))
	}
}

func TestGetFileHash_NonExistentFile(t *testing.T) {
	_, err := GetFileHash("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestUnmarshalProtobuf_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	_, err = UnmarshalProtobuf(tmpFile.Name())
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestUnmarshalProtobuf_NonExistentFile(t *testing.T) {
	_, err := UnmarshalProtobuf("/nonexistent/file.bin")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}
