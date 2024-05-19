package ssm

import (
	"context"
	"errors"
	"time"
)

// RateLimit is a state machine that executes the "state" Fn under the constraints
// of the "limitFn" LimitStrategyFn.
//
// The strategy function returns if the current execution needs to be stalled in order
// to fulfill the rate limit logic it corresponds to, together with what the corresponding
// delay should be, if it does.
func RateLimit(limitFn LimitStrategyFn, states ...Fn) Fn {
	if limitFn == nil {
		return ErrorEnd(InvalidRateLimitFn)
	}
	if len(states) == 0 {
		return End
	}
	return func(ctx context.Context) Fn {
		aggFn := batchExec
		if stall, delay := limitFn(); stall {
			aggFn = staggerExec(delay)
		}
		return aggStates(aggFn, states...)
	}
}

var InvalidRateLimitFn = errors.New("invalid rate limit method")

func staggerExec(delay time.Duration) func(states ...Fn) Fn {
	return func(states ...Fn) Fn {
		if len(states) == 0 {
			return End
		}

		return func(ctx context.Context) Fn {
			nextStates := make([]Fn, 0, len(states))

			for i, state := range states {
				if IsEnd(state) {
					continue
				}

				st := after(time.Duration(i) * delay).run(state(ctx))

				if !IsEnd(st) {
					nextStates = append(nextStates, st)
				}
			}
			return aggStates(batchExec, nextStates...)
		}
	}
}

// LimitStrategyFn are a type of functions that determine if successive calls to
// a RateLimit'ed state require stalling and how much stalling is required to fulfill
// the desired rate limit.
type LimitStrategyFn func() (bool, time.Duration)

func FixedWindow(count int, d time.Duration) func() (bool, time.Duration) {
	if count <= 0 {
		count = 1
	}
	stallTime := d / time.Duration(count)
	cnt := count

	return func() (bool, time.Duration) {
		if cnt-1 > 0 {
			cnt--
			return false, 0
		}

		cnt = count
		return true, stallTime
	}
}
