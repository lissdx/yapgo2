package pipeline

//import (
//	"context"
//	"fmt"
//	genHandlerConf "github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_generator_handler"
//	"go.uber.org/goleak"
//	"log"
//	"math/rand"
//	"testing"
//	"time"
//)
//
//// TestExample_SimplePipeline base usage
//func TestExample_SimplePipeline(t *testing.IN) {
//	defer goleak.VerifyNone(t)
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
//	defer cancel()
//
//	// Set up the data generator
//	data := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
//	// get generator function
//	genFunc := SliceGeneratorFuncFactory[[]int](data)
//	// wrap the function with handler
//	generatorHandler := GeneratorHandlerFactory[int](genFunc, genHandlerConf.WithTimesToGenerate(uint(len(data))))
//	// init generator stage
//	toStreamGeneratorStage := GeneratorStageFactory[int](generatorHandler)
//
//	// setup simple process
//	simpleProcessFn := func(i int) (string, error) { return fmt.Sprintf("%d", i), nil }
//	processHandler := ProcessHandlerFactory[int, string](simpleProcessFn)
//	processStage := ProcessStageFactory[int, string](processHandler)
//
//	genDataStream := toStreamGeneratorStage.GenerateToStream(ctx)
//	resStream := processStage.GenerateToStream(ctx, genDataStream)
//
//	evenCount := 0
//	for v := range OrDoneFnFactory[string]().GenerateToStream(ctx, resStream) {
//		t.Log(v)
//		evenCount++
//	}
//
//	t.Log("Total events: ", evenCount)
//}
//
//// TestExample_PipelineWithCustomErrorHandler add the custom error handler
//func TestExample_PipelineWithCustomErrorHandler(t *testing.IN) {
//	defer goleak.VerifyNone(t)
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	// Set up the data generator
//	data := []string{"one", "two", "three", "four", "five", "six"}
//	// get generator function
//	genFunc := SliceGeneratorFuncFactory(data)
//	// wrap the function with handler
//	generatorHandler := GeneratorHandlerFactory(genFunc, genHandlerConf.WithTimesToGenerate(uint(len(data))))
//	// init generator stage
//	toStreamGeneratorStage := GeneratorStageFactory(generatorHandler)
//
//	// setup simple process
//	// define the simple process function
//	strToIntProcessFn := func(s string) (res int, err error) {
//		switch s {
//		case "one":
//			res = 1
//		case "two":
//			res = 2
//		case "three":
//			res = 3
//		case "four":
//			res = 4
//		default:
//			err = fmt.Errorf("unknown value to tranform: %s", s)
//		}
//		return
//	}
//
//	// simple Middleware (chain) pattern implementation
//	// wrap the process function with a simple error handler
//	// function
//	errorHandler := func(next ProcessFn[string, int]) ProcessFn[string, int] {
//		return func(s string) (int, error) {
//			res, err := next(s)
//			if err != nil {
//				log.Default().Println(err.Error())
//			}
//			return res, err
//		}
//	}(strToIntProcessFn)
//
//	processHandler := ProcessHandlerFactory(errorHandler)
//	processStage := ProcessStageFactory(processHandler)
//
//	genDataStream := toStreamGeneratorStage.GenerateToStream(ctx)
//	resStream := processStage.GenerateToStream(ctx, genDataStream)
//
//	evenCount := 0
//	for v := range OrDoneFnFactory[int]().GenerateToStream(ctx, resStream) {
//		t.Log(v)
//		evenCount++
//	}
//
//	t.Log("Total events: ", evenCount)
//}
//
//// TestExample_PipelineWithFilterAndErrorIgnore filter and error ignore
//// (custom middleware implementation)
//func TestExample_PipelineWithFilterAndErrorIgnore(t *testing.IN) {
//	defer goleak.VerifyNone(t)
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	// Set up the data generator
//	data := []string{"one", "two", "three", "four", "five", "six"}
//	// get generator function
//	genFunc := SliceGeneratorFuncFactory(data)
//	// wrap the function with handler
//	generatorHandler := GeneratorHandlerFactory(genFunc, genHandlerConf.WithTimesToGenerate(uint(len(data))))
//	// init generator stage
//	toStreamGeneratorStage := GeneratorStageFactory(generatorHandler)
//
//	// setup simple process
//	// define the simple process function
//	strToIntProcessFn := func(s string) (res int, err error) {
//		switch s {
//		case "one":
//			res = 1
//		case "two":
//			res = 2
//		case "three":
//			res = 3
//		case "four":
//			res = 4
//		default:
//			err = fmt.Errorf("unknown value to tranform: %s", s)
//		}
//		return
//	}
//
//	// simple Middleware (chain) pattern implementation
//	// wrap the process function with a simple error handler
//	// function
//	errorHandler := func(next ProcessFn[string, int]) ProcessFn[string, int] {
//		return func(s string) (int, error) {
//			res, err := next(s)
//			if err != nil {
//				log.Default().Println("Error Report on process:", err.Error(), "|", fmt.Sprintf("note: the value: %+v", res), "will be passed to the next stage")
//			} else {
//				log.Default().Println(fmt.Sprintf("Process sucess: the value: %s transformed to %+v", s, res))
//			}
//			return res, err
//		}
//	}(strToIntProcessFn)
//
//	processHandler := ProcessHandlerFactory(errorHandler, WithOnErrorContinueStrategy())
//	processStage := ProcessStageFactory(processHandler)
//
//	simpleFilter := func() FilterFn[int] {
//		return func(i int) (int, bool) {
//			switch i {
//			case 0, 2, 3:
//				return i, false
//			}
//			return i, true
//		}
//	}()
//	// simple Middleware (chain) pattern implementation
//	// wrap the process function with a simple error handler
//	// function
//	filterReportHandler := func(next ProcessHandler[int, int]) ProcessHandler[int, int] {
//		return ProcessHandlerFn[int, int](func(i int) (int, bool) {
//			res, removeIt := next.Apply(i)
//			if removeIt {
//				log.Default().Println(fmt.Sprintf("Filter Report: the value: %+v", res), "FILTERED")
//			} else {
//				log.Default().Println(fmt.Sprintf("Filter Report: the value: %+v", res), "PASSED")
//			}
//			return res, removeIt
//		})
//	}
//	filterStage := ProcessStageFactory(filterReportHandler(simpleFilter))
//
//	genDataStream := toStreamGeneratorStage.GenerateToStream(ctx)
//	processStream := processStage.GenerateToStream(ctx, genDataStream)
//	filterStream := filterStage.GenerateToStream(ctx, processStream)
//
//	evenCount := 0
//	for v := range OrDoneFnFactory[int]().GenerateToStream(ctx, filterStream) {
//		t.Log("In the end ve GOT:", v)
//		evenCount++
//	}
//
//	t.Log("Total events: ", evenCount)
//}
//
//// TestExample_PipelineNoDrainGuarantee
//// in case of using the same Context there is no
//// DrainGuarantee
//// Context close our stages in the random order
//func TestExample_PipelineNoDrainGuarantee(t *testing.IN) {
//	defer goleak.VerifyNone(t)
//
//	// create a context with 1sec timeout
//	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
//	defer cancel()
//
//	// Set up the data generator
//	// get generator function (just random int generator)
//	genFunc := RandomIntGeneratorFuncFactory(100)
//	// wrap the function with handler
//	generatorHandler := GeneratorHandlerFactory(genFunc)
//
//	// create counter middleware
//	genCounter := 0
//	counterMiddleware := func(next GeneratorHandler[int]) GeneratorHandler[int] {
//		return GeneratorHandlerFunc[int](func() (int, bool) {
//			defer func() {
//				genCounter++
//			}()
//			return next.GenerateIt()
//		})
//	}
//
//	// init generator stage
//	generatorStage := GeneratorStageFactory(counterMiddleware(generatorHandler))
//
//	//
//	// Process Part
//	//
//	quiteLongProcessFunc := func() ProcessFn[int, int] {
//		return func(i int) (int, error) {
//			waitForMlSec := rand.Intn(1000-100) + 100
//			stubCtx, stubCancel := context.WithTimeout(context.Background(), time.Duration(waitForMlSec)*time.Millisecond)
//			defer stubCancel()
//			<-stubCtx.Done()
//			//time.Sleep(time.
//			//Duration(waitForMlSec) * time.Millisecond)
//			//for ii := 0; ii < 100_000; ii++ {
//			//	for bb := 0; bb < 10000; bb++ {
//			//
//			//	}
//			//}
//			return i, nil
//		}
//	}()
//	intToStringProcessFunc := func() ProcessFn[int, string] {
//		return func(i int) (string, error) {
//			waitForMlSec := rand.Intn(1000-100) + 100
//			stubCtx, stubCancel := context.WithTimeout(context.Background(), time.Duration(waitForMlSec)*time.Millisecond)
//			defer stubCancel()
//			<-stubCtx.Done()
//			return fmt.Sprintf("%d", i), nil
//		}
//	}()
//
//	quiteLongProcessFnHandler := ProcessHandlerFactory(quiteLongProcessFunc)
//	intToStringHandler := ProcessHandlerFactory(intToStringProcessFunc)
//
//	quiteLongProcessWitFanOutStage1 := ProcessStageFactoryWithFanOut(quiteLongProcessFnHandler, 1000)
//	quiteLongProcessWitFanOutStage2 := ProcessStageFactoryWithFanOut(quiteLongProcessFnHandler, 100)
//	intToStringStage := ProcessStageFactory(intToStringHandler)
//
//	//
//	// GenerateToStream Stages
//	//
//	genDataStream := generatorStage.GenerateToStream(ctx)
//	quiteLongProcessWitFanOut1OutStream := quiteLongProcessWitFanOutStage1.GenerateToStream(ctx, genDataStream)
//	quiteLongProcessWitFanOut2OutStream := quiteLongProcessWitFanOutStage2.GenerateToStream(ctx, quiteLongProcessWitFanOut1OutStream)
//	//quiteLongProcessWitFanOut2OutStream := quiteLongProcessWitFanOut1OutStream
//	resStream := intToStringStage.GenerateToStream(ctx, quiteLongProcessWitFanOut2OutStream)
//
//	//processStream := quiteLongProcessWitFanOutStage1.GenerateToStream(ctx, genDataStream)
//	//filterStream := filterStage.GenerateToStream(ctx, processStream)
//
//	evenCount := 0
//	finalContext, fcCancel := context.WithCancel(context.Background())
//	defer fcCancel()
//	for range OrDoneFnFactory[string]().GenerateToStream(finalContext, resStream) {
//		//t.Log("In the end ve GOT:", v)
//		evenCount++
//	}
//
//	t.Log("Total events generated: ", genCounter)
//	t.Log("Total events processed: ", evenCount)
//}
//
//// TestExample_PipelineDrainGuarantee
//// in case of using the same Context there is no
//// DrainGuarantee
//// so lets pass the independent Contexts
//func TestExample_PipelineDrainGuarantee(t *testing.IN) {
//	defer goleak.VerifyNone(t)
//
//	// Set up the data generator
//	// get generator function (just random int generator)
//	genFunc := RandomIntGeneratorFuncFactory(100)
//	// wrap the function with handler
//	generatorHandler := GeneratorHandlerFactory(genFunc)
//
//	// create counter middleware
//	genCounter := 0
//	counterMiddleware := func(next GeneratorHandler[int]) GeneratorHandler[int] {
//		return GeneratorHandlerFunc[int](func() (int, bool) {
//			defer func() {
//				genCounter++
//			}()
//			return next.GenerateIt()
//		})
//	}
//
//	// init generator stage
//	// Lets use UnsafeGeneratorStageFactory instead of GeneratorStageFactory
//	// UnsafeGeneratorStageFactory is potentially dangerous
//	generatorStage := UnsafeGeneratorStageFactory(counterMiddleware(generatorHandler))
//
//	//
//	// Process Part
//	//
//	quiteLongProcessFunc1 := func() ProcessFn[int, int] {
//		return func(i int) (int, error) {
//			waitForMlSec := rand.Intn(1000-100) + 100
//			<-time.After(time.Duration(waitForMlSec) * time.Millisecond)
//			return i, nil
//		}
//	}()
//	quiteLongProcessFunc2 := func() ProcessFn[int, int] {
//		return func(i int) (int, error) {
//			waitForMlSec := rand.Intn(200-20) + 20
//			<-time.After(time.Duration(waitForMlSec) * time.Millisecond)
//			return i, nil
//		}
//	}()
//	intToStringProcessFunc := func() ProcessFn[int, string] {
//		return func(i int) (string, error) {
//			waitForMlSec := rand.Intn(10-1) + 1
//			<-time.After(time.Duration(waitForMlSec) * time.Millisecond)
//			return fmt.Sprintf("to string: %d", i), nil
//		}
//	}()
//	//
//	quiteLongProcessFnHandler1 := ProcessHandlerFactory(quiteLongProcessFunc1)
//	quiteLongProcessFnHandler2 := ProcessHandlerFactory(quiteLongProcessFunc2)
//	intToStringHandler := ProcessHandlerFactory(intToStringProcessFunc)
//
//	quiteLongProcessWitFanOutStage1 := ProcessStageFactoryWithFanOut(quiteLongProcessFnHandler1, 1000)
//	quiteLongProcessWitFanOutStage2 := ProcessStageFactoryWithFanOut(quiteLongProcessFnHandler2, 300)
//	intToStringStage := ProcessStageFactory(intToStringHandler)
//
//	//
//	// GenerateToStream Stages
//	//
//	// create a context with 1sec timeout
//	ctx1, cancel1 := context.WithTimeout(context.Background(), 1*time.Second)
//	defer cancel1()
//
//	ctx2, cancel2 := context.WithCancel(context.Background())
//	defer cancel2()
//
//	genDataStream := generatorStage.GenerateToStream(ctx1)
//	quiteLongProcessWitFanOut1OutStream := quiteLongProcessWitFanOutStage1.GenerateToStream(ctx2, genDataStream)
//	quiteLongProcessWitFanOut2OutStream := quiteLongProcessWitFanOutStage2.GenerateToStream(ctx2, quiteLongProcessWitFanOut1OutStream)
//	resStream := intToStringStage.GenerateToStream(ctx2, quiteLongProcessWitFanOut2OutStream)
//
//	evenCount := 0
//	for _ = range OrDoneFnFactory[string]().GenerateToStream(ctx2, resStream) {
//		//t.Log("In the end ve GOT:", v)
//		evenCount++
//	}
//
//	t.Log("Total events generated: ", genCounter)
//	t.Log("Total events processed: ", evenCount)
//}
//
//// TestExample_PipelineNoDrainGuarantee
//// in case of using the same Context there is no
//// DrainGuarantee
//// Context close our stages in the random order
//func TestExample_PipelineNoDrainGuaranteeSpecialGenFunction(t *testing.IN) {
//	defer goleak.VerifyNone(t)
//
//	// create a context with 1sec timeout
//	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
//	defer cancel()
//
//	// Set up the data generator
//	// get generator function (just random int generator)
//	genFunc := RandomIntGeneratorFuncFactory(100)
//	// wrap the function with handler
//	generatorHandler := GeneratorHandlerFactory(genFunc)
//
//	// create counter middleware
//	genCounter := 0
//	counterMiddleware := func(next GeneratorHandler[int]) GeneratorHandler[int] {
//		return GeneratorHandlerFunc[int](func() (int, bool) {
//			defer func() {
//				genCounter++
//			}()
//			return next.GenerateIt()
//		})
//	}
//
//	// init generator stage
//	generatorStage := GeneratorStageFactory(counterMiddleware(generatorHandler))
//
//	//
//	// Process Part
//	//
//	quiteLongProcessFunc := func() ProcessFn[int, int] {
//		return func(i int) (int, error) {
//			waitForMlSec := rand.Intn(1000-100) + 100
//			stubCtx, stubCancel := context.WithTimeout(context.Background(), time.Duration(waitForMlSec)*time.Millisecond)
//			defer stubCancel()
//			<-stubCtx.Done()
//			//time.Sleep(time.
//			//Duration(waitForMlSec) * time.Millisecond)
//			//for ii := 0; ii < 100_000; ii++ {
//			//	for bb := 0; bb < 10000; bb++ {
//			//
//			//	}
//			//}
//			return i, nil
//		}
//	}()
//
//	intToStringProcessFunc := func() ProcessFn[int, string] {
//		return func(i int) (string, error) {
//			waitForMlSec := rand.Intn(1000-100) + 100
//			stubCtx, stubCancel := context.WithTimeout(context.Background(), time.Duration(waitForMlSec)*time.Millisecond)
//			defer stubCancel()
//			<-stubCtx.Done()
//			return fmt.Sprintf("%d", i), nil
//		}
//	}()
//
//	quiteLongProcessFnHandler := ProcessHandlerFactory(quiteLongProcessFunc)
//	intToStringHandler := ProcessHandlerFactory(intToStringProcessFunc)
//
//	quiteLongProcessWitFanOutStage1 := ProcessStageFactoryWithFanOut(quiteLongProcessFnHandler, 1000)
//	quiteLongProcessWitFanOutStage2 := ProcessStageFactoryWithFanOut(quiteLongProcessFnHandler, 100)
//	intToStringStage := ProcessStageFactory(intToStringHandler)
//
//	//
//	// GenerateToStream Stages
//	//
//	genDataStream := generatorStage.GenerateToStream(ctx)
//	quiteLongProcessWitFanOut1OutStream := quiteLongProcessWitFanOutStage1.GenerateToStream(ctx, genDataStream)
//	quiteLongProcessWitFanOut2OutStream := quiteLongProcessWitFanOutStage2.GenerateToStream(ctx, quiteLongProcessWitFanOut1OutStream)
//	//quiteLongProcessWitFanOut2OutStream := quiteLongProcessWitFanOut1OutStream
//	resStream := intToStringStage.GenerateToStream(ctx, quiteLongProcessWitFanOut2OutStream)
//
//	//processStream := quiteLongProcessWitFanOutStage1.GenerateToStream(ctx, genDataStream)
//	//filterStream := filterStage.GenerateToStream(ctx, processStream)
//
//	evenCount := 0
//	finalContext, fcCancel := context.WithCancel(context.Background())
//	defer fcCancel()
//	for range OrDoneFnFactory[string]().GenerateToStream(finalContext, resStream) {
//		//t.Log("In the end ve GOT:", v)
//		evenCount++
//	}
//
//	t.Log("Total events generated: ", genCounter)
//	t.Log("Total events processed: ", evenCount)
//}
