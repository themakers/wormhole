package types

var _ Type = &Function{}

type (
	Function struct {
		Args    []NameTypePair
		Results []NameTypePair
	}

	NameTypePair struct {
		Name string
		Type Type
	}
)

func (f *Function) Hash() string {
	return f.hash(map[*Definition]bool{})
}

func (f *Function) hash(prev map[*Definition]bool) string {
	s := sum("FUNC")
	for _, arg := range f.Args {
		s += arg.Type.hash(prev)
	}
	for _, result := range f.Results {
		s += result.Type.hash(prev)
	}
	return sum(s)
}

func (f *Function) String() string {
	return stringify(f)
}
