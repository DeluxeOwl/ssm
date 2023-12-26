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
		name         string
		args         args
		wantEndState bool
		wantErr      bool
	}{
		{
			name: "exec now + 100ms",
			args: args{
				t:     time.Now().Add(iterDuration),
				state: stateWithTime(t, time.Now()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := At(tt.args.t, tt.args.state)(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("At() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
