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
		astwalker.WalkAST(f, nil, InterfaceMethodWalker(func(pkg, ifcName, method, arg, res string) {
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
				Arg:       arg,
				Ret:       res,
			}

			ifc.Methods = append(ifc.Methods, meth)
			ifc.MethodsMap[meth.Name] = meth
		}))
	}

	return t, nil
}

func InterfaceMethodWalker(h func(pkg, ifc, method, arg, res string)) astwalker.VisitorFunc {
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
			//> Should have 2 params and 2 results
			if len(ft.Params.List) == 2 && len(ft.Results.List) == 2 {

				argType := fmt.Sprint(ft.Params.List[1].Type)
				resType := fmt.Sprint(ft.Results.List[0].Type)

				//> Should have context as 1st param
				if tpe, ok := ft.Params.List[0].Type.(*ast.SelectorExpr); ok && fmt.Sprint(tpe.X) == "context" && fmt.Sprint(tpe.Sel) == "Context" {
					//> Should have error as second result
					if fmt.Sprint(ft.Results.List[1].Type) == "error" {
						h(pkg, ifcName, methName, argType, resType)
					} else {
						// TODO Warn
					}
				} else {
					// TODO Warn
				}
			} else {
				// TODO Warn
			}
		} else if node.Top(0).IsIdent() &&
			node.Top(1).IsFile(){
			pkg = node.Top(0).Ident().Name
		}
		return nil
	}
}
