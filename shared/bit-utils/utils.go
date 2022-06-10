package bit_utils

import (
	"encoding/binary"
)

func BooleansToBytes(t []bool) []byte {
	b := make([]byte, (len(t)+7)/8)
	for i, x := range t {
		if x {
			b[i/8] |= 0x80 >> uint(i%8)
		}
	}
	return b
}

func BytesToBooleans(b []byte) []bool {
	t := make([]bool, 8*len(b))
	for i, x := range b {
		for j := 0; j < 8; j++ {
			if (x<<uint(j))&0x80 == 0x80 {
				t[8*i+j] = true
			}
		}
	}
	return t
}

func GetBytesFromInt(len int) []byte {
	bs := make([]byte, 2)
	binary.LittleEndian.PutUint16(bs, uint16(len))
	return bs
}

func GetIntFromBytes(lens []byte) int {
	return int(binary.LittleEndian.Uint16(lens))
}

func GetChunk(b []byte, l int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(b)/l+1)
	for len(b) >= l {
		chunk, b = b[:l], b[l:]
		chunks = append(chunks, chunk)
	}
	if len(b) > 0 {
		chunks = append(chunks, b[:len(b)])
	}
	return chunks
}

func MergeBoolLists(into *[]bool, b []bool) {
	for i := 0; i < minInt(len(*into), len(b)); i++ {
		if b[i] {
			(*into)[i] = true
		}
	}
}

func BooleanIndex(a []bool) int {
	for i := 0; i < len(a); i++ {
		if a[i] {
			return i
		}
	}
	return -1
}

func IndexInBoolean(count int, index int) []bool {
	res := make([]bool, count)
	if index >= 0 && index < count {
		res[index] = true
	}
	return res
}

func AllBooleans(a []bool) bool {
	for _, x := range a {
		if !x {
			return false
		}
	}
	return true
}

func BytesJoin(s ...[]byte) []byte {
	n := 0
	for _, v := range s {
		n += len(v)
	}

	b, i := make([]byte, n), 0
	for _, v := range s {
		i += copy(b[i:], v)
	}
	return b
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func DataTail(b []byte) [][]byte {
	var res [][]byte
	size := GetIntFromBytes(b[0:2])
	l := len(b)
	//log.Debug().Int("len", l).Int("size", size+2).Hex("size-hex", b[0:2]).Msg("len")
	if l >= size+2 {
		res = append(res, b[2:size+2])
		if l > size+2 {
			res = append(res, DataTail(b[:size+2])...)
		}
	}
	return res
}

func AddSize(data []byte) []byte {
	ld := len(data)
	return append(GetBytesFromInt(ld), data...)
}
