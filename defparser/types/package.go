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
	if prev[p] {
		return sum(
			"PACKAGE",
			p.Info.PkgPath,
			p.Info.Std,
		)
	}

	sums := []interface{}{
		"PACKAGE",
		p.Info.PkgPath,
		p.Info.Std,
	}

	sums = append(sums, "IMPORTS")
	for _, imp := range p.Imports {
		sums = append(sums, imp.Package.hash(prev))
	}

	sums = append(sums, "METHODS")
	for _, meth := range p.Methods {
		sums = append(sums, meth.hash(prev))
	}

	sums = append(sums, "DEFS")
	for _, def := range p.Definitions {
		sums = append(sums, def.hash(prev))
	}

	return sum(sums...)
}

func (p *Package) String() string {
	return stringify(p)
}
