package ssm

import (
	"math"
	"testing"
	"time"
)

func TestMaxTriesTrip(t *testing.T) {
	tests := []struct {
		name string
		max  int
		want []bool
	}{
		{
			name: "empty",
		},
		{
			name: "never tripped",
			max:  -1,
			want: []bool{false, false, false, false, false, false},
		},
		{
			name: "three not tripped",
			max:  3,
			want: []bool{false, false, true, true},
		},
		{
			name: "never tripped",
			max:  math.MaxInt32, // this looks quite suspicious
			want: []bool{false, false, false, false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := MaxTriesTrip(tt.max)
			for i, want := range tt.want {
				if got := fn(); got != want {
					t.Errorf("MaxTriesTrip() iter %d = %v, want %v", i, got, want)
				}
			}
		})
	}
}

func div3Trip(t *testing.T, start time.Time) func() TripStrategyFn {
	return func() TripStrategyFn {
		c := 0
		t.Logf("reset at %s", time.Now().Sub(start).Truncate(100*time.Microsecond))
		return func() bool {
			status := c > 0 && c%3 == 0
			c++
			return status
		}
	}
}

func TestStatus(t *testing.T) {
	tests := []struct {
		name string
		want []bool
	}{
		{
			name: "empty",
			want: nil,
		},
		{
			name: "check 13 successive calls",
			want: []bool{false, false, false, true, false, false, true, false, false, true, false, false, true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := div3Trip(t, time.Now())()
			for i, want := range tt.want {
				if got := fn(); got != want {
					t.Errorf("div3Trip() iter %d = %v, want %v", i, got, want)
				}
			}
		})
	}
}

func TestTimedTrip(t *testing.T) {
	type args struct {
		d  time.Duration
		fn TripStrategyFn
	}

	resetMs := 5 * time.Millisecond

	tests := []struct {
		name string
		args args
		want []bool
	}{
		{
			name: "empty",
			args: args{
				time.Millisecond,
				nil,
			},
			want: nil,
		},
		{
			name: "check calls",
			args: args{
				resetMs,
				func() bool {
					return false
				},
			},
			want: []bool{false, false, false, false, false, false, false, false, false, false, false, false, false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := TimedTrip(tt.args.d, tt.args.fn)
			for i, want := range tt.want {
				if got := fn(); got != want {
					t.Errorf("TimedTrip() iter %d = %v, want %v", i, got, want)
				}
			}
		})
	}
}
