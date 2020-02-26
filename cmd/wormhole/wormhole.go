package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/themakers/wormhole/defparser"
)

func PWD() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}

func parse(wd string) *defparser.Result {
	files := listSourceFiles(wd)
	log.Println(files)

	/*
		History, lol
		p, err := parsex.Parse(wd)
		pkg, err := wormparse.Parse(wd)
		res, err := decparse.Parse(wd)
	*/

	res, err := defparser.Parse(wd)
	if err != nil {
		switch v := err.(type) {
		case defparser.Loop:
			fmt.Println("LOOP")
			spew.Dump(v)
		default:
			panic(err)
		}
	}

	// func() {
	// 	f, err := os.Create("debug.log")
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	defer f.Close()
	// 	logger := log.New(f, "", 0)

	// 	log := func(v interface{}) {
	// 		ts := v.([]types.Type)
	// 		for _, t := range ts {
	// 			logger.Printf(
	// 				"====\n%s\n%s\n====\n\n",
	// 				t.Hash(),
	// 				t,
	// 			)
	// 		}
	// 		logger.Print("##############\n\n")
	// 	}

	// 	log(res.Packages)
	// 	log(res.STDPackages)
	// 	log(res.Definitions)
	// 	log(res.STDDefinitions)
	// 	log(res.Methods)
	// }()

	// spew.Dump(res.Packages[0].Info.PkgPath)

	fmt.Println("\n\n\n\n###### DONE ########\n\n\n\n")

	spew.Dump(res)

	panic("FINISH")

	return nil
}

func main() {
	// writeCode := func(fname string, code []byte) {
	// 	err := ioutil.WriteFile(fname, code, 0777)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

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
		_ = parse(PWD())
		// if len(ifaces) == 0 {
		// 	return
		// }

		// code := []byte(Render(pkg))
		// writeCode(outFile, code)
		// code, err := imports.Process(outFile, code, &imports.Options{
		// 	Fragment:   false,
		// 	AllErrors:  true,
		// 	Comments:   true,
		// 	TabIndent:  true,
		// 	TabWidth:   8,
		// 	FormatOnly: false,
		// })
		// if err != nil {
		// 	panic(err)
		// }
		// writeCode(outFile, code)

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
