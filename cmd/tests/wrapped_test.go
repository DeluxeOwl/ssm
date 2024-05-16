package tests

import (
	"context"
	"git.sr.ht/~mariusor/ssm/cmd/internal"
	"git.sr.ht/~mariusor/ssm/cmd/internal/dot"
	"runtime"
	"time"

	"git.sr.ht/~mariusor/ssm"
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
	//		n8[label="After"];
	//		n6[label="End"];
	//		n4[label="NonBlocking"];
	//		n9[label="after.run"];
	//		n5[label="nb.run"];
	//		n7[label="nb.wait"];
	//		n10[label="runAfter"];
	//
	//	}
	//	subgraph cluster_s1 {
	//		label="tests";
	//		n2[label="Wait"];
	//
	//	}
	//
	//	n8->n9;
	//	n4->n5;
	//	n4->n8;
	//	n2->n4;
	//	n9->n6;
	//	n9->n10;
	//	n5->n6;
	//	n5->n7;
	//	n7->n7;
	//
	//}
}
