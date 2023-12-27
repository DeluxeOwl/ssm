package ssm

import (
	"context"
	"time"
)

func At(t time.Time, state Fn) Fn {
	return timer(t).run(state)
}

type timer time.Time

func (t timer) run(states ...Fn) Fn {
	run := aggStates(states...)
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
