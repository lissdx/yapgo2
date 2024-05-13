package pipeline

import (
	"context"
)

//type Runnable[T, RS any] interface {
//	Run(context.Context, ReadOnlyStream[T]) ReadOnlyStream[RS]
//}

//type ProcessResult struct {
//	Result interface{}
//	Error error
//}

//type ErrorHandler interface {
//	ErrorHandler(err error)
//}

//type DefaultErrorHandler struct {}
//func (DefaultErrorHandler)ErrorHandler(err error)  {
//	fmt.Printf("pipeline error: %s \n", err.Error())
//}

type ReadOnlyStream[T any] <-chan T
type WriteOnlyStream[T any] chan<- T
type BidirectionalStream[T any] chan T

// //type ReadOnlyErrorStream <-chan error
//

// StageFn is a lower level function type that chains together multiple
// stages using channels.
// In general meaning the code that is wrapping the ProcessFn
// It should get an input ReadOnly stream and returns output stream
// (means input stream for the next stage)
type StageFn[T, RS any] func(context.Context, ReadOnlyStream[T]) ReadOnlyStream[RS]

// ProcessFn is the primary function types defined by users of this
// package and passed in to instantiate a meaningful pipeline.
// In case of error is presented (err != nil)
// the event will be removed from pipeline
type ProcessFn[T, RS any] func(T) (RS, error)

// FilterFn is the primary function types defined by users of this
// package and passed in to instantiate a meaningful pipeline.
// Removes all elements from the pipeline for which returns true.
type FilterFn[T any] func(T) (T, bool)

type ProcessHandler[T, RS any] interface {
	ProcessHandlerFunc(T) (RS, error)
}

// Run every stage implements the Runnable interface
// just to stress the point of start running of the stage
func (sf StageFn[T, RS]) Run(ctx context.Context, inStream ReadOnlyStream[T]) ReadOnlyStream[RS] {
	return sf(ctx, inStream)
}

//type StagesFn[T, RS any] []StageFn[T, RS]
//
//func (s StageFn[T, RS]) Stage() {
//
//}

//type Stageable interface {
//	Stage()
//}

// // FilterFn are the primary function types defined by users of this
// // package and passed in to instantiate a meaningful pipeline.
// // It used by Filter stage, in case of success the object will be passed
// // to the next stage
// type FilterFn func(inObj interface{}) (outObj interface{}, success bool)
//
// // ErrorProcessFn are the primary function types defined by users of this package
// // Simple example may be:
// //func errorHandler(err error)  {
// //	fmt.Printf("pipeline error: %s \n", err.Error())
// //}
type ErrorProcessFn func(error)

// // Pipeline type defines a pipeline to which processing "stages" can
// // be added and configured to fan-out. Pipelines are meant to be long-running
// // as they continuously process data as it comes in.
// //
// // A pipeline can be simultaneously run multiple times with different
// // input channels by invoking the Run() method multiple times.
// // A running pipeline shouldn't be copied.
//type Pipeline []Stageable

//type Pipeline[T, RS any] StagesFn[T, any]

//
//// NewPipeline is a convenience method that creates a new Pipeline
//func NewPipeline() Pipeline {
//	return Pipeline{}
//}
//
// AddStage is a convenience method for adding a stage with fanSize = 1.
// See AddStageWithFanOut for more information.

//func (p *Pipeline) AddStage(inFunc ProcessFn[], errorProcessFn ErrorProcessFn) {
//	stgFactoryRes := ProcessStageFactory(inFunc, errorProcessFn)
//	*p = append(*p, ProcessStageFactory(inFunc, errorProcessFn))
//}

//func AddStageToPipeline[T, RS any](p *Pipeline, processFn ProcessFn[T, RS], errorProcessFn ErrorProcessFn) {
//	stgFactoryRes := ProcessStageFactory[T, RS](processFn, errorProcessFn)
//	*p = append(*p, stgFactoryRes)
//}

//func (p *Pipeline) Run(doneCh, inStream ReadOnlyStream) (outChan ReadOnlyStream) {
//
//	for _, stage := range *p {
//		inStream = stage(doneCh, inStream)
//	}
//
//	outChan = inStream
//	return
//}

