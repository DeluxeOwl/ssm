package ssm

import (
	"context"
	"sync"
)

func parallelStates(states ...Fn) Fn {
	return aggStates(Parallel, states...)
}

func Parallel(states ...Fn) Fn {
	if len(states) == 0 {
		return nil
	}

	return func(ctx context.Context) Fn {
		nextStates := make([]Fn, 0, len(states))

		wg := sync.WaitGroup{}
		for _, state := range states {
			if state == nil {
				continue
			}

			go func(st Fn) {
				wg.Add(1)
				if next := st(ctx); next != nil {
					nextStates = append(nextStates, next)
				}
				wg.Done()
			}(state)
		}
		wg.Wait()

		return parallelStates(nextStates...)
	}
}
