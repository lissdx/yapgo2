package pipeline

import (
	"context"
	"github.com/lissdx/yapgo2/pkg/logger"
	confProducer "github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_producer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"math/rand"
	"slices"
	"testing"
	"time"
)

var gentLogger = func() logger.ILogger {
	return logger.LoggerFactory(
		logger.WithZapLoggerImplementer(),
		logger.WithLoggerLevel("DEBUG"),
		logger.WithZapColored(),
		logger.WithZapConsoleEncoding(),
		logger.WithZapColored(),
	)
}()

func Test_SliceGenerator(t *testing.T) {
	defer goleak.VerifyNone(t)

	tests := []struct {
		name        string
		getGenFn    GeneratorFuncHandler[int]
		contextFunc func() (context.Context, context.CancelFunc)
		want        []int
		options     []confProducer.Option
		wantPanic   bool
		checkResult bool
	}{
		{name: "[generate from slice]: nil slice given ", getGenFn: func() GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return SliceGeneratorFuncFactory([]int(nil))
		}(),
			contextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			wantPanic: false, want: []int{0}},
		{name: "[generate from slice]: an empty slice given", getGenFn: func() GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return SliceGeneratorFuncFactory([]int{})
		}(),
			contextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			wantPanic: false, want: []int{0}},
		{name: "[generate from slice]: one element given", getGenFn: func() GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return SliceGeneratorFuncFactory([]int{777})
		}(),
			contextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			wantPanic: false, want: []int{777}},
		{name: "[generate from slice]: 2 elements given", getGenFn: func() GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return SliceGeneratorFuncFactory([]int{777, 999})
		}(), contextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithTimeout(context.Background(), time.Microsecond*1000)
		},
			wantPanic: false, want: []int{777, 999}},
		{name: "[generate from slice]: more than 2 elements", getGenFn: func() GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return SliceGeneratorFuncFactory([]int{-1, 888, 1001, 1002, 777, 999})
		}(), contextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithTimeout(context.Background(), time.Microsecond*1000)
		},
			wantPanic: false, want: []int{-1, 777, 888, 999, 1001, 1002}},
		{name: "[generate from slice]: generate given slice once, nil slice", getGenFn: func() GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return SliceGeneratorFuncFactory([]int(nil))
		}(), contextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithTimeout(context.Background(), time.Microsecond*1000)
		},
			// so 0 is param of WithTimesToGenerate - means run the generator forever
			options:   []confProducer.Option{confProducer.WithTimesToGenerate(uint(len([]int(nil))))},
			wantPanic: false, want: []int{0}},
		{name: "[generate from slice]: generate given slice once, one element", getGenFn: func() GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return SliceGeneratorFuncFactory([]int{999})
		}(), contextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
			options:     []confProducer.Option{confProducer.WithTimesToGenerate(uint(len([]int{999})))},
			checkResult: true,
			wantPanic:   false, want: []int{999}},
		{name: "[generate from slice]: generate given slice once, 2 elements", getGenFn: func() GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return SliceGeneratorFuncFactory([]int{999, -1})
		}(), contextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
			options:     []confProducer.Option{confProducer.WithTimesToGenerate(uint(len([]int{999, -1})))},
			checkResult: true,
			wantPanic:   false, want: []int{999, -1}},
		{name: "[generate from slice]: generate given slice once, more than 2 elements", getGenFn: func() GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return SliceGeneratorFuncFactory([]int{999, -1, 777, 3, 0, 777})
		}(), contextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
			options:     []confProducer.Option{confProducer.WithTimesToGenerate(uint(len([]int{999, -1, 777, 3, 0, 777})))},
			checkResult: true,
			wantPanic:   false, want: []int{999, -1, 777, 3, 0, 777}},
		{name: "[generate from slice]: generate given slice more than once, more than 2 elements", getGenFn: func() GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return SliceGeneratorFuncFactory([]int{999, -1, 777, -1})
		}(), contextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
			options:     []confProducer.Option{confProducer.WithTimesToGenerate(uint(11))},
			checkResult: true,
			wantPanic:   false, want: []int{999, -1, 777, -1, 999, -1, 777, -1, 999, -1, 777}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("recover = %v, wantPanic = %v", r, tt.wantPanic)
				}
			}()

			ctxWithCancel, cancelFn := tt.contextFunc()
			defer cancelFn()
			generatorHandler := GeneratorProducerFactory(tt.getGenFn, tt.options...)
			var got []int
			valStream := generatorHandler.GenerateToStream(ctxWithCancel)

			for num := range valStream {
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

func Test_CustomFuncGenerator(t *testing.T) {
	defer goleak.VerifyNone(t)

	tests := []struct {
		name        string
		getGenFn    GeneratorFuncHandler[string]
		contextFunc func() (context.Context, context.CancelFunc)
		want        []string
		options     []confProducer.Option
		wantPanic   bool
		checkResult bool
	}{
		{name: "[generate from custom function]: without check result", getGenFn: func() GeneratorFuncHandler[string] {
			time.Sleep(time.Microsecond * 10)
			return func() string {
				sourceSlice := []string{"one", "two", "three", "four", "five", "six", "seven"}
				return sourceSlice[rand.Intn(len(sourceSlice))]
			}
		}(),
			contextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			wantPanic: false,
			want:      []string{"one", "two", "three", "four", "five", "six", "seven"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("recover = %v, wantPanic = %v", r, tt.wantPanic)
				}
			}()

			ctxWithCancel, cancelFn := tt.contextFunc()
			defer cancelFn()
			generatorHandler := GeneratorProducerFactory(tt.getGenFn, tt.options...)
			var got []string
			valStream := generatorHandler.GenerateToStream(ctxWithCancel)

			for num := range valStream {
				got = append(got, num)
			}

			res := func() []string {
				if tt.checkResult {
					return got
				}
				slices.Sort(got)
				slices.Sort(tt.want)
				return slices.Compact(got)
			}()

			assert.Equal(t, tt.want, res)
		})
	}
}

