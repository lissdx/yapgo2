package tests

import (
	"context"
	"github.com/lissdx/yapgo2/pkg/logger"
	pl "github.com/lissdx/yapgo2/pkg/pipeline"
	"github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_producer"
	config_stream_handler "github.com/lissdx/yapgo2/pkg/pipeline/config/stream_handler"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"slices"
	"testing"
	"time"
)

//
//import (
//	"context"
//	pl "github.com/lissdx/yapgo2/pkg/pipeline"
//	genHandlerConf "github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_generator_handler"
//	"github.com/stretchr/testify/assert"
//	"go.uber.org/goleak"
//	"slices"
//	"testing"
//	"time"
//)

var htLogger = func() logger.ILogger {
	return logger.LoggerFactory(
		logger.WithZapLoggerImplementer(),
		logger.WithLoggerLevel("DEBUG"),
		logger.WithZapColored(),
		logger.WithZapConsoleEncoding(),
		logger.WithZapColored(),
	)
}()

func TestOrDoneCloseByChannel(t *testing.T) {
	defer goleak.VerifyNone(t)

	tests := []struct {
		name           string
		dataList       []int
		options        []config_producer.Option
		genContextFunc func() (context.Context, context.CancelFunc)
		checkResult    bool
		want           []int
	}{
		{name: "oneInt Infinitely",
			dataList: []int{1},
			want:     []int{1},
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
		},
		{name: "multiInt Infinitely",
			dataList: []int{1, -1, 2, 777, 999, 7, 777, -1},
			want:     []int{-1, 1, 2, 7, 777, 999},
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
		},
		{name: "multiInt + WithTimesToGenerate",
			dataList: []int{1, -1, 2, 777, 999, 7, 777, -1},
			want:     []int{-1, 1, 2, 7, 777, 999},
			options:  []config_producer.Option{config_producer.WithTimesToGenerate(uint(20))},
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := tt.genContextFunc()
			defer cancelFn()

			genHandler := pl.GeneratorProducerFactory(pl.SliceGeneratorFuncFactory(tt.dataList), tt.options...)
			dataStream := genHandler.GenerateToStream(ctx)

			var got []int
			for num := range pl.OrDoneFnFactory[int](config_stream_handler.WithLogger(htLogger)).Run(context.Background(), dataStream) {
				got = append(got, num)
			}

			res := func() []int {
				if tt.checkResult {
					return got
				}
				slices.Sort(got)
				return slices.Compact(got)
			}()

			assert.Equal(t, tt.want, res)
		})
	}
}

func TestOrDoneCloseByContext(t *testing.T) {
	defer goleak.VerifyNone(t)

	tests := []struct {
		name              string
		dataList          []int
		options           []config_producer.Option
		orDoneContextFunc func() (context.Context, context.CancelFunc)
		checkResult       bool
		want              []int
	}{
		{name: "oneInt Infinitely",
			dataList: []int{1},
			want:     []int{1},
			orDoneContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
		},
		{name: "oneInt Infinitely",
			dataList: []int{1, -1, 2, 777, 999, 7, 777, -1},
			want:     []int{-1, 1, 2, 7, 777, 999},
			orDoneContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			genContext, genContextCancel := context.WithCancel(context.Background())

			genHandler := pl.GeneratorProducerFactory(pl.SliceGeneratorFuncFactory(tt.dataList), tt.options...)
			dataStream := genHandler.GenerateToStream(genContext)

			var got []int
			orDoneCtx, orDoneCtxCancel := tt.orDoneContextFunc()
			defer orDoneCtxCancel()
			for num := range pl.OrDoneFnFactory[int](config_stream_handler.WithLogger(htLogger)).Run(orDoneCtx, dataStream) {
				got = append(got, num)
			}

			// stop a generator after OrDone was stopped
			genContextCancel()

			res := func() []int {
				if tt.checkResult {
					return got
				}
				slices.Sort(got)
				return slices.Compact(got)
			}()

			assert.Equal(t, tt.want, res)
		})
	}
}

