package ssm

import (
	"context"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func mockSelf(_ context.Context) Fn {
	return mockSelf
}

func TestParallel(t *testing.T) {
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
			name:   "with explicit nil",
			states: []Fn{nil},
		},
		{
			name:   "one fn",
			states: []Fn{mockEmpty},
			want:   Parallel(mockEmpty),
		},
		{
			name:   "two mock fns",
			states: []Fn{mockEmpty, mockEmpty},
			want:   Parallel(mockEmpty, mockEmpty),
		},
		{
			name:   "one self fns",
			states: []Fn{mockSelf},
			want:   Parallel(mockSelf),
		},
		{
			name:   "two self fns",
			states: []Fn{mockSelf, mockSelf},
			want:   Parallel(mockSelf, mockSelf),
		},
		{
			name:   "with End",
			states: []Fn{End},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parallel(tt.states...)
			if !sameFns(got, tt.want) {
				t.Errorf("Parallel() returned end state %q, expected %q", nameOf(got), nameOf(tt.want))
			}
		})
	}
}

func sameFns(f1, f2 any) bool {
	p1 := reflect.ValueOf(f1).Pointer()
	p2 := reflect.ValueOf(f2).Pointer()
	if p1 == p2 {
		return true
	}
	if p1 == 0 || p2 == 0 {
		return false
	}
	s1, l1 := runtime.FuncForPC(p1).FileLine(p1)
	s2, l2 := runtime.FuncForPC(p2).FileLine(p2)
	return s1 == s2 && l1 == l2
}

func nameOf(f Fn) string {
	p := reflect.ValueOf(f).Pointer()
	if p == 0 {
		return "End"
	}
	name := filepath.Base(runtime.FuncForPC(p).Name())
	return name
}
