package ssm

import (
	"context"
	"testing"
	"time"
)

var iterDuration = 100 * time.Millisecond

func stateWithTime(t *testing.T, startTime time.Time) Fn {
	return func(ctx context.Context) (Fn, error) {
		diff := time.Now().Sub(startTime)
		if diff.Truncate(time.Millisecond) != iterDuration {
			t.Errorf("Execution time: %s, expected %s", diff, iterDuration)
		} else {
			t.Logf("Executed after %s", diff)
		}
		return End, nil
	}
}

func TestEvery(t *testing.T) {
	type args struct {
		e     time.Duration
		state Fn
	}
	tests := []struct {
		name         string
		args         args
		wantEndState bool
		wantErr      bool
	}{
		{
			name: "exec in 100ms",
			args: args{
				e:     iterDuration,
				state: stateWithTime(t, time.Now()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Every(tt.args.e, tt.args.state)(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Every()() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
