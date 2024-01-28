package ssm

import (
	"context"
	"sync"
)

// NonBlocking executes states in a goroutine and until it resolves it returns a wait state
func NonBlocking(states ...Fn) Fn {
	c := make(nb)
	return c.run(states...)
}

type nb chan Fn

func (n nb) run(states ...Fn) Fn {
	run := batchStates(states...)
	if IsEnd(run) {
		return End
	}

	do := func(ctx context.Context, run Fn) func() {
		return func() {
			go func(ctx context.Context, run Fn) {
				n <- run(ctx)
			}(ctx, run)
		}
	}

	return func(ctx context.Context) Fn {
		new(sync.Once).Do(do(ctx, run))

		// On the first iteration we execute the "wait" state in order to check
		// if the channel already returned.
		return n.wait(ctx)
	}
}

func (n nb) wait(ctx context.Context) Fn {
	select {
	case next := <-n:
		return next
	default:
		return n.wait
	}
}
