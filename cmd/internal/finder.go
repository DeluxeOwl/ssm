package internal

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/emicklei/dot"
)

var (
	SSMStateType     string
	SSMModulePath    string
	SSMName          string
	SSMModuleVersion string
)

func LoadStates(targets []string) ([]stateNode, error) {
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

func shakeStates(states []stateNode) []stateNode {
	finalStates := make([]stateNode, 0, len(states))
	for _, s := range states {
	top:
		for _, ss := range states {
			for _, ns := range ss.NextStates {
				if ns == s.Name || s.Group != SSMName {
					finalStates = append(finalStates, s)
					break top
				}
			}
		}
	}
	return finalStates
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
	*dot.Node

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

	if !returnIsValid(result, group) {
		return nil, false
	}

	name := getFuncNameFromNode(fn)

	res := stateNode{
		Group: group,
		Name:  name,
	}
	res.NextStates = make([]string, 0)
	ast.Walk(walker(getReturns(&res.NextStates)), fn)

	return &res, true
}

func appendFuncNameFromIdent(states *[]string, n ast.Node) {
	id, ok := n.(*ast.Ident)
	if !ok {
		return
	}
	*states = append(*states, id.String())
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
				for _, arg := range cfn.Args {
					appendFuncNameFromIdent(nextStates, arg)
				}
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
	switch ee := ex.(type) {
	case *ast.SelectorExpr:
		name := ee.Sel.String()
		if ee.X != nil {
			if ident, ok := ee.X.(*ast.Ident); ok {
				if ident.Obj != nil {
					if ident.Obj.Decl != nil {
						if f, ok := ident.Obj.Decl.(*ast.Field); ok {
							name = getFuncNameFromExpr(f.Type) + "." + name
						}
					}
				} else {
					name = ident.Name + "." + name
				}
			}
		}
		return name
	case *ast.StarExpr:
		if typ, ok := ee.X.(*ast.Ident); ok {
			return "*" + typ.String()
		}
	case *ast.Ident:
		return ee.String()
	}
	return ""
}

func returnIsValid(r ast.Node, group string) bool {
	par, ok := r.(*ast.Field)
	if !ok {
		return false
	}
	if group == SSMName {
		return SSMStateType == SSMName+"."+getFuncNameFromExpr(par.Type)
	}
	return SSMStateType == getFuncNameFromExpr(par.Type)
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
