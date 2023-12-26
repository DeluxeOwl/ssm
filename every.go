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
	if len(states) == 0 {
		return nil
	}

	var run Fn
	if len(states) > 1 {
		run = Batch(states...)
	} else {
		run = states[0]
	}

	var err error

	return func(ctx context.Context) (Fn, error) {
		done := make(chan struct{})
		time.AfterFunc(time.Duration(e), func() {
			run, err = run(ctx)
			done <- struct{}{}
		})
		select {
		case <-done:
			break
		}
		return run, err
	}
}