func Test_AValueGeneratorFunc(t *testing.T) {
	defer goleak.VerifyNone(t)

	tests := []struct {
		name        string
		getGenFn    GeneratorFuncHandler[any]
		contextFunc func() (context.Context, context.CancelFunc)
		want        []any
		options     []confProducer.Option
		wantPanic   bool
		checkResult bool
	}{
		{name: "[generate from custom function]: without check result", getGenFn: func() GeneratorFuncHandler[any] {
			time.Sleep(time.Microsecond * 10)
			return AValueGeneratorFuncFactory[any](nil)
		}(),
			contextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			wantPanic: false,
			options:   []confProducer.Option{confProducer.WithTimesToGenerate(5)},
			want:      []any{nil, nil, nil, nil, nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("recover = %v, wantPanic = %v", r, tt.wantPanic)
				}
			}()

			ctxWithCancel, cancelFn := tt.contextFunc()
			defer cancelFn()
			generatorHandler := GeneratorProducerFactory(tt.getGenFn, tt.options...)
			var got []any
			valStream := generatorHandler.GenerateToStream(ctxWithCancel)

			for num := range valStream {
				got = append(got, num)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_AValueGeneratorFuncFactory(t *testing.T) {

	type testCase[T any] struct {
		name                string
		aValueGeneratorFunc GeneratorFuncHandler[T]
		want                int
	}
	tests := []testCase[int]{
		{
			name:                "simple Once [1] generator",
			aValueGeneratorFunc: AValueGeneratorFuncFactory(1),
			want:                1,
		},
		{
			name:                "simple Once [-1] generator",
			aValueGeneratorFunc: AValueGeneratorFuncFactory(-1),
			want:                -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.aValueGeneratorFunc.GenerateData())
		})
	}
}

func Test_SliceGeneratorFuncFactory(t *testing.T) {

	type testCase[T any] struct {
		name               string
		sliceGeneratorFunc GeneratorFuncHandler[T]
		want               []int
	}
	tests := []testCase[int]{
		{
			name: "simple slice generator",
			sliceGeneratorFunc: func() GeneratorFuncHandler[int] {
				dataSlice := []int{1, 2, 3}
				return SliceGeneratorFuncFactory(dataSlice)
			}(),
			want: []int{1, 2, 3},
		},
		{
			name: "simple empty slice generator",
			sliceGeneratorFunc: func() GeneratorFuncHandler[int] {
				dataSlice := []int{}
				return SliceGeneratorFuncFactory(dataSlice)
			}(),
			want: []int{0},
		},
		{
			name: "simple nil slice generator",
			sliceGeneratorFunc: func() GeneratorFuncHandler[int] {
				dataSlice := []int(nil)
				return SliceGeneratorFuncFactory(dataSlice)
			}(),
			want: []int{0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := make([]int, 0, 100)

			for i := 0; i < 100; i++ {
				res = append(res, tt.sliceGeneratorFunc.GenerateData())
			}

			res = func() []int {
				slices.Sort(res)
				return slices.Compact(res)
			}()

			assert.Equal(t, tt.want, res)
		})
	}
}

func Test_GeneratorProducerFactoryWithTimesToGenerate(t *testing.T) {

	type testCase[T any] struct {
		name               string
		sliceGeneratorFunc GeneratorFuncHandler[T]
		genContextFunc     func() (context.Context, context.CancelFunc)
		options            []confProducer.Option
		normalizeResult    bool
		want               []int
	}
	tests := []testCase[int]{
		{
			name: "simple int producer",
			sliceGeneratorFunc: func() GeneratorFuncHandler[int] {
				dataSlice := []int{1, 2, 3}
				return SliceGeneratorFuncFactory(dataSlice)
			}(),
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			options: []confProducer.Option{confProducer.WithTimesToGenerate(uint(7)), confProducer.WithLogger(gentLogger)},
			want:    []int{1, 2, 3, 1, 2, 3, 1},
		},
		{
			name: "one int producer",
			sliceGeneratorFunc: func() GeneratorFuncHandler[int] {
				dataSlice := []int{1, 2, 3}
				return SliceGeneratorFuncFactory(dataSlice)
			}(),
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			options: []confProducer.Option{confProducer.WithTimesToGenerate(uint(1)), confProducer.WithLogger(gentLogger)},
			want:    []int{1},
		},
		{
			name: "WithTimesToGenerate == 0 int producer (Infinitely)",
			sliceGeneratorFunc: func() GeneratorFuncHandler[int] {
				dataSlice := []int{1, 2, 3}
				return SliceGeneratorFuncFactory(dataSlice)
			}(),
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Millisecond*1000)
			},
			options:         []confProducer.Option{confProducer.WithTimesToGenerate(0), confProducer.WithLogger(gentLogger)},
			normalizeResult: true,
			want:            []int{1, 2, 3},
		},
		{
			name: "WithTimesToGenerate == 10_000_000 int producer",
			sliceGeneratorFunc: func() GeneratorFuncHandler[int] {
				dataSlice := []int{1, 2, 3}
				return SliceGeneratorFuncFactory(dataSlice)
			}(),
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Millisecond*1)
			},
			options:         []confProducer.Option{confProducer.WithTimesToGenerate(1_000_000), confProducer.WithLogger(gentLogger)},
			normalizeResult: true,
			want:            []int{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := make([]int, 0, 100)
			ctx, cancelFn := tt.genContextFunc()
			defer cancelFn()

			genHandler := GeneratorProducerFactory(tt.sliceGeneratorFunc, tt.options...)

			for i := range genHandler.GenerateToStream(ctx) {
				res = append(res, i)
			}

			res = func() []int {
				if tt.normalizeResult {
					slices.Sort(res)
					return slices.Compact(res)
				}
				return res
			}()

			assert.Equal(t, tt.want, res)
		})
	}
}

