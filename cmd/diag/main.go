package main

import (
	"flag"
	"log"

	"git.sr.ht/~mariusor/ssm/cmd/internal"
	"git.sr.ht/~mariusor/ssm/cmd/internal/dot"
)

func main() {
	var output string
	flag.StringVar(&output, "o", "", "The file in which to save the dot file.\nThe type is inferred from the extension (.dot for Graphviz and .mmd for Mermaid)")
	flag.Parse()

	targets := internal.FindAllTargets(flag.Args()...)
	states, err := internal.LoadStates(targets...)
	if err != nil {
		log.Panicf("Error: %s", err)
	}
	if err := dot.Dot(output, states...); err != nil {
		log.Panicf("Error: %s", err)
	}
}
