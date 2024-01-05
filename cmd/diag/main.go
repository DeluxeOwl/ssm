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
	"strings"
)

const (
	ssmModulePath = "git.sr.ht/~mariusor/ssm"
	ssmStateType  = "Fn"
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
		importsSsm := false
		for _, imp := range f.Imports {
			if strings.Trim(imp.Path.Value, `"`) == ssmModulePath {
				importsSsm = true
			}
		}
		if !importsSsm {
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

	var name string
	if fn.Recv != nil {
		recv := fn.Recv.List[0]
		if t, ok := recv.Type.(*ast.StarExpr); ok {
			if id, ok := t.X.(*ast.Ident); ok {
				name = fmt.Sprintf("*%s", id.String())
			}
		}
		if id, ok := recv.Type.(*ast.Ident); ok {
			name = fmt.Sprintf("%s", id.String())
		}
		name = name + "." + fn.Name.String()
	} else {
		name = fn.Name.String()
	}

	res := StateNode{
		Name: name,
	}
	res.NextStates = make([]string, 0)
	ast.Walk(walker(getReturns(&res.NextStates)), n)

	fmt.Printf("  %v\n", res)
	return &res, true
}

func getReturns(nextStates *[]string) func(n ast.Node) bool {
	return func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, rr := range ret.Results {
			if cfn, ok := rr.(*ast.CallExpr); ok {
				*nextStates = append(*nextStates, getFuncName(cfn.Fun))
			} else if fn, ok := rr.(*ast.Ident); ok {
				*nextStates = append(*nextStates, fn.String())
				//} else if fn, ok := rr.(*ast.FuncLit); ok {
				//	for _, rs := range fn.Body.List {
				//		ast.Walk(walker(getReturns(nextStates)), rs)
				//	}
			} else if fn, ok := rr.(*ast.FuncType); ok {
				fmt.Printf("  func type: %v", fn)
			} else if sel, ok := rr.(*ast.SelectorExpr); ok {
				*nextStates = append(*nextStates, sel.Sel.String())
			}
		}
		return true
	}
}

func getFuncName(ex ast.Expr) string {
	var ident *ast.Ident
	if typ, ok := ex.(*ast.SelectorExpr); ok {
		ident = typ.Sel
	}
	if typ, ok := ex.(*ast.Ident); ok {
		ident = typ
	}
	if ident == nil {
		return ""
	}
	return ident.String()
}

func returnIsValid(r ast.Node) bool {
	par, ok := r.(*ast.Field)
	if !ok {
		return false
	}
	return getFuncName(par.Type) == ssmStateType
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
