package register

import (
	"go/ast"

	"github.com/themakers/wormhole/defparser/types"
)

func (r *Register) Registrate(ast *ast.Package) error {
	err := r.registrate(ast, map[string]*types.Definition{}, "")
	if err != nil {
		if v, ok := err.(ErrUnableRegistrate); ok && v.DefName == "" {
			return nil
		}
	}
	return err
}

func (r *Register) registrate(
	ast *ast.Package,
	prevDefs map[string]*types.Definition,
	target string,
) error {

	return nil
}