func TestMergeFnFactoryCloseByChannel(t *testing.T) {
	tests := []struct {
		name           string
		options        []config_producer.Option
		genContextFunc func() (context.Context, context.CancelFunc)
		dataList       [][]int
		want           []int
	}{
		{name: "oneInt", dataList: [][]int{{1}}, want: []int{1}},
		{name: "twoInt", dataList: [][]int{{-1, 0}}, want: []int{-1, 0}},
		{name: "multiInt", dataList: [][]int{{1}, {-1, 0}, {777, 555, 999, 999, 999, 999, 999, 999}}, want: []int{-1, 0, 1, 555, 777, 999}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := tt.genContextFunc()
			defer cancelFn()

			dataStreams := make([]pl.ReadOnlyStream[int], 0, len(tt.dataList))
			for i := 0; i < len(tt.dataList); i++ {
				genHandler := pl.GeneratorProducerFactory(pl.SliceGeneratorFuncFactory(tt.dataList[i]), tt.options...)
				dataStream := genHandler.GenerateToStream(ctx)
				dataStreams = append(dataStreams, dataStream)
			}

			resultDataStream := pl.MergeFnFactory[[]pl.ReadOnlyStream[int]]().Run(ctx, dataStreams)

			var got []int
			for num := range pl.OrDoneFnFactory[int]().GenerateToStream(ctx, resultDataStream) {
				got = append(got, num)
			}

			t.Logf("%s got: %v", t.Name(), got)
			slices.Sort(got)
			res := slices.Compact(got)

			assert.Equal(t, tt.want, res)
		})
	}
}

//func TestOrDoneContextWithTimeout(t *testing.T) {
//	tests := []struct {
//		name     string
//		dataList []int
//		want     []int
//	}{
//		{name: "oneInt", dataList: []int{1}, want: []int{1}},
//		{name: "twoInt", dataList: []int{-1, 0}, want: []int{-1, 0}},
//		{name: "multiInt", dataList: []int{777, 555, 999, 999, 999, 999, 999, 999}, want: []int{555, 777, 999}},
//	}
//
//	for _, tt := range tests {
//		t.GenerateToStream(tt.name, func(t *testing.T) {
//
//			ctx, cancelFn := context.WithTimeout(context.Background(), time.Microsecond*1000)
//			defer cancelFn()
//
//			genHandler := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList))
//			dataStream := pl.GeneratorStageFactory(genHandler).GenerateToStream(ctx)
//
//			var got []int
//			for num := range pl.OrDoneFnFactory[int]().GenerateToStream(ctx, dataStream) {
//				got = append(got, num)
//			}
//
//			t.Logf("%s got: %v", t.Name(), got)
//			slices.Sort(got)
//			res := slices.Compact(got)
//
//			assert.Equal(t, tt.want, res)
//		})
//	}
//}

//// Close channel
//func TestOrDoneReadFromClosedChannel(t *testing.T) {
//	tests := []struct {
//		name     string
//		dataList []int
//		want     []int
//	}{
//		{name: "oneInt", dataList: []int{1}, want: []int{}},
//		{name: "twoInt", dataList: []int{-1, 0}, want: []int{}},
//		{name: "multiInt", dataList: []int{777, 555, 999, 999, 999, 999, 999, 999}, want: []int{}},
//	}
//
//	for _, tt := range tests {
//		t.GenerateToStream(tt.name, func(t *testing.T) {
//
//			ctx, cancelFn := context.WithTimeout(context.Background(), time.Microsecond*1000)
//			defer cancelFn()
//
//			genHandler := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList))
//			dataStream := pl.GeneratorStageFactory(genHandler).GenerateToStream(ctx)
//
//			// drain and wait for
//			// channel closing
//			for range dataStream {
//				// do nothing
//			}
//
//			orDoneCtx, orDoneCancel := context.WithCancel(context.Background())
//			defer orDoneCancel()
//			var got = make([]int, 0)
//			for num := range pl.OrDoneFnFactory[int]().GenerateToStream(orDoneCtx, dataStream) {
//				got = append(got, num)
//			}
//
//			t.Logf("%s got: %v", t.Name(), got)
//			slices.Sort(got)
//			res := slices.Compact(got)
//
//			assert.Equal(t, tt.want, res)
//		})
//	}
//}

