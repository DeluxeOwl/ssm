package ssm

import "context"

// Batch executes the received states sequentially, and accumulates the next states.
// The resulting next state is returned as a sequential batch of all the non End states resolved.
func Batch(states ...Fn) Fn {
	return aggStates(batchExec, states...)
}

func batchExec(states ...Fn) Fn {
	if len(states) == 0 {
		return End
	}

	return func(ctx context.Context) Fn {
		nextStates := make([]Fn, 0, len(states))

		for _, state := range states {
			if IsEnd(state) {
				continue
			}

			st := state(ctx)

			if !IsEnd(st) {
				nextStates = append(nextStates, st)
			}
		}
		return aggStates(batchExec, nextStates...)
	}
}
