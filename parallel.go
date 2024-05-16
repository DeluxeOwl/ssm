package ssm

import (
	"context"
)

// Parallel executes the received states in parallel goroutines, and accumulates the next states.
// The resulting next state is returned as a parallel batch of all the non End states resolved.
func Parallel(states ...Fn) Fn {
	return aggStates(parallelStates, states...)
}

func parallelStates(states ...Fn) Fn {
	for i, state := range states {
		if IsEnd(state) {
			states = append(states[:i], states[i+1:]...)
		}
	}

	if len(states) == 0 {
		return End
	}

	return func(ctx context.Context) Fn {
		nextStates := make([]Fn, 0, len(states))
		c := make(chan Fn, len(states))

		for _, state := range states {
			go func(st Fn) {
				c <- st(ctx)
			}(state)
		}

		for range states {
			nextStates = append(nextStates, <-c)
		}
		return aggStates(parallelStates, nextStates...)
	}
}
