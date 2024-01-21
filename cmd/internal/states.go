package internal

import (
	"go/ast"
	"strings"
	"unicode"

	"github.com/emicklei/dot"
)

// Connectable is a dot.Node or a *internal.StateNode
type Connectable interface {
	Attr(label string, value interface{}) dot.Node
}

func functionIsNotExported(name string) bool {
	if name == "" {
		return false
	}
	return unicode.IsLower(rune(name[0]))
}

func New(fn ast.Node, group string) *StateNode {
	name := getFuncNameFromNode(fn)

	if strings.EqualFold(group, ssmName) && functionIsNotExported(name) {
		return nil
	}

	res := StateNode{
		Group: group,
		Name:  name,
	}
	res.NextStates = make([]string, 0)

	ast.Walk(walker(getReturns(&res.NextStates)), fn)
	return &res
}

func getReturns(nextStates *[]string) func(n ast.Node) bool {
	return func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, rr := range ret.Results {
			switch r := rr.(type) {
			case *ast.CallExpr:
				// TODO(marius): we need to do this recursively to load all states of form:
				//   return State1(State2(State3(...)))
				if st := getFuncNameFromExpr(r.Fun); st != "" {
					*nextStates = append(*nextStates, st)
				}
				for _, arg := range r.Args {
					appendFuncNameFromArg(nextStates, arg)
				}
			case *ast.Ident:
				appendFuncNameFromArg(nextStates, rr)
			case *ast.FuncLit:
				for _, rs := range r.Body.List {
					ast.Walk(walker(getReturns(nextStates)), rs)
				}
			case *ast.FuncType:
				if st := getFuncNameFromNode(r); st != "" {
					*nextStates = append(*nextStates, st)
				}
			case *ast.SelectorExpr:
				if st := getFuncNameFromExpr(r); st != "" {
					*nextStates = append(*nextStates, st)
				}
			}
		}
		return true
	}
}

func getFuncNameFromExpr(ex ast.Expr) string {
	switch ee := ex.(type) {
	case *ast.SelectorExpr:
		name := ee.Sel.String()
		if ee.X != nil {
			switch tt := ee.X.(type) {
			case *ast.CallExpr:
				name = getFuncNameFromExpr(tt.Fun) + "." + name
			case *ast.Ident:
				ident := tt
				if ident.Obj != nil {
					if ident.Obj.Decl != nil {
						if f, ok := ident.Obj.Decl.(*ast.Field); ok {
							name = getFuncNameFromExpr(f.Type) + "." + name
						}
					}
				} else if ident.Name != ssmName {
					name = ident.Name + "." + name
				}
			}
		}
		return name
	case *ast.StarExpr:
		if typ, ok := ee.X.(*ast.Ident); ok {
			return typ.String()
		}
	case *ast.Ident:
		return ee.String()
	}
	return ""
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

func appendFuncNameFromArg(states *[]string, n ast.Node) {
	switch nn := n.(type) {
	case *ast.Ident:
		if st := nn.String(); st != "" {
			*states = append(*states, st)
		}
	case *ast.CallExpr:
		if st := getFuncNameFromExpr(nn.Fun); st != "" {
			*states = append(*states)
		}
	}
}

func (s stateSearch) returnIsValid(r ast.Node, imp map[string]string) bool {
	par, ok := r.(*ast.Field)
	if !ok {
		return false
	}
	alias := ssmName
	for n, p := range imp {
		if strings.Trim(p, "\"") == ssmModulePath {
			alias = n
			break
		}
	}
	n := getFuncNameFromExpr(par.Type)
	if strings.Contains(n, alias) {
		n = strings.Replace(n, alias, ssmName, 1)
	}
	return ssmStateType == n || ssmStateType == ssmName+"."+n
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
