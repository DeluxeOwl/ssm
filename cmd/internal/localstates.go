package internal

import "strings"

// SSMStates is a static slice containing the states in the SSM package.
// They get parsed dynamically at module.init() time
var SSMStates = make([]Connectable, 0)

type Flag int8

const (
	FlagExported = 1 << iota
	FlagTerminator

	FlagNone Flag = 0
)

type StateNode struct {
	Name       string
	Group      string
	Flags      Flag
	InStates   []Connectable
	NextStates []Connectable
}

func (s *StateNode) Equals(ss Connectable) bool {
	if sn, ok := ss.(*StateNode); ok {
		return s.Match(sn.Group, sn.Name)
	}
	return false
}

func (s *StateNode) Children() []Connectable {
	return s.NextStates
}

func (s *StateNode) Append(n ...Connectable) {
	s.NextStates = append(s.NextStates, n...)
}

func (s *StateNode) Match(group, name string) bool {
	if group == "" {
		if b, a, ok := strings.Cut(name, "."); ok {
			group = b
			name = a
		}
	}
	return s.Name == name && (group == "" || s.Group == group)
}
