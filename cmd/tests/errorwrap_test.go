package tests

import (
	"context"
	"fmt"
	"runtime"
	"time"

	sm "git.sr.ht/~mariusor/ssm"
	"git.sr.ht/~mariusor/ssm/cmd/internal"
	"git.sr.ht/~mariusor/ssm/cmd/internal/dot"
)

// LogAndStop -> ssm.ErrorEnd -> ssm.End
func LogAndStop(err error) sm.Fn {
	return sm.ErrorEnd(err)
}

// Stop -> ssm.End
func Stop() sm.Fn {
	return func(_ context.Context) sm.Fn {
		return sm.End
	}
}

type ts string

// run -> LogAndStop -> ssm.ErrorEnd -> ssm.End
func (s ts) run(_ context.Context) sm.Fn {
	if time.Now().IsZero() {
		return Stop()
	}
	return LogAndStop(fmt.Errorf("test"))
}

func Example_ts_run() {
	_, f, _, _ := runtime.Caller(0) // f will be the current file path

	states, _ := internal.LoadStates(f)
	_ = dot.Dot("", states...)

	// Output: digraph  {
	//	subgraph cluster_s1 {
	//		label="tests";
	//		n2[label="LogAndStop"];
	//		n3[label="Stop"];
	//		n4[label="ts.run"];
	//
	//	}
	//
	//	n4->n3;
	//	n4->n2;
	//
	//}
}
