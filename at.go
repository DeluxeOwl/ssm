package ssm

import (
	"time"
)

// At runs the received state at t time.Time.
// This function blocks until the time is reached, when it returns the next resolved state.
func At(t time.Time, state Fn) Fn {
	return alarm(t).run(state)
}

type alarm time.Time

func (t alarm) run(states ...Fn) Fn {
	run := batchExec(states...)
	if IsEnd(run) {
		return End
	}

	return runAfter(time.Time(t).Sub(time.Now()), run)
}
