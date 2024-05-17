package ssm

import (
	"context"
	"errors"
	"testing"
)

func TestWrap(t *testing.T) {
	tests := []struct {
		name string
		fn   func(ctx context.Context) error
		want Fn
	}{
		{
			name: "nil",
			want: nil,
		},
		{
			name: "one repeat",
			fn:   repeat(1, t),
			want: ErrorEnd(stop),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Wrap(tt.fn); !sameEndStates(got, tt.want) {
				t.Errorf("Wrap() = %v, want %v", nameOf(got), nameOf(tt.want))
			}
		})
	}
}

func repeat(times int, t *testing.T) func(context.Context) error {
	cnt := 0
	tmp := times
	return func(ctx context.Context) error {
		cnt++
		times--
		if times > 0 {
			return nil
		}
		if cnt != tmp {
			t.Errorf("Repeated for different times %d, than expected %d", cnt, tmp)
		}
		return stop
	}
}

var stop = errors.New("stop")

func TestWrapRepeat(t *testing.T) {
	tests := []struct {
		name string
		fn   func(ctx context.Context) error
		want Fn
	}{
		{
			name: "nil",
		},
		{
			name: "one repeat",
			fn:   repeat(1, t),
			want: ErrorEnd(stop),
		},
		{
			name: "two repeats",
			fn:   repeat(2, t),
			want: ErrorEnd(stop),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WrapRepeat(tt.fn); !sameEndStates(got, tt.want) {
				t.Errorf("WrapRepeat() = %v, want %v", nameOf(got), nameOf(tt.want))
			}
		})
	}
}
