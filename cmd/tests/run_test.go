package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"git.sr.ht/~mariusor/ssm"
	"git.sr.ht/~mariusor/ssm/cmd/internal"
	"git.sr.ht/~mariusor/ssm/cmd/internal/dot"
)

var cwd, _ = os.Getwd()

func example() {
	ssm.Run(context.Background(), ssm.End, ssm.ErrorEnd(fmt.Errorf("text")))
}

func Example_Run() {
	states, _ := internal.LoadStates(filepath.Join(cwd, "run_test.go"))
	dot.Dot("", states...)

	// Output: digraph  {
	//	subgraph cluster_s1 {
	//		label="ssm";
	//		n3[label="End"];
	//		n4[label="ErrorEnd"];
	//		n2[label="ssm.Run"];
	//
	//	}
	//
	//	n2->n3;
	//	n2->n4;
	//
	//}
}
