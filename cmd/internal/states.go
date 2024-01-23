package internal

import (
	"go/ast"
	"strings"
	"unicode"
)

// Connectable is a dot.Node or a *internal.StateNode
type Connectable interface {
	Children() []Connectable
	Match(group, name string) bool
	Append(n ...Connectable)
}

func functionIsNotExported(name string) bool {
	if name == "" {
		return false
	}
	return unicode.IsLower(rune(name[0]))
}

func fromNode(fn ast.Node, group string) Connectable {
	name := getFuncNameFromNode(fn)

	if strings.EqualFold(group, ssmName) && functionIsNotExported(name) {
		return nil
	}

	res := &StateNode{
		Name:       name,
		Group:      group,
		NextStates: make([]Connectable, 0),
	}

	return res
}

func findState(states []Connectable, group, n string) (Connectable, bool) {
	for _, ss := range states {
		if ss.Match(group, n) {
			return ss, true
		}
	}
	for _, s := range SSMStates {
		if s.Match(group, n) {
			return s, true
		}
	}
	return nil, false
}

func (s stateSearch) getReturns(res Connectable) func(n ast.Node) bool {
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
				nm := getFuncNameFromExpr(r.Fun)
				st, ok := findState(s.states, "", nm)
				if ok {
					res.Append(st)
				}
				for _, arg := range r.Args {
					s.appendFuncNameFromArg(st, arg)
				}
			case *ast.Ident:
				s.appendFuncNameFromArg(res, rr)
			case *ast.FuncLit:
				for _, rs := range r.Body.List {
					ast.Walk(walker(s.getReturns(res)), rs)
				}
			case *ast.FuncType:
				nm := getFuncNameFromNode(r)
				if st, ok := findState(s.states, "", nm); ok {
					res.Append(st)
				}
			case *ast.SelectorExpr:
				nm := getFuncNameFromExpr(r)
				if st, ok := findState(s.states, "", nm); ok {
					res.Append(st)
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

func (s stateSearch) appendFuncNameFromArg(res Connectable, n ast.Node) {
	name := ""
	switch nn := n.(type) {
	case *ast.Ident:
		name = nn.String()
	case *ast.CallExpr:
		name = getFuncNameFromExpr(nn.Fun)
	}
	st, ok := findState(s.states, "", name)
	if ok {
		res.Append(st)
	}
	if nn, ok := n.(*ast.CallExpr); ok {
		ast.Walk(walker(s.getReturns(st)), nn)
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
