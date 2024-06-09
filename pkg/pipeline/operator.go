package pipeline

// Generator The generator function takes in a variadic slice of interface{},
// constructs a buffered channel of integers with a length equal to the incoming
// integer slice, starts a goroutine, and returns the constructed channel.
// Then, on the goroutine that was created, generator ranges over the variadic slice that was passed
// in and sends the slices’ values on the channel it created.
// Note - send on the channel shares a select statement with a selection on the ctx.Done().
// Again, this is the pattern we established in “Preventing Goroutine Leaks” to guard against leaking goroutines.
//func Generator[IN any](ctx context.Context, values ...IN) ReadOnlyStream[IN] {
//	outStream := make(chan IN, len(values))
//	go func() {
//		defer close(outStream)
//		for _, v := range values {
//			select {
//			case <-ctx.Done():
//				return
//			case outStream <- v:
//			}
//		}
//	}()
//	return outStream
//}

//
//// Repeat
//// This function will repeat the values you pass to created channel infinitely until
//// you tell it to stop.
//func Repeat(doneCh ReadOnlyStream, values ...interface{}) ReadOnlyStream {
//	valueStream := make(chan interface{})
//	go func() {
//		defer close(valueStream)
//		for {
//			for _, v := range values {
//				select {
//				case <-doneCh:
//					return
//				case valueStream <- v:
//				}
//			}
//		}
//	}()
//	return valueStream
//}
//
//// RepeatFn like Repeat but
//// this function will repeat the values generated by function
//func RepeatFn(doneCh ReadOnlyStream, fn func() interface{}) ReadOnlyStream {
//	valueStream := make(chan interface{})
//	go func() {
//		defer close(valueStream)
//		for {
//			select {
//			case <-doneCh:
//				return
//			case valueStream <- fn():
//			}
//		}
//	}()
//	return valueStream
//}
//
//// Take
//// This pipeline stage will only take the first num items off of
//// its incoming valueStream and then exit
//// maxTimesTake num of taking times. If maxTimesTake <=
//func Take(doneCh ReadOnlyStream, inStream ReadOnlyStream, maxTimesTake int) ReadOnlyStream {
//	valueStream := make(chan interface{})
//	go func() {
//		defer close(valueStream)
//		if maxTimesTake <= 0 {
//			return
//		}
//		for i := 0; i < maxTimesTake; i++ {
//			select {
//			case <-doneCh:
//				return
//			case h, ok := <-inStream:
//				if !ok {
//					return
//				}
//				select {
//				case <-doneCh:
//					return
//				case valueStream <- h:
//				}
//			}
//		}
//	}()
//	return valueStream
//}
//
//// OrDone
//// wrap our read from the channel with a select statement that
//// also selects from a done channel.
//// for val := range orDone(done, myChan) {
////   ...Do something with val
//// }
//func OrDone(doneCh, inStream ReadOnlyStream) ReadOnlyStream {
//	valStream := make(chan interface{})
//	go func() {
//		defer close(valStream)
//		for {
//			select {
//			case <-doneCh:
//				return
//			case v, ok := <-inStream:
//				if !ok {
//					return
//				}
//				select {
//				case valStream <- v:
//				case <-doneCh:
//					return // I've added return here
//				}
//			}
//		}
//	}()
//	return valStream
//}
//
//// Tee split values coming in from a channel so
//// that you can send them off into two separate areas of your codebase.
//func Tee(doneCh, inStream ReadOnlyStream) (ReadOnlyStream, ReadOnlyStream) {
//	out1 := make(chan interface{})
//	out2 := make(chan interface{})
//	go func() {
//		defer close(out1)
//		defer close(out2)
//		for val := range OrDone(doneCh, inStream) {
//			var hOut1, hOut2 = out1, out2
//			for i := 0; i < 2; i++ {
//				select {
//				case <-doneCh:
//					return // I've added return here
//				case hOut1 <- val:
//					hOut1 = nil
//				case hOut2 <- val:
//					hOut2 = nil
//				}
//			}
//		}
//	}()
//	return out1, out2
//}
//
//// MergeChannels merges an array of channels into a single channel. This utility
//// function can also be used independently outside a pipeline.
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
//
////func TeeTo(doneCh, inStream ReadOnlyStream, outStreamLeft, outStreamRight BidirectionalStream)(ReadOnlyStream, ReadOnlyStream){
////	go func() {
////		for val := range OrDone(doneCh, inStream) {
////			var hOutL, hOutR = outStreamLeft, outStreamRight
////			for i := 0; i < 2; i++ {
////				select {
////				case <- doneCh:
////					return // I've added return here
////				case hOutL <- val:
////					hOutL = nil
////				case hOutR <- val:
////					hOutR = nil
////				}
////			}
////		}
////	}()
////	return outStreamLeft, outStreamRight
////}
