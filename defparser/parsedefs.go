package defparser

import (
	"fmt"
	"go/ast"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/themakers/wormhole/defparser/types"
)

func parseDefs(tc *typeChecker, pkg *ast.Package) error {
	var (
		parseTypeDefinition   func(*ast.TypeSpec) (*types.Definition, error)
		parseMethodDefinition func(*ast.FuncDecl) (*types.Method, error)
		parseTypeDeclaration  func(ast.Node) (types.Type, error)
		fileName              string
	)

	parseFuncSignature := func(node *ast.FuncType) (*types.Function, error) {
		var args []types.NameTypePair
		for _, param := range node.Params.List {
			var n string
			if len(param.Names) > 0 {
				n = param.Names[0].Name
			}
			v, err := parseTypeDeclaration(param.Type)
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
			v, err := parseTypeDeclaration(res.Type)
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

	// parseMethod := func(node *ast.FuncType) (*types.Method, error) {
	// 	return nil, nil
	// }

	isIgnorable := func(node ast.Node) bool {
		switch node.(type) {
		case *ast.ImportSpec:
		case *ast.CommentGroup:
		case *ast.Comment:
		default:
			return false
		}
		return true
	}

	parseTypeDefinition = func(d *ast.TypeSpec) (*types.Definition, error) {
		name := d.Name.Name

		t, err := parseTypeDeclaration(d.Type)
		if err != nil {
			return nil, err
		}
		return tc.def(name, t), nil
	}

	parseTypeDeclaration = func(node ast.Node) (types.Type, error) {
		switch n := node.(type) {
		case *ast.MapType:
			k, err := parseTypeDeclaration(n.Key)
			if err != nil {
				return nil, err
			}
			v, err := parseTypeDeclaration(n.Value)
			if err != nil {
				return nil, err
			}

			return tc.implMap(k, v), nil

		case *ast.ArrayType:
			t, err := parseTypeDeclaration(n.Elt)
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

			// s := make(types.Struct)
			for i, field := range n.Fields.List {
				t, err := parseTypeDeclaration(field.Type)
				if err != nil {
					return nil, err
				}

				var tag string
				if field.Tag != nil {
					tag = field.Tag.Value
				}

				fields[i] = types.StructField{
					Tag:  tag,
					Type: t,
				}
			}
			return tc.implStruct(fields), nil

		case *ast.FuncType:
			return parseFuncSignature(n)

		case *ast.InterfaceType:
			meths := make([]*types.Method, len(n.Methods.List))
			for i, field := range n.Methods.List {
				f, err := parseFuncSignature(field.Type.(*ast.FuncType))
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
			t, err := types.String2Builtin(n.Name)
			if err == nil {
				return t, nil
			}
			return nil, fmt.Errorf(
				"Unknown ast.Ident: %s",
				spew.Sdump(n),
			)
			// return Type{
			// 	Name: n.Name,
			// }, nil

		case *ast.SelectorExpr:
			return tc.defRef(
				n.Sel.Name,
				n.X.(*ast.Ident).Name,
			)
			return Type{
				From: 
				Name: 
			}, nil

		case *ast.StarExpr:
			t, err := parseTypeDeclaration(n.X)
			if err != nil {
				return nil, err
			}
			return Pointer{
				Type: t,
			}, nil

		case *ast.ChanType:
			t, err := parseTypeDeclaration(
				n.Value.(*ast.Ident),
			)
			if err != nil {
				return nil, err
			}
			return Chan{
				Type: t,
			}, nil

		default:
			return nil, fmt.Errorf(
				"No match for type declaration: %s",
				spew.Sdump(n),
			)
		}
	}

	parseMethodDefinition = func(dec *ast.FuncDecl) (Method, error) {
		var meth Method

		meth.Name = dec.Name.Name

		t, err := parseTypeDeclaration(dec.Recv.List[0].Type)
		if err != nil {
			return Method{}, err
		}
		meth.Receiver = t

		err = parseFuncSignature(dec.Type, &meth.Signature)
		return meth, err
	}

	parseFuncDefinition := func(dec *ast.FuncDecl) (Function, error) {
		var f Function
		if dec.Name == nil {
			return Function{}, fmt.Errorf(
				"Invalid function definition: no name were specified: %s",
				spew.Sdump(dec),
			)
		}
		f.Name = dec.Name.Name
		err := parseFuncSignature(dec.Type, &f)
		return f, err
	}

	var (
		types     []Type
		methods   []Method
		functions []Function
		parse     func(node ast.Node) error
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
			t, err := parseTypeDefinition(n)
			if err != nil {
				return err
			}
			types = append(types, t)

		case *ast.FuncDecl:
			if n.Recv == nil {
				f, err := parseFuncDefinition(n)
				if err != nil {
					return nil
				}
				functions = append(functions, f)
			} else {
				meth, err := parseMethodDefinition(n)
				if err != nil {
					return err
				}
				methods = append(methods, meth)
			}

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

	for _, file := range pkg.Files {
		fileName = file.Name.Name
		for _, decl := range file.Decls {
			if err := parse(decl); err != nil {
				return nil, nil, nil, err
			}
		}
	}

	return types, methods, functions, nil
}
