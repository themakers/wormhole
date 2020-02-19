package types

import (
	"fmt"
)

var _ Type = &Array{}

type Array struct {
	Len  int
	Type Type
}

func (a *Array) Hash() string {
	return hash(a.String())
}

const arrayTmpl = "[%d]%s"

func (a *Array) String() string {
	return fmt.Sprintf(arrayTmpl, a.Len, a.Type)
}
