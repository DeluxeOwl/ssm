package tests

import (
	"context"
	"runtime"
	"time"

	"git.sr.ht/~mariusor/ssm"
	"git.sr.ht/~mariusor/ssm/cmd/internal"
	"git.sr.ht/~mariusor/ssm/cmd/internal/dot"
)

func Wait(ctx context.Context) ssm.Fn {
	return ssm.NonBlocking(
		ssm.After(10*time.Millisecond, Wait),
	)
}

func Example_Wait() {
	_, f, _, _ := runtime.Caller(0) // f will be the current file path

	states, _ := internal.LoadStates(f)
	_ = dot.Dot("", states...)
	// Output: digraph  {
	//	subgraph cluster_s3 {
	//		label="ssm";
	//		n10[label="After"];
	//		n6[label="End"];
	//		n8[label="ErrorEnd"];
	//		n4[label="NonBlocking"];
	//		n11[label="after.run"];
	//		n9[label="errState.stop"];
	//		n5[label="nb.run"];
	//		n7[label="nb.wait"];
	//		n12[label="runAfter"];
	//
	//	}
	//	subgraph cluster_s1 {
	//		label="tests";
	//		n2[label="Wait"];
	//
	//	}
	//
	//	n10->n11;
	//	n8->n9;
	//	n4->n5;
	//	n4->n10;
	//	n2->n4;
	//	n11->n6;
	//	n11->n12;
	//	n9->n6;
	//	n5->n6;
	//	n5->n7;
	//	n7->n8;
	//	n7->n6;
	//	n7->n7;
	//	n12->n8;
	//	n12->n6;
	//
	//}
}
