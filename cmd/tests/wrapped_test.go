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
	_, f, _, _ := runtime.Caller(0)
	states, _ := internal.LoadStates(f)
	_ = dot.Dot("", states...)

	// Output: digraph  {
	//	subgraph cluster_s3 {
	//		label="ssm";
	//		n5[label="After"];
	//		n4[label="NonBlocking"];
	//
	//	}
	//	subgraph cluster_s1 {
	//		label="tests";
	//		n2[label="Wait"];
	//
	//	}
	//
	//	n4->n5;
	//	n2->n4;
	//
	//}
}
