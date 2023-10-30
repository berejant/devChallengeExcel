package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCellBinarySerializer_Marshal(t *testing.T) {
	serializer := &CellBinarySerializer{}
	serialized := serializer.Marshal("key1", "value1")
	assert.NotNil(t, serialized)
	assert.Greater(t, len(serialized), 8)
}

func TestCellBinarySerializer_Unmarshal(t *testing.T) {
	serializer := &CellBinarySerializer{}

	t.Run("valid_data", func(t *testing.T) {
		assertMarshalAndUnmarshal := func(expectedKey string, expectedValue string) {
			serialized := serializer.Marshal(expectedKey, expectedValue)
			actualKey, actualValue, err := serializer.Unmarshal(serialized)

			assert.NoError(t, err)
			assert.Equal(t, expectedKey, actualKey)
			assert.Equal(t, expectedValue, actualValue)
		}

		assertMarshalAndUnmarshal("key1", "value1")

		assertMarshalAndUnmarshal(
			"key1_should be any URL-compatible text that represents a cell (variable) and can be generated on the client",
			"value1_Data should be persisted and available between docker containers restarts",
		)
	})

	t.Run("empty_data", func(t *testing.T) {
		key, value, err := serializer.Unmarshal([]byte{})

		assert.Error(t, err)
		assert.Equal(t, "", key)
		assert.Equal(t, "", value)
	})

	t.Run("invalid_data", func(t *testing.T) {
		key, value, err := serializer.Unmarshal([]byte{' ', 'q', 'r'})

		assert.Error(t, err)
		assert.Equal(t, "", key)
		assert.Equal(t, "", value)
	})

}
