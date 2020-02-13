package wormparse

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Parse(pkgPath string) (*Package, error) {
	var (
		index          int
		parsedPackages = make(map[PackageInfo]*Package)
		parse          func(pkgFullPath, pkgPath string, prev map[string]int) (*Package, error)
	)

	parse = func(pkgFullPath, pkgPath string, prev map[string]int) (*Package, error) {
		pkgFullPath = filepath.Clean(pkgFullPath)
		if !filepath.IsAbs(pkgFullPath) {
			return nil, ErrNotAbsoluteFilePath
		}

		pkgIndx, ok := prev[pkgFullPath]
		if ok {
			return nil, Loop{
				index: pkgIndx,
			}
		}

		var res Package
		m := make(map[string]int)
		{
			pkgIndx = index
			index++
			for k, v := range prev {
				m[k] = v
			}
			m[pkgFullPath] = pkgIndx
		}

		pkgs, err := parser.ParseDir(
			token.NewFileSet(),
			pkgFullPath,
			nil,
			0,
		)
		if err != nil {
			return nil, err
		}

		{
			var pkgName string
			{
				var (
					fmtStr string
					i      int
				)
				for pkg := range pkgs {
					if !strings.HasSuffix(pkg, "_test") {
						fmtStr += fmt.Sprintf("\n%s", pkg)
						pkgName = pkg
						i++
					}
				}
				if i == 0 {
					return nil, PackagingError(fmt.Errorf(""+
						"No Go packages were defined "+
						" in specified directory: %s",
						pkgFullPath,
					))
				} else if i > 1 {
					return nil, PackagingError(fmt.Errorf("" +
						"More than 1 package were defined:" +
						fmtStr +
						"in specified directory: %s" +
						pkgFullPath,
					))
				}
			}

			imps := make(map[string]string)
			for _, file := range pkgs[pkgName].Files {
				for _, imp := range file.Imports {
					s := imp.Path.Value
					s = s[1 : len(s)-1]
					if imp.Name != nil {
						imps[s] = imp.Name.Name
					} else {
						imps[s] = ""
					}
				}
			}

			res.Info = PackageInfo{
				PkgName:     pkgName,
				PkgPath:     pkgPath,
				PkgFullPath: pkgFullPath,
			}
			{
				res, ok := parsedPackages[res.Info]
				if ok {
					return res, nil
				}
			}

			res.Imports = make([]Import, len(imps))

			var (
				loops = make(map[int]Loop)
				i     int

				mergeLoops = func(src Loop) {
					dst, ok := loops[src.index]
					if !ok {
						loops[src.index] = src
						return
					}

					for _, v := range dst.Nodes {
						f := true
						for _, w := range src.Nodes {
							if v == w {
								f = false
								break
							}
						}
						if f {
							src.Nodes = append(src.Nodes, v)
						}
					}

					loops[src.index] = src
				}
			)

			for imp, alias := range imps {
				if _, err := os.Stat(path.Join(GOSRC, imp)); !os.IsNotExist(err) {
					impPath := path.Join(GOSRC, imp)
					pkg, err := parse(impPath, imp, m)

					switch v := err.(type) {
					case Loop:
						mergeLoops(v)

					case Loops:
						for _, loop := range v {
							mergeLoops(loop)
						}

					case error:
						return nil, v

					default:
						res.Imports[i] = Import{
							Alias:   alias,
							Package: pkg,
						}
					}
				} else if _, err := os.Stat(path.Join(GOSTD, imp)); !os.IsNotExist(err) {
					impPath := path.Join(GOSTD, imp)
					var name string
					{
						s := strings.Split(imp, "/")
						name = s[len(s)-1]
					}

					res.Imports[i] = Import{
						Alias: alias,
						Package: &Package{
							Info: PackageInfo{
								PkgName:     name,
								PkgPath:     imp,
								PkgFullPath: impPath,
								Std:         true,
							},
						},
					}
				} else {
					return nil, PackagingError(fmt.Errorf(
						"Package weren't found: %s",
						imp,
					))
				}

				i++
			}

			if len(loops) > 0 {
				loop, ok := loops[pkgIndx]
				if ok {
					loop.Nodes = append(loop.Nodes, res.Info)
					loops[pkgIndx] = loop
				}

				var res Loops
				for _, loop := range loops {
					res = append(res, loop)
				}

				return nil, res
			}
		}

		parsedPackages[res.Info] = &res
		return &res, err
	}

	return parse(pkgPath, "", make(map[string]int))
}
