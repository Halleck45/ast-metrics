package namer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

func loadW2V(data []byte) (map[string][]float32, int, error) {
	r := bytes.NewReader(data)

	magic := make([]byte, 4)
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, 0, err
	}
	if string(magic) != "W2V1" {
		return nil, 0, errors.New("invalid W2V1 magic header")
	}

	var dim uint32
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &dim); err != nil {
		return nil, 0, err
	}
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, 0, err
	}

	vectors := make(map[string][]float32, count)

	for i := uint32(0); i < count; i++ {
		var wlen uint16
		if err := binary.Read(r, binary.LittleEndian, &wlen); err != nil {
			return nil, 0, err
		}

		word := make([]byte, wlen)
		if _, err := io.ReadFull(r, word); err != nil {
			return nil, 0, err
		}

		vec := make([]float32, dim)
		if err := binary.Read(r, binary.LittleEndian, &vec); err != nil {
			return nil, 0, err
		}

		vectors[string(word)] = vec
	}

	return vectors, int(dim), nil
}
