package types

import (
	"crypto/sha256"

	"github.com/getlantern/hex"
)

type (
	Type interface {
		Hash() string
		hash(prev map[*Definition]bool) string
		String() string
	}

	Selector interface {
		Select(string) (Type, error)
		Type
	}
)

var hasher = sha256.New()

func sum(v string) string {
	return hex.DefaultEncoding.EncodeToString(hasher.Sum([]byte(v)))
}

var Untyped Type = untyped{}

type untyped struct{}

func (u untyped) Hash() string {
	return u.hash(map[*Definition]bool{})
}

func (u untyped) hash(_ map[*Definition]bool) string {
	return sum("untyped")
}

func (_ untyped) String() string {
	return ""
}
