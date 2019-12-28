package parsex

import (
	"fmt"
	"github.com/themakers/wormhole/parsex/astwalker"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// TODO Implement post loading of unresolved types. i.e. when dir name differ from pkg name, or pkg imported as .

var (
	GOROOT = filepath.Clean(os.Getenv("GOROOT"))
	GOPATH = filepath.Clean(os.Getenv("GOPATH"))
	GOSRC  = filepath.Join(GOPATH, "src")
)

func PWD() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}

func Parse(files ...string) (*Parsed, error) {
	t := &Parsed{
		InterfacesMap: map[string]*Interface{},
	}

	fset := token.NewFileSet()

	for _, file := range files {
		file := filepath.Clean(file)

		if !filepath.IsAbs(file) {
			panic("Only absolute paths are allowed. " + file)
		}

		log.Println("parsing", file)

		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments) // |parser.Trace
		if err != nil {
			return nil, err
		}

		//fileStr := (func() string {
		//	data, err := ioutil.ReadFile(file)
		//	if err != nil {
		//		panic(err)
		//	}
		//	return string(data)
		//})()

		//pkgPath := strings.TrimPrefix(strings.TrimSuffix(strings.TrimSuffix(file, filepath.Base(file)), "/"), GOSRC+"/")
		//pkgName := f.Message.Message

		fmt.Println(strings.Repeat("•", 64))
		astwalker.WalkAST(f, nil, testWalker()).Root().Children()[0].Children()
		fmt.Println(strings.Repeat("•", 64))
		astwalker.WalkAST(f, nil, InterfaceMethodWalker(func(pkg, ifcName, method string, arg, res []Param) {
			t.Pkg = pkg

			ifc := t.InterfacesMap[ifcName]
			if ifc == nil {
				ifc = &Interface{
					Name:       ifcName,
					MethodsMap: map[string]*Method{},
				}
				t.Interfaces = append(t.Interfaces, ifc)
				t.InterfacesMap[ifcName] = ifc
			}

			meth := &Method{
				Interface: ifcName,
				Name:      method,
				Args:      arg,
				Rets:      res,
			}

			ifc.Methods = append(ifc.Methods, meth)
			ifc.MethodsMap[meth.Name] = meth
		}))
	}

	return t, nil
}

func typeName(e ast.Expr) string {
	if sel, ok := e.(*ast.SelectorExpr); ok {
		tx := fmt.Sprint(sel.X)
		tsel := fmt.Sprint(sel.Sel)
		if tx != "" {
			return fmt.Sprintf("%s.%s", tx, tsel)
		} else {
			return tsel
		}
	} else if idn, ok := e.(*ast.Ident); ok {
		return idn.Name
	} else if sexp, ok := e.(*ast.StarExpr); ok {
		return fmt.Sprintf("*%s", typeName(sexp.X))
	} else {
		panic(fmt.Sprintf("unknown element: %#v", e))
	}
}

func InterfaceMethodWalker(h func(pkg, ifc, method string, args, ress []Param)) astwalker.VisitorFunc {
	pkg := ""
	return func(node *astwalker.Node) astwalker.VisitorFunc {

		//> Should be interface method
		if node.Top(0).IsFuncType() &&
			len(node.Top(1).Children()) == 2 && node.Top(1).Children()[0].IsIdent() &&
			node.Top(3).IsInterfaceType() &&
			node.Top(4).IsTypeSpec() &&
			node.Top(5).IsGenDecl() &&
			len(node.Top(4).Children()) == 2 && node.Top(4).Children()[0].IsIdent() {

			ifcName := node.Top(4).Children()[0].Ident().Name
			methName := node.Top(1).Children()[0].Ident().Name

			ft := node.FuncType()

			var i = 0
			getParams := func(fields []*ast.Field) (params []Param) {
				for _, p := range fields {
					var par Param
					if len(p.Names) > 0 {
						for _, n := range p.Names {
							i++
							if n.Name == "" {
								n.Name = fmt.Sprintf("__%d", i)
							}
							par.Names = append(par.Names, n.Name)
						}
					} else {
						i++
						par.Names = append(par.Names, fmt.Sprintf("__%d", i))
					}
					par.Type = typeName(p.Type)
					params = append(params, par)
				}

				return params
			}

			var args []Param

			var ress []Param

			if len(ft.Params.List) > 1 {
				args = getParams(ft.Params.List[1:])
			}

			if len(ft.Results.List) > 0 {
				ress = getParams(ft.Results.List)
			}

			h(pkg, ifcName, methName, args, ress)
		} else if node.Top(0).IsIdent() &&
			node.Top(1).IsFile() {
			pkg = node.Top(0).Ident().Name
		}
		return nil
	}
}
