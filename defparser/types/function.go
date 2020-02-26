package types

var _ Type = &Function{}

type (
	Function struct {
		Args    []*NameTypePair
		ArgsMap map[string]*NameTypePair

		Results    []*NameTypePair
		ResultsMap map[string]*NameTypePair
	}

	NameTypePair struct {
		Name string
		Type Type
	}
)

func (f *Function) Hash() string {
	return f.hash(map[Type]bool{})
}

func (f *Function) hash(prev map[Type]bool) string {
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
