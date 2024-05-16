package ssm

import "context"

type Fn func(context.Context) Fn

type aggregatorFn func(...Fn) Fn

func aggStates(batch aggregatorFn, states ...Fn) Fn {
	if len(states) == 0 {
		return End
	}
	if len(states) == 1 {
		return states[0]
	}
	return batch(states...)
}

func Run(ctx context.Context, states ...Fn) error {
	return loop(ctx, aggStates(batchStates, states...))
}

func RunParallel(ctx context.Context, states ...Fn) error {
	return loop(ctx, aggStates(parallelStates, states...))
}

func loop(ctx context.Context, state Fn) error {
	if IsEnd(state) {
		return nil
	}
	var cancel context.CancelCauseFunc
	var err error

	ctx, cancel = context.WithCancelCause(ctx)
	ctx = context.WithValue(ctx, __start, state)
	if cancel != nil {
		ctx = context.WithValue(ctx, __cancel, cancel)
	}

	for {
		select {
		case <-ctx.Done():
			if err = context.Cause(ctx); err != nil {
				state = ErrorEnd(err)
				break
			}
			state = End
			break
		default:
			if next := state(ctx); !IsEnd(next) {
				state = next
				continue
			}
		}
		break
	}
	return context.Cause(ctx)
}
