package pipeline

import (
	"context"
	"fmt"
	"github.com/lissdx/yapgo2/pkg/logger"
	genProdConf "github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_producer"
	cnfProcHandler "github.com/lissdx/yapgo2/pkg/pipeline/config/process/config_handler"
	cnfProcStage "github.com/lissdx/yapgo2/pkg/pipeline/config/process/config_stage"
	"go.uber.org/goleak"
	"math/rand"
	"testing"
	"time"
)

var simpleFilter = func() FilterFn[int] {
	return func(i int) (int, bool) {
		switch i {
		case 0, 2, 3:
			return i, false
		}
		return i, true
	}
}()

var strToIntProcessFn = func() ProcessFn[string, int] {
	return func(s string) (res int, err error) {
		switch s {
		case "one":
			res = 1
		case "two":
			res = 2
		case "three":
			res = 3
		case "four":
			res = 4
		default:
			err = fmt.Errorf("unknown value to tranform: %s", s)
		}
		return
	}
}()

// simple Middleware (or chain) pattern implementation
// wrap the process function with a simple error handler
// function
var errorHandler = func(next ProcessFn[string, int]) ProcessFn[string, int] {
	return func(s string) (res int, err error) {
		defer func(err *error) {
			if *err != nil {
				plLogger.Error(*err)
			}
		}(&err)
		//res, err = next(s)
		return next(s)
	}
}

// simple Middleware (chain) pattern implementation
// wrap the process function with a simple error handler
// function
func isOmittedReportHandler[T any](next ProcessHandler[T, T]) ProcessHandler[T, T] {
	return ProcessHandlerFn[T, T](func(t T) ProcessResultCarrier[T] {
		processResultCarrier := next.Apply(t)
		if processResultCarrier.IsOmitted() {
			plLogger.Debug("isOmitted Report: the value: %+v OMITTED", processResultCarrier.ProcessResult())
		} else {
			plLogger.Debug("isOmitted Report: the value: %+v PASSED", processResultCarrier.ProcessResult())
		}
		return processResultCarrier
	})
}

var plLogger = func() logger.ILogger {
	return logger.LoggerFactory(
		logger.WithZapLoggerImplementer(),
		logger.WithLoggerLevel("DEBUG"),
		logger.WithZapColored(),
		logger.WithZapConsoleEncoding(),
		logger.WithZapColored(),
	)
}()

