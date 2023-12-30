package ssm

import (
	"context"
	"testing"
	"time"
)

const defaultDelay = 100 * time.Millisecond

func stateWithTime(t *testing.T, startTime time.Time) Fn {
	return func(ctx context.Context) Fn {
		diff := time.Now().Sub(startTime)
		if diff.Truncate(time.Millisecond) != defaultDelay {
			t.Errorf("Execution time: %s, expected %s", diff, defaultDelay)
		} else {
			t.Logf("Executed after %s", diff)
		}
		return End
	}
}

func TestEvery(t *testing.T) {
	type args struct {
		e     time.Duration
		state Fn
	}
	tests := []struct {
		name     string
		args     args
		endState Fn
	}{
		{
			name: "exec in 100ms",
			args: args{
				e:     defaultDelay,
				state: stateWithTime(t, time.Now()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Every(tt.args.e, tt.args.state)(context.Background())
			if ptrOf(got) != ptrOf(tt.endState) {
				t.Errorf("Every()() = %v, wantErr %v", got, tt.endState)
			}
		})
	}
}
