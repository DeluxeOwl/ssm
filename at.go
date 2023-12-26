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
		time.AfterFunc(time.Time(t).Sub(time.Now()), func() {
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
