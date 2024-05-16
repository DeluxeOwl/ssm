package tests

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"git.sr.ht/~mariusor/ssm"
	"git.sr.ht/~mariusor/ssm/cmd/internal"
	"git.sr.ht/~mariusor/ssm/cmd/internal/dot"
)

var cwd, _ = os.Getwd()

func example() {
	ssm.Run(context.Background(), ssm.End, ssm.ErrorEnd(fmt.Errorf("text")))
}

func Example_example() {
	_, f, _, _ := runtime.Caller(0) // f will be the current file path

	states, _ := internal.LoadStates(f)
	_ = dot.Dot("", states...)
	// Output: digraph  {
	//	subgraph cluster_s1 {
	//		label="ssm";
	//		n3[label="End"];
	//		n4[label="ErrorEnd"];
	//		n5[label="errState.stop"];
	//		n2[label="ssm.Run"];
	//
	//	}
	//
	//	n4->n5;
	//	n5->n3;
	//	n2->n3;
	//	n2->n4;
	//
	//}
}
