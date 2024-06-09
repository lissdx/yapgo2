package pipeline

import (
	"context"
	"github.com/lissdx/yapgo2/pkg/logger"
	"github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_producer"
	config_stream_handler "github.com/lissdx/yapgo2/pkg/pipeline/config/stream_handler"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"slices"
	"testing"
	"time"
)

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

			genHandler := GeneratorProducerFactory(SliceGeneratorFuncFactory(tt.dataList), tt.options...)
			dataStream := genHandler.GenerateToStream(ctx)

			var got []int
			for num := range OrDoneFnFactory[int](config_stream_handler.WithLogger(htLogger)).Run(context.Background(), dataStream) {
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

			genHandler := GeneratorProducerFactory(SliceGeneratorFuncFactory(tt.dataList), tt.options...)
			dataStream := genHandler.GenerateToStream(genContext)

			var got []int
			orDoneCtx, orDoneCtxCancel := tt.orDoneContextFunc()
			defer orDoneCtxCancel()
			for num := range OrDoneFnFactory[int](config_stream_handler.WithLogger(htLogger)).Run(orDoneCtx, dataStream) {
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
		genContextFunc func() (context.Context, context.CancelFunc)
		dataList       [][]int
		want           []int
	}{
		{name: "oneInt",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			dataList: [][]int{{1}}, want: []int{1}},
		{name: "twoInt",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			dataList: [][]int{{-1, 0}}, want: []int{-1, 0}},
		{name: "multiInt", genContextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
			dataList: [][]int{{1}, {-1, 0}, {777, 555, 999, 999, 999, 999, 999, 999}},
			want:     []int{-1, 0, 1, 555, 777, 999, 999, 999, 999, 999, 999},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := tt.genContextFunc()
			defer cancelFn()

			dataStreams := make([]ReadOnlyStream[int], 0, len(tt.dataList))
			for i := 0; i < len(tt.dataList); i++ {
				genHandler := GeneratorProducerFactory(SliceGeneratorFuncFactory(tt.dataList[i]),
					config_producer.WithTimesToGenerate(uint(len(tt.dataList[i]))))
				dataStream := genHandler.GenerateToStream(ctx)
				dataStreams = append(dataStreams, dataStream)
			}

			resultDataStream := MergeFnFactory[[]ReadOnlyStream[int]](config_stream_handler.WithLogger(htLogger)).
				Run(ctx, dataStreams)

			var got []int
			for v := range resultDataStream {
				got = append(got, v)
			}

			slices.Sort(got)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMergeFnFactoryCloseByContext(t *testing.T) {
	tests := []struct {
		name             string
		genContextFunc   func() (context.Context, context.CancelFunc)
		mergeContextFunc func() (context.Context, context.CancelFunc)
		dataList         [][]int
		checkResult      bool
		want             []int
	}{
		{name: "oneInt",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			mergeContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			dataList: [][]int{{1}}, want: []int{1}},
		{name: "twoInt",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			mergeContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			dataList: [][]int{{-1, 0}},
			want:     []int{-1, 0},
		},
		{name: "multiInt",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			mergeContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			dataList: [][]int{{1}, {-1, 0}, {777, 555, 999, 999, 999, 999, 999, 999}},
			want:     []int{-1, 0, 1, 555, 777, 999},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := tt.genContextFunc()
			defer cancelFn()

			dataStreams := make([]ReadOnlyStream[int], 0, len(tt.dataList))
			for i := 0; i < len(tt.dataList); i++ {
				genHandler := GeneratorProducerFactory(SliceGeneratorFuncFactory(tt.dataList[i]))
				dataStream := genHandler.GenerateToStream(ctx)
				dataStreams = append(dataStreams, dataStream)
			}

			mCtx, mCancelFn := tt.mergeContextFunc()
			defer mCancelFn()
			resultDataStream := MergeFnFactory[[]ReadOnlyStream[int]](config_stream_handler.WithLogger(htLogger)).
				Run(mCtx, dataStreams)

			var got []int
			for v := range resultDataStream {
				got = append(got, v)
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

func TestFlatSlicesToStreamFnFactoryCloseByChannel(t *testing.T) {
	tests := []struct {
		name           string
		genContextFunc func() (context.Context, context.CancelFunc)
		dataList       [][]int
		want           []int
	}{
		{name: "oneInt one slice",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			dataList: [][]int{{1}},
			want:     []int{1},
		},
		{name: "2 int one slice",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			dataList: [][]int{{1, 2}},
			want:     []int{1, 2},
		},
		{name: "1 int 3 slices",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			dataList: [][]int{{1}, {2}, {3}},
			want:     []int{1, 2, 3},
		},
		{name: "2 int 3 slices",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			dataList: [][]int{{1, 2}, {3, 4}, {5, 6}},
			want:     []int{1, 2, 3, 4, 5, 6},
		},
		{name: "multi int multi slices",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			dataList: [][]int{{1, 2, 7, 100}, {3, 4}, {777, 5, 6, 999, 101}},
			want:     []int{1, 2, 7, 100, 3, 4, 777, 5, 6, 999, 101},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := tt.genContextFunc()
			defer cancelFn()

			genHandler := GeneratorProducerFactory(SliceGeneratorFuncFactory(tt.dataList),
				config_producer.WithTimesToGenerate(uint(len(tt.dataList))))
			dataStream := genHandler.GenerateToStream(ctx)

			resultDataStream := FlatSlicesToStreamFnFactory[[]int](config_stream_handler.WithLogger(htLogger)).
				Run(ctx, dataStream)

			var got []int
			for v := range resultDataStream {
				got = append(got, v)
			}

			//slices.Sort(got)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFlatSlicesToStreamFnFactoryCloseByContext(t *testing.T) {
	tests := []struct {
		name           string
		genContextFunc func() (context.Context, context.CancelFunc)
		fsContextFunc  func() (context.Context, context.CancelFunc)
		dataList       [][]int
		checkResult    bool
		want           []int
	}{
		{name: "oneInt one slice",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			fsContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			dataList: [][]int{{1}},
			want:     []int{1},
		},
		{name: "2 int one slice",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			fsContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			dataList: [][]int{{1, 2}},
			want:     []int{1, 2},
		},
		{name: "1 int 3 slices",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			fsContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			dataList: [][]int{{1}, {2}, {3}},
			want:     []int{1, 2, 3},
		},
		{name: "2 int 3 slices",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			fsContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			dataList: [][]int{{1, 2}, {3, 4}, {5, 6}},
			want:     []int{1, 2, 3, 4, 5, 6},
		},
		{name: "multi int multi slices",
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			fsContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			dataList: [][]int{{1, 2, 7, 100}, {3, 4}, {777, 5, 6, 999, 101}},
			want:     []int{1, 2, 3, 4, 5, 6, 7, 100, 101, 777, 999},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := tt.genContextFunc()
			defer cancelFn()

			genHandler := GeneratorProducerFactory(SliceGeneratorFuncFactory(tt.dataList))
			dataStream := genHandler.GenerateToStream(ctx)

			fsCtx, fsCancelFn := tt.fsContextFunc()
			defer fsCancelFn()
			resultDataStream := FlatSlicesToStreamFnFactory[[]int](config_stream_handler.WithLogger(htLogger)).
				Run(fsCtx, dataStream)

			var got []int
			for v := range resultDataStream {
				got = append(got, v)
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
