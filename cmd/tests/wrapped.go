package tests

import (
	"context"
	"time"

	"git.sr.ht/~mariusor/ssm"
)

func Wait(ctx context.Context) ssm.Fn {
	return ssm.NonBlocking(
		ssm.After(10*time.Millisecond, Wait),
	)
}
