package pipeline

import (
	"context"
	"runtime"
	"sync"
)

// StreamToStreamHandler should implement the handle processing
// of certain data and sends it to output stream
// returns the outStream
type StreamToStreamHandler[T, RS any] func(context.Context, ReadOnlyStream[T]) ReadOnlyStream[RS]
type StreamsToStreamHandler[S ~[]ReadOnlyStream[T], T any] func(context.Context, S) ReadOnlyStream[T]
type StubFunc[T any] func(context.Context, ReadOnlyStream[T]) context.Context

func (s StubFunc[T]) Run(ctx context.Context, inStream ReadOnlyStream[T]) context.Context {
	return s(ctx, inStream)
}

func (ss StreamToStreamHandler[T, RS]) Run(ctx context.Context, inStream ReadOnlyStream[T]) ReadOnlyStream[RS] {
	return ss(ctx, inStream)
}
func (sss StreamsToStreamHandler[S, RS]) Run(ctx context.Context, inStreams S) ReadOnlyStream[RS] {
	return sss(ctx, inStreams)
}

// OrDoneFnFactory
// wrap our read from the channel with a select statement that
// also selects from a done context.
// Example:
//
//	for val := range OrDoneFnFactory[int]().Run(ctx, inStream) {
//	  ...Do something with val
//	}
func OrDoneFnFactory[T any]() StreamToStreamHandler[T, T] {
	return func(ctx context.Context, inStream ReadOnlyStream[T]) ReadOnlyStream[T] {
		outStream := make(chan T)
		go func() {
			defer close(outStream)
			for {
				select {
				case <-ctx.Done():
					return
				case vData, ok := <-inStream:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case outStream <- vData:
					}
				}
				runtime.Gosched()
			}
		}()
		return outStream
	}
}

// MergeFnFactory
// wrap our reading from multiple channels with a select statement that
// also selects from a context (ctx.Done()).
// sends all given data from channel list ([]ReadOnlyStream[T])
// to output channel only
func MergeFnFactory[S ~[]ReadOnlyStream[T], T any]() StreamsToStreamHandler[S, T] {

	return func(ctx context.Context, inStreams S) ReadOnlyStream[T] {
		outStream := make(chan T)
		var wg sync.WaitGroup
		wg.Add(len(inStreams))

		for _, inStream := range inStreams {
			go func(dataStream ReadOnlyStream[T]) {
				defer wg.Done()
				for vData := range OrDoneFnFactory[T]().Run(ctx, dataStream) {
					select {
					case <-ctx.Done():
						return
					case outStream <- vData:
					}
				}
			}(inStream)
		}

		go func() {
			defer close(outStream)
			wg.Wait()
		}()

		return outStream
	}
}

// FlatSlicesToStreamFnFactory flats slices []T given via inStream
// into flat data T and sends them one by one to the outStream
func FlatSlicesToStreamFnFactory[S ~[]T, T any]() StreamToStreamHandler[S, T] {
	return func(ctx context.Context, inStream ReadOnlyStream[S]) ReadOnlyStream[T] {
		outStream := make(chan T)
		go func() {
			defer close(outStream)
			for slicedData := range OrDoneFnFactory[S]().Run(ctx, inStream) {
				select {
				case <-ctx.Done():
					return
				default:
					for i := 0; i < len(slicedData); i += 1 {
						select {
						case <-ctx.Done():
							return
						case outStream <- slicedData[i]:
						}
					}
				}
			}
		}()
		return outStream
	}
}

func NoOpSubFnFactory[T any]() StubFunc[T] {
	return func(ctx context.Context, inStream ReadOnlyStream[T]) context.Context {
		childContext, closeCtx := context.WithCancel(ctx)
		go func() {
			defer closeCtx()
			for range OrDoneFnFactory[T]().Run(ctx, inStream) {
				// do nothing
				// just drain the input stream
			}
		}()
		return childContext
	}
}
