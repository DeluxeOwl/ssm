package dot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git.sr.ht/~mariusor/ssm/cmd/internal"
	"github.com/emicklei/dot"
)

func Dot(output string, states ...internal.Connectable) error {
	dd := dotBuilder{
		g:  dot.NewGraph(dot.Directed),
		v:  make(map[string]*dot.Node),
		sg: make(map[string]*dot.Graph),
		e:  make(map[string]*dot.Edge),
	}
	_ = dd.addStates(states...)
	g := dd.g
	if output == "" {
		fmt.Print(g.String())
		return nil
	}
	var data string
	if filepath.Ext(output) == mermaidExt {
		data = dot.MermaidGraph(g, dot.MermaidTopDown)
	} else {
		data = g.String()
	}
	return os.WriteFile(output, []byte(data), 0666)
}

const mermaidExt = ".mmd"

type dotBuilder struct {
	g     *dot.Graph
	sg    map[string]*dot.Graph
	v     map[string]*dot.Node
	e     map[string]*dot.Edge
	depth int
}

func (d *dotBuilder) node(n *internal.StateNode) (*dot.Node, bool) {
	g := d.g
	if len(n.Group) > 0 {
		g = d.subgraph(n.Group)
	}
	nodeKey := strings.Join([]string{n.Group, n.Name}, ".")
	if nn, ok := d.v[nodeKey]; ok {
		return nn, false
	}
	nn := g.Node(n.Name)
	d.v[nodeKey] = &nn
	return d.v[nodeKey], true
}

func (d *dotBuilder) addStates(states ...internal.Connectable) []*dot.Node {
	if d.depth > 10 {
		// NOTE(Marius): this should be handled some other way. We should be able to detect if a pair of nodes has
		// already been linked and we should skip
		return nil
	}
	d.depth++
	defer func() { d.depth-- }()

	result := make([]*dot.Node, 0)
	for _, st := range states {
		state, ok := st.(*internal.StateNode)
		if !ok {
			continue
		}
		n1, _ := d.node(state)
		result = append(result, n1)
		if len(state.NextStates) > 0 {
			nodes := d.addStates(state.NextStates...)
			for _, n2 := range nodes {
				d.edge(n1, n2)
			}
		}
	}
	return result
}

func (d *dotBuilder) edge(n1, n2 *dot.Node) (*dot.Edge, bool) {
	edgeKey := n1.ID() + "-" + n2.ID()
	if edge, ok := d.e[edgeKey]; ok {
		return edge, false
	}
	edge := d.g.Edge(*n1, *n2)
	d.e[edgeKey] = &edge
	return d.e[edgeKey], true
}

func (d *dotBuilder) subgraph(n string) *dot.Graph {
	if g, ok := d.sg[n]; ok {
		return g
	}
	d.sg[n] = d.g.Subgraph(n, dot.ClusterOption{})
	return d.sg[n]
}
