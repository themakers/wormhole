package types

var (
	_ Type     = &Definition{}
	_ Selector = &Definition{}
)

// Definition represent any named type defined by
// "type <definition name> <declaration>" construction and then used this way:
// packageName.definitionName.
type Definition struct {
	// Definition's name.
	Name string

	// Store what this definition represents.
	Declaration Type

	// Package where this definition where defined.
	Package *Package

	// Methods defined on this definition.
	Methods []*Method

	// Fast access alternative to Definition.Methods
	// with method names as keys.
	MethodsMap map[string]*Method

	// Is this type exported.
	Exported bool

	// Is this definition from standart library. Definitions with Std == true
	// always have nil Declaration and Methods fields.
	Std bool

	Underlying []struct {
		name string
		t    Type
	}
}

func (d *Definition) Select(name string) (Type, error) {
	if m, ok := d.MethodsMap[name]; ok {
		return m, nil
	}

	if s, ok := d.Declaration.(Selector); ok {
		return s.Select(name)
	}

	return nil, ErrSelectorUndefined{
		Sel: name,
	}
}

func (d *Definition) Hash() string {
	return d.hash(map[*Definition]bool{})
}

func (d *Definition) hash(prev map[*Definition]bool) string {
	if prev[d] {
		return sum(sum("DEF") + d.Package.hash(prev) + sum(d.Name))
	}

	prev[d] = true
	return sum(sum("DEF") +
		sum(d.Package.Info.PkgPath) +
		sum(d.Name) +
		d.Declaration.hash(prev),
	)
}

func (d *Definition) String() string {
	return stringify(d)
}
