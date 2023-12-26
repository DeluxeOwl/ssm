package ssm

import (
	"context"
	"testing"
)

func TestRun(t *testing.T) {

	iter := func(_ context.Context) (Fn, error) {
		return End, nil
	}

	tests := []struct {
		name    string
		states  []Fn
		wantErr bool
	}{
		{
			name:    "just start",
			states:  []Fn{iter},
			wantErr: false,
		},
		{
			name:    "just start",
			states:  []Fn{iter},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Run(context.Background(), tt.states...); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
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
			states:       []Fn{func(_ context.Context) (Fn, error) { return End, nil }},
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
				gotSt, err := got(context.Background())
				if err != nil {
					t.Errorf("Batch()() error = %v, expected nil", err)
				}
				if gotSt != nil {
					t.Errorf("Batch()() final state = %v, expected nil", gotSt)
				}
			}
		})
	}
}
