package ssm

import (
	"context"
	"errors"
)

type Fn func(context.Context) (Fn, error)

var End Fn = nil

func Batch(states ...Fn) Fn {
	if len(states) == 0 {
		return nil
	}
	return func(ctx context.Context) (Fn, error) {
		nextStates := make([]Fn, 0, len(states))
		errs := make([]error, 0)

		for _, state := range states {
			if state == nil {
				continue
			}

			ns, err := state(ctx)
			if err != nil {
				errs = append(errs, err)
			}

			if ns != nil {
				nextStates = append(nextStates, ns)
			}
		}

		var next Fn
		if len(nextStates) > 0 {
			next = Batch(nextStates...)
		}
		if len(errs) > 0 {
			return next, errors.Join(errs...)
		}

		return next, nil
	}
}

func Run(ctx context.Context, states ...Fn) error {
	var err error

	state := Batch(states...)
	for {
		state, err = state(ctx)
		if err != nil {
			return err
		}
		if state == nil {
			break
		}
	}
	return nil
}
