package defparser

import (
	"errors"
	"fmt"
	"go/ast"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/themakers/wormhole/defparser/types"
)

func aggregateDefinitions(tr *typeRegister, pkg *ast.Package) error {
	var (
		do func(ast.Node) error

		typeDef   func(*ast.TypeSpec) (func() error, error)
		funcDef   func(*ast.FuncDecl) (func() error, error)
		methodDef func(*ast.FuncDecl) func() error

		typeDec       func(ast.Node) (types.Type, error)
		funcSignature func(*ast.FuncType) (*types.Function, error)
		identifier    func(ast.Node) (*types.Definition, error)
		isIgnorable   func(ast.Node) bool

		fileName  string
		callbacks []func() error
	)

	do = func(node ast.Node) error {
		switch n := node.(type) {
		case *ast.GenDecl:
			for _, spec := range n.Specs {
				if err := do(spec); err != nil {
					return err
				}
			}
			return nil

		case *ast.TypeSpec:
			clbk, err := typeDef(n)
			callbacks = append([]func() error{clbk}, callbacks...)
			return err

		case *ast.FuncDecl:
			if n.Recv == nil {
				clbk, err := funcDef(n)
				callbacks = append([]func() error{clbk}, callbacks...)
				return err
			}
			callbacks = append([]func() error{methodDef(n)}, callbacks...)
			return nil

		default:
			if isIgnorable(node) {
				return nil
			}

			return fmt.Errorf(
				"No matches %s %s",
				fileName,
				spew.Sdump(n),
			)
		}
	}

	typeDef = func(d *ast.TypeSpec) (func() error, error) {
		def, err := tr.define(d.Name.Name)
		return func() error {
			def.Declaration, err = typeDec(d.Type)
			return err
		}, err
	}

	funcDef = func(dec *ast.FuncDecl) (func() error, error) {
		def, err := tr.define(dec.Name.Name)
		return func() error {
			def.Declaration, err = funcSignature(dec.Type)
			return err
		}, err
	}

	methodDef = func(dec *ast.FuncDecl) func() error {
		return func() error {
			r, err := identifier(dec.Recv.List[0].Type)
			if err != nil {
				return err
			}
			f, err := funcSignature(dec.Type)
			if err != nil {
				return err
			}
			return tr.method(dec.Name.Name, r, f)
		}
	}

	typeDec = func(node ast.Node) (types.Type, error) {
		switch n := node.(type) {
		case *ast.MapType:
			k, err := typeDec(n.Key)
			if err != nil {
				return nil, err
			}
			v, err := typeDec(n.Value)
			if err != nil {
				return nil, err
			}

			return tr.implMap(k, v), nil

		case *ast.ArrayType:
			t, err := typeDec(n.Elt)
			if err != nil {
				return nil, err
			}

			if n.Len == nil {
				return tr.implSlice(t), nil
			}

			l, err := strconv.Atoi(n.Len.(*ast.BasicLit).Value)
			if err != nil {
				return nil, err
			}
			return tr.implArray(l, t), nil

		case *ast.StructType:
			fields := make([]*types.StructField, len(n.Fields.List))
			for i, field := range n.Fields.List {
				t, err := typeDec(field.Type)
				if err != nil {
					return nil, err
				}

				var tag string
				if field.Tag != nil {
					tag = field.Tag.Value
				}

				if len(field.Names) == 1 {
					fields[i] = tr.mkStructField(
						field.Names[0].Name,
						tag,
						t,
					)
				} else if len(field.Names) == 0 {
					if def, ok := t.(*types.Definition); ok {
						fields[i] = tr.mkEmbeddedStructField(
							tag,
							def,
						)
					}
					return nil, fmt.Errorf(
						"cannot use type %s as embedded",
						t,
					)
				} else {
					return nil, errors.New("Unable to parse struct field")
				}
			}

			s, clbk := tr.implStruct(fields)
			if clbk != nil {
				callbacks = append(callbacks, clbk)
			}
			return s, nil

		case *ast.FuncType:
			return funcSignature(n)

		case *ast.InterfaceType:
			meths := make([]*types.Method, len(n.Methods.List))
			for i, field := range n.Methods.List {
				f, err := funcSignature(field.Type.(*ast.FuncType))
				if err != nil {
					return nil, err
				}
				meths[i], err = tr.implMethod(
					field.Names[0].Name,
					f,
				)
				if err != nil {
					return nil, err
				}
			}

			i, clbk := tr.implInter(meths)
			if clbk != nil {
				callbacks = append(callbacks, clbk)
			}
			return i, nil

		case *ast.Field:
			ident, ok := n.Type.(*ast.Ident)
			if ok {
				t, err := types.String2Builtin(ident.Name)
				if err == nil {
					return t, nil
				}
			}
			return nil, nil

		case *ast.Ident:
			var (
				t   types.Type
				err error
			)
			t, err = tr.regBuiltin(n.Name)
			if err != nil {
				return identifier(n)
			}
			return t, nil

		case *ast.SelectorExpr:
			return identifier(n)

		case *ast.StarExpr:
			t, err := typeDec(n.X)
			if err != nil {
				return nil, err
			}
			return tr.implPtr(t), nil

		case *ast.ChanType:
			t, err := typeDec(
				n.Value,
			)
			if err != nil {
				return nil, err
			}
			return tr.implChan(t), nil

		default:
			return nil, fmt.Errorf(
				"No match for type declaration: %s",
				spew.Sdump(n),
			)
		}
	}

	funcSignature = func(node *ast.FuncType) (*types.Function, error) {
		args := make([]*types.NameTypePair, len(node.Params.List))
		for i, param := range node.Params.List {
			var n string
			if len(param.Names) > 0 {
				n = param.Names[0].Name
			}
			t, err := typeDec(param.Type)
			if err != nil {
				return nil, err
			}
			args[i] = tr.mkNameTypePair(n, t)
		}

		if node.Results == nil {
			return tr.implFunc(args, nil), nil
		}

		results := make([]*types.NameTypePair, len(node.Results.List))
		for i, res := range node.Results.List {
			var n string
			if len(res.Names) > 0 {
				n = res.Names[0].Name
			}
			t, err := typeDec(res.Type)
			if err != nil {
				return nil, err
			}
			results[i] = tr.mkNameTypePair(n, t)
		}

		return tr.implFunc(args, results), nil
	}

	identifier = func(node ast.Node) (*types.Definition, error) {
		switch n := node.(type) {
		case *ast.SelectorExpr:
			return tr.definitionRef(
				n.Sel.Name,
				n.X.(*ast.Ident).Name,
			)

		case *ast.Ident:
			return tr.definitionRef(n.Name, "")

		default:
			return nil, fmt.Errorf(
				"Can't parse identifier: %s",
				spew.Sdump(node),
			)
		}
	}

	isIgnorable = func(node ast.Node) bool {
		switch node.(type) {
		case *ast.ImportSpec:
		case *ast.CommentGroup:
		case *ast.Comment:
		default:
			return false
		}
		return true
	}

	for _, file := range pkg.Files {
		fileName = file.Name.Name
		for _, decl := range file.Decls {
			if err := do(decl); err != nil {
				return err
			}
		}
	}

	// shitty
	var i int
	for len(callbacks) > 0 {
		if i > 10000 {
			panic("unable to solve types")
		}
		clbk := callbacks[0]
		callbacks = callbacks[1:]
		if err := clbk(); err == errClbkRetry {
			callbacks = append(callbacks, clbk)
		} else if err != nil {
			return err
		}
		i++
	}

	return nil
}