func Test_GeneratorProducerFactoryWithOutTimesToGenerate(t *testing.T) {

	type testCase[T any] struct {
		name               string
		sliceGeneratorFunc GeneratorFuncHandler[T]
		genContextFunc     func() (context.Context, context.CancelFunc)
		options            []confProducer.Option
		normalizeResult    bool
		want               []int
	}
	tests := []testCase[int]{
		{
			name: "simple int producer",
			sliceGeneratorFunc: func() GeneratorFuncHandler[int] {
				dataSlice := []int{1, 2, 3, 4, -1, -5, 101}
				return SliceGeneratorFuncFactory(dataSlice)
			}(),
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Millisecond*1000)
			},
			options:         []confProducer.Option{confProducer.WithLogger(gentLogger)},
			normalizeResult: true,
			want:            []int{-5, -1, 1, 2, 3, 4, 101},
		},
		{
			name: "slice with a one element producer",
			sliceGeneratorFunc: func() GeneratorFuncHandler[int] {
				dataSlice := []int{1}
				return SliceGeneratorFuncFactory(dataSlice)
			}(),
			genContextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Millisecond*1000)
			},
			options:         []confProducer.Option{confProducer.WithLogger(gentLogger)},
			normalizeResult: true,
			want:            []int{1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := make([]int, 0, 100)
			ctx, cancelFn := tt.genContextFunc()
			defer cancelFn()

			genHandler := GeneratorProducerFactory(tt.sliceGeneratorFunc, tt.options...)

			for i := range genHandler.GenerateToStream(ctx) {
				res = append(res, i)
			}

			res = func() []int {
				if tt.normalizeResult {
					slices.Sort(res)
					return slices.Compact(res)
				}
				return res
			}()

			assert.Equal(t, tt.want, res)
		})
	}
}
