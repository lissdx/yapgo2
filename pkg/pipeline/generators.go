package pipeline

import (
	"context"
	"errors"
)

var EmptySliceToGenerateFrom = errors.New("the Slice provided to the SliceGenerator function is empty")

//type GeneratorType int
//
//const (
//	SliceToStreamGenerator GeneratorType = iota
//	MaxGeneratorType
//)
//
//func (t GeneratorType) IsValid() bool {
//	return t >= 0 && t < MaxGeneratorType
//}

// type SliceToStreamGeneratorFn[S ~[]T, T any] func(context.Context, S) ReadOnlyStream[T]
// type FnStreamGeneratorFn[T any] func(context.Context, GeneratorFn[T]) ReadOnlyStream[T]

type GeneratorFn[T any] func() T
type ToStreamGenerator[T, RS any] func(context.Context, T) ReadOnlyStream[RS]

func (tsg ToStreamGenerator[T, RS]) Run(ctx context.Context, t T) ReadOnlyStream[RS] {
	return tsg(ctx, t)
}

// GeneratorFnToStreamGeneratorFactory is the main approach of the data generation
// We are going to use the GeneratorFn to generate any kind of data
// see the generators_example_test and learn how we can use it
// The data may be generated in 2 modes:
//  1. GenerateInfinitely - means never ended generation
//  2. Generate N data-items only - in this case we should provide
//     the N via GenConfigOption.WithTimesToGenerate
//
// Example:
// genStage := GeneratorFnToStreamGeneratorFactory[GeneratorFn[int]](WithTimesToGenerate(7))
// outStream := genStage.Run(ctx, func () int {return rand.Intn(100)})
// ...
// NOTE: WithTimesToGenerate(0) means GenerateInfinitely mode
func GeneratorFnToStreamGeneratorFactory[GF GeneratorFn[T], T any](options ...GenConfigOption) ToStreamGenerator[GF, T] {
	genConfig := generatorConfig{}
	for _, option := range options {
		option.apply(&genConfig)
	}

	return func(ctx context.Context, genFunc GF) ReadOnlyStream[T] {
		outStream := make(chan T)
		go func() {
			maxTimesToGenerate := genConfig.timesToGenerate
			defer close(outStream)
			for genConfig.timesToGenerate == GenerateInfinitely || maxTimesToGenerate > 0 {
				select {
				case <-ctx.Done():
					return
				case outStream <- genFunc():
				}
				if maxTimesToGenerate > 0 {
					maxTimesToGenerate -= 1
				}
			}
		}()
		return outStream
	}
}

// SliceToStreamOnePassGeneratorFactory
// simple wrapper of GeneratorFnToStreamGeneratorFactory
// generates data by len of the given slice and returns
// the slice one by one element
func SliceToStreamOnePassGeneratorFactory[S ~[]T, T any]() ToStreamGenerator[S, T] {
	return func(ctx context.Context, vData S) ReadOnlyStream[T] {
		genFunc := GeneratorFnFromSliceFactory[S, T](vData)
		return GeneratorFnToStreamGeneratorFactory[GeneratorFn[T]](WithTimesToGenerate(uint(len(vData)))).
			Run(ctx, genFunc)
	}
}

// SliceToStreamInfinitelyGeneratorFactory
// simple wrapper of GeneratorFnToStreamGeneratorFactory
// generates data from the given slice and returns
// the elements of the slice repetitively
func SliceToStreamInfinitelyGeneratorFactory[S ~[]T, T any]() ToStreamGenerator[S, T] {
	return func(ctx context.Context, vData S) ReadOnlyStream[T] {
		genFunc := GeneratorFnFromSliceFactory[S, T](vData)
		return GeneratorFnToStreamGeneratorFactory[GeneratorFn[T]]().
			Run(ctx, genFunc)
	}
}

// SliceToStreamNGeneratorFactory
// simple wrapper of GeneratorFnToStreamGeneratorFactory
// generates data from the given slice and returns
// the elements N times (maybe repetitively in case of N > len(given_slice))
func SliceToStreamNGeneratorFactory[S ~[]T, T any](n uint) ToStreamGenerator[S, T] {
	return func(ctx context.Context, vData S) ReadOnlyStream[T] {
		genFunc := GeneratorFnFromSliceFactory[S, T](vData)
		return GeneratorFnToStreamGeneratorFactory[GeneratorFn[T]](WithTimesToGenerate(n)).
			Run(ctx, genFunc)
	}
}

// FlatSlicesToStreamFnFactory flats slices []T given by inStream
// into flat data T and sends them one by one via outStream
//func FlatSlicesToStreamFnFactory[IS ReadOnlyStream[S], S ~[]T, T any]() ToStreamGenerator[IS, T] {
//	return func(ctx context.Context, inStream IS) ReadOnlyStream[T] {
//		outStream := make(chan T)
//		go func() {
//			defer close(outStream)
//			for {
//				select {
//				case <-ctx.Done():
//					return
//				case h, ok := <-inStream:
//					if !ok {
//						return
//					}
//					select {
//					case <-ctx.Done():
//						return
//					default:
//						for _, v := range h {
//							select {
//							case <-ctx.Done():
//								return
//							case outStream <- v:
//							}
//						}
//					}
//				}
//				runtime.Gosched()
//			}
//		}()
//		return outStream
//	}
//}

//func MergeChannels(doneCh ReadOnlyStream, inChans []ReadOnlyStream) ReadOnlyStream {
//	var wg sync.WaitGroup
//	wg.Add(len(inChans))
//
//	outChan := make(chan interface{})
//	for _, inChan := range inChans {
//		go func(ch <-chan interface{}) {
//			defer wg.Done()
//			for obj := range OrDone(doneCh, ch) {
//				select {
//				case <-doneCh:
//					return
//				case outChan <- obj:
//				}
//
//			}
//		}(inChan)
//	}
//
//	go func() {
//		defer close(outChan)
//		wg.Wait()
//	}()
//	return outChan
//}

//func OrDone[T any](ctx context.Context, inStream ReadOnlyStream[T]) ReadOnlyStream[T] {
//	outStream := make(chan T)
//	go func() {
//		defer close(outStream)
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			case v, ok := <-inStream:
//				if !ok {
//					return
//				}
//				select {
//				case <-ctx.Done():
//					return
//				case outStream <- v:
//				}
//			}
//		}
//	}()
//	return outStream
//}

/**
 *  Helpers part
 */

// GeneratorFnFromSliceFactory simple "Infinitely" item-generator
// from the given slice of Data
func GeneratorFnFromSliceFactory[S ~[]T, T any](s S) GeneratorFn[T] {
	if len(s) <= 0 {
		panic(EmptySliceToGenerateFrom)
	}
	currentIndx := 0
	return func() T {
		defer func() {
			currentIndx = (currentIndx + 1) % len(s)
		}()

		return s[currentIndx]
	}
}
