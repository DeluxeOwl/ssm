package internal

var SSMStates = make([]Connectable, 0)

type StateNode struct {
	Name       string
	Group      string
	NextStates []Connectable

	visited bool
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
	return s.Name == name && (group == "" || s.Group == group)
}

func (s *StateNode) Visited() bool {
	return s.visited
}
