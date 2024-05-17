package ssm

import (
	"context"
	"testing"
	"time"
)

func TestFixedWindow(t *testing.T) {
	type args struct {
		count int
		d     time.Duration
	}
	tests := []struct {
		name           string
		iter           int
		args           args
		stalledPattern []bool
		delayPattern   []time.Duration
	}{
		{
			name: "nil",
			args: args{},
		},
		{
			name: "10 iter / ms",
			args: args{
				count: 10,
				d:     time.Millisecond,
			},
			iter: 20,
			stalledPattern: []bool{
				false, false, false, false, false, false, false, false, false, true,
				false, false, false, false, false, false, false, false, false, true,
			},
			delayPattern: []time.Duration{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 100 * time.Microsecond,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 100 * time.Microsecond,
			},
		},
		{
			name: "4 iter / 10ms",
			args: args{
				count: 4,
				d:     10 * time.Millisecond,
			},
			iter: 12,
			stalledPattern: []bool{
				false, false, false, true,
				false, false, false, true,
				false, false, false, true,
			},
			delayPattern: []time.Duration{
				0, 0, 0, 2500 * time.Microsecond,
				0, 0, 0, 2500 * time.Microsecond,
				0, 0, 0, 2500 * time.Microsecond,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FixedWindow(tt.args.count, tt.args.d)
			for i := 0; i < tt.iter; i++ {
				gotStalled, gotDelay := got()
				if i >= len(tt.stalledPattern) {
					t.Errorf("iter %d, only %d length for stalled want slice", i, len(tt.stalledPattern))
					return
				}
				if i >= len(tt.delayPattern) {
					t.Errorf("iter %d, only %d length for delay want slice", i, len(tt.delayPattern))
					return
				}
				if gotStalled != tt.stalledPattern[i] {
					t.Errorf("iter %d invalid status %t, wanted %t", i, gotStalled, tt.stalledPattern[i])
				}
				if gotDelay != tt.delayPattern[i] {
					t.Errorf("iter %d invalid delay %s, wanted %s", i, gotDelay, tt.delayPattern[i])
				}
			}
		})
	}
}

func TestRateLimit(t *testing.T) {
	type args struct {
		limitFn LimitStrategyFn
		state   Fn
	}
	tests := []struct {
		name string
		args args
		want Fn
	}{
		{
			name: "nil",
		},
		{
			name: "rate limit",
			args: args{
				limitFn: FixedWindow(1, time.Millisecond),
				state:   mockEmpty,
			},
			want: mockEmpty,
		},
		{
			name: "rate limit",
			args: args{
				limitFn: FixedWindow(1, time.Microsecond),
				state: func(_ context.Context) Fn {
					<-time.After(10 * time.Millisecond)
					return End
				},
			},
			want: After(10*time.Millisecond, End),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			to := RateLimit(tt.args.limitFn, tt.args.state)
			if IsEnd(to) {
				if !IsEnd(tt.want) {
					t.Errorf("RateLimit() = %v, wanted %v", nameOf(to), nameOf(tt.want))
				}
				return
			}
			if got := to(context.Background()); !sameEndStates(got, tt.want) {
				t.Errorf("RateLimit() = %v, want %v", nameOf(got), nameOf(tt.want))
			}
		})
	}
}
