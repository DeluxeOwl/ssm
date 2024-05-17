package ssm

import "context"

type Fn func(context.Context) Fn

type aggregatorFn func(...Fn) Fn

// aggStates filters out End states from the received "states" list.
// If the resulting list is empty, the End state is returned.
// If it's one element long it gets returned directly as there's no need to call the batch aggregatorFn on it.
// For other cases, the batch aggregatorFn function gets called.
func aggStates(batchFn aggregatorFn, states ...Fn) Fn {
	for i, state := range states {
		if IsEnd(state) {
			states = append(states[:i], states[i+1:]...)
		}
	}

	if len(states) == 0 {
		return End
	}

	if len(states) == 1 {
		return states[0]
	}

	return batchFn(states...)
}

// Run executes the received states machine in a loop in sequential fashion
// until it's reduced to a single End, or ErrorEnd state, when it stops and
// returns the corresponding error.
func Run(ctx context.Context, states ...Fn) error {
	return run(ctx, aggStates(batchExec, states...))
}

// RunParallel executes the received states machine in a loop in parallel fashion
// until it's reduced to a single End, or ErrorEnd state, when it stops and
// returns the corresponding error.
func RunParallel(ctx context.Context, states ...Fn) error {
	return run(ctx, aggStates(parallelExec, states...))
}

func run(ctx context.Context, state Fn) error {
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
