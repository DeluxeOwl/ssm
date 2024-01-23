package internal

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strings"

	"git.sr.ht/~mariusor/ssm"
)

var (
	_fnReflectType = reflect.ValueOf(ssm.Fn(nil)).Type()

	ssmStateType  = _fnReflectType.String()
	ssmModulePath = _fnReflectType.PkgPath()
	ssmName       = filepath.Base(ssmModulePath)

	build, _ = debug.ReadBuildInfo()
	//ssmModulePath    = build.Main.Path
	ssmModuleVersion = build.Main.Version
)

type stateSearch struct {
	p *ast.Package

	states []Connectable
}

func LoadStates(targets ...string) ([]Connectable, error) {
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

	s := stateSearch{states: make([]Connectable, 0)}
	for _, pack := range packages {
		if packageIsValid(pack) {
			s.p = pack
			s.states = append(s.states, s.loadStatesFromPackage()...)
		}
	}
	for _, pack := range packages {
		if packageIsValid(pack) {
			s.p = pack
			s.loadNextStatesFromPackage(pack.Name)
		}
	}
	//states = shakeStates(states)
	return s.states, errors.Join(errs...)
}

func packageIsUs(p *ast.Package) bool {
	return strings.EqualFold(p.Name, ssmName)
}

func packageIsValid(p *ast.Package) bool {
	if packageIsUs(p) {
		return true
	}
	for _, f := range p.Files {
		for _, imp := range f.Imports {
			if importIsValid(imp) {
				return true
			}
		}
	}
	return false
}

func importIsValid(imp *ast.ImportSpec) bool {
	return strings.Trim(imp.Path.Value, `"`) == ssmModulePath
}

func shakeStates(states []Connectable) []Connectable {
	finalStates := make([]Connectable, 0, len(states))
	for _, st := range states {
		s, ok := st.(*StateNode)
		if !ok {
			continue
		}
	top:
		for _, ss := range states {
			for _, ns := range ss.Children() {
				if s.Equals(ns) {
					finalStates = append(finalStates, s)
					break top
				}
			}
		}
	}
	return finalStates
}

func (s stateSearch) loadStatesFromPackage() []Connectable {
	states := make([]Connectable, 0)
	imports := make(map[string]string)
	ast.Walk(walker(func(n ast.Node) bool {
		switch nn := n.(type) {
		case *ast.File:
			for _, i := range nn.Imports {
				if i == nil || i.Name == nil {
					continue
				}
				imports[i.Name.Name] = i.Path.Value
			}
		case *ast.FuncDecl:
			if state := s.loadStateFromFuncDecl(nn, imports); state != nil {
				states = append(states, state)
			}
		}
		return true
	}), s.p)
	return states
}

func (s stateSearch) loadNextStatesFromPackage(group string) {
	ast.Walk(walker(func(n ast.Node) bool {
		switch fn := n.(type) {
		case *ast.FuncDecl:
			// Find the state
			res, ok := findState(s.states, group, getFuncNameFromNode(fn))
			if ok {
				// extract next states from its return values
				ast.Walk(walker(s.getReturns(res)), fn)
			}
		}
		return true
	}), s.p)
}

func (s stateSearch) loadStateFromFuncDecl(n ast.Node, imp map[string]string) Connectable {
	fn, ok := n.(*ast.FuncDecl)
	if !ok {
		return nil
	}
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return nil
	}

	result := fn.Type.Results.List[0]

	group := s.p.Name
	if !s.returnIsValid(result, imp) {
		return nil
	}

	return fromNode(fn, group)
}
