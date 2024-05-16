package tests

import (
	"runtime"
	"time"

	"git.sr.ht/~mariusor/ssm"
	"git.sr.ht/~mariusor/ssm/cmd/internal"
	"git.sr.ht/~mariusor/ssm/cmd/internal/dot"
)

// AfterTwoSeconds -> ssm.After -> state
func AfterTwoSeconds(state ssm.Fn) ssm.Fn {
	return ssm.After(2*time.Second, state)
}

func Example_AfterTwoSeconds() {
	_, f, _, _ := runtime.Caller(0) // f will be the current file path

	states, _ := internal.LoadStates(f)
	_ = dot.Dot("", states...)
	// Output: digraph  {
	//	subgraph cluster_s3 {
	//		label="ssm";
	//		n4[label="After"];
	//		n6[label="End"];
	//		n5[label="after.run"];
	//		n7[label="runAfter"];
	//
	//	}
	//	subgraph cluster_s1 {
	//		label="tests";
	//		n2[label="AfterTwoSeconds"];
	//
	//	}
	//
	//	n4->n5;
	//	n2->n4;
	//	n5->n6;
	//	n5->n7;
	//
	//}
}
