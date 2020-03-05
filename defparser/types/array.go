package types

var _ Type = &Array{}

type Array struct {
	Len  uint64
	Type Type
}

func (a *Array) Hash() Sum {
	return a.hash(map[Type]bool{})
}

func (a *Array) hash(prev map[Type]bool) Sum {
	return sum("ARRAY", a.Len, "OF", a.Type.hash(prev))
}

func (a *Array) String() string {
	return stringify(a)
}
