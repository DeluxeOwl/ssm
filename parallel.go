package ssm

import (
	"context"
	"sync"
)

// Parallel executes the received states in parallel goroutines, and accumulates the next states.
// The resulting next state is returned as a parallel batch of all the non End states resolved.
func Parallel(states ...Fn) Fn {
	if len(states) == 0 {
		return End
	}

	return func(ctx context.Context) Fn {
		nextStates := make([]Fn, 0, len(states))

		wg := sync.WaitGroup{}
		m := sync.Mutex{}
		for _, state := range states {
			if IsEnd(state) {
				continue
			}
			wg.Add(1)

			go func(st Fn) {
				if next := st(ctx); next != nil {
					m.Lock()
					nextStates = append(nextStates, next)
					m.Unlock()
				}
				wg.Done()
			}(state)
		}
		wg.Wait()

		return parallelStates(nextStates...)
	}
}

func parallelStates(states ...Fn) Fn {
	return aggStates(Parallel, states...)
}
