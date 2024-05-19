package ssm

import (
	"context"
	"testing"
	"time"
)

var timedState = After(defaultDelay, mockEmpty)

var expectedNonBlockingWaitState = new(nb).wait

func TestNonBlocking(t *testing.T) {
	tests := []struct {
		name   string
		states []Fn
		want   []Fn
	}{
		{
			name:   "nil",
			states: nil,
			want:   []Fn{expectedNonBlockingWaitState, End},
		},
		{
			name:   "mock empty",
			states: []Fn{mockEmpty},
			want: []Fn{
				expectedNonBlockingWaitState, End,
			},
		},
		{
			name:   "timed return mock empty",
			states: []Fn{timedState},
			want: []Fn{
				// 10 times wait, then mockEmpty, then end state
				expectedNonBlockingWaitState, expectedNonBlockingWaitState, expectedNonBlockingWaitState,
				expectedNonBlockingWaitState, expectedNonBlockingWaitState, expectedNonBlockingWaitState,
				expectedNonBlockingWaitState, expectedNonBlockingWaitState, expectedNonBlockingWaitState,
				expectedNonBlockingWaitState, End,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NonBlocking(tt.states...)
			testLoopStates(t, got, tt.want...)
		})
	}
}

func testLoopStates(t *testing.T, start Fn, expected ...Fn) {
	if IsEnd(start) {
		return
	}

	ctx := context.Background()

	for i, exp := range expected {
		next := start(ctx)
		if !sameFns(next, exp) {
			t.Errorf("Invalid state at iteration %d = %v, want %v", i, nameOf(next), nameOf(exp))
		}
		if IsEnd(next) {
			break
		}
		start = next
		time.Sleep(defaultDelay / 10)
	}
}