//// with closed channels
//func TestMergeFnFactory2(t *testing.T) {
//	tests := []struct {
//		name          string
//		dataList      [][]int
//		want          []int
//		timesGenerate int
//	}{
//		{name: "one", dataList: [][]int{{1}}, want: []int{1}},
//		{name: "one close", dataList: [][]int{{2}}, want: []int{}},
//		{name: "two", dataList: [][]int{{1}, {-1, 0}}, want: []int{-1, 0, 1}},
//		{name: "multi", timesGenerate: 4, dataList: [][]int{{1}, {-1, 0}, {777, 555, 888, 111, 33333, 999, 999, 999}}, want: []int{-1, 0, 111, 555, 777, 888}},
//	}
//
//	for _, tt := range tests {
//		t.GenerateToStream(tt.name, func(t *testing.T) {
//
//			ctx, cancelFn := context.WithTimeout(context.Background(), time.Millisecond*1000)
//			defer cancelFn()
//
//			dataStreams := make([]pl.ReadOnlyStream[int], 0, len(tt.dataList))
//
//			switch tt.name {
//			case "one":
//				genHandler := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList[0]))
//				dataStream := pl.GeneratorStageFactory(genHandler).GenerateToStream(ctx)
//				dataStreams = append(dataStreams, dataStream)
//			case "one close":
//				genHandler := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList[0]), genHandlerConf.WithTimesToGenerate(1000))
//				dataStream := pl.GeneratorStageFactory(genHandler).GenerateToStream(ctx)
//				dataStreams = append(dataStreams, dataStream)
//
//				waitFor1 := pl.NoOpSubFnFactory[int]()(ctx, dataStream)
//				<-waitFor1.Done()
//			case "two":
//				for i := 0; i < len(tt.dataList); i++ {
//					genHandler := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList[i]))
//					dataStream := pl.GeneratorStageFactory(genHandler).GenerateToStream(ctx)
//					dataStreams = append(dataStreams, dataStream)
//				}
//			case "multi":
//				mCtx1, mCancelFn := context.WithTimeout(context.Background(), time.Millisecond*10)
//				defer mCancelFn()
//				genHandler1 := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList[0]))
//				dataStream1 := pl.GeneratorStageFactory(genHandler1).GenerateToStream(mCtx1)
//				dataStreams = append(dataStreams, dataStream1)
//				// wait and close dataStream1
//				waitFor1 := pl.NoOpSubFnFactory[int]()(ctx, dataStream1)
//				select {
//				case <-waitFor1.Done():
//					t.Log("dataStream1 is closed")
//				}
//
//				// dataStream2 will be closed after 1 time generation of all elements
//				genHandler2 := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList[1]),
//					genHandlerConf.WithTimesToGenerate(1000))
//				dataStream2 := pl.GeneratorStageFactory(genHandler2).GenerateToStream(ctx)
//				dataStreams = append(dataStreams, dataStream2)
//
//				// dataStream2 will be closed after 1 time generation of all elements
//				genHandler3 := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList[2]),
//					genHandlerConf.WithTimesToGenerate(4))
//				dataStream3 := pl.GeneratorStageFactory(genHandler3).GenerateToStream(ctx)
//				dataStreams = append(dataStreams, dataStream3)
//			}
//
//			resultDataStream := pl.MergeFnFactory[[]pl.ReadOnlyStream[int]]().GenerateToStream(ctx, dataStreams)
//
//			t.Logf("%s data resiving", t.Name())
//			var got = make([]int, 0, 100)
//			for num := range pl.OrDoneFnFactory[int]().GenerateToStream(ctx, resultDataStream) {
//				got = append(got, num)
//			}
//
//			//log.Default().Println(got)
//			slices.Sort(got)
//			res := slices.Compact(got)
//
//			assert.Equal(t, tt.want, res)
//		})
//	}
//}
//
//func TestFlatSlicesToStreamFnFactory1(t *testing.T) {
//	tests := []struct {
//		name          string
//		args          [][]int
//		want          []int
//		timesGenerate int
//	}{
//		{name: "one", args: [][]int{{1}}, want: []int{1}},
//		{name: "two", args: [][]int{{1}, {-1, 0}}, want: []int{-1, 0, 1}},
//		{name: "multi", args: [][]int{{1}, {-1, 0}, {777, 555, 888, 111, 33333, 999, 999, 999}}, want: []int{-1, 0, 1, 111, 555, 777, 888, 999, 33333}},
//	}
//
//	for _, tt := range tests {
//		t.GenerateToStream(tt.name, func(t *testing.T) {
//
//			ctx, cancelFn := context.WithTimeout(context.Background(), time.Millisecond*300)
//			defer cancelFn()
//
//			genHandler := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.args))
//			dataStream := pl.GeneratorStageFactory(genHandler).GenerateToStream(ctx)
//			resultStream := pl.FlatSlicesToStreamFnFactory[[]int]().GenerateToStream(ctx, dataStream)
//
//			var got []int
//			for num := range pl.OrDoneFnFactory[int]().GenerateToStream(ctx, resultStream) {
//				got = append(got, num)
//			}
//
//			//log.Default().Println(got)
//			slices.Sort(got)
//			res := slices.Compact(got)
//
//			assert.Equal(t, tt.want, res)
//		})
//	}
//}
