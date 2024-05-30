package tests

import (
	"context"
	"github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_producer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"math/rand"
	"slices"
	"testing"
	"time"

	pl "github.com/lissdx/yapgo2/pkg/pipeline"
)

func Test_SliceGenerator(t *testing.T) {
	defer goleak.VerifyNone(t)

	tests := []struct {
		name        string
		getGenFn    pl.GeneratorFuncHandler[int]
		contextFunc func() (context.Context, context.CancelFunc)
		want        []int
		options     []config_producer.Option
		wantPanic   bool
		checkResult bool
	}{
		{name: "[generate from slice]: nil slice given ", getGenFn: func() pl.GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return pl.SliceGeneratorFuncFactory([]int(nil))
		}(),
			contextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			wantPanic: false, want: []int{0}},
		{name: "[generate from slice]: an empty slice given", getGenFn: func() pl.GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return pl.SliceGeneratorFuncFactory([]int{})
		}(),
			contextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			wantPanic: false, want: []int{0}},
		{name: "[generate from slice]: one element given", getGenFn: func() pl.GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return pl.SliceGeneratorFuncFactory([]int{777})
		}(),
			contextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Microsecond*1000)
			},
			wantPanic: false, want: []int{777}},
		{name: "[generate from slice]: 2 elements given", getGenFn: func() pl.GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return pl.SliceGeneratorFuncFactory([]int{777, 999})
		}(), contextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithTimeout(context.Background(), time.Microsecond*1000)
		},
			wantPanic: false, want: []int{777, 999}},
		{name: "[generate from slice]: more than 2 elements", getGenFn: func() pl.GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return pl.SliceGeneratorFuncFactory([]int{-1, 888, 1001, 1002, 777, 999})
		}(), contextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithTimeout(context.Background(), time.Microsecond*1000)
		},
			wantPanic: false, want: []int{-1, 777, 888, 999, 1001, 1002}},
		{name: "[generate from slice]: generate given slice once, nil slice", getGenFn: func() pl.GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return pl.SliceGeneratorFuncFactory([]int(nil))
		}(), contextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithTimeout(context.Background(), time.Microsecond*1000)
		},
			// so 0 is param of WithTimesToGenerate - means run the generator forever
			options:   []config_producer.Option{config_producer.WithTimesToGenerate(uint(len([]int(nil))))},
			wantPanic: false, want: []int{0}},
		{name: "[generate from slice]: generate given slice once, one element", getGenFn: func() pl.GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return pl.SliceGeneratorFuncFactory([]int{999})
		}(), contextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
			options:     []config_producer.Option{config_producer.WithTimesToGenerate(uint(len([]int{999})))},
			checkResult: true,
			wantPanic:   false, want: []int{999}},
		{name: "[generate from slice]: generate given slice once, 2 elements", getGenFn: func() pl.GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return pl.SliceGeneratorFuncFactory([]int{999, -1})
		}(), contextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
			options:     []config_producer.Option{config_producer.WithTimesToGenerate(uint(len([]int{999, -1})))},
			checkResult: true,
			wantPanic:   false, want: []int{999, -1}},
		{name: "[generate from slice]: generate given slice once, more than 2 elements", getGenFn: func() pl.GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return pl.SliceGeneratorFuncFactory([]int{999, -1, 777, 3, 0, 777})
		}(), contextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
			options:     []config_producer.Option{config_producer.WithTimesToGenerate(uint(len([]int{999, -1, 777, 3, 0, 777})))},
			checkResult: true,
			wantPanic:   false, want: []int{999, -1, 777, 3, 0, 777}},
		{name: "[generate from slice]: generate given slice more than once, more than 2 elements", getGenFn: func() pl.GeneratorFuncHandler[int] {
			time.Sleep(time.Microsecond * 10)
			return pl.SliceGeneratorFuncFactory([]int{999, -1, 777, -1})
		}(), contextFunc: func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
			options:     []config_producer.Option{config_producer.WithTimesToGenerate(uint(11))},
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
			generatorHandler := pl.GeneratorProducerFactory(tt.getGenFn, tt.options...)
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
		getGenFn    pl.GeneratorFuncHandler[string]
		contextFunc func() (context.Context, context.CancelFunc)
		want        []string
		options     []config_producer.Option
		wantPanic   bool
		checkResult bool
	}{
		{name: "[generate from custom function]: without check result", getGenFn: func() pl.GeneratorFuncHandler[string] {
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
			generatorHandler := pl.GeneratorProducerFactory(tt.getGenFn, tt.options...)
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
		getGenFn    pl.GeneratorFuncHandler[any]
		contextFunc func() (context.Context, context.CancelFunc)
		want        []any
		options     []config_producer.Option
		wantPanic   bool
		checkResult bool
	}{
		{name: "[generate from custom function]: without check result", getGenFn: func() pl.GeneratorFuncHandler[any] {
			time.Sleep(time.Microsecond * 10)
			return pl.AValueGeneratorFuncFactory[any](nil)
		}(),
			contextFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			wantPanic: false,
			options:   []config_producer.Option{config_producer.WithTimesToGenerate(5)},
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
			generatorHandler := pl.GeneratorProducerFactory(tt.getGenFn, tt.options...)
			var got []any
			valStream := generatorHandler.GenerateToStream(ctxWithCancel)

			for num := range valStream {
				got = append(got, num)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
