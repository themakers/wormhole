package types

import (
	"crypto/sha256"
	h "hash"
)

type Type interface {
	Hash() string
	String() string
}

var hash h.Hash = sha256.New()

var Untyped Type = untyped{}

type untyped struct{}

func (u untyped) Hash() string {
	return string(
		hash.Sum([]byte(u.String())),
	)
}

func (_ untyped) String() string {
	return "???"
}
