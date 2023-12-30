package ssm

import (
	"context"
	"time"
)

// Every runs the received state after d time.Duration has elapsed.
// This function blocks until the timer elapses, when it returns the next resolved state.
func Every(d time.Duration, state Fn) Fn {
	return timer(d).run(state)
}

type timer time.Duration

func (e timer) run(states ...Fn) Fn {
	run := batchStates(states...)
	if run == nil {
		return End
	}

	return func(ctx context.Context) Fn {
		done := make(chan Fn)
		time.AfterFunc(time.Duration(e), func() {
			done <- run(ctx)
		})
		select {
		case next := <-done:
			return next
		}
	}
}
