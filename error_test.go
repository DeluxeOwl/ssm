package ssm

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func mockEmpty(_ context.Context) Fn {
	return End
}

func ptrOf(t Fn) int64 {
	return int64(reflect.ValueOf(t).Pointer())
}

func TestStartState(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want Fn
	}{
		{
			name: "empty",
			ctx:  context.Background(),
			want: nil,
		},
		{
			name: "some func",
			ctx:  context.WithValue(context.Background(), __start, mockEmpty),
			want: mockEmpty,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StartState(tt.ctx); ptrOf(got) != ptrOf(tt.want) {
				t.Errorf("StartState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Fn
	}{
		{
			name: "empty",
			err:  nil,
			want: errState{}.runStop,
		},
		{
			name: "random err",
			err:  fmt.Errorf("test"),
			want: errState{fmt.Errorf("test")}.runStop,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EndError(tt.err)
			if ptrOf(got) != ptrOf(tt.want) {
				t.Errorf("EndError() = %v, want %v", got, tt.want)
			}
			if st := got(context.Background()); ptrOf(st) != ptrOf(End) {
				t.Errorf("Post run state for EndError() = %v, want %v", st, End)
			}
		})
	}
}

func TestErrorRestart(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		ctx      context.Context
		want     Fn
		endState Fn
	}{
		{
			name: "empty",
			err:  nil,
			want: errState{}.runRestart,
		},
		{
			name: "random err, background context",
			err:  fmt.Errorf("test"),
			want: errState{fmt.Errorf("test")}.runRestart,
		},
		{
			name:     "random err, start state in context",
			err:      fmt.Errorf("test"),
			ctx:      context.WithValue(context.Background(), __start, mockEmpty),
			want:     errState{fmt.Errorf("test")}.runRestart,
			endState: mockEmpty,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorRestart(tt.err)
			if ptrOf(got) != ptrOf(tt.want) {
				t.Errorf("ErrorRestart() = %v, want %v", got, tt.want)
			}
			if st := got(context.Background()); ptrOf(st) != ptrOf(End) {
				t.Errorf("Post run state for ErrorRestart() = %v, want %v", st, End)
			}
		})
	}
}
