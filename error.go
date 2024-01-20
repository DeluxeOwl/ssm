package ssm

import (
	"context"
	"reflect"
)

// ErrorEnd represents an error state which returns an End state.
func ErrorEnd(err error) Fn {
	return errState{err}.runStop
}

// ErrorRestart represents an error state which returns the first iteration passed.
// This iteration is loaded from the context, and is saved there by the Run and RunParallel functions.
func ErrorRestart(err error) Fn {
	return errState{err}.runRestart
}

// StartState retrieves the initial state from ctx context.Context.
// If nothing is found it returns the End state.
func StartState(ctx context.Context) Fn {
	start := ctx.Value(__start)
	if start == nil {
		return End
	}
	if state, ok := start.(func(context.Context) Fn); ok {
		return state
	}
	return End
}

type smKeys string

const __start smKeys = "__start"

// Error this state
func (e errState) Error() string {
	return e.error.Error()
}

type errState struct {
	error
}

type ErrorFn Fn

func (f ErrorFn) Error() string {
	return "error"
}

func ptrOf(t Fn) uintptr {
	return reflect.ValueOf(t).Pointer()
}

var (
	_ptrEndStop    = ptrOf(errState{}.runStop)
	_ptrEndRestart = ptrOf(errState{}.runRestart)
)

// IsError ver grubby API to check if a state Fn is an error state
func IsError(f Fn) bool {
	p := ptrOf(f)
	return p == _ptrEndStop || p == _ptrEndRestart
}

func (e errState) runStop(_ context.Context) Fn {
	return End
}

func (e errState) runRestart(ctx context.Context) Fn {
	return StartState(ctx)
}
