package pipeline

import (
	"context"
	"errors"
	"fmt"
	confProcessHandler "github.com/lissdx/yapgo2/pkg/pipeline/config/process/config_handler"
	cnfProcStage "github.com/lissdx/yapgo2/pkg/pipeline/config/process/config_stage"
	"golang.org/x/sync/errgroup"
)

// ProcessFn is the primary function types defined by users of this
// package and passed in to instantiate a meaningful pipeline.
// It is the data processing function
// In general: if an error is presented (err != nil)
// the event will be removed from pipeline
// We are able to overwrite it with onErrorStrategy. See: ProcessHandlerFactory
type ProcessFn[T, RS any] func(T) (RS, error)

// FilterFn is the primary function types defined by users of this
// package and passed in to instantiate a meaningful pipeline.
// Returns true if the event should be removed from the pipeline
type FilterFn[T any] func(T) (T, bool)

// ProcessHandlerFn - a wrapper for the result of
// ProcessFn / FilterFn
// should be used in the pipeline stage
// Returns the object implements ProcessResultCarrier
// interface
// IsOmitted() true
// means that the object should be ignored in
// a pipeline
type ProcessHandlerFn[T, RS any] func(T) ProcessResultCarrier[RS]

// StageFn is a lower level function type that chains together multiple
// stages using channels.
// It calls to the ProcessFn / FilterFn via ProcessHandler interface
// It should get an input ReadOnly stream and returns output stream
// (means input stream for the next stage)
type StageFn[IN, RS any] func(context.Context, ReadOnlyStream[IN]) ReadOnlyStream[RS]

// Run every stage implements the Runnable interface
// just to stress the point of start running of the stage
func (sf StageFn[IN, RS]) Run(ctx context.Context, inStream ReadOnlyStream[IN]) ReadOnlyStream[RS] {
	return sf(ctx, inStream)
}

// ProcessHandler interface
// implemented by ProcessHandlerFn and FilterFn
type ProcessHandler[T, RS any] interface {
	Apply(T) ProcessResultCarrier[RS]
}

func (phf ProcessHandlerFn[T, RS]) Apply(t T) ProcessResultCarrier[RS] {
	return phf(t)
}

// ProcessStageFactoryWithFanOut creates the ProcessStageFactory in FanOut mode
func ProcessStageFactoryWithFanOut[T, RS any](processFunc ProcessHandler[T, RS], fanSize uint64, options ...cnfProcStage.Option) StageFn[T, RS] {
	config := cnfProcStage.NewProcessStageConfig(options...)
	loggerPref := fmt.Sprintf("%s", config.Name())

	return func(ctx context.Context, inStream ReadOnlyStream[T]) ReadOnlyStream[RS] {
		outStreams := make([]ReadOnlyStream[RS], 0, fanSize)
		config.Logger().Info("%s: ProcessStageFactoryWithFanOut going to start. Given FanOut factor: %d", loggerPref, fanSize)
		for i := uint64(0); i < fanSize; i++ {
			processStageName := fmt.Sprintf("%s_%d", loggerPref, i)
			opt := append(options, cnfProcStage.WithName(processStageName))
			outStreams = append(outStreams, ProcessStageFactory[T, RS](processFunc, opt...).Run(ctx, inStream))
		}

		config.Logger().Info("%s: ProcessStageFactoryWithFanOut created %d ProcessStages.", loggerPref, fanSize)
		return MergeFnFactory[[]ReadOnlyStream[RS]]().Run(ctx, outStreams)
	}
}

// ProcessStageFactory makes a standard stage function. (wraps a given ProcessHandler).
// StageFn functions types accepts an inStream and returns an outStream
// (to chain multiple stages into a pipeline)
func ProcessStageFactory[IN, RS any](processHandler ProcessHandler[IN, RS], options ...cnfProcStage.Option) StageFn[IN, RS] {

	config := cnfProcStage.NewProcessStageConfig(options...)
	loggerPref := fmt.Sprintf("%s", config.Name())

	return func(ctx context.Context, inStream ReadOnlyStream[IN]) ReadOnlyStream[RS] {
		outStream := make(chan RS)
		g, ctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			config.Logger().Info("%s: ProcessStage started...", loggerPref)
			for {
				select {
				case <-ctx.Done():
					config.Logger().Debug("%s: Got <-ctx.Done(). ProcessStage was interrupted", loggerPref)
					return ctx.Err()
				case vData, ok := <-inStream:
					// if input stream closed
					if !ok {
						config.Logger().Debug("%s: ProcessStage the input stream was closed", loggerPref)
						return nil
					}
					processResultCarrier := processHandler.Apply(vData)
					if processResultCarrier.IsOmitted() {
						config.Logger().Trace("%s: ProcessStage process result object is Omitted. obj: %+v", loggerPref, processResultCarrier)
						continue
					}
					select {
					case <-ctx.Done():
						if config.Logger().TraceIsEnabled() {
							config.Logger().Trace("%s: Got <-ctx.Done(). ProcessStage was interrupted, but result was provided: %+v", loggerPref, processResultCarrier.ProcessResult())
						} else if config.Logger().DebugIsEnabled() {
							config.Logger().Debug("%s: Got <-ctx.Done(). ProcessStage was interrupted, but result was provided", loggerPref)
						}
						return ctx.Err()
					case outStream <- processResultCarrier.ProcessResult():
						config.Logger().Trace("%s: ProcessStage was sent to the next stage: %+v", loggerPref, processResultCarrier.ProcessResult())
					}
				}
			}
		})

		go func() {
			defer close(outStream)
			if err := g.Wait(); err != nil {
				if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
					config.Logger().Error("%s: ProcessStage error: %s", loggerPref, err.Error())
				} else {
					config.Logger().Info("%s: ProcessStage interruption cause: %s", loggerPref, err.Error())
				}
			}
			config.Logger().Info("%s: ProcessStage stopped", loggerPref)
		}()

		return outStream
	}
}

// ProcessHandlerFactory wraps ProcessFn[IN, RS] and adapt it to ProcessHandler[IN, RS] interface
func ProcessHandlerFactory[IN, RS any](processFn ProcessFn[IN, RS], options ...confProcessHandler.Option) ProcessHandler[IN, RS] {
	config := confProcessHandler.NewProcessHandlerConfig(options...)

	if config.IsOnErrorContinueStrategy() {
		return func() ProcessHandlerFn[IN, RS] {
			return func(t IN) ProcessResultCarrier[RS] {
				res, _ := processFn(t)
				return NewProcessResult(res, false)
			}
		}()
	}

	return func() ProcessHandlerFn[IN, RS] {
		return func(t IN) ProcessResultCarrier[RS] {
			res, err := processFn(t)
			return NewProcessResult(res, err != nil)
		}
	}()
}

func ProcessHandlerFilterFactory[IN any](filter FilterFn[IN]) ProcessHandler[IN, IN] {

	return func() ProcessHandlerFn[IN, IN] {
		return func(t IN) ProcessResultCarrier[IN] {
			res, isOmitted := filter(t)
			return NewProcessResult(res, isOmitted)
		}
	}()

}
