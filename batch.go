package ssm

import "context"

// Batch executes the received states sequentially, and accumulates the next states.
// The resulting next state is returned as a sequential batch of all the non End states resolved.
func Batch(states ...Fn) Fn {
	if len(states) == 0 {
		return End
	}
	return func(ctx context.Context) Fn {
		nextStates := make([]Fn, 0, len(states))

		for _, state := range states {
			if IsEnd(state) {
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
