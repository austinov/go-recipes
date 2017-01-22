package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n\n")
		fmt.Fprintf(os.Stderr, "  %s [flags] [source directories]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	var (
		dstPath string
		dstPack string
	)
	flag.StringVar(&dstPath, "dst-path", "", "destination path to store files")
	flag.StringVar(&dstPack, "dst-pack", "", "destination package name")
	flag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("args: %v\n", flag.Args())

	for _, arg := range flag.Args() {
		var (
			pack    *build.Package
			newPath = dstPath
			newPack = dstPack
		)
		s, err := os.Stat(arg)
		if err == nil && s.IsDir() {
			pack, err = build.ImportDir(arg, 0)
		}
		if os.IsNotExist(err) {
			err = nil
		}
		if pack == nil && err == nil {
			pack, err = build.Import(arg, wd, 0)
		}
		if err != nil {
			log.Fatalf("%s: %s", arg, err)
		}
		if dstPath == "" {
			newPath = pack.Dir
		}
		if dstPack == "" {
			newPack = pack.Name
		}

		importPath := ""
		if pack.ImportPath != "" && pack.ImportPath != "." {
			importPath = fmt.Sprintf("\n%q", pack.ImportPath)
		}
		var changed bool
		for _, f := range pack.GoFiles {
			err = processFile(pack.Dir, f, pack.Name, newPath, newPack, importPath)
			if err != nil {
				log.Fatalf("%s %s: %s", arg, f, err)
			}
			changed = true
		}

		if changed {
			gofmt(pack.Dir)
		}
	}
}

func processFile(path, file, pack, newPath, newPack, importPath string) error {
	log.Printf("processFile:\n\tsources path=%q file=%q pack=%q\n\tdestinations path=%q pack=%q importPath=%q\n", path, file, pack, newPath, newPack, importPath)

	structs, err := Structs(filepath.Join(path, file), pack, newPack)
	if err != nil {
		return err
	}

	if len(structs) == 0 {
		return nil
	}

	ext := filepath.Ext(file)
	base := strings.TrimSuffix(file, ext)
	f, err := os.Create(filepath.Join(newPath, base+"_genorm"+ext))
	if err != nil {
		return err
	}
	defer f.Close()

	if err = importTemplate.Execute(f, []string{newPack, importPath}); err != nil {
		return err
	}
	if err = bodyTemplate.Execute(f, structs); err != nil {
		return err
	}
	return nil
}

func Structs(path, pack, dstPack string) ([]StructInfo, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	v := &visitor{
		pack:     pack,
		dstPack:  dstPack,
		file:     file,
		fset:     fset,
		structs:  make(map[string]StructInfo),
		notfound: make(map[string][]string),
	}
	ast.Walk(v, file)

	if len(v.errors) > 0 {
		return nil, fmt.Errorf("%+v", v.errors)
	}

	for embedType, structTypes := range v.notfound {
		structInfo, ok := v.structs[embedType]
		if !ok {
			return nil, fmt.Errorf("embedded type %s not found in data", embedType)
		}
		for _, st := range structTypes {
			structType, ok := v.structs[st]
			if !ok {
				return nil, fmt.Errorf("owner %s of embedded type %s not found in data", st, embedType)
			}
			if structInfo.PKIndex >= 0 && structType.PKIndex >= 0 {
				v.errors = append(v.errors, fmt.Errorf(`%s has field with duplicate "pk" label in tag, it is not allowed`, structType.Type))
			}
			// append embedded fields to head of array
			structType.Fields = append(structType.Fields[:0], append(structInfo.Fields, structType.Fields[0:]...)...)
			structType.PKIndex = structInfo.PKIndex
			v.structs[st] = structType
		}
	}

	for _, si := range v.structs {
		if err = checkFields(&si); err != nil {
			return nil, err
		}
	}

	if len(v.errors) > 0 {
		return nil, fmt.Errorf("%+v", v.errors)
	}

	i := 0
	res := make([]StructInfo, len(v.structs))
	for _, v := range v.structs {
		res[i] = v
		i++
	}

	return res, nil
}

func checkFields(res *StructInfo) error {
	if len(res.Fields) == 0 {
		return fmt.Errorf(`%s has no fields with "genorm:" tag`, res.Type)
	}

	dupes := make(map[string]string)
	for _, f := range res.Fields {
		if f2, ok := dupes[f.Column]; ok {
			return fmt.Errorf(`%s has field %s with "genorm:" tag with duplicate column name %s (used by %s)`, res.Type, f.Name, f.Column, f2)
		}
		dupes[f.Column] = f.Name
	}

	return nil
}

func gofmt(path string) {
	cmd := exec.Command("gofmt", "-s", "-w", path)
	log.Printf(strings.Join(cmd.Args, " "))
	b, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("gofmt error: %s", err)
	}
	log.Printf("gofmt output: %s", b)
}
