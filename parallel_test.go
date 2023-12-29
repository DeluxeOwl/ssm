package ssm

import (
	"context"
	"testing"
)

func TestParallel(t *testing.T) {
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
			got := Parallel(tt.states...)
			if (got != nil) != tt.wantEndState {
				t.Errorf("Parallel() wantEndState %t", tt.wantEndState)
			} else if got != nil {
				gotSt := got(context.Background())
				if gotSt != nil {
					t.Errorf("Batch()() final state = %v, expected nil", gotSt)
				}
			}
		})
	}
}
