package ssm

import (
	"context"
	"sync"
)

// Parallel executes the received states in parallel goroutines, and accumulates the next states.
// The resulting next state is returned as a parallel batch of all the non End states resolved.
func Parallel(states ...Fn) Fn {
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
		wg := sync.WaitGroup{}

		for _, state := range states {
			if IsEnd(state) {
				continue
			}

			wg.Add(1)
			go func(st Fn) {
				wg.Done()
				c <- st(ctx)
			}(state)
		}

		select {
		case st, _ := <-c:
			if !IsEnd(st) {
				nextStates = append(nextStates, st)
			}
		}
		wg.Wait()
		return parallelStates(nextStates...)
	}
}

func parallelStates(states ...Fn) Fn {
	return aggStates(Parallel, states...)
}
