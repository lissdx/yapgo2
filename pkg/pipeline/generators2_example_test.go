package pipeline

//
//import (
//	"context"
//	genConf "github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_generator"
//	"go.uber.org/goleak"
//	"log"
//	"math/rand"
//	"testing"
//)
//
///**
// *  Generators Examples
// */
//
//// Random int generation (10 times generation) example
//func TestExample_Generator2FactoryWithTimesToGenerate(t *testing.IN) {
//	defer goleak.VerifyNone(t)
//
//	timesToGenerate := 10
//	// Custom generator
//	genFunc := func() int {
//		return rand.Intn(100)
//	}
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	// Create the generator handler
//	// It should generate values timesToGenerate times
//	// Values should be random int in interval [0-100)
//	// see genFunc
//	generatorHandler := GeneratorProducerFactory[int](genFunc, genConf.WithTimesToGenerate(uint(timesToGenerate)))
//
//	// Set up the generator stage with the generatorHandler
//	toStreamGeneratorStage := GeneratorStageFactory2(generatorHandler)
//
//	// GenerateToStream the stage
//	outStream := toStreamGeneratorStage.GenerateToStream(ctx)
//
//	dataCounter := 0
//	// We are waiting until the outStream
//	// will be closed by the Generator
//	for v := range outStream {
//		dataCounter += 1
//		log.Default().Println("Got value:", v)
//	}
//
//	log.Default().Println("Data was fetched:", dataCounter, "times")
//
//}

////// TestExample_Generator2FactoryWithGenCounter
////// middleware (counter og generated values) added
////func TestExample_Generator2FactoryWithGenCounter(t *testing.IN) {
////	defer goleak.VerifyNone(t)
////
////	// Custom generator
////	genFunc := func() int {
////		return rand.Intn(100)
////	}
////
////	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
////	defer cancel()
////
////	// Create the generator handler
////	// It should generate values timesToGenerate times
////	// Values should be random int in interval [0-100)
////	// see genFunc
////	generatorHandler := GeneratorProducerFactory[int](genFunc)
////
////	// create custom counter middleware
////	genCounter := 0
////	counterMiddleware := func(next GeneratorHandler2[int]) GeneratorHandler2[int] {
////		return GeneratorHandlerFunc2[int](func(mCtx context.Context) ProcessResultCarrier[int] {
////			defer func() {
////				genCounter++
////			}()
////			return next.GenerateIt2(mCtx)
////		})
////	}
////
////	// Set up the generator stage with the generatorHandler
////	toStreamGeneratorStage := GeneratorStageFactory2(counterMiddleware(generatorHandler))
////
////	// GenerateToStream the stage
////	outStream := toStreamGeneratorStage.GenerateToStream(ctx)
////
////	dataCounter := 0
////	// We are waiting until the outStream
////	// will be closed by the Generator
////	for _ = range outStream {
////		dataCounter += 1
////		//log.Default().Println("Got value:", v)
////	}
////
////	log.Default().Println("Data was generated:", genCounter, "times")
////	log.Default().Println("Data was fetched:", dataCounter, "times")
////
////}
////
////// TestExample_Generator2FactorySlowGenerator
////// middleware (counter og generated values) added
////func TestExample_Generator2FactorySlowGenerator(t *testing.IN) {
////	defer goleak.VerifyNone(t)
////
////	// Custom generator
////	genFunc := func() int {
////		waitForMlSec := rand.Intn(1000-100) + 100
////		stubCtx, stubCancel := context.WithTimeout(context.Background(), time.Duration(waitForMlSec)*time.Millisecond)
////		defer stubCancel()
////		<-stubCtx.Done()
////		return rand.Intn(100)
////	}
////
////	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond*1)
////	defer cancel()
////
////	// Create the generator handler
////	// It should generate values timesToGenerate times
////	// Values should be random int in interval [0-100)
////	// see genFunc
////	generatorHandler := GeneratorProducerFactory[int](genFunc)
////
////	// create custom counter middleware
////	genCounter := 0
////	counterMiddleware := func(next GeneratorHandler2[int]) GeneratorHandler2[int] {
////		return GeneratorHandlerFunc2[int](func(mCtx context.Context) ProcessResultCarrier[int] {
////			defer func() {
////				genCounter++
////			}()
////			return next.GenerateIt2(mCtx)
////		})
////	}
////
////	// Set up the generator stage with the generatorHandler
////	toStreamGeneratorStage := GeneratorStageFactory2(counterMiddleware(generatorHandler))
////
////	// GenerateToStream the stage
////	outStream := toStreamGeneratorStage.GenerateToStream(ctx)
////
////	dataCounter := 0
////	// We are waiting until the outStream
////	// will be closed by the Generator
////	for _ = range outStream {
////		dataCounter += 1
////		//log.Default().Println("Got value:", v)
////	}
////
////	log.Default().Println("Data was generated:", genCounter, "times")
////	log.Default().Println("Data was fetched:", dataCounter, "times")
////
////}
