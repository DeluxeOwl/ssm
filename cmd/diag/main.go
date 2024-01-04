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

			// NOTE(marius) we're looking for functions that return a ssm.Fn
			if _, ok := isStateFn(fn); ok {
			}
			return true
		}), f)
	}
}

type StateNode struct {
	Name       string
	Alias      string
	NextStates []string
}

func isStateFn(n ast.Node) (*StateNode, bool) {
	fn, ok := n.(*ast.FuncDecl)
	if !ok {
		return nil, false
	}
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return nil, false
	}

	result := fn.Type.Results.List[0]

	if !returnIsValid(result) {
		return nil, false
	}

	if fn.Recv != nil {
		fmt.Printf("%v - %v\n", fn.Recv, fn.Name.String())
	} else {
		fmt.Printf("%v\n", fn.Name.String())
	}
	res := StateNode{
		Name: fn.Name.String(),
	}
	ast.Walk(walker(func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, rr := range ret.Results {
			if _, ok := rr.(*ast.CallExpr); ok {
				//fmt.Printf("  callexpr: %v", rr)
				continue
			}
			if fn, ok := rr.(*ast.Ident); ok {
				res.NextStates = append(res.NextStates, fn.String())
			} else if fn, ok := rr.(*ast.FuncLit); ok {
				fmt.Printf("  literal: %v", fn)
			} else if fn, ok := rr.(*ast.FuncType); ok {
				fmt.Printf("  func type: %v", fn)
			} else {
				fmt.Printf("  %T: %v", rr, rr)
			}

		}
		return true
	}), n)
	return &res, true
}

func returnIsValid(r ast.Node) bool {
	par, ok := r.(*ast.Field)
	if !ok {
		return false
	}
	if typ, ok := par.Type.(*ast.Ident); ok {
		if typ.Name == "Fn" {
			return true
		}
	}
	return false
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
