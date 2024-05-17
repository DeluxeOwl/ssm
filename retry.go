package ssm

import (
	"context"
	"math/rand"
	"sync/atomic"
	"time"
)

// Retry is a way to construct a state machine out of repeating the execution of received state "fn",
// when it returns an IsError state until the number of "retries" has been reached.
//
// The "fn" parameter can be one of the functions accepting a StrategyFn parameters,
// which wrap the original state Fn, and which provide a way to delay the execution between retries.
//
// The Retry state machine is reentrant, therefore can be used from multiple goroutines.
func Retry(count int, fn Fn) Fn {
	return retries(count).run(fn)
}

type ar atomic.Int32

func retries(count int) *ar {
	i := atomic.Int32{}
	i.Store(int32(count - 1))
	return (*ar)(&i)
}

func (r *ar) run(fn Fn) Fn {
	i := (*atomic.Int32)(r)
	return func(ctx context.Context) Fn {
		for {
			next := fn(ctx)
			if !IsError(next) {
				return next
			}
			if i.Load() > 0 {
				i.Add(-1)
				return r.run(fn)
			}
			return next
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
		// linearly increase duration for next run
		d = time.Duration(m) * d
		return t
	}
}

// Jitter adds random jitter of "max" time.Duration for the fn StrategyFn
func Jitter(max time.Duration, fn StrategyFn) StrategyFn {
	var d time.Duration
	return func() time.Duration {
		j := rand.Int63n(int64(max))
		if fn != nil {
			d = fn()
		}

		return d + time.Duration(j)
	}
}
