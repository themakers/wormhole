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
	var sums []byte
	for _, arg := range args {
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

		default:
			panic(
				fmt.Errorf(
					"failed to take hash sum of \"%v\" of type \"%T\"",
					a,
					a,
				),
			)
		}

		sums = append(sums, s[:]...)
	}

	return sha256.Sum256(sums)
}
