package main

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var SerializerError = errors.New("invalid serialized data")

type CellBinarySerializer struct {
}

func NewCellBinarySerializer() *CellBinarySerializer {
	return &CellBinarySerializer{}
}

func (s *CellBinarySerializer) Marshal(key string, value string) []byte {
	keyBytes := []byte(key)

	serializedData := make([]byte, 0, 2+len(keyBytes)+len(value))

	serializedData = binary.LittleEndian.AppendUint16(serializedData, uint16(len(keyBytes)))
	serializedData = append(serializedData, keyBytes...)
	serializedData = append(serializedData, []byte(value)...)
	return serializedData
}

func (s *CellBinarySerializer) Unmarshal(data []byte) (key string, value string, err error) {
	if len(data) < 2 {
		return "", "", fmt.Errorf("%w: should be more than 2 bytes (data: %v)", SerializerError, string(data))
	}

	keyLength := binary.LittleEndian.Uint16(data)
	if len(data) < int(keyLength)+2 {
		return "", "", fmt.Errorf("%w: key size is less than bytes amount (keySize: %d; data: %v)", SerializerError, keyLength, string(data))
	}

	key = string(data[2 : keyLength+2])
	value = string(data[keyLength+2:])
	return
}
