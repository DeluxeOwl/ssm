package ssm

import "context"

// EndError represents an error state which returns an End state.
func EndError(err error) Fn {
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
		return Fn(state)
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

func (e errState) runStop(_ context.Context) Fn {
	return End
}

func (e errState) runRestart(ctx context.Context) Fn {
	return StartState(ctx)
}
