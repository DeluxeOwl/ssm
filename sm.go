package ssm

import (
	"context"
)

type Fn func(context.Context) Fn

type aggregatorFn func(...Fn) Fn

var End Fn = nil

var _endPtr = ptrOf(End)

func IsEnd(f Fn) bool {
	return ptrOf(f) == _endPtr
}

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
	loop(ctx, batchStates(states...))
}

func RunParallel(ctx context.Context, states ...Fn) {
	loop(ctx, parallelStates(states...))
}

func loop(ctx context.Context, state Fn) {
	if IsEnd(state) {
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
