package pipeline

//
//import (
//	"log"
//	"time"
//)
//
//func GeneratorLoggerMiddlewareExample[IN any](name string, next GeneratorHandler[IN]) GeneratorHandler[IN] {
//	return func() GeneratorHandlerFunc[IN] {
//		return func() (IN, bool) {
//			defer simpleLogger(name)()
//			return next.GenerateIt()
//		}
//	}()
//}
//
//func GeneratorTracerMiddlewareExample[IN any](name string, next GeneratorHandler[IN]) GeneratorHandler[IN] {
//	return func() GeneratorHandlerFunc[IN] {
//		return func() (IN, bool) {
//			defer measureTime(name)()
//			return next.GenerateIt()
//		}
//	}()
//}
//
//func measureTime(process string) func() {
//	log.Default().Println("Start", process)
//	start := time.Now()
//	return func() {
//		log.Default().Printf("Time taken by %s is %v\n", process, time.Since(start))
//	}
//}
//
//func simpleLogger(process string) func() {
//	log.Default().Println("Start", process)
//	return func() {
//		log.Default().Println("End", process)
//	}
//}
