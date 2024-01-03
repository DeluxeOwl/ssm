package ssm

import (
	"context"
	"time"
)

// Retry is a way to construct a state machine out of repeating the execution of received state "fn", until the number of
// "retries" has been reached, or "fn" returns the End state.
// The "strategy" parameter can be used as a way to delay the execution between retries. If nil is passed the Batch
// aggregator function is used, which just executes the code sequentially without pause.
func Retry(retries int, fn Fn) Fn {
	return retry(retries, fn)
}

// BackOff returns an aggregator function which can be used to execute the received state with increasing delays.
// The function for determining the delay is passed in the StrategyFn "dur" parameter.
//
// There is no end condition, so take care to limit the execution through some external method.
func BackOff(dur StrategyFn, fn Fn) Fn {
	return func(ctx context.Context) Fn {
		return after(dur()).run(fn)(ctx)
	}
}

func retry(retries int, fn Fn) Fn {
	return func(ctx context.Context) Fn {
		for {
			if next := fn(ctx); next == nil {
				break
			}
			if retries-1 <= 0 {
				break
			}
			return retry(retries-1, fn)
		}
		return End
	}
}

// StrategyFn is the type that returns the desired time.Duration for the BackOff function.
type StrategyFn func() time.Duration

// Constant returns a constant time.Duration for every call.
func Constant(d time.Duration) StrategyFn {
	return func() time.Duration {
		return d
	}
}

// Double returns the double of the time.Duration for every call.
func Double(d time.Duration) StrategyFn {
	return func() time.Duration {
		t := d
		d *= 2
		return t
	}
}
