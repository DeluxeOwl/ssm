package ssm

import (
	"context"
	"time"
)

func Every(e time.Duration, state Fn) Fn {
	return every(e).run(state)
}

type every time.Duration

func (e every) run(states ...Fn) Fn {
	run := batchStates(states...)
	if run == nil {
		return End
	}

	return func(ctx context.Context) Fn {
		done := make(chan struct{})
		time.AfterFunc(time.Duration(e), func() {
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
