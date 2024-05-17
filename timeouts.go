package ssm

import (
	"context"
	"time"
)

// Timeout is a state machine that calls the "state" Fn limiting its execution time
// to a maximum of "max" time.Duration.
// When the timeout is reached the execution is canceled.
func Timeout(max time.Duration, state Fn) Fn {
	if IsEnd(state) {
		return state
	}

	var cancel func()
	return func(ctx context.Context) Fn {
		ctx, cancel = context.WithTimeout(ctx, max)
		defer cancel()

		next := make(chan Fn)
		go func() {
			next <- state(ctx)
		}()

		select {
		case <-ctx.Done():
			if err := context.Cause(ctx); err != nil {
				return TimeoutExceeded()
			}
		case st := <-next:
			return st
		}

		return End
	}
}

// TimeoutExceeded is an error state that is used when a Timeout is reached.
func TimeoutExceeded() Fn {
	return errState{context.DeadlineExceeded}.stop
}
