package ssm

import (
	"context"
)

type Fn func(context.Context) Fn

type aggregatorFn func(...Fn) Fn

var End Fn = nil

func aggStates(batch aggregatorFn, states ...Fn) Fn {
	if len(states) == 0 {
		return End
	}
	if len(states) == 1 {
		return states[0]
	}
	return batch(states...)
}

func Run(ctx context.Context, states ...Fn) {
	loopStates(ctx, batchStates(states...))
}

func RunParallel(ctx context.Context, states ...Fn) {
	loopStates(ctx, parallelStates(states...))
}

func loopStates(ctx context.Context, state Fn) {
	if state == nil {
		return
	}

	ctx = context.WithValue(ctx, __start, state)

	for {
		if next := state(ctx); next != nil {
			state = next
			continue
		}
		break
	}
}
