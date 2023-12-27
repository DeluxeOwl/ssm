package ssm

import (
	"context"
)

type Fn func(context.Context) Fn

var End Fn = nil

func Batch(states ...Fn) Fn {
	if len(states) == 0 {
		return nil
	}
	return func(ctx context.Context) Fn {
		nextStates := make([]Fn, 0, len(states))

		for _, state := range states {
			if state == nil {
				continue
			}

			ns := state(ctx)

			if ns != nil {
				nextStates = append(nextStates, ns)
			}
		}

		return aggStates(nextStates...)
	}
}

func aggStates(states ...Fn) Fn {
	if len(states) == 0 {
		return End
	}
	if len(states) == 1 {
		return states[0]
	}
	return Batch(states...)
}

func Run(ctx context.Context, states ...Fn) {
	state := aggStates(states...)

	if state == nil {
		return
	}

	ctx = context.WithValue(ctx, __start, state)

	for {
		if state = state(ctx); state == nil {
			break
		}
	}
}
