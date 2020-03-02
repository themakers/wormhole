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

func (i *Interface) Hash() Sum {
	return i.hash(map[Type]bool{})
}

func (i *Interface) hash(prev map[Type]bool) Sum {
	s := make([][]byte, len(i.Methods)*2+1)
	s[0] = []byte("INTERFACE")

	for i, meth := range i.Methods {
		s[i*2+1] = []byte(meth.Name)
		v := meth.Signature.hash(prev)
		s[i*2+2] = v[:]
	}

	return sum(s...)
}

func (i *Interface) String() string {
	return stringify(i)
}
