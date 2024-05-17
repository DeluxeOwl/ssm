package ssm

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	type args struct {
		when  time.Duration
		state Fn
	}
	tests := []struct {
		name string
		args args
		ctx  context.Context
		want Fn
	}{
		{
			name: "nil",
			args: args{},
			ctx:  context.Background(),
		},
		{
			name: "timeout exceeded",
			args: args{
				when: time.Microsecond,
				state: func(ctx context.Context) Fn {
					<-time.After(time.Millisecond)
					return ErrorEnd(errors.New("waited for 1ms"))
				},
			},
			ctx:  context.Background(),
			want: TimeoutExceeded(),
		},
		{
			name: "timeout not reached",
			args: args{
				when: time.Second,
				state: func(ctx context.Context) Fn {
					<-time.After(200 * time.Microsecond)
					return ErrorEnd(errors.New("waited for 200μs"))
				},
			},
			ctx:  context.Background(),
			want: ErrorEnd(errors.New("waited for 200μs")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			to := Timeout(tt.args.when, tt.args.state)
			if IsEnd(to) {
				if !IsEnd(tt.want) {
					t.Errorf("Timeout() = %v, wanted %v", nameOf(to), nameOf(tt.want))
				}
				return
			}
			if got := to(tt.ctx); !sameEndStates(tt.want, got) {
				t.Errorf("Timeout() = %v, want %v", nameOf(got), nameOf(tt.want))
			}
		})
	}
}

func sameEndStates(s1, s2 Fn) bool {
	e1 := Run(context.Background(), s1)
	e2 := Run(context.Background(), s2)

	if e1 == nil {
		return e2 == nil
	}
	return e2 != nil && e1.Error() == e2.Error()
}
