package types

import (
	"fmt"
)

var _ Type = &Array{}

type Array struct {
	Len  int
	Type interface{}
}

func (a *Array) Hash() string {
	return string(
		hash.Sum([]byte(a.String())),
	)
}

const arrayTmpl = "[%d]%s"

func (a *Array) String() string {
	return fmt.Sprintf(arrayTmpl, a.Len, a.Type)
}
