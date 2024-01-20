package internal

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

var (
	SSMStateType     string
	SSMModulePath    string
	SSMName          string
	SSMModuleVersion string
)

func LoadStates(targets []string) ([]Connectable, error) {
	states := make([]Connectable, 0)
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
	states = shakeStates(states)
	return states, errors.Join(errs...)
}

func packageIsUs(p *ast.Package) bool {
	return strings.EqualFold(p.Name, SSMName)
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

func validImport(imp *ast.ImportSpec) bool {
	return strings.Trim(imp.Path.Value, `"`) == SSMModulePath
}

func shakeStates(states []Connectable) []Connectable {
	finalStates := make([]Connectable, 0, len(states))
	for _, st := range states {
		s, ok := st.(StateNode)
		if !ok {
			continue
		}
	top:
		for _, sss := range states {
			if ss, ok := sss.(StateNode); ok {
				for _, ns := range ss.NextStates {
					if ns == s.Name || s.Group != SSMName {
						finalStates = append(finalStates, s)
						break top
					}
				}
			}
		}
	}
	return finalStates
}

func loadStatesFromPackage(p *ast.Package, group string) []Connectable {
	states := make([]Connectable, 0)
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

func loadStateFromFuncDecl(n ast.Node, group string) (*StateNode, bool) {
	fn, ok := n.(*ast.FuncDecl)
	if !ok {
		return nil, false
	}
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return nil, false
	}

	result := fn.Type.Results.List[0]

	if !returnIsValid(result, group) {
		return nil, false
	}

	res := New(fn, group)
	return res, true
}
