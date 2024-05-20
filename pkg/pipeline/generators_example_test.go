package pipeline

import (
	"context"
	genHandlerConf "github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_generator_handler"
	"go.uber.org/goleak"
	"log"
	"math/rand"
	"testing"
	"time"
)

/**
 *  Generators Examples
 */

// Random int generation (10 times generation) example
func TestExample_GeneratorFactoryWithTimesToGenerate(t *testing.T) {
	defer goleak.VerifyNone(t)

	timesToGenerate := 10
	// Custom generator
	genFunc := func() int {
		return rand.Intn(100)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the generator handler
	// It should generate values timesToGenerate times
	// Values should be random int in interval [0-100)
	// see genFunc
	generatorHandler := GeneratorHandlerFactory[int](genFunc, genHandlerConf.WithTimesToGenerate(uint(timesToGenerate)))

	// Set up the generator stage with the generatorHandler
	toStreamGeneratorStage := GeneratorStageFactory[int](generatorHandler)

	// Run the stage
	outStream := toStreamGeneratorStage.Run(ctx)

	dataCounter := 0
	for v := range outStream {
		dataCounter += 1
		log.Default().Println("Got value:", v)
	}

	log.Default().Println("Data was generated:", dataCounter, "times")

}

// Random int generation never ended  example
func TestExample_GeneratorFactoryWithOutTimesToGenerate(t *testing.T) {
	defer goleak.VerifyNone(t)

	// Use predefined GeneratorFunc factory
	genFunc := RandomIntGeneratorFuncFactory(1001)
	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond*300)
	defer cancel()

	// Create (wrap) the genFunc with GeneratorHandler
	generatorHandler := GeneratorHandlerFactory[int](genFunc)

	// Set up the generator stage
	toStreamGeneratorStage := GeneratorStageFactory[int](generatorHandler)
	// ... run it and get the output stream
	outStream := toStreamGeneratorStage.Run(ctx)

	dataCounter := 0
	for v := range outStream {
		dataCounter += 1
		log.Default().Println("Got value:", v)
	}

	log.Default().Println("Data was generated:", dataCounter, "times")

}

// We are able to wrap the main GeneratorHandler
// with a Middleware function(s)
// Random int generation never ended
func TestExample_GeneratorFactoryWithMiddleWare(t *testing.T) {
	defer goleak.VerifyNone(t)

	genFunc := RandomIntGeneratorFuncFactory(1001)

	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond*1000)
	defer cancel()

	generatorHandler := GeneratorHandlerFactory[int](genFunc)
	middleware1 := GeneratorTracerMiddlewareExample[int]("generator tracer", generatorHandler)
	middleware2 := GeneratorLoggerMiddlewareExample[int]("generator logger", middleware1)

	toStreamGeneratorStage := GeneratorStageFactory[int](middleware2)
	outStream := toStreamGeneratorStage.Run(ctx)

	dataCounter := 0
	for v := range outStream {
		dataCounter += 1
		log.Default().Println("Got value:", v)
	}

	log.Default().Println("Data was generated:", dataCounter, "times")
}

// TestExample_SliceGeneratorWithMiddleWare Generate data
// from the given slice of data
// with a Middleware function(s)
// generate data len(slice) times
// simplest way (without type annotation)
func TestExample_SliceGeneratorWithMiddleWare(t *testing.T) {
	defer goleak.VerifyNone(t)

	data := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	genFunc := SliceGeneratorFuncFactory(data)

	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond*1000)
	defer cancel()

	generatorHandler := GeneratorHandlerFactory(genFunc, genHandlerConf.WithTimesToGenerate(uint(len(data))))
	traceMiddleware1 := GeneratorTracerMiddlewareExample("Gen traces", generatorHandler)
	loggerMiddleware := GeneratorLoggerMiddlewareExample("generator logger", traceMiddleware1)
	traceMiddleware2 := GeneratorTracerMiddlewareExample("External tracer", loggerMiddleware)

	toStreamGeneratorStage := GeneratorStageFactory(traceMiddleware2)
	outStream := toStreamGeneratorStage.Run(ctx)

	dataCounter := 0
	for v := range outStream {
		dataCounter += 1
		log.Default().Println("Got value:", v)
	}

	log.Default().Println("Data was generated:", dataCounter, "times")

}
