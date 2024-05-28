package pipeline

import (
	"context"
	"fmt"
	cnfStreamHandler "github.com/lissdx/yapgo2/pkg/pipeline/config/stream_handler"
	"golang.org/x/sync/errgroup"
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
//	for val := range OrDoneFnFactory[int]().GenerateToStream(ctx, inStream) {
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
				//runtime.Gosched()
			}
		}()
		return outStream
	}
}

// MergeFnFactory
// wrap our reading from multiple channels with a select statement that
// also selects from a context (ctx.Done()).
// sends all given data from channel list ([]ReadOnlyStream[IN])
// to output channel only
func MergeFnFactory[S ~[]ReadOnlyStream[T], T any](options ...cnfStreamHandler.Option) StreamsToStreamHandler[S, T] {

	conf := cnfStreamHandler.NewStreamHandlerConfig(options...)
	loggerPref := fmt.Sprintf("MergeFnFactory_%s", conf.Name())

	return func(ctx context.Context, inStreams S) ReadOnlyStream[T] {
		outStream := make(chan T)

		g, ctx := errgroup.WithContext(ctx)

		for i, inStream := range inStreams {
			lInStream := inStream
			indx := i
			g.Go(func() error {

				logPref := fmt.Sprintf("%s_%d", loggerPref, indx)
				conf.Logger().Info("%s_%d: MergeFnFactory started", logPref, indx)

			exit:
				for {
					select {
					case <-ctx.Done():
						conf.Logger().Debug("%s: Got <-ctx.Done(). MergeFnFactory", logPref)
						return ctx.Err()
					case vData, ok := <-lInStream:
						if !ok {
							break exit
						}
						select {
						case <-ctx.Done():
							return fmt.Errorf("%s: MergeFnFactory was interrupted. Data was fetched but got <-ctx.Done(). context error: %w",
								logPref, ctx.Err())
						case outStream <- vData:
						}
					}
				}

				conf.Logger().Debug("%s: MergeFnFactory input stream closed", logPref)
				return nil
			})
		}

		go func() {
			defer close(outStream)
			if err := g.Wait(); err != nil {
				conf.Logger().Error("%s: MergeFnFactory error: %s", loggerPref, err.Error())
			}
			conf.Logger().Info("%s: MergeFnFactory all processes was stopped", loggerPref)
			conf.Logger().Info("%s: MergeFnFactory close outStream", loggerPref)
		}()

		return outStream
	}
}

// FlatSlicesToStreamFnFactory flats slices []IN given via inStream
// into flat data IN and sends them one by one to the outStream
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
