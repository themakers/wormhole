package types

var _ Type = &Package{}

type (
	Package struct {
		Info           PackageInfo
		Imports        []Import
		ImportsMap     map[string]Import
		Definitions    []*Definition
		DefinitionsMap map[string]*Definition
		Methods        []*Method
		MethodsMap     map[string]*Method
	}

	PackageInfo struct {
		PkgName     string
		PkgPath     string
		PkgFullPath string
		Std         bool
	}

	Import struct {
		Package *Package
		Alias   string
	}
)

func (p *Package) Hash() Sum {
	return p.hash(map[Type]bool{})
}

func (p *Package) hash(prev map[Type]bool) Sum {
	s := make(
		[][]byte,
		3+len(p.Imports)+len(p.Definitions)+len(p.Methods),
	)
	s[0] = []byte("PACKAGE")
	if p.Info.Std {
		s[1] = []byte("STD")
	} else {
		s[1] = []byte("NON-STD")
	}
	s[2] = []byte(p.Info.PkgPath)

	if prev[p] {
		return sum(s...)
	}
	prev[p] = true

	i := 3
	for _, imp := range p.Imports {
		v := imp.Package.hash(prev)
		s[i] = v[:]
		i++
	}

	for _, def := range p.Definitions {
		v := def.hash(prev)
		s[i] = v[:]
		i++
	}

	for _, meth := range p.Methods {
		v := meth.hash(prev)
		s[i] = v[:]
		i++
	}

	return sum(s...)
}

func (p *Package) String() string {
	return stringify(p)
}
