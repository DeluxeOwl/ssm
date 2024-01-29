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

	ssmRun         = ssmName + ".Run"
	ssmRunParallel = ssmName + ".RunParallel"
)

type stateSearch struct {
	p *ast.Package

	loadInternal bool

	imports map[string]string
	states  []Connectable
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

	s := stateSearch{states: make([]Connectable, 0), imports: make(map[string]string), loadInternal: true}
	for _, pack := range packages {
		if packageIsValid(pack) {
			s.p = pack
			appendStates(&s.states, s.loadStatesFromPackage()...)
		}
	}
	for _, pack := range packages {
		if packageIsValid(pack) {
			s.p = pack
			s.loadNextStatesFromPackage(pack.Name)
		}
	}
	for _, pack := range packages {
		if packageIsValid(pack) {
			s.p = pack
			appendStates(&s.states, s.loadStatesFromPackage()...)
		}
	}
	for _, pack := range packages {
		if packageIsValid(pack) {
			s.p = pack
			appendStates(&s.states, s.loadStartFromPackage()...)
		}
	}
	for _, pack := range packages {
		if packageIsValid(pack) {
			s.p = pack
			s.loadNextStatesFromPackage(pack.Name)
		}
	}
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

func (s stateSearch) loadStartFromPackage() []Connectable {
	states := make([]Connectable, 0)
	ast.Walk(walker(func(n ast.Node) bool {
		if start := s.loadStartFromNode(n); start != nil {
			appendStates(&states, start)
		}
		return true
	}), s.p)
	return states
}

func (s stateSearch) loadStatesFromPackage() []Connectable {
	states := make([]Connectable, 0)
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
				appendStates(&states, state)
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
				appendStates(&states, state)
			}
		}
		return true
	}), s.p)
	return states
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

func (s stateSearch) loadNextStatesFromPackage(group string) {
	ast.Walk(walker(func(n ast.Node) bool {
		switch fn := n.(type) {
		case *ast.FuncDecl:
			// Find the state
			name := getStateNameFromNode(fn)
			res, ok := findState(s.states, group, name)
			if ok {
				// extract next states from its return values
				ast.Walk(walker(s.getReturns(res)), fn)
			}
		}
		return true
	}), s.p)
}

func (s stateSearch) loadStartFromNode(n ast.Node) Connectable {
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
	ast.Walk(walker(s.getParams(&start)), fn)
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