// // AddFilterStage is a convenience method for adding a stage with fanSize = 1.
// // See AddFilterStageWithFanOut for more information.
//
//	func (p *Pipeline) AddFilterStage(inFunc FilterFn) {
//		*p = append(*p, filterStageFnFactory(inFunc))
//	}
//
// // AddStageWithFanOut adds a parallel fan-out ProcessFn to the pipeline. The
// // fanSize number indicates how many instances of this stage will read from the
// // previous stage and process the data flowing through simultaneously to take
// // advantage of parallel CPU scheduling.
// //
// // Most pipelines will have multiple stages, and the order in which AddStage()
// // and AddStageWithFanOut() is invoked matters -- the first invocation indicates
// // the first stage and so forth.
// //
// // Since discrete goroutines process the inChan for FanOut > 1, the order of
// // objects flowing through the FanOut stages can't be guaranteed.
//
//	func (p *Pipeline) AddStageWithFanOut(inFunc ProcessFn, errorProcessFn ErrorProcessFn, fanSize uint64) {
//		*p = append(*p, fanningStageFnFactory(inFunc, errorProcessFn, fanSize))
//	}
//
//	func (p *Pipeline) AddFilterStageWithFanOut(inFunc FilterFn, fanSize uint64) {
//		*p = append(*p, fanningFilterStageFnFactory(inFunc, fanSize))
//	}
//
// // fanningStageFnFactory makes a stage function that fans into multiple
// // goroutines increasing the stage throughput depending on the CPU.
//
//	func fanningStageFnFactory(inFunc ProcessFn, errorProcessFn ErrorProcessFn, fanSize uint64) (outFunc StageFn) {
//		return func(done ReadOnlyStream, inChan ReadOnlyStream) (outChan ReadOnlyStream) {
//			var channels []ReadOnlyStream
//			for i := uint64(0); i < fanSize; i++ {
//				//channels = append(channels, ProcessStageFactory(inFunc)(inChan))
//				channels = append(channels, ProcessStageFactory(inFunc, errorProcessFn)(done, inChan))
//			}
//			outChan = MergeChannels(done, channels)
//			return
//		}
//	}
//
// // fanningFilterStageFnFactory makes a stage function that fans into multiple
// // goroutines increasing the stage throughput depending on the CPU.
//
//	func fanningFilterStageFnFactory(inFunc FilterFn, fanSize uint64) (outFunc StageFn) {
//		return func(done ReadOnlyStream, inChan ReadOnlyStream) (outChan ReadOnlyStream) {
//			var channels []ReadOnlyStream
//			for i := uint64(0); i < fanSize; i++ {
//				//channels = append(channels, ProcessStageFactory(inFunc)(inChan))
//				channels = append(channels, filterStageFnFactory(inFunc)(done, inChan))
//			}
//			outChan = MergeChannels(done, channels)
//			return
//		}
//	}
//

// ProcessStageFactory makes a standard stage function. (wraps a given ProcessFn).
// StageFn functions types accepts an inStream and returns an outStream
// (to chain multiple stages into a pipeline)
func ProcessStageFactory[T, RS any](processFunc ProcessFn[T, RS], errorProcessFn ErrorProcessFn) StageFn[T, RS] {
	return func(ctx context.Context, inStream ReadOnlyStream[T]) ReadOnlyStream[RS] {
		outStream := make(chan RS)
		go func() {
			defer close(outStream)

			// NOTE in general we may ignore the additional
			// context creating
			//orCtx, orCancel := context.WithCancel(ctx)
			//localCtx, localCancel := context.WithCancel(orCtx)
			//defer orCancel()
			//defer localCancel()

			for inObj := range OrDoneFnFactory[T]().Run(ctx, inStream) {
				outObj, err := processFunc(inObj)
				if err != nil {
					errorProcessFn(err)
					continue
				}
				select {
				case <-ctx.Done():
					return
				case outStream <- outObj:
				}
			}
		}()
		return outStream
	}
}

