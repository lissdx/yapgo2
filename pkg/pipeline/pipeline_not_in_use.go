package pipeline

//
//import (
//	"context"
//)
//
//// ProcessFn is the primary function types defined by users of this
//// package and passed in to instantiate a meaningful pipeline.
//// In general, it is the data processing function
//// In general: if an error is presented (err != nil)
//// the event will be removed from pipeline
//// We are able to overwrite it. See: ProcessHandlerFactory
//type ProcessFn[IN, RS any] func(IN) (RS, error)
//
//// FilterFn is the primary function types defined by users of this
//// package and passed in to instantiate a meaningful pipeline.
//// Removes all elements from the pipeline for which returns true.
//type FilterFn[IN any] func(IN) (IN, bool)
//
//// ProcessHandlerFn - an orchestrator for the ProcessFn / FilterFn
//// should be used in the pipeline stage
//// The bool param
//// is able to ignore the processed event
//// returns:
//// RS - result data
//// bool - ignore the data
//// if true means the RS should be ignored
//type ProcessHandlerFn[IN, RS any] func(IN) (RS, bool)
//
//// StageFn is a lower level function type that chains together multiple
//// stages using channels.
//// It calls to the ProcessFn / FilterFn via ProcessHandler interface
//// It should get an input ReadOnly stream and returns output stream
//// (means input stream for the next stage)
//type StageFn[IN, RS any] func(context.Context, ReadOnlyStream[IN]) ReadOnlyStream[RS]
//
//// GenerateToStream every stage implements the Runnable interface
//// just to stress the point of start running of the stage
//func (sf StageFn[IN, RS]) GenerateToStream(ctx context.Context, inStream ReadOnlyStream[IN]) ReadOnlyStream[RS] {
//	return sf(ctx, inStream)
//}
//
//// ProcessHandler interface
//// implemented by ProcessHandlerFn and FilterFn
//type ProcessHandler[IN, RS any] interface {
//	Apply(IN) (RS, bool)
//}
//
//func (phf ProcessHandlerFn[IN, RS]) Apply(t IN) (RS, bool) {
//	return phf(t)
//}
//
//func (phf FilterFn[IN]) Apply(t IN) (IN, bool) {
//	return phf(t)
//}
//
//// ProcessStageFactory makes a standard stage function. (wraps a given ProcessFn).
//// StageFn functions types accepts an inStream and returns an outStream
//// (to chain multiple stages into a pipeline)
//func ProcessStageFactory[IN, RS any](processFunc ProcessHandler[IN, RS]) StageFn[IN, RS] {
//	return func(ctx context.Context, inStream ReadOnlyStream[IN]) ReadOnlyStream[RS] {
//		outStream := make(chan RS)
//		go func() {
//			defer close(outStream)
//			for inObj := range OrDoneFnFactory[IN]().GenerateToStream(ctx, inStream) {
//				{
//					select {
//					case <-ctx.Done():
//						return
//					default:
//						outObj, ignoreIt := processFunc.Apply(inObj)
//						if ignoreIt {
//							continue
//						}
//						select {
//						case <-ctx.Done():
//							return
//						case outStream <- outObj:
//						}
//					}
//				}
//
//			}
//		}()
//		return outStream
//	}
//}
//
//// ProcessStageFactory makes a standard stage function. (wraps a given ProcessFn).
//// StageFn functions types accepts an inStream and returns an outStream
//// (to chain multiple stages into a pipeline)
//func ProcessStageFactoryTest[IN, RS any](processFunc ProcessHandler[IN, RS]) StageFn[IN, RS] {
//	return func(ctx context.Context, inStream ReadOnlyStream[IN]) ReadOnlyStream[RS] {
//		outStream := make(chan RS)
//		go func() {
//			defer close(outStream)
//			for inObj := range OrDoneFnFactory[IN]().GenerateToStream(ctx, inStream) {
//				{
//					select {
//					case <-ctx.Done():
//						return
//					default:
//						//genFunc := func() ReadOnlyStream[RS] {
//						//	tempOutStream := make(chan RS)
//						//
//						//	return func() RS {
//						//		outObj, ignoreIt := processFunc.Apply(inObj)
//						//	}
//						//	return tempOutStream
//						//}
//						//GeneratorHandlerFactory()
//						outObj, ignoreIt := processFunc.Apply(inObj)
//						if ignoreIt {
//							continue
//						}
//						select {
//						case <-ctx.Done():
//							return
//						case outStream <- outObj:
//						}
//					}
//				}
//
//			}
//		}()
//		return outStream
//	}
//}
//
//// ProcessStageFactoryWithFanOut creates the ProcessStageFactory in FanOut mode
//func ProcessStageFactoryWithFanOut[IN, RS any](processFunc ProcessHandler[IN, RS], fanSize uint64) StageFn[IN, RS] {
//	return func(ctx context.Context, inStream ReadOnlyStream[IN]) ReadOnlyStream[RS] {
//		outStreams := make([]ReadOnlyStream[RS], 0, fanSize)
//		for i := uint64(0); i < fanSize; i++ {
//			outStreams = append(outStreams, ProcessStageFactory[IN, RS](processFunc).GenerateToStream(ctx, inStream))
//		}
//
//		return MergeFnFactory[[]ReadOnlyStream[RS]]().GenerateToStream(ctx, outStreams)
//	}
//}
//
//// ProcessHandlerFactory wraps ProcessFn[IN, RS] and adapt it to ProcessHandler[IN, RS] interface
//func ProcessHandlerFactory[IN, RS any](processFn ProcessFn[IN, RS], options ...Option) ProcessHandler[IN, RS] {
//	config := yapgo2Config{}
//	for _, option := range options {
//		option.apply(&config)
//	}
//
//	if config.onErrorStrategy == OnErrorContinue {
//		return func() ProcessHandlerFn[IN, RS] {
//			return func(t IN) (RS, bool) {
//				res, _ := processFn(t)
//				return res, false
//			}
//		}()
//	}
//
//	return func() ProcessHandlerFn[IN, RS] {
//		return func(t IN) (RS, bool) {
//			res, err := processFn(t)
//			return res, err != nil
//		}
//	}()
//
//}
