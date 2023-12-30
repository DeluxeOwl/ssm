package ssm

import (
	"context"
	"testing"
	"time"
)

const defaultDelay = 100 * time.Millisecond

var defaultTime = time.Now().Add(defaultDelay)

func timedState(ctx context.Context) Fn {
	return mockEmpty
}

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
			want:   []Fn{expectedNonBlockingWaitState, End},
		},
		{
			name:   "timed return mock empty",
			states: []Fn{timedState},
			want:   []Fn{expectedNonBlockingWaitState, mockEmpty, End},
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
	if start == nil {
		return
	}

	ctx := context.Background()

	for i, exp := range expected {
		next := start(ctx)
		if ptrOf(next) != ptrOf(exp) {
			t.Errorf("Invalid state at iteration %d = %v, want %v", i, next, exp)
		}
		if next == nil {
			break
		}
		start = next
		time.Sleep(defaultDelay / 10)
	}
}
