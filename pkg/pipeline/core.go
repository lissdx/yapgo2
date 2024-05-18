package pipeline

import "context"

type ReadOnlyStream[T any] <-chan T
type WriteOnlyStream[T any] chan<- T
type BidirectionalStream[T any] chan T

type Runnable[T, RS any] interface {
	Run(context.Context, T) RS
}

type ProcessHandlerInterface[T, RS any] interface {
	// ProcessHandler ServerHTTP Response
	ProcessHandler(T) (RS, error)
}
