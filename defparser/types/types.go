package types

import (
	"crypto/sha256"

	"github.com/getlantern/hex"
)

type (
	Type interface {
		Hash() string
		hash(prev map[Type]bool) string
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
