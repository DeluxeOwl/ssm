package mocks

import (
	"time"

	"git.sr.ht/~mariusor/ssm"
)

// AfterTwoSeconds -> ssm.After -> state
func AfterTwoSeconds(state ssm.Fn) ssm.Fn {
	return ssm.After(2*time.Second, state)
}
