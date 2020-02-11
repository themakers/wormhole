package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/themakers/wormhole/parsex"
	"golang.org/x/tools/imports"
)

func PWD() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}

func parse(wd string) (string, []*parsex.Interface) {
	files := listSourceFiles(wd)
	log.Println(files)
	p, err := parsex.Parse(wd)
	if err != nil {
		panic(err)
	}
	return p.Pkg, p.Interfaces
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
		outFile := filepath.Join(PWD(), "wormhole.gen.go")
		bcpFile := filepath.Join(PWD(), "wormhole.gen.go.bak")
		{
			os.Remove(bcpFile)
			os.Rename(outFile, bcpFile)
			//defer (func() {
			//	if rec := recover(); rec != nil {
			//		stack := string(debug.Stack())
			//		if err, ok := rec.(template.ExecError); ok {
			//			log.Printf("PANIC: %#v\n", err.Err)
			//		} else {
			//			log.Printf("PANIC: %#v\n%s", rec, stack)
			//		}
			//		//os.Remove(outFile)
			//		//os.Rename(bcpFile, outFile)
			//	} else {
			//		os.Remove(bcpFile)
			//	}
			//})()
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
	var files []string
	for _, arbitraryFile := range arbitraryFiles {
		if !arbitraryFile.IsDir() &&
			strings.HasSuffix(arbitraryFile.Name(), ".go") &&
			!strings.HasSuffix(arbitraryFile.Name(), "_test.go") &&
			!strings.HasSuffix(arbitraryFile.Name(), "wormhole.gen.go") {

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
