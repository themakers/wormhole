package types

import "fmt"

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

func (d *Definition) Hash() Sum {
	return d.hash(map[Type]bool{})
}

func (d *Definition) hash(prev map[Type]bool) Sum {
	if prev[d] {
		return sum("DEF", d.Name, "FROM", d.Package.hash(prev), "SELF")
	}

	if d.Declaration == nil {
		panic(fmt.Errorf("definition \"%s\" have a nil declaration", d))
	}

	prev[d] = true
	return sum(
		"DEF",
		d.Name,
		"FROM",
		d.Package.hash(prev),
		"DECL",
		d.Declaration.hash(prev),
	)
}

func (d *Definition) String() string {
	return stringify(d)
}
