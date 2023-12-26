package ssm

import (
	"context"
	"time"
)

func At(t time.Time, states ...Fn) Fn {
	return timer(t).run(states...)
}

type timer time.Time

func (t timer) run(states ...Fn) Fn {
	run := Batch(states...)

	var next Fn
	var err error

	return func(ctx context.Context) (Fn, error) {
		done := make(chan struct{})
		time.AfterFunc(time.Time(t).Sub(time.Now()), func() {
			next, err = run(ctx)
			done <- struct{}{}
		})
		select {
		case <-done:
			break
		}
		return next, err
	}
}
