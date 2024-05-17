package ssm

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

// Breaker is a state machine that can be used for disabling execution of incoming "fn" state if
// its returning state is an error state and the conditions of the "trip" TripStrategyFn are fulfilled.
//
// Currently, there is no method for closing the Breaker once opened.
func Breaker(trip TripStrategyFn, fn Fn) Fn {
	b := b{
		tripCheck: trip,
		fn:        fn,
	}
	return b.check
}

// OpenBreaker is an error state that is used to trip open the Breaker.
func OpenBreaker() Fn {
	return errState{errors.New("open breaker")}.stop
}

// TripStrategyFn is used by the Circuit Breaker state machine to determine if current run
// requires the breaker to trip open.
//
// Our API doesn't allow these functions to do the error check on the returned state of
// the breaker state. It is assumed that the Breaker itself calls the TripStrategyFn function
// only for error states.
type TripStrategyFn func() bool

func neverTrip() bool {
	return false
}

type mt atomic.Int32

func (m *mt) check() bool {
	i := (*atomic.Int32)(m)
	if st := i.Load(); st == 0 {
		return true
	}
	i.Add(-1)
	return false
}

// MaxTriesTrip returns false for "max" invocations, then returns true.
// It is the simplest form of count based circuit breaking.
//
// The check function itself can be run safely in parallel, so if multiple
// checks are needed, MaxTriesTrip must be invoked multiple times.
func MaxTriesTrip(max int) TripStrategyFn {
	if max < 0 {
		return neverTrip
	}
	i := atomic.Int32{}
	i.Store(int32(max - 1))
	m := (*mt)(&i)
	return m.check
}

// TimedTrip uses "fn" TripStrategyFn for returning the status of the Breaker, but it resets it
// every "d" time.Duration.
func TimedTrip(d time.Duration, fn TripStrategyFn) TripStrategyFn {
	if fn == nil {
		// Run at least once
		return MaxTriesTrip(1)
	}

	t := &fn
	// When the timer expires, it means that the passed trip strategy has not opened the breaker, so we reset
	// both the timer and the trip strategy.
	timer := time.NewTimer(d)
	return func() bool {
		select {
		case <-timer.C:
			fn = *t
			timer.Reset(d)
		default:
		}
		return fn()
	}
}

type b struct {
	tripCheck func() bool
	fn        Fn
}

func (b b) check(ctx context.Context) Fn {
	next := b.fn(ctx)
	if IsError(next) && b.tripCheck() {
		return OpenBreaker()
	}
	return b.check
}
