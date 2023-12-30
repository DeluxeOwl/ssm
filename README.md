# Simple State Machine

This is a very basic API for creating State Machines.

A state is represented by any function which accepts a context as a parameter and returns another state function.

The states are run iteratively until the `End` state is reached, when the state machine stops.

Error handling is provided through two `Error` states:
* The simple `Error` state which stops the execution
* The `RestartError` state which restarts the execution from the first received state.

```go

type Fn func(ctx context.Context) Fn

```

To start the state machine you can pass a state to one of the `Run` or `RunParallel` functions.
The later runs the received states in parallel and the former does so sequentially.

