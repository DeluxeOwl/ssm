package ssm

import (
	"context"
	"time"
)

func Every(e time.Duration, states ...Fn) Fn {
	return every(e).run(states...)
}

type every time.Duration

func (e every) run(states ...Fn) Fn {
	run := Batch(states...)

	var next Fn
	var err error

	return func(ctx context.Context) (Fn, error) {
		done := make(chan struct{})
		time.AfterFunc(time.Duration(e), func() {
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
