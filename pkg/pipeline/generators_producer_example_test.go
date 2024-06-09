package pipeline

import (
	"context"
	"github.com/lissdx/yapgo2/pkg/logger"
	genProdConf "github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_producer"
	"go.uber.org/goleak"
	"math/rand"
	"testing"
	"time"
)

var gtLogger = func() logger.ILogger {
	return logger.LoggerFactory(
		logger.WithZapLoggerImplementer(),
		logger.WithLoggerLevel("DEBUG"),
		logger.WithZapColored(),
		logger.WithZapConsoleEncoding(),
		logger.WithZapColored(),
	)
}()

/**
*  Generators Examples
 */

// Random int generation (10 times generation) example
func TestExample_ProducerWithTimesToGenerate(t *testing.T) {
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
	genProducer := GeneratorProducerFactory(genFunc,
		genProdConf.WithTimesToGenerate(uint(timesToGenerate)),
		genProdConf.WithLogger(gtLogger))

	// We are waiting until the outStream
	// will be closed by the Generator
	var dataCounter = 0
	for v := range genProducer.GenerateToStream(ctx) {
		dataCounter += 1
		gtLogger.Trace("Got value: %+v", v)
	}

	gtLogger.Info("Data was fetched: %d times", dataCounter)

}

// Random int generation (1000 times generation) but Interrupted example
func TestExample_SimpleProducerInterruptedWithTimesToGenerate(t *testing.T) {
	defer goleak.VerifyNone(t)

	timesToGenerate := 1000
	// Custom generator emulates
	// "slow" producer
	slowGenFunc := func() int {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		<-ctx.Done()
		return rand.Intn(100)
	}

	// Context with timeout
	// will interrupt the generator
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Create the generator handler
	// It should generate values timesToGenerate times
	// Values should be random int in interval [0-100)
	// see slowGenFunc
	genProducer := GeneratorProducerFactory(slowGenFunc,
		genProdConf.WithTimesToGenerate(uint(timesToGenerate)),
		genProdConf.WithLogger(gtLogger))

	// We are waiting until the outStream
	// will be closed by the ctx
	var dataCounter = 0
	for v := range genProducer.GenerateToStream(ctx) {
		dataCounter += 1
		gtLogger.Trace("Got value: %+v", v)
	}

	gtLogger.Info("Data was fetched: %d times", dataCounter)

}

// Random int generation extra slow generator but Interrupted WithTimeout example
func TestExample_SimpleProducerInterruptedExtraSlowGeneratorWithoutTimesToGenerate(t *testing.T) {
	defer goleak.VerifyNone(t)
	// Custom generator
	genFunc := func() int {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		<-ctx.Done()
		return rand.Intn(100)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Create the generator handler
	// It should generate values timesToGenerate times
	// Values should be random int in interval [0-100)
	// see genFunc
	genProducer := GeneratorProducerFactory(genFunc, genProdConf.WithLogger(gtLogger))

	// We are waiting until the outStream
	// will be closed by the Generator
	var dataCounter = 0
	for v := range genProducer.GenerateToStream(ctx) {
		dataCounter += 1
		gtLogger.Trace("Got value: %+v", v)
	}

	gtLogger.Info("Data was fetched: %d times", dataCounter)

}

// Random int generation extra slow generator but Interrupted cancel example
func TestExample_SimpleProducerInterruptedWithCancelExtraSlowGeneratorWithoutTimesToGenerate(t *testing.T) {
	defer goleak.VerifyNone(t)
	// Custom generator
	genFunc := func() int {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		<-ctx.Done()
		return rand.Intn(100)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		internalCtx, internalCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer internalCancel()
		<-internalCtx.Done()

		cancel()

	}()
	defer cancel()

	// Create the generator handler
	// It should generate values timesToGenerate times
	// Values should be random int in interval [0-100)
	// see genFunc
	genProducer := GeneratorProducerFactory(genFunc, genProdConf.WithLogger(gtLogger))

	// We are waiting until the outStream
	// will be closed by the Generator
	var dataCounter = 0
	for v := range genProducer.GenerateToStream(ctx) {
		dataCounter += 1
		gtLogger.Trace("Got value: %+v", v)
	}

	gtLogger.Info("Data was fetched: %d times", dataCounter)

}

// GenerateToStream data from slice Infinitely
func TestExample_SliceProducer(t *testing.T) {
	defer goleak.VerifyNone(t)
	// Slice generator
	data := []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}
	genFunc := SliceGeneratorFuncFactory(data)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		interruptCtx, interruptCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer interruptCancel()
		<-interruptCtx.Done()
		cancel()
	}()

	// Create the generator handler
	// It should generate values from data slice
	genProducer := GeneratorProducerFactory(genFunc, genProdConf.WithLogger(gtLogger))

	// We are waiting until the outStream
	// will be closed by the Generator
	var dataCounter = 0
	for v := range genProducer.GenerateToStream(ctx) {
		dataCounter += 1
		gtLogger.Trace("Got value: %+v", v)
	}

	gtLogger.Info("Data was fetched: %d times", dataCounter)

}

// GenerateToStream data from slice a time only
func TestExample_SliceProducerOnce(t *testing.T) {
	defer goleak.VerifyNone(t)
	// Slice generator
	data := []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}
	genFunc := SliceGeneratorFuncFactory(data)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the generator handler
	// It should generate values from data slice
	// one time only
	genProducer := GeneratorProducerFactory(genFunc, genProdConf.WithLogger(gtLogger),
		genProdConf.WithTimesToGenerate(uint(len(data))))

	// We are waiting until the outStream
	// will be closed by the Generator
	var dataCounter = 0
	for v := range genProducer.GenerateToStream(ctx) {
		dataCounter += 1
		gtLogger.Trace("Got value: %+v", v)
	}

	gtLogger.Info("Data was fetched: %d times", dataCounter)

}

// GenerateToStream data from slice a time only
// with middleware
func TestExample_SliceProducerOnceWithTraceHandlerFactoryExample(t *testing.T) {
	defer goleak.VerifyNone(t)
	// Slice generator
	data := []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}
	genFunc := SliceGeneratorFuncFactory(data)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the generator handler
	// It should generate values from data slice
	// one time only
	name := "SliceGeneratorExample"
	genProducer := GeneratorProducerFactory(WithTraceHandlerFactory(name, gtLogger, genFunc),
		genProdConf.WithName(name),
		genProdConf.WithLogger(gtLogger),
		genProdConf.WithTimesToGenerate(uint(len(data))))

	// We are waiting until the outStream
	// will be closed by the Generator
	var dataCounter = 0
	for v := range genProducer.GenerateToStream(ctx) {
		dataCounter += 1
		gtLogger.Trace("Got value: %+v", v)
	}

	gtLogger.Info("Data was fetched: %d times", dataCounter)

}
