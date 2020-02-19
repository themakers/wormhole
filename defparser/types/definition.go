package types

import "fmt"

var _ Type = &Definition{}

type Definition struct {
	Name        string
	Declaration Type
	Package     *Package
	Methods     []*Method
	Exported    bool
	Std         bool
}

func (d *Definition) Hash() string {
	return string(
		hash.Sum([]byte(d.String())),
	)
}

const (
	defInterfaceTmpl = "type %s"
	defTmpl          = "%s.type %s %s"
)

func (d *Definition) String() string {
	if _, ok := d.Declaration.(*Interface); ok {
		return fmt.Sprintf(
			defInterfaceTmpl,
			d.Declaration,
		)
	}
	return fmt.Sprintf(
		defTmpl,
		d.Package,
		d.Name,
		d.Declaration,
	)
}
