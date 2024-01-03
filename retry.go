package ssm

import (
	"context"
	"time"
)

// Retry is a way to construct a state machine out of repeating the execution of received "fn", until the number of
// "retries" has been reached, or "fn" returns no error.
// The "strategy" parameter can be used as a way to delay the execution between retries. If nil is passed the Batch
// aggregator function is used, which just executes the code sequentially without pause.
func Retry(retries int, strategy aggregatorFn, fn func(ctx context.Context) error) Fn {
	return retry(retries, strategy, fn)
}

func retry(times int, strategy aggregatorFn, fn func(context.Context) error) Fn {
	if strategy == nil {
		strategy = Batch
	}
	return func(ctx context.Context) Fn {
		for {
			if err := fn(ctx); err == nil {
				break
			}
			if times-1 <= 0 {
				break
			}
			return strategy(retry(times-1, strategy, fn))
		}
		return End
	}
}

// BackOff returns an aggregator function which can be used to execute the received states with increasing delays.
// The initial delay is passed through the delay time.Duration parameter, and the method of increase is delay *= 5
//
// There is no end condition, so take care to limit the execution through some external method.
func BackOff(delay time.Duration) aggregatorFn {
	bb := backoff(delay)
	return bb.run
}

type backoff time.Duration

func (b *backoff) run(states ...Fn) Fn {
	nextRun := time.Now().Add(time.Duration(*b))
	*b = 5 * (*b)
	return alarm(nextRun).run(states...)
}