// ProcessStageFactoryWithFanOut makes the FanOut stage. (wraps and repeats given ProcessFn fanSize times).
// Will establish StageFn by fanSize, the result of processing will be merged into
// a single output stream
// (to chain multiple stages into a pipeline)
func ProcessStageFactoryWithFanOut[T, RS any](processFunc ProcessFn[T, RS], errorProcessFn ErrorProcessFn, fanSize uint64) StageFn[T, RS] {
	return func(ctx context.Context, inStream ReadOnlyStream[T]) ReadOnlyStream[RS] {
		outStreams := make([]ReadOnlyStream[RS], 0, fanSize)
		for i := uint64(0); i < fanSize; i++ {
			outStreams = append(outStreams, ProcessStageFactory[T, RS](processFunc, errorProcessFn).Run(ctx, inStream))
		}

		return MergeFnFactory[[]ReadOnlyStream[RS]]().Run(ctx, outStreams)
	}
}

// FilterStageFactory wraps a given FilterFn.
// Returns StageFn functions types accepts an inStream and returns an outStream
// (to chain multiple stages into a pipeline)
// In case the FilterFn[T] will return true, the stage will filter it
func FilterStageFactory[T any](filterFn FilterFn[T]) StageFn[T, T] {
	return func(ctx context.Context, inStream ReadOnlyStream[T]) ReadOnlyStream[T] {
		outStream := make(chan T)
		go func() {
			defer close(outStream)

			for inObj := range OrDoneFnFactory[T]().Run(ctx, inStream) {
				outObj, filterIt := filterFn(inObj)
				if filterIt {
					continue
				}
				select {
				case <-ctx.Done():
					return
				case outStream <- outObj:
				}
			}

		}()
		return outStream
	}
}

// ProcessFilterStageFactoryWithFanOut makes the FanOut FilterStage. (wraps and repeats given FilterFn fanSize times).
// Will establish StageFn by fanSize, the result of processing will be merged into
// a single output stream
// (to chain multiple stages into a pipeline)
func ProcessFilterStageFactoryWithFanOut[T any](filterFn FilterFn[T], fanSize uint64) StageFn[T, T] {
	return func(ctx context.Context, inStream ReadOnlyStream[T]) ReadOnlyStream[T] {
		outStreams := make([]ReadOnlyStream[T], 0, fanSize)
		for i := uint64(0); i < fanSize; i++ {
			outStreams = append(outStreams, FilterStageFactory[T](filterFn).Run(ctx, inStream))
		}
		return MergeFnFactory[[]ReadOnlyStream[T]]().Run(ctx, outStreams)
	}
}

//
//// filterStageFnFactory makes a standard stage function from a given FilterFn.
//func filterStageFnFactory(inFunc FilterFn) (outFunc StageFn) {
//	return func(doneCh, inChan ReadOnlyStream) ReadOnlyStream {
//		outChan := make(chan interface{})
//		go func() {
//			defer close(outChan)
//			for inObj := range OrDone(doneCh, inChan) {
//				outObj, success := inFunc(inObj)
//				if !success {
//					continue
//				}
//				select {
//				case <-doneCh:
//					return
//				case outChan <- outObj:
//				}
//			}
//		}()
//		return outChan
//	}
//}

//// Run starts the pipeline with all the stages that have been added. and returns
//// the result channel with 'pipelined' data
//// We be able to use the channel to connect another pipeline or use the pipelined data
//func (p *Pipeline) Run(doneCh, inStream ReadOnlyStream) (outChan ReadOnlyStream) {
//
//	for _, stage := range *p {
//		inStream = stage(doneCh, inStream)
//	}
//
//	outChan = inStream
//	return
//}
//
//// RunPlug starts the pipeline with all the stages that have been added. RunPlug is not
//// a blocking function and will return immediately with a doneChan. Consumers
//// can wait on the doneChan for an indication of when the pipeline has completed
//// processing.
////
//// The pipeline runs until its `inStream` channel is open. Once the `inStream` is closed,
//// the pipeline stages will sequentially complete from the first stage to the last.
//// Once all stages are complete, the last outChan is drained and the doneChan is closed.
////
//// RunPlug() can be invoked multiple times to start multiple instances of a pipeline
//// that will typically process different incoming channels.
//func (p *Pipeline) RunPlug(doneCh, inStream ReadOnlyStream) ReadOnlyStream {
//
//	for _, stage := range *p {
//		inStream = stage(doneCh, inStream)
//	}
//
//	doneChan := make(chan interface{})
//	go func() {
//		defer close(doneChan)
//		for range OrDone(doneCh, inStream) {
//			// pull objects from inChan so that the gc marks them
//		}
//	}()
//
//	return doneChan
//}
