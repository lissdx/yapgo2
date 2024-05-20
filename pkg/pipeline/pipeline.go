package pipeline

import (
	"context"
)

// ProcessFn is the primary function types defined by users of this
// package and passed in to instantiate a meaningful pipeline.
// In general, it is the data processing function
// In general: if an error is presented (err != nil)
// the event will be removed from pipeline
// We are able to overwrite it. See: ProcessHandlerFactory
type ProcessFn[T, RS any] func(T) (RS, error)

// FilterFn is the primary function types defined by users of this
// package and passed in to instantiate a meaningful pipeline.
// Removes all elements from the pipeline for which returns true.
type FilterFn[T any] func(T) (T, bool)

// ProcessHandlerFn - an orchestrator for the ProcessFn / FilterFn
// should be used in the pipeline stage
// The bool param
// is able to ignore the processed event
// returns:
// RS - result data
// bool - ignore the data
// if true means the RS should be ignored
type ProcessHandlerFn[T, RS any] func(T) (RS, bool)

// StageFn is a lower level function type that chains together multiple
// stages using channels.
// It calls to the ProcessFn / FilterFn via ProcessHandler interface
// It should get an input ReadOnly stream and returns output stream
// (means input stream for the next stage)
type StageFn[T, RS any] func(context.Context, ReadOnlyStream[T]) ReadOnlyStream[RS]

// Run every stage implements the Runnable interface
// just to stress the point of start running of the stage
func (sf StageFn[T, RS]) Run(ctx context.Context, inStream ReadOnlyStream[T]) ReadOnlyStream[RS] {
	return sf(ctx, inStream)
}

// ProcessHandler interface
// implemented by ProcessHandlerFn and FilterFn
type ProcessHandler[T, RS any] interface {
	HandleIt(T) (RS, bool)
}

func (phf ProcessHandlerFn[T, RS]) HandleIt(t T) (RS, bool) {
	return phf(t)
}

func (phf FilterFn[T]) HandleIt(t T) (T, bool) {
	return phf(t)
}

// ProcessStageFactory makes a standard stage function. (wraps a given ProcessFn).
// StageFn functions types accepts an inStream and returns an outStream
// (to chain multiple stages into a pipeline)
func ProcessStageFactory[T, RS any](processFunc ProcessHandler[T, RS]) StageFn[T, RS] {
	return func(ctx context.Context, inStream ReadOnlyStream[T]) ReadOnlyStream[RS] {
		outStream := make(chan RS)
		go func() {
			defer close(outStream)
			for inObj := range OrDoneFnFactory[T]().Run(ctx, inStream) {
				{
					select {
					case <-ctx.Done():
						return
					default:
						outObj, ignoreIt := processFunc.HandleIt(inObj)
						if ignoreIt {
							continue
						}
						select {
						case <-ctx.Done():
							return
						case outStream <- outObj:
						}
					}
				}

			}
		}()
		return outStream
	}
}

// ProcessStageFactory makes a standard stage function. (wraps a given ProcessFn).
// StageFn functions types accepts an inStream and returns an outStream
// (to chain multiple stages into a pipeline)
func ProcessStageFactoryTest[T, RS any](processFunc ProcessHandler[T, RS]) StageFn[T, RS] {
	return func(ctx context.Context, inStream ReadOnlyStream[T]) ReadOnlyStream[RS] {
		outStream := make(chan RS)
		go func() {
			defer close(outStream)
			for inObj := range OrDoneFnFactory[T]().Run(ctx, inStream) {
				{
					select {
					case <-ctx.Done():
						return
					default:
						//genFunc := func() ReadOnlyStream[RS] {
						//	tempOutStream := make(chan RS)
						//
						//	return func() RS {
						//		outObj, ignoreIt := processFunc.HandleIt(inObj)
						//	}
						//	return tempOutStream
						//}
						//GeneratorHandlerFactory()
						outObj, ignoreIt := processFunc.HandleIt(inObj)
						if ignoreIt {
							continue
						}
						select {
						case <-ctx.Done():
							return
						case outStream <- outObj:
						}
					}
				}

			}
		}()
		return outStream
	}
}

// ProcessStageFactoryWithFanOut creates the ProcessStageFactory in FanOut mode
func ProcessStageFactoryWithFanOut[T, RS any](processFunc ProcessHandler[T, RS], fanSize uint64) StageFn[T, RS] {
	return func(ctx context.Context, inStream ReadOnlyStream[T]) ReadOnlyStream[RS] {
		outStreams := make([]ReadOnlyStream[RS], 0, fanSize)
		for i := uint64(0); i < fanSize; i++ {
			outStreams = append(outStreams, ProcessStageFactory[T, RS](processFunc).Run(ctx, inStream))
		}

		return MergeFnFactory[[]ReadOnlyStream[RS]]().Run(ctx, outStreams)
	}
}

// ProcessHandlerFactory wraps ProcessFn[T, RS] and adapt it to ProcessHandler[T, RS] interface
func ProcessHandlerFactory[T, RS any](processFn ProcessFn[T, RS], options ...Option) ProcessHandler[T, RS] {
	config := yapgo2Config{}
	for _, option := range options {
		option.apply(&config)
	}

	if config.onErrorStrategy == OnErrorContinue {
		return func() ProcessHandlerFn[T, RS] {
			return func(t T) (RS, bool) {
				res, _ := processFn(t)
				return res, false
			}
		}()
	}

	return func() ProcessHandlerFn[T, RS] {
		return func(t T) (RS, bool) {
			res, err := processFn(t)
			return res, err != nil
		}
	}()

}
