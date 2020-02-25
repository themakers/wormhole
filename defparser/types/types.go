package types

import (
	"crypto/sha256"

	"github.com/getlantern/hex"
)

type (
	Type interface {
		Hash() string
		hash(prev map[*Definition]bool) string
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
	return u.hash(nil)
}

func (u untyped) hash(_ map[*Definition]bool) string {
	return sum("untyped")
}

func (u untyped) String() string {
	return stringify(u)
}
