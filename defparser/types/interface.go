package types

import (
	"fmt"
	"strings"
)

var _ Type = &Interface{}

type (
	Interface struct {
		Methods []*Method
	}
)

func (i *Interface) Hash() string {
	return hash(i.String())
}

const interTmpl = "inter{%s}"

func (i *Interface) String() string {
	methods := make([]string, len(i.Methods))

	for i, meth := range i.Methods {
		methods[i] = meth.String()
	}

	return fmt.Sprintf(
		interTmpl,
		strings.Join(methods, ","),
	)
}
