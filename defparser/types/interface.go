package types

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
	return i.hash(map[*Definition]bool{})
}

func (i *Interface) hash(prev map[*Definition]bool) string {
	s := sum("INTERFACE")
	for _, meth := range i.Methods {
		s += meth.hash(prev)
	}
	return sum(s)
}

func (i *Interface) String() string {
	return stringify(i)
}
