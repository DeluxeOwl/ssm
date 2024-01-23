package mocks

import (
	"context"
	"fmt"

	"git.sr.ht/~mariusor/ssm"
)

func example() {
	ssm.Run(context.Background(), ssm.End, ssm.ErrorEnd(fmt.Errorf("text")))
}
