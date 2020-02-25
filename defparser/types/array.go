package types

var _ Type = &Array{}

type Array struct {
	Len  int
	Type Type
}

func (a *Array) Hash() string {
	return a.hash(map[*Definition]bool{})
}

func (a *Array) hash(prev map[*Definition]bool) string {
	return sum(sum("ARRAY") + sum(string(a.Len)) + a.Type.hash(prev))
}

func (a *Array) String() string {
	return stringify(a)
}
