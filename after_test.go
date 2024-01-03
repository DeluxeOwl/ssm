package ssm

import (
	"context"
	"testing"
	"time"
)

const defaultDelay = 100 * time.Millisecond

func mockStateWithTime(t *testing.T, startTime time.Time) Fn {
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

func TestAfter(t *testing.T) {
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
				state: mockStateWithTime(t, time.Now()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := After(tt.args.e, tt.args.state)(context.Background())
			if ptrOf(got) != ptrOf(tt.endState) {
				t.Errorf("After()() = %v, wantErr %v", got, tt.endState)
			}
		})
	}
}
