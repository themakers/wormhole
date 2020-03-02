package register

import (
	"go/ast"

	"github.com/themakers/wormhole/defparser/types"
)

func (r *Register) Registrate(pkg *ast.Package) error {
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			err := r.registrate(pkg, map[string]*types.Definition{}, decl)
			if err != nil {
				return err
			}
		}
	}

	err := r.registrate(pkg, map[string]*types.Definition{}, "")
	if err != nil {
		if v, ok := err.(ErrUnableRegistrate); ok && v.DefName == "" {
			return nil
		}
	}
	return err
}

func (r *Register) registrate(
	pkg *ast.Package,
	prevDefs map[string]*types.Definition,
	dec ast.Decl,
) error {

	return nil
}

func declMap(pkg *ast.Package, f func(fname string, decl ast.Decl) error) error {
	for fname, file := range pkg.Files {
		for _, decl := range file.Decls {

		}
	}
}
