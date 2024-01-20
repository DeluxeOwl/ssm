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

func appendFuncNameFromIdent(states *[]string, n ast.Node) {
	id, ok := n.(*ast.Ident)
	if !ok {
		return
	}
	*states = append(*states, id.String())
}

func returnIsValid(r ast.Node, group string) bool {
	par, ok := r.(*ast.Field)
	if !ok {
		return false
	}
	n := getFuncNameFromExpr(par.Type)
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
