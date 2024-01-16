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
	"golang.org/x/mod/module"
)

const ssmStateType = "Fn"

var (
	build, _         = debug.ReadBuildInfo()
	ssmModulePath    = build.Main.Path
	ssmModuleVersion = build.Main.Version
)

func getModulePath(name, version string) (string, error) {
	// first we need GOMODCACHE
	cache, ok := os.LookupEnv("GOMODCACHE")
	if !ok {
		cache = filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
	}

	// then we need to escape path
	escapedPath, err := module.EscapePath(name)
	if err != nil {
		return "", err
	}

	if version == "(devel)" {
		versions := make([]module.Version, 0)
		potentials, _ := filepath.Glob(filepath.Join(cache, ssmModulePath+"*"))
		for _, path := range potentials {
			d, err := os.Stat(path)
			if err != nil {
				continue
			}
			if !d.IsDir() || !strings.Contains(path, escapedPath) {
				continue
			}
			if pieces := strings.Split(path, "@"); len(pieces) == 2 {
				if strings.Count(pieces[1], "/") > 0 {
					continue
				}
				versions = append(versions, module.Version{Path: name, Version: pieces[1]})
			}
		}
		if len(versions) > 0 {
			module.Sort(versions)
			version = versions[len(versions)-1].Version
		}
	}
	// version also
	escapedVersion, err := module.EscapeVersion(version)
	if err != nil {
		return "", err
	}

	return filepath.Join(cache, escapedPath+"@"+escapedVersion), nil
}

func findAllTargets(args []string) []string {
	targets := make([]string, 0)
	for _, arg := range args {
		if abs, err := filepath.Abs(filepath.Clean(arg)); err == nil {
			arg = abs
		}
		fi, err := os.Stat(arg)
		if err != nil {
			continue
		}
		if fi.IsDir() {
			_ = filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
				if !d.IsDir() {
					return nil
				}
				if strings.Contains(path, ".git") {
					return filepath.SkipDir
				}
				targets = append(targets, path)
				return nil
			})
		} else {
			targets = append(targets, arg)
		}
	}
	if ssmPath, err := getModulePath(ssmModulePath, ssmModuleVersion); err == nil {
		targets = append(targets, ssmPath)
	}
	return targets
}

const mermaidExt = ".mmd"

func shakeStates(states []stateNode) []stateNode {
	finalStates := make([]stateNode, 0, len(states))
	for _, s := range states {
	top:
		for _, ss := range states {
			for _, ns := range ss.NextStates {
				if ns == s.Name || s.Group != filepath.Base(ssmModulePath) {
					finalStates = append(finalStates, s)
					break top
				}
			}
		}
	}
	return finalStates
}

func main() {
	var output string
	flag.StringVar(&output, "o", "", "The file in which to save the dot file.\nThe type is inferred from the extension (.dot for Graphviz and .mmd for Mermaid)")
	flag.Parse()

	targets := findAllTargets(flag.Args())
	states, err := loadStates(targets)
	if err != nil {
		log.Panicf("Error: %s", err)
	}
	states = shakeStates(states)

	references := make(map[string]dot.Node)

	g := dot.NewGraph(dot.Directed)
	for _, state := range states {
		sg := g
		if len(state.Group) > 0 {
			sg = g.Subgraph(state.Group, dot.ClusterOption{})
		}
		references[state.Name] = sg.Node(state.Name)
	}

	for _, state := range states {
		if n1, ok := references[state.Name]; ok {
			for _, next := range state.NextStates {
				if n2, ok := references[next]; ok {
					g.Edge(n1, n2)
				}
			}
		}
	}
	if output == "" {
		fmt.Print(g.String())
	}
	var data string
	if filepath.Ext(output) == mermaidExt {
		data = dot.MermaidGraph(g, dot.MermaidTopDown)
	} else {
		data = g.String()
	}
	os.WriteFile(output, []byte(data), 0666)
}

func validImport(imp *ast.ImportSpec) bool {
	return strings.Trim(imp.Path.Value, `"`) == ssmModulePath
}

func packageIsUs(p *ast.Package) bool {
	return strings.EqualFold(p.Name, filepath.Base(ssmModulePath))
}

func packageIsValid(p *ast.Package) bool {
	if packageIsUs(p) {
		return true
	}
	for _, f := range p.Files {
		for _, imp := range f.Imports {
			if validImport(imp) {
				return true
			}
		}
	}
	return false
}

func validGoFile(info fs.FileInfo) bool {
	n := info.Name()
	return !strings.Contains(n, "_test") && filepath.Ext(n) == ".go"
}

func loadStates(targets []string) ([]stateNode, error) {
	states := make([]stateNode, 0)
	errs := make([]error, 0)
	packages := make(map[string]*ast.Package)

	for _, target := range targets {
		fi, err := os.Stat(target)
		if err != nil {
			continue
		}

		if fi.IsDir() {
			fset := token.NewFileSet()
			pp, err := parser.ParseDir(fset, target, validGoFile, parser.ParseComments)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			for pn, p := range pp {
				packages[pn] = p
			}
		} else {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, target, nil, parser.ParseComments)
			parent := filepath.Base(filepath.Dir(target))
			if err != nil {
				errs = append(errs, err)
				continue
			}
			packages[parent] = &ast.Package{Name: parent, Files: map[string]*ast.File{target: f}}
		}
	}

	for _, pack := range packages {
		if packageIsValid(pack) {
			states = append(states, loadStatesFromPackage(pack, pack.Name)...)
		}
	}
	return states, errors.Join(errs...)
}

func loadStatesFromPackage(p *ast.Package, group string) []stateNode {
	states := make([]stateNode, 0)
	ast.Walk(walker(func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if state, ok := loadStateFromFuncDecl(fn, group); ok {
				states = append(states, *state)
			}
		}
		return true
	}), p)
	return states
}

type stateNode struct {
	Name       string
	Group      string
	Alias      string
	NextStates []string
}

func loadStateFromFuncDecl(n ast.Node, group string) (*stateNode, bool) {
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
		Group: group,
		Name:  name,
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
					name = id.String()
				}
			}
			if id, ok := recv.Type.(*ast.Ident); ok {
				name = id.String()
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
					//} else {
					//	name = ident.Name + "." + name
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
