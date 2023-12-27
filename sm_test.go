package ssm

import (
	"context"
	"testing"
)

func TestRun(t *testing.T) {

	iter := func(_ context.Context) Fn {
		return End
	}

	tests := []struct {
		name   string
		states []Fn
	}{
		{
			name:   "just start",
			states: []Fn{iter},
		},
		{
			name:   "just start",
			states: []Fn{iter},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Run(context.Background(), tt.states...)
		})
	}
}

func TestBatch(t *testing.T) {
	tests := []struct {
		name         string
		states       []Fn
		wantEndState bool
	}{
		{
			name:   "nil",
			states: nil,
		},
		{
			name:   "empty",
			states: []Fn{},
		},
		{
			name:         "one fn",
			states:       []Fn{mockEmpty},
			wantEndState: true,
		},
		{
			name:         "with nil",
			states:       []Fn{nil},
			wantEndState: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Batch(tt.states...)
			if (got != nil) != tt.wantEndState {
				t.Errorf("Batch() wantEndState %t", tt.wantEndState)
			} else if got != nil {
				gotSt := got(context.Background())
				if gotSt != nil {
					t.Errorf("Batch()() final state = %v, expected nil", gotSt)
				}
			}
		})
	}
}
