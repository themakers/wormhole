package defparser

import (
	"fmt"
	"go/ast"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/themakers/wormhole/defparser/types"
)

func aggregateDefinitions(tc *typeChecker, pkg *ast.Package) error {
	var (
		parse           func(node ast.Node) error
		typeDef         func(d *ast.TypeSpec) error
		funcDef         func(dec *ast.FuncDecl) error
		methodDef       func(dec *ast.FuncDecl) error
		isIgnorable     func(node ast.Node) bool
		funcSignature   func(node *ast.FuncType) (*types.Function, error)
		typeDeclaration func(ast.Node) (types.Type, error)

		fileName string
	)

	parse = func(node ast.Node) error {
		switch n := node.(type) {
		case *ast.GenDecl:
			for _, spec := range n.Specs {
				if err := parse(spec); err != nil {
					return err
				}
			}

		case *ast.TypeSpec:
			return typeDef(n)

		case *ast.FuncDecl:
			if n.Recv == nil {
				return funcDef(n)
			}
			return methodDef(n)

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

		return nil
	}

	typeDef = func(d *ast.TypeSpec) error {
		t, err := typeDeclaration(d.Type)
		if err != nil {
			return err
		}
		_, err = tc.def(d.Name.Name, t)
		return err
	}

	funcDef = func(dec *ast.FuncDecl) error {
		t, err := funcSignature(dec.Type)
		if err != nil {
			return err
		}
		_, err = tc.def(dec.Name.Name, t)
		return err
	}

	methodDef = func(dec *ast.FuncDecl) error {
		t, err := typeDeclaration(dec.Recv.List[0].Type)
		if err != nil {
			return err
		}
		f, err := funcSignature(dec.Type)
		if err != nil {
			return err
		}

		_, err = tc.def(dec.Name.Name, tc.meth(dec.Name.Name, t, f))
		return err
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

	funcSignature = func(node *ast.FuncType) (*types.Function, error) {
		var args []types.NameTypePair
		for _, param := range node.Params.List {
			var n string
			if len(param.Names) > 0 {
				n = param.Names[0].Name
			}
			v, err := typeDeclaration(param.Type)
			if err != nil {
				return nil, err
			}
			args = append(args, types.NameTypePair{
				Name: n,
				Type: v,
			})
		}

		if node.Results == nil {
			return tc.implFunc(args, nil), nil
		}

		var results []types.NameTypePair
		for _, res := range node.Results.List {
			var n string
			if len(res.Names) > 0 {
				n = res.Names[0].Name
			}
			v, err := typeDeclaration(res.Type)
			if err != nil {
				return nil, err
			}
			results = append(results, types.NameTypePair{
				Name: n,
				Type: v,
			})
		}

		return tc.implFunc(args, results), nil
	}

	typeDeclaration = func(node ast.Node) (types.Type, error) {
		switch n := node.(type) {
		case *ast.MapType:
			k, err := typeDeclaration(n.Key)
			if err != nil {
				return nil, err
			}
			v, err := typeDeclaration(n.Value)
			if err != nil {
				return nil, err
			}

			return tc.implMap(k, v), nil

		case *ast.ArrayType:
			t, err := typeDeclaration(n.Elt)
			if err != nil {
				return nil, err
			}

			if n.Len == nil {
				return tc.implSlice(t), nil
			}

			l, err := strconv.Atoi(n.Len.(*ast.BasicLit).Value)
			if err != nil {
				return nil, err
			}
			return tc.implArray(l, t), nil

		case *ast.StructType:
			fields := make([]types.StructField, len(n.Fields.List))
			for i, field := range n.Fields.List {
				t, err := typeDeclaration(field.Type)
				if err != nil {
					return nil, err
				}

				var tag string
				if field.Tag != nil {
					tag = field.Tag.Value
				}

				var name string
				if len(field.Names) == 1 {
					name = field.Names[0].Name
				} else {
					def := t.(*types.Definition)
					name = def.Name
				}

				fields[i] = types.StructField{
					Name: name,
					Tag:  tag,
					Type: t,
				}
			}
			return tc.implStruct(fields), nil

		case *ast.FuncType:
			return funcSignature(n)

		case *ast.InterfaceType:
			meths := make([]*types.Method, len(n.Methods.List))
			for i, field := range n.Methods.List {
				f, err := funcSignature(field.Type.(*ast.FuncType))
				if err != nil {
					return nil, err
				}
				meths[i] = tc.meth(field.Names[0].Name, types.Untyped, f)
			}
			return tc.implInter(meths), nil

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
			t, err = tc.regBuiltin(n.Name)
			if err != nil {
				t, err = tc.defRef(n.Name, "")
				if err != nil {
					return nil, fmt.Errorf(
						"Failed to parse %s :: %s",
						spew.Sdump(n),
						err,
					)
				}
			}
			return t, nil

		case *ast.SelectorExpr:
			return tc.defRef(
				n.Sel.Name,
				n.X.(*ast.Ident).Name,
			)

		case *ast.StarExpr:
			t, err := typeDeclaration(n.X)
			if err != nil {
				return nil, err
			}
			return tc.implPtr(t), nil

		case *ast.ChanType:
			t, err := typeDeclaration(
				n.Value,
			)
			if err != nil {
				return nil, err
			}
			return tc.implChan(t), nil

		default:
			return nil, fmt.Errorf(
				"No match for type declaration: %s",
				spew.Sdump(n),
			)
		}
	}

	for _, file := range pkg.Files {
		fileName = file.Name.Name
		for _, decl := range file.Decls {
			if err := parse(decl); err != nil {
				return err
			}
		}
	}

	for name := range tc.undefinedIdentifiers {
		return fmt.Errorf(
			"Undefined identifier \"%s\" in package %s",
			name,
			tc.pkg.Info.PkgPath,
		)
	}

	return nil
}
