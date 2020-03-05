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

func (f *Function) Hash() Sum {
	return f.hash(map[Type]bool{})
}

func (f *Function) hash(prev map[Type]bool) Sum {
	s := []interface{}{"FUNC"}
	for _, arg := range f.Args {
		s = append(s, arg.Type.hash(prev))
	}

	s = append(s, "RETURNS")
	for _, result := range f.Results {
		s = append(s, result.Type.hash(prev))
	}

	return sum(s...)
}

func (f *Function) String() string {
	return stringify(f)
}
