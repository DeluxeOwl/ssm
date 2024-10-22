package ssm

import (
	"context"
	"time"
)

// After runs the received state after d time.Duration has elapsed.
// This function blocks until the timer elapses, when it returns the next resolved state.
func After(d time.Duration, state Fn) Fn {
	return after(d).run(state)
}

type after time.Duration

func (e after) run(states ...Fn) Fn {
	run := aggStates(batchExec, states...)
	if IsEnd(run) {
		return End
	}

	return runAfter(time.Duration(e), run)
}

func runAfter(d time.Duration, run Fn) Fn {
	return func(ctx context.Context) Fn {
		done := make(chan Fn)
		time.AfterFunc(d, func() {
			done <- run(ctx)
		})
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				return ErrorEnd(err)
			}
			return End
		case next := <-done:
			return next
		}
	}
}
