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
	s := make([][]byte, len(f.Args)+len(f.Results)+2)
	s[0] = []byte("FUNC")

	i := 1
	for _, arg := range f.Args {
		v := arg.Type.hash(prev)
		s[i] = v[:]
		i++
	}

	s[i] = []byte("RETURNS")
	i++

	for _, result := range f.Results {
		v := result.Type.hash(prev)
		s[i] = v[:]
		i++
	}

	return sum(s...)
}

func (f *Function) String() string {
	return stringify(f)
}
