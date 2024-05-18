package pipeline

import (
	"log"
	"time"
)

func GeneratorLoggerMiddlewareExample[T any](name string, next GeneratorHandler[T]) GeneratorHandler[T] {
	return func() GeneratorHandlerFunc[T] {
		return func() (T, bool) {
			defer simpleLogger(name)()
			return next.GenerateIt()
		}
	}()
}

func GeneratorTracerMiddlewareExample[T any](name string, next GeneratorHandler[T]) GeneratorHandler[T] {
	return func() GeneratorHandlerFunc[T] {
		return func() (T, bool) {
			defer measureTime(name)()
			return next.GenerateIt()
		}
	}()
}

func measureTime(process string) func() {
	log.Default().Println("Start", process)
	start := time.Now()
	return func() {
		log.Default().Printf("Time taken by %s is %v\n", process, time.Since(start))
	}
}

func simpleLogger(process string) func() {
	log.Default().Println("Start", process)
	return func() {
		log.Default().Println("End", process)
	}
}
