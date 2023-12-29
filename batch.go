package ssm

import "context"

func Batch(states ...Fn) Fn {
	if len(states) == 0 {
		return nil
	}
	return func(ctx context.Context) Fn {
		nextStates := make([]Fn, 0, len(states))

		for _, state := range states {
			if state == nil {
				continue
			}

			ns := state(ctx)

			if ns != nil {
				nextStates = append(nextStates, ns)
			}
		}

		return batchStates(nextStates...)
	}
}

func batchStates(states ...Fn) Fn {
	return aggStates(Batch, states...)
}
