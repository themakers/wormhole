package types

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

type Sum [32]byte

func (s Sum) String() string {
	return fmt.Sprintf("%x", s[:])
}

func sum(args ...interface{}) Sum {
	sums := make([]byte, len(args)*32)
	for i, arg := range args {
		var s Sum
		switch a := arg.(type) {
		case int:
			l := make([]byte, 8)
			binary.LittleEndian.PutUint64(l, uint64(a))
			s = sum(l)

		case string:
			s = sum([]byte(a))

		case bool:
			if a {
				s = sum(int(1))
			} else {
				s = sum(int(0))
			}

		case []byte:
			s = sha256.Sum256(a)

		case Sum:
			s = a
		}
	}

	var length int
	for _, arg := range args {
		length += len(arg)
	}

	var (
		data = make([]byte, length)
		i    int
	)
	for _, arg := range args {
		for _, b := range arg {
			data[i] = b
			i++
		}
	}

	return sha256.Sum256(data)
}
