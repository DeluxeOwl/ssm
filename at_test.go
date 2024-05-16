package ssm

import (
	"context"
	"testing"
	"time"
)

func TestAt(t *testing.T) {
	type args struct {
		t     time.Time
		state Fn
	}
	tests := []struct {
		name     string
		args     args
		endState Fn
	}{
		{
			name: "exec now + 100ms",
			args: args{
				t:     time.Now().Add(defaultDelay),
				state: mockStateWithTime(t, time.Now()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := At(tt.args.t, tt.args.state)(context.Background())
			if !sameFns(got, tt.endState) {
				t.Errorf("At()() = %v, wantErr %v", nameOf(got), nameOf(tt.endState))
			}
		})
	}
}
