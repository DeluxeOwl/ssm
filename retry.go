package ssm

import (
	"context"
	"math/rand"
	"time"
)

// Retry is a way to construct a state machine out of repeating the execution of received state "fn",
// until the number of "retries" has been reached, or "fn" returns the End state.
//
// The "fn" parameter can be one of the functions accepting a StrategyFn parameters,
// which wrap the original state Fn, and which provide a way to delay the execution between retries.
func Retry(retries int, fn Fn) Fn {
	return retry(retries, fn)
}

func retry(retries int, fn Fn) Fn {
	return func(ctx context.Context) Fn {
		for {
			next := fn(ctx)
			if !IsError(next) {
				return next
			}
			if retries-1 <= 0 {
				return next
			}
			return retry(retries-1, fn)
		}
	}
}

// StrategyFn is the type that returns the desired time.Duration for the BackOff function.
type StrategyFn func() time.Duration

// BackOff returns an aggregator function which can be used to execute the received state with increasing delays.
// The function for determining the delay is passed in the StrategyFn "dur" parameter.
//
// There is no end condition, so take care to limit the execution through some external method.
func BackOff(dur StrategyFn, fn Fn) Fn {
	return func(ctx context.Context) Fn {
		return after(dur()).run(fn)(ctx)
	}
}

// Constant returns a constant time.Duration for every call.
func Constant(d time.Duration) StrategyFn {
	return func() time.Duration {
		return d
	}
}

// Linear returns the linear function of the time.Duration multiplied by mul for every call.
func Linear(d time.Duration, m float64) StrategyFn {
	return func() time.Duration {
		t := d
		d = time.Duration(m) * d
		return t
	}
}

// Jitter adds random jitter of "max" time.Duration for the fn StrategyFn
func Jitter(max time.Duration, fn StrategyFn) StrategyFn {
	return func() time.Duration {
		j := rand.Int63n(int64(max))
		return fn() + time.Duration(j)
	}
}
