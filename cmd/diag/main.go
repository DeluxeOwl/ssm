package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"git.sr.ht/~mariusor/ssm/cmd/internal"

	"github.com/emicklei/dot"
)

var ()

const mermaidExt = ".mmd"

func main() {
	var output string
	flag.StringVar(&output, "o", "", "The file in which to save the dot file.\nThe type is inferred from the extension (.dot for Graphviz and .mmd for Mermaid)")
	flag.Parse()

	targets := internal.FindAllTargets(flag.Args())
	states, err := internal.LoadStates(targets)
	if err != nil {
		log.Panicf("Error: %s", err)
	}

	references := make(map[string]dot.Node)

	g := dot.NewGraph(dot.Directed)
	for _, st := range states {
		state, ok := st.(internal.StateNode)
		if !ok {
			continue
		}
		sg := g
		if len(state.Group) > 0 {
			sg = g.Subgraph(state.Group, dot.ClusterOption{})
		}
		references[state.Name] = sg.Node(state.Name)
	}

	for _, st := range states {
		state, ok := st.(internal.StateNode)
		if !ok {
			continue
		}
		if n1, ok := references[state.Name]; ok {
			for _, next := range state.NextStates {
				if strings.Index(next, ".") > 0 {
					if _, a, ok := strings.Cut(next, "."); ok {
						next = a
					}
				}
				if n2, ok := references[next]; ok {
					g.Edge(n1, n2)
				}
			}
		}
	}
	if output == "" {
		fmt.Print(g.String())
	}
	var data string
	if filepath.Ext(output) == mermaidExt {
		data = dot.MermaidGraph(g, dot.MermaidTopDown)
	} else {
		data = g.String()
	}
	os.WriteFile(output, []byte(data), 0666)
}
