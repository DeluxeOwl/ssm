package mocks

import (
	"context"
	"flag"
	"fmt"

	"git.sr.ht/~mariusor/ssm"
	"git.sr.ht/~mariusor/ssm/cmd/internal"
	"git.sr.ht/~mariusor/ssm/cmd/internal/dot"
)

func example() {
	ssm.Run(context.Background(), ssm.End, ssm.ErrorEnd(fmt.Errorf("text")))
}

func Example_Run() {
	targets := internal.FindAllTargets(flag.Args())
	states, _ := internal.LoadStates(targets...)
	dot.Dot("", states...)

	// Output: digraph  {
	//        subgraph cluster_s1 {
	//                label="ssm";
	//                n3[label="End"];
	//                n4[label="ErrorEnd"];
	//                n2[label="ssm.Run"];
	//
	//        }
	//
	//        n2->n3;
	//        n2->n4;
	//
	//}
}
