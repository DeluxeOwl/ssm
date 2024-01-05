package mocks

import (
	"context"
	"fmt"
	"time"

	sm "git.sr.ht/~mariusor/ssm"
)

// LogAndStop -> ssm.ErrorEnd -> ssm.End
func LogAndStop(err error) sm.Fn {
	return sm.ErrorEnd(err)
}

// Stop -> ssm.End
func Stop() sm.Fn {
	return func(_ context.Context) sm.Fn {
		return sm.End
	}
}

type ts string

// run -> LogAndStop -> ssm.ErrorEnd -> ssm.End
func (s ts) run(_ context.Context) sm.Fn {
	if time.Now().IsZero() {
		return Stop()
	}
	return LogAndStop(fmt.Errorf("test"))
}
