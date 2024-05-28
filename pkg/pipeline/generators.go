package pipeline

//
//import (
//	"context"
//	genFuncConf "github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_generator_function"
//	genHandlerConf "github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_generator_handler"
//	"math/rand"
//)
//
////const stopIt = true
////
////var EmptySliceToGenerateFrom = errors.New("the Slice provided to the SliceGenerator function is empty")
////
////// GeneratorFunc should generate any
////// kind of data
////// returns:
////// IN - generated data
////type GeneratorFunc[IN any] func() IN
//
//// GeneratorHandlerFunc - an orchestrator for the GeneratorFunc
//// should be used in generator stage
//// The bool param
//// is able to stop the stage
//// returns:
//// IN - generated data
//// bool - DONE (stopIt) if true means no more data should be generated.
////
////	In this case we should stop the generator stage
//type GeneratorHandlerFunc[IN any] func() (IN, bool)
//
//// ToStreamGeneratorFn func should be returned by the stage creation factory
//// (in general means stage that ready to run it)
//// after the "stage" is created we may start the stage as result we will
//// get the "output stream" (channel) and read the generated data from the "output stream"
////type ToStreamGeneratorFn[IN any] func(context.Context) ReadOnlyStream[IN]
//
//// GeneratorHandler interface
//// implemented by GeneratorHandlerFunc
//type GeneratorHandler[IN any] interface {
//	GenerateIt() (IN, bool)
//}
//
//// GeneratorStageRunner interface
//// implemented by ToStreamGeneratorFn
//type GeneratorStageRunner[IN any] interface {
//	GenerateToStream(context.Context) ReadOnlyStream[IN]
//}
//
//func (gsr ToStreamGeneratorFn[IN]) GenerateToStream(ctx context.Context) ReadOnlyStream[IN] {
//	return gsr(ctx)
//}
//
//func (g GeneratorHandlerFunc[IN]) GenerateIt() (IN, bool) {
//	return g()
//}
//
//// GeneratorStageFactory is the main stage of the data generation
//// We are going to use the GeneratorFunc to generate any kind of data
//// see the generators_example_test and learn how we can use it
//// To generate a data we should provide the GeneratorHandlerFunc
////
//// Example:
//// generatorHandler := GeneratorHandlerFactory(genFunc)
//// genStage := GeneratorStageFactory(generatorHandler)
//// outStream := genStage.GenerateToStream(ctx)
//func GeneratorStageFactory[IN any](g GeneratorHandler[IN]) ToStreamGeneratorFn[IN] {
//
//	return func(ctx context.Context) ReadOnlyStream[IN] {
//		// We own the outStream and return it as a read-only channel.
//		// In this case, we can guarantee proper closure of it.
//		outStream := make(chan IN)
//		go func() {
//			defer close(outStream)
//			for {
//				select {
//				case <-ctx.Done():
//					return
//				default:
//					vData, stop := g.GenerateIt()
//					// We do not want to send any more data to a consumer.
//					// Stop the data generation process and
//					// close the output channel.
//					if stop {
//						return
//					}
//					// NOTE: The last generated data may not
//					// be provided if the consumer is slow.
//					// If you sure that the consumer alwais presented
//					// you may youse
//					select {
//					case outStream <- vData:
//					case <-ctx.Done():
//						return
//					}
//				}
//			}
//		}()
//		return outStream
//	}
//}
//
//// UnsafeGeneratorStageFactory is similar to the
//// GeneratorStageFactory, but it guarantees that all generated
//// values should be sent.
//func UnsafeGeneratorStageFactory[IN any](g GeneratorHandler[IN]) ToStreamGeneratorFn[IN] {
//
//	return func(ctx context.Context) ReadOnlyStream[IN] {
//		// We own the outStream and return it as a read-only channel.
//		// In this case, we can guarantee proper closure of it.
//		outStream := make(chan IN)
//		go func() {
//			defer close(outStream)
//			for {
//				select {
//				case <-ctx.Done():
//					return
//				default:
//					vData, stop := g.GenerateIt()
//					// We do not want to send any more data to a consumer.
//					// Stop the data generation process and
//					// close the output channel.
//					if stop {
//						return
//					}
//					// ATTENTION: This is a critical part of the code.
//					// We might be stuck here indefinitely. Please ensure
//					// that the consumer is still functioning, albeit slowly.
//					select {
//					case outStream <- vData:
//					}
//				}
//			}
//		}()
//		return outStream
//	}
//}
//
//// GeneratorHandlerFactory factory wraps the GeneratorFunc
//// if the WithTimesToGenerate option is provided we will wrap
//// GeneratorFunc with special method (which calculates the generation times)
//// if WithTimesToGenerate == 0 - it means run the generator forever
//func GeneratorHandlerFactory[IN any](genFunc GeneratorFunc[IN], options ...genHandlerConf.Option) GeneratorHandler[IN] {
//
//	handlerConfig := genHandlerConf.NewGeneratorHandlerConfig(options...)
//
//	switch handlerConfig.GenerateInfinitely() {
//	case false:
//		// wrap the GeneratorFunc
//		// and run it handlerConfig.TimesToGenerate() times only
//		return func() GeneratorHandlerFunc[IN] {
//			generatedTimes := uint(0)
//			return func() (IN, bool) {
//				defer func() { generatedTimes += 1 }()
//				if generatedTimes >= handlerConfig.TimesToGenerate() {
//					// return the default init/empty IN, and stopIt values
//					return func() (t IN) { return }(), stopIt
//				}
//				// generate the IN value and return with continue
//				return genFunc(), false
//			}
//		}()
//	default:
//		// return the never ended generator
//		return GeneratorHandlerFunc[IN](func() (IN, bool) {
//			return genFunc(), false
//		})
//	}
//}
//
//// RandomIntGeneratorFuncFactory simple random int generator
//func RandomIntGeneratorFuncFactory(interval int) GeneratorFunc[int] {
//	if interval <= 0 {
//		return func() int {
//			return rand.Int()
//		}
//	}
//
//	return func() int {
//		return rand.Intn(interval)
//	}
//}

//// SliceGeneratorFuncFactory repeatedly generates values
//// from the given slice
//// if fnConfig.IgnoreEmptySliceLength() is true and the given slice is empty
//// the generator will generate the default init value of the IN type
//// if fnConfig.IgnoreEmptySliceLength() is false the exception will be triggered
//func SliceGeneratorFuncFactory[S ~[]IN, IN any](s S, options ...genFuncConf.Option) GeneratorFunc[IN] {
//
//	fnConfig := genFuncConf.NewGeneratorFunctionConfig(options...)
//
//	// if the given slice is empty
//	// check the generatorConfig.ignoreEmptySliceLength
//	if len(s) <= 0 {
//		if !fnConfig.IgnoreEmptySliceLength() {
//			panic(EmptySliceToGenerateFrom)
//		}
//		return AValueGeneratorFuncFactory[IN](func() (t IN) { return }())
//	}
//
//	// generate the next value from the given slice
//	// in case of the end of the slice
//	// reset the index and start the generation from
//	// the beginning
//	return func() GeneratorFunc[IN] {
//		currentIndx := 0
//		return func() IN {
//			defer func() {
//				currentIndx = (currentIndx + 1) % len(s)
//			}()
//
//			return s[currentIndx]
//		}
//	}()
//}

//// AValueGeneratorFuncFactory generates
//// the passed value only
//func AValueGeneratorFuncFactory[IN any](t IN) GeneratorFunc[IN] {
//	return func() GeneratorFunc[IN] {
//		return func() IN {
//			return t
//		}
//	}()
//}
