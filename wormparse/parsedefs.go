package wormparse

import (
	"errors"
	"fmt"
	"go/ast"
	"strconv"

	"github.com/davecgh/go-spew/spew"
)

func parseDefs(pkgInfo PackageInfo, pkg *ast.Package) ([]Type, []Method, error) {
	var (
		parseTypeDefinition   func(*ast.TypeSpec) (Type, error)
		parseMethodDefinition func(*ast.FuncDecl) (Method, error)
		parseTypeDeclaration  func(ast.Node) (interface{}, error)
	)

	matchBasicType := func(t string) (Type, error) {
		res := Type{
			Basic: true,
		}
		switch t {
		case "bool":
			fallthrough
		case "int":
			fallthrough
		case "int64":
			fallthrough
		case "int32":
			fallthrough
		case "uint":
			fallthrough
		case "uint64":
			fallthrough
		case "uint32":
			fallthrough
		case "string":
			fallthrough
		case "rune":
			fallthrough
		case "byte":
			fallthrough
		case "error":
			res.Name = t
		default:
			return res, fmt.Errorf("No std type matches: %s", t)
		}

		return res, nil
	}

	parseFuncSignature := func(node *ast.FuncType, f *Function) error {
		for _, param := range node.Params.List {
			var n string
			if len(param.Names) > 0 {
				n = param.Names[0].Name
			}
			v, err := parseTypeDeclaration(param.Type)
			if err != nil {
				return err
			}
			f.Args = append(f.Args, NameTypePair{
				Name: n,
				Type: v.(Type),
			})
		}

		if node.Results == nil {
			return nil
		}

		for _, res := range node.Results.List {
			var n string
			if len(res.Names) > 0 {
				n = res.Names[0].Name
			}
			v, err := parseTypeDeclaration(res.Type)
			if err != nil {
				return err
			}
			f.Return = append(f.Return, NameTypePair{
				Name: n,
				Type: v,
			})
		}

		return nil
	}

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

	parseTypeDefinition = func(d *ast.TypeSpec) (res Type, err error) {
		res.Name = d.Name.Name

		switch n := d.Type.(type) {
		case *ast.InterfaceType:
			var err error
			res.Definition, err = parseTypeDeclaration(n)
			if err != nil {
				return Type{}, err
			}

		case *ast.StructType:
			var err error
			res.Definition, err = parseTypeDeclaration(n)
			if err != nil {
				return Type{}, err
			}
		}

		return res, nil
	}

	parseTypeDeclaration = func(node ast.Node) (res interface{}, err error) {
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

			return Map{
				Key:   k,
				Value: v,
			}, nil

		case *ast.ArrayType:
			t, err := parseTypeDeclaration(n.Elt)
			if err != nil {
				return nil, err
			}

			if n.Len == nil {
				return Slice{
					Type: t,
				}, nil
			}

			l, err := strconv.Atoi(n.Len.(*ast.BasicLit).Value)
			if err != nil {
				return nil, err
			}
			return Array{
				Len:  l,
				Type: t,
			}, nil

		case *ast.StructType:
			s := make(Struct)
			for _, field := range n.Fields.List {
				t, err := parseTypeDeclaration(field.Type)
				if err != nil {
					return Type{}, err
				}

				var tag string
				if field.Tag != nil {
					tag = field.Tag.Value
				}

				s[field.Names[0].Name] = TagTypePair{
					Tag:  tag,
					Type: t,
				}
			}
			return s, err

		case *ast.FuncType:
			var f Function
			err := parseFuncSignature(n, &f)
			return f, err

		case *ast.InterfaceType:
			var i Interface
			for _, field := range n.Methods.List {
				var f Function
				f.Name = field.Names[0].Name
				if err := parseFuncSignature(
					field.Type.(*ast.FuncType),
					&f,
				); err != nil {
					return Type{}, err
				}
				i.Methods = append(i.Methods, f)
			}
			return i, nil

		case *ast.Field:
			ident, ok := n.Type.(*ast.Ident)
			if ok {
				t, err := matchBasicType(ident.Name)
				if err == nil {
					return t, nil
				}
			}
			return nil, nil

		case *ast.Ident:
			t, err := matchBasicType(n.Name)
			if err == nil {
				return t, nil
			}
			return Type{
				Name: n.Name,
			}, nil

		case *ast.SelectorExpr:
			return Type{
				From: n.X.(*ast.Ident).Name,
				Name: n.Sel.Name,
			}, nil

		case *ast.StarExpr:
			t, err := parseTypeDeclaration(n.X)
			if err != nil {
				return nil, nil
			}
			return Pointer{
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
		meth.Reciever = t

		err = parseFuncSignature(dec.Type, &meth.Signature)
		return meth, err
	}

	var (
		types   []Type
		methods []Method
		parse   func(node ast.Node) error
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
				return nil
			}
			meth, err := parseMethodDefinition(n)
			if err != nil {
				return err
			}
			methods = append(methods, meth)

		default:
			if isIgnorable(node) {
				return nil
			}

			return errors.New("No matches")
		}

		return nil
	}

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			if err := parse(decl); err != nil {
				return nil, nil, err
			}
		}
	}

	return types, methods, nil
}
