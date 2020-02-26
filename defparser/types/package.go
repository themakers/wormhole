package types

var _ Type = &Package{}

type (
	Package struct {
		Info           PackageInfo
		Imports        []Import
		ImportsMap     map[PackageInfo]Import
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

func (p *Package) Hash() string {
	return p.hash(map[Type]bool{})
}

func (p *Package) hash(prev map[Type]bool) string {
	if p.Info.Std {
		return sum(sum("PACKAGE") + sum("STD") + sum(p.Info.PkgPath))
	}

	if prev[p] {
		return sum(sum("PACKAGE") + sum(p.Info.PkgPath))
	}

	prev[p] = true

	s := sum("PACKAGE") + sum(p.Info.PkgPath)

	for _, imp := range p.Imports {
		s += imp.Package.hash(prev)
	}

	for _, def := range p.Definitions {
		s += def.hash(prev)
	}

	for _, meth := range p.Methods {
		s += meth.hash(prev)
	}

	return sum(s)
}

func (p *Package) String() string {
	return stringify(p)
}
