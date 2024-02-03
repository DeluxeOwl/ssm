package internal

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strings"

	"git.sr.ht/~mariusor/ssm"
)

var (
	_fnReflectType = reflect.ValueOf(ssm.Fn(nil)).Type()
	build, _       = debug.ReadBuildInfo()

	ssmStateType     = _fnReflectType.String()
	ssmModulePath    = _fnReflectType.PkgPath()
	ssmName          = filepath.Base(ssmModulePath)
	ssmModuleVersion = build.Main.Version

	ssmRun         = ssmName + ".Run"
	ssmRunParallel = ssmName + ".RunParallel"
)

type stateSearch struct {
	p *ast.Package

	loadInternal bool

	imports map[string]string
}

func parseTargetPackages(targets ...string) (map[string]*ast.Package, error) {
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
				if !packageIsValid(p) {
					logFn("no states in package %q, skipping", p.Name)
					continue
				}
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
			p := &ast.Package{Name: parent, Files: map[string]*ast.File{target: f}}
			if !packageIsValid(p) {
				logFn("no states in package %q, skipping", p.Name)
				continue
			}
			packages[parent] = p
		}
	}
	return packages, errors.Join(errs...)
}

// loadStateNames returns a flat list of states found in the received packages
func (s stateSearch) loadStateNames(packages map[string]*ast.Package) []Connectable {
	states := make([]Connectable, 0)
	for _, pack := range packages {
		s.p = pack
		s.loadStatesFromDeclarations(&states)
	}
	return states
}

var logFn = log.New(os.Stderr, "dot: ", log.LstdFlags).Printf

func (s stateSearch) loadNextStates(states *[]Connectable, packages map[string]*ast.Package) []Connectable {
	for _, pack := range packages {
		s.p = pack
		s.loadNextStatesFromPackage(states, pack.Name)
	}
	return *states
}

func (s stateSearch) findStartStates(states *[]Connectable, packages map[string]*ast.Package) []Connectable {
	for _, pack := range packages {
		s.p = pack
		s.loadStartFromPackage(states)
	}
	return *states
}

func loadInternalStates(ssmPath string) ([]Connectable, error) {
	packages, err := parseTargetPackages(ssmPath)
	if err != nil {
		return nil, err
	}

	s := stateSearch{imports: make(map[string]string), loadInternal: true}

	return s.loadStateNames(packages), nil
}

func LoadStates(targets ...string) ([]Connectable, error) {
	packages, err := parseTargetPackages(targets...)
	if err != nil {
		return nil, err
	}

	s := stateSearch{imports: make(map[string]string), loadInternal: true}

	// NOTE(marius): we now have all ssm.Fn declared in the target packages.
	states := s.loadStateNames(packages)

	// NOTE(marius): we can iterate again and find each state's following states.
	// This requires to look at the definitions of the states.
	states = s.loadNextStates(&states, packages)

	// NOTE(marius): we can now find where the targets are calling ssm.Run/ssm.RunParallel
	// and build the state graph out of that.
	states = s.findStartStates(&states, packages)

	return states, nil
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
			if importIsUs(imp) {
				return true
			}
		}
	}
	return false
}

func importIsUs(imp *ast.ImportSpec) bool {
	return strings.Trim(imp.Path.Value, `"`) == ssmModulePath
}

func (s stateSearch) loadStartFromPackage(states *[]Connectable) []Connectable {
	ast.Walk(walker(func(n ast.Node) bool {
		if start := s.loadStartFromNode(states, n); start != nil {
			appendStates(states, start)
		}
		return true
	}), s.p)
	return *states
}

func (s stateSearch) loadStatesFromDeclarations(states *[]Connectable) []Connectable {
	ast.Walk(walker(func(n ast.Node) bool {
		switch nn := n.(type) {
		case *ast.Ident:
			if nn.Obj == nil || nn.Obj.Kind != ast.Var {
				return true
			}
			if !s.declIsValid(nn.Obj.Decl, s.imports) {
				return true
			}
			if state := s.loadStateFromIdent(nn); state != nil {
				appendStates(states, state)
			}
		case *ast.File:
			for _, i := range nn.Imports {
				if i == nil || i.Name == nil {
					continue
				}
				s.imports[i.Name.Name] = i.Path.Value
			}
		case *ast.FuncDecl:
			if state := s.loadStateFromFuncDecl(nn); state != nil {
				appendStates(states, state)
			}
		}
		return true
	}), s.p)
	return *states
}

func appendStates(states *[]Connectable, toAppend ...Connectable) {
	for _, st := range toAppend {
		add := true
		for _, check := range *states {
			if check.Equals(st) {
				add = false
			}
		}
		if add {
			*states = append(*states, st)
		}
	}
}

func (s stateSearch) loadNextStatesFromPackage(states *[]Connectable, group string) {
	ast.Walk(walker(func(n ast.Node) bool {
		switch fn := n.(type) {
		case *ast.FuncDecl:
			// Find the state
			name := getStateNameFromNode(fn)
			res, ok := findState(*states, group, name)
			if ok {
				// extract next states from its return values
				ast.Walk(walker(s.getReturns(states, res)), fn)
			}
		}
		return true
	}), s.p)
}

func (s stateSearch) loadStartFromNode(states *[]Connectable, n ast.Node) Connectable {
	fn, ok := n.(*ast.ExprStmt)
	if !ok {
		return nil
	}
	name := getFuncNameFromExpr(fn.X)
	if name != ssmRun && name != ssmRunParallel {
		return nil
	}

	start := StateNode{Name: name, Group: ssmName}
	// extract next states from its return values
	ast.Walk(walker(s.getParams(states, &start)), fn)
	return &start
}

func (s stateSearch) loadStateFromIdent(n ast.Node) Connectable {
	id, ok := n.(*ast.Ident)
	if !ok {
		return nil
	}
	decl, ok := id.Obj.Decl.(*ast.ValueSpec)
	if !ok {
		return nil
	}
	if !typeIsValid(decl.Type, s.imports) {
		return nil
	}
	group := s.p.Name
	return s.fromNode(id, group)
}

func (s stateSearch) fromFuncBody(fn *ast.BlockStmt) Connectable {
	//for _, st := range fn.List {
	//	//fmt.Printf("%v", st)
	//}
	return nil
}

func (s stateSearch) fromStateFuncDecl(fn *ast.FuncDecl) Connectable {
	result := fn.Type.Results.List[0]

	group := s.p.Name
	if !s.returnIsValid(result, s.imports) {
		return nil
	}
	return s.fromNode(fn, group)
}

func (s stateSearch) loadStateFromFuncDecl(n ast.Node) Connectable {
	fn, ok := n.(*ast.FuncDecl)
	if !ok {
		return nil
	}
	maybeIsState := fn.Type.Results != nil && len(fn.Type.Results.List) == 1
	if maybeIsState {
		return s.fromStateFuncDecl(fn)
	}
	return nil //s.fromFuncBody(fn.Body)
}
