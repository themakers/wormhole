package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"golang.org/x/tools/imports"
)

func PWD() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}

type InterfaceMethodArgument struct {
	Name string
	Type string
}

type InterfaceMethod struct {
	Name string
	Args []InterfaceMethodArgument
	Rets []InterfaceMethodArgument
}

type Interface struct {
	Name    string
	Methods []InterfaceMethod
}

func parseMethod(ln string) (im InterfaceMethod) {
	ln = strings.Trim(ln, "\n\t\r ")

	r := bufio.NewReader(bytes.NewBufferString(ln))
	methName, err := r.ReadString(byte('('))
	if err != nil {
		panic(err)
	}
	im.Name = methName[:len(methName)-1]

	// r.UnreadByte()

	parseArgs := func() (imas []InterfaceMethodArgument) {
		for {
			b, _ := r.ReadByte()
			if string(b) != " " {
				if string(b) != "(" {
					r.UnreadByte()
				}
				break
			}
		}

		{
			var (
				plvl      = 0
				stateName = true
				ima       = new(InterfaceMethodArgument)
			)
			put := func(c rune) {
				if stateName {
					ima.Name += string(c)
				} else {
					ima.Type += string(c)
				}
			}
			flush := func() {
				imas = append(imas, *ima)
				ima = new(InterfaceMethodArgument)
				stateName = true
			}

		L:
			for {
				c := (func() rune {
					b, err := r.ReadByte()
					if err != nil {
						if err == io.EOF {
							return 0
						}
						panic(err)
					}
					return rune(b)
				})()

				if c == 0 {
					flush()
					break L
				}

				switch c {
				case '(', '{':
					plvl++
					if plvl != 0 {
						put(c)
					}
				case ')', '}':
					plvl--
					if plvl == -1 {
						flush()
						break L
					} else {
						put(c)
					}
				case ',':
					if plvl == 0 {
						flush()
						r.ReadByte()
					} else {
						put(c)
					}
				case ' ':
					if stateName {
						stateName = false
					} else {
						put(c)
					}
				default:
					put(c)
				}
			}
		}

		return
	}

	im.Args = parseArgs()

	im.Rets = parseArgs()

	// FIXME: Hack
	if len(im.Rets) == 1 && im.Rets[0].Name == "" && im.Rets[0].Type == "" {
		im.Rets = []InterfaceMethodArgument{}
	}

	normalize := func(imas []InterfaceMethodArgument, suffix string) {
		mixed := false
		for _, ima := range imas {
			if ima.Type != "" {
				mixed = true
				break
			}
		}

		if !mixed {
			for i := range imas {
				imas[i].Type = imas[i].Name
				imas[i].Name = fmt.Sprintf("%s%d", suffix, i)
			}
		}
	}

	normalize(im.Args, "a")
	normalize(im.Rets, "r")

	log.Printf("ARGS: %#v", im.Args)
	log.Printf("RETS: %#v", im.Rets)

	return
}

func parseFile(path string) (pkg string, ifcs []Interface) {
	lines := (func() []string {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		return strings.Split(string(data), "\n")
	})()

	for i := 0; i < len(lines); i++ {
		ln := strings.Trim(lines[i], "\n\r\t ")

		spl := strings.Split(ln, " ")
		if len(spl) == 4 && spl[0] == "type" && spl[2] == "interface" && spl[3] == "{" {
			ifc := Interface{
				Name: spl[1],
			}
			log.Println("IFC", ifc.Name)

			for i++; i < len(lines); i++ {
				ln := strings.Trim(lines[i], "\n\r\t ")
				if ln == "}" {
					break
				} else if ln != "" {
					ifc.Methods = append(ifc.Methods, parseMethod(ln))
				}
			}

			ifcs = append(ifcs, ifc)
		} else if len(spl) == 2 && spl[0] == "package" {
			pkg = spl[1]
		}
	}

	return
}

func parse(wd string) (pkg string, interfaces []Interface) {
	files := listSourceFiles(wd)
	log.Println(files)

	for _, file := range files {
		p, ifcs := parseFile(file)
		pkg = p
		interfaces = append(interfaces, ifcs...)
	}

	return
}

func main() {

	writeCode := func(fname string, code []byte) {
		err := ioutil.WriteFile(fname, code, 0777)
		if err != nil {
			panic(err)
		}
	}

	switch os.Args[1] {
	case "go":
		outFile := filepath.Join(PWD(), "nowire.gen.go")
		bcpFile := filepath.Join(PWD(), "nowire.gen.go.bak")
		{
			os.Remove(bcpFile)
			os.Rename(outFile, bcpFile)
			defer (func() {
				if rec := recover(); rec != nil {
					log.Printf("PANIC: %#v\n%s", rec, debug.Stack())
					//os.Remove(outFile)
					//os.Rename(bcpFile, outFile)
				} else {
					os.Remove(bcpFile)
				}
			})()
		}
		pkg, ifaces := parse(PWD())
		if len(ifaces) == 0 {
			return
		}

		code := []byte(Render(pkg, ifaces))
		writeCode(outFile, code)
		code, err := imports.Process(outFile, code, &imports.Options{
			Fragment:   false,
			AllErrors:  true,
			Comments:   true,
			TabIndent:  true,
			TabWidth:   8,
			FormatOnly: false,
		})
		if err != nil {
			panic(err)
		}
		writeCode(outFile, code)

	default:
		log.Println("usage:?")
	}
}

func listSourceFiles(dir string) []string {
	arbitraryFiles, err := ioutil.ReadDir(dir)
	perr(err)
	files := []string{}
	for _, arbitraryFile := range arbitraryFiles {
		if !arbitraryFile.IsDir() &&
			strings.HasSuffix(arbitraryFile.Name(), ".go") &&
			!strings.HasSuffix(arbitraryFile.Name(), "_test.go") &&
			!strings.HasSuffix(arbitraryFile.Name(), "warp.gen.go") {

			files = append(files, filepath.Join(dir, arbitraryFile.Name()))
		}
	}
	return files
}

func perr(err error) {
	if err != nil {
		panic(err)
	}
}
