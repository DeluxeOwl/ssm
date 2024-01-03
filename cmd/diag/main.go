package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func loadPathsFromArgs() []string {
	dirs := make([]string, 0)
	for _, arg := range flag.Args() {
		if abs, err := filepath.Abs(filepath.Clean(arg)); err == nil {
			arg = abs
		}
		if _, err := os.Stat(arg); err != nil {
			continue
		}
		dirs = append(dirs, arg)
	}
	return dirs
}

func filesFromArgs() []string {
	dirs := loadPathsFromArgs()
	files := make([]string, 0)
	isGoFile := func(path string) bool {
		fi, _ := os.Stat(path)
		return !fi.IsDir() && filepath.Ext(path) == ".go"
	}
	for _, dir := range dirs {
		if isGoFile(dir) {
			files = append(files, dir)
			continue
		}
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if !isGoFile(path) {
				return nil
			}
			files = append(files, path)
			return nil
		})
		if err != nil {
			log.Printf("error: %s", err)
		}
	}
	return files
}

func main() {
	flag.Parse()

	files := filesFromArgs()

	fset := token.NewFileSet()
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("error: %s", err)
			continue
		}

		f, err := parser.ParseFile(fset, file, data, parser.ParseComments)
		if err != nil {
			log.Printf("error: %s", err)
			continue
		}
		ast.Walk(walker(func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			// NOTE(marius) we're looking for functions that have one parameter which is a context.Context
			// and a return which is a ssm.Fn
			if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
				return true
			}
			if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
				return true
			}
			fmt.Printf("%s (", fn.Name)
			for _, par := range fn.Type.Params.List {
				fmt.Printf("%s ", par.Type)
			}
			fmt.Print(") ")
			for _, res := range fn.Type.Results.List {
				fmt.Printf("%s ", res.Type)
			}
			fmt.Print("\n")

			return true
		}), f)
	}
}

// walker adapts a function to satisfy the ast.Visitor interface.
// The function return whether the walk should proceed into the node's children.
type walker func(ast.Node) bool

func (w walker) Visit(node ast.Node) ast.Visitor {
	if w(node) {
		return w
	}
	return nil
}
