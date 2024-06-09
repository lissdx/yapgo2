package pipeline

import (
	"context"
	"errors"
	"fmt"
	"github.com/lissdx/yapgo2/pkg/logger"
	genConf "github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_producer"
	"golang.org/x/sync/errgroup"
	"math/rand"
	"time"
)

// GeneratorFunc should generate any
// kind of data
// returns:
// IN - generated data
type GeneratorFunc[T any] func() T
type GeneratorFuncHandler[T any] GeneratorFunc[T]

// ToStreamGeneratorFn func should be returned by the stage creation factory
// (in general means stage that ready to run it)
// after the "stage" is created we may start the stage as result we will
// get the "output stream" (channel) and read the generated data from the "output stream"
//type ToStreamGeneratorFn[IN any, ROS ReadOnlyStream[ProcessResultCarrier[IN]]] func(context.Context) ROS

// ToStreamGenerator interface
// implemented by GeneratorHandlerFunc
type ToStreamGenerator[T any, ROS ReadOnlyStream[T]] interface {
	GenerateToStream(context.Context) ROS
}

// ToStreamGenerateFunc - should wrap the GeneratorFunc
// the result of GeneratorFunc will be wrapped as ProcessResultCarrier
// and will be sent to the output stream
type ToStreamGenerateFunc[T any, ROS ReadOnlyStream[T]] func(ctx context.Context) ROS

func (gpf ToStreamGenerateFunc[T, ROS]) GenerateToStream(ctx context.Context) ROS {
	return gpf(ctx)
}

// DataGenerator interface
// implemented by GeneratorFunc
type DataGenerator[T any] interface {
	GenerateData() T
}

func (g GeneratorFuncHandler[T]) GenerateData() T {
	return g()
}

// GeneratorProducerFactory factory creates the ToStreamGenerateFunc
// if the WithTimesToGenerate option is provided we will wrap
// GeneratorFunc with special method (which calculates the generation times)
// if WithTimesToGenerate == 0 - means run the generator in the Infinitely mode
func GeneratorProducerFactory[T any, ROS ReadOnlyStream[T]](genFuncHandler GeneratorFuncHandler[T], options ...genConf.Option) ToStreamGenerator[T, ROS] {
	factoryName := "GeneratorProducerFactory"
	conf := genConf.NewGeneratorProducerConfig(options...)
	loggerPref := fmt.Sprintf("%s", conf.Name())

	switch conf.GenerateInfinitely() {
	case false:
		// Create the ToStreamGenerateFunc with TimesToGenerate option
		return ToStreamGenerateFunc[T, ROS](func(ctx context.Context) ROS {
			outStream := make(chan T)
			g, ctx := errgroup.WithContext(ctx)

			g.Go(func() error {
				conf.Logger().Info("%s: %s started with TimesToGenerate: %d", loggerPref, factoryName, conf.TimesToGenerate())
				conf.Logger().Info("%s: %s outStream created", loggerPref, factoryName)
				genCount := uint(0)
				for ; genCount < conf.TimesToGenerate(); genCount++ {
					select {
					case <-ctx.Done():
						conf.Logger().Debug("%s: Got <-ctx.Done(). %s sent %d items", loggerPref, factoryName, genCount)
						return fmt.Errorf("interrupted. Sent items: %d (insead of %d). context error: %w",
							genCount, conf.TimesToGenerate(), ctx.Err())
					case outStream <- genFuncHandler.GenerateData():
					}
				}

				conf.Logger().Debug("%s: %s successfully sent %d items", loggerPref, factoryName, genCount)
				conf.Logger().Debug("%s: %s input stream was closed", loggerPref, factoryName)

				return nil
			})

			go func() {
				defer close(outStream)
				if err := g.Wait(); err != nil {
					if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
						conf.Logger().Error("%s: %s error: %s", loggerPref, factoryName, err.Error())
					} else {
						conf.Logger().Warn("%s: %s interrupted: %s", loggerPref, factoryName, err.Error())
					}
				}
				conf.Logger().Info("%s: %s was stopped", loggerPref, factoryName)
				conf.Logger().Info("%s: %s close outStream", loggerPref, factoryName)
			}()

			return outStream
		})
	default:
		// Create the ToStreamGenerateFunc without TimesToGenerate option
		return ToStreamGenerateFunc[T, ROS](func(ctx context.Context) ROS {
			outStream := make(chan T)
			g, ctx := errgroup.WithContext(ctx)

			g.Go(func() error {
				conf.Logger().Info("%s: %s started with TimesToGenerate: ProduceInfinitely", loggerPref, factoryName)
				conf.Logger().Info("%s: %s outStream created", loggerPref, factoryName)
				for {
					select {
					case <-ctx.Done():
						conf.Logger().Debug("%s: Got <-ctx.Done(). %s was interrupted", loggerPref, factoryName)
						return ctx.Err()
					case outStream <- genFuncHandler.GenerateData():
					}
				}
			})

			go func() {
				defer close(outStream)
				if err := g.Wait(); err != nil {
					if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
						conf.Logger().Error("%s: %s error: %s", loggerPref, factoryName, err.Error())
					} else {
						conf.Logger().Info("%s: %s interruption cause: %s", loggerPref, factoryName, err.Error())
					}
				}
				conf.Logger().Info("%s: %s stopped", loggerPref, factoryName)
			}()

			return outStream
		})
	}
}

/*
 *
 * Simple Generator function part
 * Predefined generator examples
 *
 */

// SliceGeneratorFuncFactory creates the GeneratorFuncHandler
// which should repeatedly generates values
// from the given slice
// if given slice is empty or nil
// the generator will generate the default init value of the T type
func SliceGeneratorFuncFactory[S ~[]T, T any](s S) GeneratorFuncHandler[T] {

	// if the given slice is empty
	if len(s) <= 0 {
		return AValueGeneratorFuncFactory[T](func() (t T) { return }())
	}

	// generate the next value from the given slice
	// in case of the end of the slice
	// reset the index and start the generation from
	// the beginning
	return GeneratorFuncHandler[T](func() GeneratorFunc[T] {
		currentIndx := 0
		return func() T {
			defer func() {
				currentIndx = (currentIndx + 1) % len(s)
			}()

			return s[currentIndx]
		}
	}())
}

// AValueGeneratorFuncFactory generates
// the passed value only
func AValueGeneratorFuncFactory[T any](t T) GeneratorFuncHandler[T] {
	return GeneratorFuncHandler[T](func() GeneratorFunc[T] {
		return func() T {
			return t
		}
	}())
}

// RandomIntGeneratorFuncFactory simple random int generator
func RandomIntGeneratorFuncFactory(interval int) GeneratorFuncHandler[int] {
	if interval <= 0 {
		return func() int {
			return rand.Int()
		}
	}

	return func() int {
		return rand.Intn(interval)
	}
}

func WithTraceHandlerFactory[T any](name string, lg logger.ILogger, next GeneratorFuncHandler[T]) GeneratorFuncHandler[T] {
	return func() T {
		defer measureTime(name, lg)()

		return next.GenerateData()
	}
}

func measureTime(process string, lg logger.ILogger) func() {
	lg.Debug("Start %s", process)
	start := time.Now()
	return func() {
		lg.Debug("Time taken by %s is %v", process, time.Since(start))
	}
}
