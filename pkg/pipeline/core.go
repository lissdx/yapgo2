package pipeline

import "context"

type ReadOnlyStream[T any] <-chan T
type WriteOnlyStream[T any] chan<- T
type BidirectionalStream[T any] chan T

type Runnable[T, RS any] interface {
	Run(context.Context, T) RS
}

type ProcessHandlerInterface[T, RS any] interface {
	ProcessHandler(T) (RS, error)
}

type ProcessResultCarrier[T any] interface {
	ProcessResult() T
	IsOmitted() bool
}

type processResult[T any] struct {
	data    T
	omitted bool
}

func (p *processResult[T]) ProcessResult() T {
	return p.data
}

func (p *processResult[T]) IsOmitted() bool {
	return p.omitted
}

func NewOmittedProcessResult[T any]() ProcessResultCarrier[T] {
	return &processResult[T]{omitted: true}
}

func NewProcessResult[T any](data T, omitted bool) ProcessResultCarrier[T] {
	return &processResult[T]{data: data, omitted: omitted}
}
