package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"

	"github.com/emicklei/dot"

	"git.sr.ht/~mariusor/ssm"
	"git.sr.ht/~mariusor/ssm/cmd/internal"
)

var (
	_fnReflectType = reflect.ValueOf(ssm.Fn(nil)).Type()

	ssmStateType  = _fnReflectType.String()
	ssmModulePath = _fnReflectType.PkgPath()
	ssmName       = filepath.Base(ssmModulePath)

	build, _ = debug.ReadBuildInfo()
	//ssmModulePath    = build.Main.Path
	ssmModuleVersion = build.Main.Version
)

// Connectable is a dot.Node or a *dotx.Composite
type Connectable interface {
	Attr(label string, value interface{}) dot.Node
}

const mermaidExt = ".mmd"

func main() {
	var output string
	flag.StringVar(&output, "o", "", "The file in which to save the dot file.\nThe type is inferred from the extension (.dot for Graphviz and .mmd for Mermaid)")
	flag.Parse()

	internal.SSMModuleVersion = ssmModuleVersion
	internal.SSMModulePath = ssmModulePath
	internal.SSMName = ssmName
	internal.SSMStateType = ssmStateType

	targets := internal.FindAllTargets(flag.Args())
	states, err := internal.LoadStates(targets)
	if err != nil {
		log.Panicf("Error: %s", err)
	}

	references := make(map[string]dot.Node)

	g := dot.NewGraph(dot.Directed)
	for _, state := range states {
		sg := g
		if len(state.Group) > 0 {
			sg = g.Subgraph(state.Group, dot.ClusterOption{})
		}
		references[state.Name] = sg.Node(state.Name)
	}

	for _, state := range states {
		if n1, ok := references[state.Name]; ok {
			for _, next := range state.NextStates {
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
