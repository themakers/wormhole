package register

import (
	"go/ast"
	"strconv"

	"github.com/themakers/wormhole/defparser/types"
)

func (r *Register) Registrate(pkg *ast.Package) error {
	return declMap(
		pkg,
		func(spec ast.Decl) error {
			return r.registrate(
				pkg,
				map[string]*types.Definition{},
				spec,
			)
		},
	)
}

func (r *Register) registrate(
	pkg *ast.Package,
	prevDefs map[string]*types.Definition,
	decl ast.Decl,
) error {
	var getType func(expr ast.Expr) (types.Type, error)

	getType = func(expr ast.Expr) (types.Type, error) {
		switch e := expr.(type) {
		case *ast.MapType:
			k, err := getType(e.Key)
			if err != nil {
				return nil, err
			}

			v, err := getType(e.Value)
			if err != nil {
				return nil, err
			}

			return *&types.Map{
				k,
				v,
			}, nil

		case *ast.ArrayType:
			if e.Len == nil {

			}

			len, err := strconv.Atoi(e.Len.(*ast.BasicLit).Value)
			if err != nil {
				return nil, err
			}

		case *ast.StructType:
		case *ast.FuncType:
		case *ast.InterfaceType:
		// case *ast.Field:
		case *ast.Ident:
		case *ast.SelectorExpr:
		case *ast.StarExpr:
		case *ast.ChanType:
		}

		return nil, nil
	}

	switch d := decl.(type) {
	case *ast.GenDecl:
		for _, spec := range d.Specs {
			s, ok := spec.(*ast.TypeSpec)
			if !ok {
				return nil
			}

			n := s.Name.Name
			if _, ok := r.DefinitionsMap[n]; ok {
				return nil
			}

			t, err := getType(s.Type)
			if err != nil {
				return err
			}

			if err := r.regDef(n, t, r.Package); err != nil {
				return err
			}

		}
		return nil

	case *ast.FuncDecl:
		return nil

	case *ast.BadDecl:
		return ErrBadSyntax{
			from: r.FileSet.Position(d.From),
			to:   r.FileSet.Position(d.To),
		}

	default:
		panic("something went wrong")
	}
}

func declMap(
	pkg *ast.Package,
	f func(decl ast.Decl) error,
) error {
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			if err := f(decl); err != nil {
				return err
			}
		}
	}
	return nil
}
