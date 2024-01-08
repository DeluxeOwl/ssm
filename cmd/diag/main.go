package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/emicklei/dot"
)

const ssmStateType = "Fn"

var (
	build, _            = debug.ReadBuildInfo()
	ssmModulePath       = build.Main.Path
	ssmModuleImportPath = build.Main.Path + "@" + build.Main.Version
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

func filesFromArgs() ([]string, error) {
	dirs := loadPathsFromArgs()
	files := make([]string, 0)
	isGoFile := func(path string) bool {
		fi, _ := os.Stat(path)
		return !fi.IsDir() && filepath.Ext(path) == ".go"
	}
	errs := make([]error, 0)
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
			errs = append(errs, err)
		}
	}
	return files, errors.Join(errs...)
}

func main() {
	flag.Parse()

	files, err := filesFromArgs()
	if err != nil {
		log.Panicf("Error: %s", err)
	}

	states, err := loadStatesFromFiles(files)
	if err != nil {
		log.Panicf("Error: %s", err)
	}
	g := dot.NewGraph(dot.Directed)
	references := make(map[string]dot.Node)
	for _, state := range states {
		references[state.Name] = g.Node(state.Name)
	}
	for _, state := range states {
		n1, ok := references[state.Name]
		if !ok {
			continue
		}
		for _, next := range state.NextStates {
			n2, ok := references[next]
			if !ok {
				continue
			}
			g.Edge(n1, n2)
		}
	}
	fmt.Println(g.String())
}

func loadStatesFromFiles(files []string) ([]stateNode, error) {
	states := make([]stateNode, 0)
	fset := token.NewFileSet()
	errs := make([]error, 0)
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		f, err := parser.ParseFile(fset, file, data, parser.ParseComments)
		if err != nil {
			errs = append(errs, err)
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
		states = append(states, loadStatesFromFile(f)...)
	}
	return states, errors.Join(errs...)
}

func loadStatesFromFile(f *ast.File) []stateNode {
	states := make([]stateNode, 0)
	ast.Walk(walker(func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if state, ok := loadStateFromFuncDecl(fn); ok {
			states = append(states, *state)
		}
		return true
	}), f)
	return states
}

type stateNode struct {
	Name       string
	Alias      string
	NextStates []string
}

func loadStateFromFuncDecl(n ast.Node) (*stateNode, bool) {
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

	name := getFuncNameFromNode(fn)

	res := stateNode{
		Name: name,
	}
	res.NextStates = make([]string, 0)
	ast.Walk(walker(getReturns(&res.NextStates)), n)

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
				*nextStates = append(*nextStates, getFuncNameFromExpr(cfn.Fun))
			} else if fn, ok := rr.(*ast.Ident); ok {
				*nextStates = append(*nextStates, fn.String())
			} else if fn, ok := rr.(*ast.FuncLit); ok {
				for _, rs := range fn.Body.List {
					ast.Walk(walker(getReturns(nextStates)), rs)
				}
			} else if fn, ok := rr.(*ast.FuncType); ok {
				*nextStates = append(*nextStates, getFuncNameFromNode(fn))
			} else if sel, ok := rr.(*ast.SelectorExpr); ok {
				*nextStates = append(*nextStates, getFuncNameFromExpr(sel))
			}
		}
		return true
	}
}

func getFuncNameFromNode(n ast.Node) string {
	var name string
	if fn, ok := n.(*ast.FuncDecl); ok {
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
	}
	return name
}

func getFuncNameFromExpr(ex ast.Expr) string {
	if fun, ok := ex.(*ast.SelectorExpr); ok {
		name := fun.Sel.String()
		if fun.X != nil {
			if ident, ok := fun.X.(*ast.Ident); ok {
				if ident.Obj != nil && ident.Obj.Decl != nil {
					if f, ok := ident.Obj.Decl.(*ast.Field); ok {
						name = getFuncNameFromExpr(f.Type) + "." + name
					}
				} else {
					name = ident.Name + "." + name
				}
			}
		}
		return name
	}
	if st, ok := ex.(*ast.StarExpr); ok {
		if typ, ok := st.X.(*ast.Ident); ok {
			return "*" + typ.String()
		}
	}
	if typ, ok := ex.(*ast.SelectorExpr); ok {
		return typ.Sel.String()
	}
	if typ, ok := ex.(*ast.Ident); ok {
		return typ.String()
	}
	return ""
}

func returnIsValid(r ast.Node) bool {
	par, ok := r.(*ast.Field)
	if !ok {
		return false
	}
	return getFuncNameFromExpr(par.Type) == ssmStateType
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
