package types

import (
	"encoding/binary"
)

var _ Type = &Array{}

type Array struct {
	Len  uint64
	Type Type
}

func (a *Array) Hash() Sum {
	return a.hash(map[Type]bool{})
}

func (a *Array) hash(prev map[Type]bool) Sum {
	l := make([]byte, 8)
	binary.LittleEndian.PutUint64(l, a.Len)

	t := a.Type.hash(prev)

	return sum([]byte("ARRAY"), l, t[:])
}

func (a *Array) String() string {
	return stringify(a)
}
