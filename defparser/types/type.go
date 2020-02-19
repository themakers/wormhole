package types

import (
	"crypto/sha256"
	"github.com/getlantern/hex"
)

type Type interface {
	Hash() string
	String() string
}

var hasher = sha256.New()

func hash(v string) string {
	return hex.DefaultEncoding.EncodeToString(hasher.Sum([]byte(v)))
}

var Untyped Type = untyped{}

type untyped struct{}

func (u untyped) Hash() string {
	return hash(u.String())
}

func (_ untyped) String() string {
	return "???"
}
