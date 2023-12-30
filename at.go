package ssm

import (
	"context"
	"time"
)

// At runs the received state at t time.Time.
// This function blocks until the time is reached, when it returns the next resolved state.
func At(t time.Time, state Fn) Fn {
	return alarm(t).run(state)
}

type alarm time.Time

func (t alarm) run(states ...Fn) Fn {
	run := batchStates(states...)
	if run == nil {
		return End
	}

	return func(ctx context.Context) Fn {
		done := make(chan struct{})
		time.AfterFunc(time.Time(t).Sub(time.Now()), func() {
			run = run(ctx)
			done <- struct{}{}
		})
		select {
		case <-done:
			break
		}
		return run
	}
}
