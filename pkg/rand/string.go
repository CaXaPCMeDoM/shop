package rand

import (
	"crypto/rand"
	"encoding/binary"
)

var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func MustStr(n int) string {
	b := make([]byte, n)
	charsetLen := len(charset)

	for i := 0; i < n; {
		var buf [8]byte
		_, err := rand.Read(buf[:])
		if err != nil {
			panic(err)
		}

		val := binary.LittleEndian.Uint64(buf[:])
		charsetLen64 := uint64(charsetLen)

		for val > 0 && i < n {
			idx := int(val % charsetLen64)
			b[i] = charset[idx]
			i++
			val /= charsetLen64
		}
	}

	return string(b)
}
