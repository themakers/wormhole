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
