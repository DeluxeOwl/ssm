package ssm

import (
	"testing"
)

func TestBatch(t *testing.T) {
	tests := []struct {
		name   string
		states []Fn
		want   Fn
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
			name:   "one fn",
			states: []Fn{mockEmpty},
			want:   Batch(mockEmpty),
		},
		{
			name:   "with explicit nil",
			states: []Fn{nil},
		},
		{
			name:   "two mock fns",
			states: []Fn{mockEmpty, mockEmpty},
			want:   Batch(mockEmpty, mockEmpty),
		},
		{
			name:   "one self fns",
			states: []Fn{mockSelf},
			want:   Batch(mockSelf),
		},
		{
			name:   "two self fns",
			states: []Fn{mockSelf, mockSelf},
			want:   Batch(mockSelf, mockSelf),
		},
		{
			name:   "with End",
			states: []Fn{End},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Batch(tt.states...)
			if !sameFns(got, tt.want) {
				t.Errorf("Batch() returned end state %q, expected %q", nameOf(got), nameOf(tt.want))
			}
		})
	}
}
