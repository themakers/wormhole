package types

import (
	"fmt"
	"strings"
)

var (
	_ Type     = &Interface{}
	_ Selector = &Interface{}
)

// Interface represent Go's interfaces.
type Interface struct {
	// All methods defined in the interface.
	Methods []*Method

	// Fast access alternative to Interface.Methods with method names as keys.
	MethodsMap map[string]*Method
}

func (i *Interface) Select(name string) (Type, error) {
	if m, ok := i.MethodsMap[name]; ok {
		return m, nil
	}
	return nil, ErrSelectorUndefined{
		Sel: name,
	}
}

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