// TestExample_SimplePipeline base usage
func TestExample_SimplePipeline(t *testing.T) {
	defer goleak.VerifyNone(t)

	timesToGenerate := 10
	stageName := "IntToString_Stage"
	// Set up the data generator
	data := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	// get generator function
	genFunc := SliceGeneratorFuncFactory[[]int](data)
	// process function
	var processFunc ProcessFn[int, string] = func(i int) (string, error) {
		return fmt.Sprintf("%d", i), nil
	}

	// handlers part
	// wrap the generator function with a handler
	genProducer := GeneratorProducerFactory[int](genFunc, genProdConf.WithTimesToGenerate(uint(timesToGenerate)),
		genProdConf.WithLogger(plLogger))
	// create process handler
	processHandler := ProcessHandlerFactory(processFunc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// stage part
	// get the generated data stream
	genStream := genProducer.GenerateToStream(ctx)
	// apply the process function to the data stream
	resultStream := ProcessStageFactory(processHandler,
		cnfProcStage.WithLogger(plLogger),
		cnfProcStage.WithName(stageName)).
		Run(ctx, genStream)

	evenCount := 0
	for v := range resultStream {
		evenCount += 1
		plLogger.Trace("Got value: %+v", v)
	}

	plLogger.Info("Total events: %d", evenCount)
}

// TestExample_PipelineWithCustomErrorHandler add the custom error handler
func TestExample_PipelineWithCustomErrorHandler(t *testing.T) {
	defer goleak.VerifyNone(t)

	stageName := "StringToInt_Stage"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up the data generator
	data := []string{"one", "two", "three", "four", "five", "six"}
	// get generator function
	genFunc := SliceGeneratorFuncFactory(data)
	// wrap the function with handler
	genProducer := GeneratorProducerFactory(genFunc, genProdConf.WithTimesToGenerate(uint(len(data))),
		genProdConf.WithLogger(plLogger))

	// stage part
	// get the generated data stream
	genStream := genProducer.GenerateToStream(ctx)
	// apply the process function to the data stream
	processHandler := ProcessHandlerFactory(errorHandler(strToIntProcessFn))
	resultStream := ProcessStageFactory(processHandler,
		cnfProcStage.WithLogger(plLogger),
		cnfProcStage.WithName(stageName)).
		Run(ctx, genStream)

	evenCount := 0
	resSlice := make([]int, 0)
	for v := range resultStream {
		plLogger.Debug("Got value: %+v", v)
		resSlice = append(resSlice, v)
		evenCount++
	}

	plLogger.Info("Total events: %d", evenCount)
	plLogger.Info("Result slice: %+v", resSlice)
}

// TestExample_PipelineWithFilterAndErrorIgnore filter and error ignore
// (custom middleware implementation)
func TestExample_PipelineWithFilterAndErrorIgnore(t *testing.T) {
	defer goleak.VerifyNone(t)

	stageName := "StringToInt_ErrorIgnore_Stage"
	filterStageName := "StringToInt_Filter_Stage"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up the data generator
	data := []string{"one", "two", "three", "four", "five", "six"}
	// get generator function
	genFunc := SliceGeneratorFuncFactory(data)
	// wrap the gen function with handler (ToStreamGenerator)
	genProducer := GeneratorProducerFactory(genFunc, genProdConf.WithTimesToGenerate(uint(len(data))),
		genProdConf.WithLogger(plLogger))

	// init the process handler with errorHandler
	// the errorHandler is a simple wrapper of
	// strToIntProcessFn process function
	// will transform "one" -> 1, "two" -> 2, "three" -> 3, "four" -> 4
	// otherwise an error will be generated
	// The result value should be the default result value (in our case it should be 0)
	processHandler := ProcessHandlerFactory(errorHandler(strToIntProcessFn),
		cnfProcHandler.WithOnErrorContinueStrategy())

	// init the first stage in the pipeline
	processStage := ProcessStageFactory(processHandler, cnfProcStage.WithName(stageName))

	// init the filter handler
	// will pass 0, 2, 3
	filterHandlerWithReport := isOmittedReportHandler(ProcessHandlerFilterFactory(simpleFilter))
	filterStage := ProcessStageFactory(filterHandlerWithReport,
		cnfProcStage.WithLogger(plLogger),
		cnfProcStage.WithName(filterStageName))

	// run pipeline
	genDataStream := genProducer.GenerateToStream(ctx)
	processStream := processStage.Run(ctx, genDataStream)
	resultStream := filterStage.Run(ctx, processStream)

	evenCount := 0
	resSlice := make([]int, 0, len(data))
	for v := range resultStream {
		plLogger.Debug("Got value: %+v", v)
		resSlice = append(resSlice, v)
		evenCount++
	}

	plLogger.Info("Total events: %d", evenCount)
	plLogger.Info("Result slice: %+v", resSlice)
}

// TestExample_PipelineWithFanOutAndDrainGuarantee
// in case of using the same Context there is no
// DrainGuarantee
// Context close our stages in the random order

func TestExample_PipelineWithFanOutAndDrainGuarantee(t *testing.T) {
	defer goleak.VerifyNone(t)
	timesToGenerate := uint(100)
	stageName1 := "quiteLongProcessFunc"
	//stageName_2 := "intToStringProcessFunc"
	// create a context with 1sec timeout
	//ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	//defer cancel()

	// Set up the data generator
	// get generator function (just random int generator)
	genFunc := RandomIntGeneratorFuncFactory(100)
	// wrap the function with handler
	generatorHandler := GeneratorProducerFactory(genFunc,
		genProdConf.WithTimesToGenerate(timesToGenerate),
		genProdConf.WithLogger(plLogger))

	//
	// Process Part
	//
	quiteLongProcessFunc := func() ProcessFn[int, int] {
		return func(i int) (int, error) {
			waitForMlSec := rand.Intn(1000-100) + 100
			stubCtx, stubCancel := context.WithTimeout(context.Background(), time.Duration(waitForMlSec)*time.Millisecond)
			defer stubCancel()
			<-stubCtx.Done()
			return i, nil
		}
	}()
	//intToStringProcessFunc := func() ProcessFn[int, string] {
	//	return func(i int) (string, error) {
	//		waitForMlSec := rand.Intn(1000-100) + 100
	//		stubCtx, stubCancel := context.WithTimeout(context.Background(), time.Duration(waitForMlSec)*time.Millisecond)
	//		defer stubCancel()
	//		<-stubCtx.Done()
	//		return fmt.Sprintf("%d", i), nil
	//	}
	//}()

	quiteLongProcessFnHandler := ProcessHandlerFactory(quiteLongProcessFunc)
	//intToStringHandler := ProcessHandlerFactory(intToStringProcessFunc,
	//	cnfProcStage.WithLogger(plLogger),
	//	cnfProcStage.WithName(stageName_2))

	quiteLongProcessWitFanOutStage1 := ProcessStageFactoryWithFanOut(quiteLongProcessFnHandler, 5,
		cnfProcStage.WithName(stageName1), cnfProcStage.WithLogger(plLogger))
	//quiteLongProcessWitFanOutStage2 := ProcessStageFactoryWithFanOut(quiteLongProcessFnHandler, 100)
	//intToStringStage := ProcessStageFactory(intToStringHandler)

	//
	// GenerateToStream Stages
	//
	genCtx, genCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer genCancel()
	genDataStream := generatorHandler.GenerateToStream(genCtx)
	stg1Ctx, stg1Cancel := context.WithCancel(context.Background())
	defer stg1Cancel()
	quiteLongProcessWitFanOut1OutStream := quiteLongProcessWitFanOutStage1.Run(stg1Ctx, genDataStream)
	//quiteLongProcessWitFanOut2OutStream := quiteLongProcessWitFanOutStage2.GenerateToStream(ctx, quiteLongProcessWitFanOut1OutStream)
	//quiteLongProcessWitFanOut2OutStream := quiteLongProcessWitFanOut1OutStream
	//resStream := intToStringStage.GenerateToStream(ctx, quiteLongProcessWitFanOut2OutStream)

	//processStream := quiteLongProcessWitFanOutStage1.GenerateToStream(ctx, genDataStream)
	//filterStream := filterStage.GenerateToStream(ctx, processStream)

	evenCount := 0
	//finalContext, fcCancel := context.WithCancel(context.Background())
	//defer fcCancel()
	for v := range quiteLongProcessWitFanOut1OutStream {
		plLogger.Trace("Got value: %+v", v)
		evenCount++
	}

	plLogger.Debug("Total events processed: ", evenCount)
}

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
