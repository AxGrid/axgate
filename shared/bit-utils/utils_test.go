package bit_utils

import (
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBooleansBytes(t *testing.T) {

	mask := []bool{true, false, false, true, true}
	maskBytes := BooleansToBytes(mask)
	assert.Equal(t, len(maskBytes), 1)
	assert.EqualValues(t, mask, BytesToBooleans(maskBytes)[:5])

	mask = []bool{true, false, false, true, true, false, false, false, true}
	maskBytes = BooleansToBytes(mask)
	assert.Equal(t, len(maskBytes), 2)
	assert.EqualValues(t, mask, BytesToBooleans(maskBytes)[:9])
}

func TestLength(t *testing.T) {
	size := 65_535
	lenBytes := GetBytesFromInt(size)
	log.Info().Hex("size", lenBytes).Msg("bytes")
	assert.Equal(t, size, GetIntFromBytes(lenBytes))
}

func TestMergeList(t *testing.T) {
	a := []bool{false, false, true, false}
	b := []bool{true, false, false, true}
	MergeBoolLists(&a, b)
	assert.EqualValues(t, a, []bool{true, false, true, true})
}

func TestChunk(t *testing.T) {
	a := []byte{1, 2, 3, 4, 5}
	s := GetChunk(a, 3)
	assert.Equal(t, len(s), 2)
	assert.EqualValues(t, s[0], []byte{1, 2, 3})
	assert.EqualValues(t, s[1], []byte{4, 5})
	s = GetChunk(a, 5)
	assert.Equal(t, len(s), 1)
	assert.EqualValues(t, s[0], []byte{1, 2, 3, 4, 5})
	s = GetChunk(a, 10)
	assert.Equal(t, len(s), 1)
	assert.EqualValues(t, s[0], []byte{1, 2, 3, 4, 5})

	a = []byte{}
	s = GetChunk(a, 10)
	assert.Equal(t, len(s), 0)

}
