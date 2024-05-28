package tests

//
//import (
//	"context"
//	pl "github.com/lissdx/yapgo2/pkg/pipeline"
//	confGenFunc "github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_generator_function"
//	confGenHandler "github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_generator_handler"
//	"github.com/stretchr/testify/assert"
//	"go.uber.org/goleak"
//	"slices"
//	"testing"
//	"time"
//)
//
//func TestMain(m *testing.M) {
//	goleak.VerifyTestMain(m)
//}
//
//func Test_GeneratorFnToStreamGenerator(t *testing.T) {
//	tests := []struct {
//		name      string
//		getGenFn  func() pl.GeneratorFunc[int]
//		want      []int
//		wantPanic bool
//	}{
//		{name: "nil", getGenFn: func() pl.GeneratorFunc[int] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]int(nil))
//		}, wantPanic: true},
//		{name: "nil with ignore empty slice", getGenFn: func() pl.GeneratorFunc[int] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]int(nil), confGenFunc.WithIgnoreEmptySlice(true))
//		}, wantPanic: false, want: []int{0}},
//		{name: "empty", getGenFn: func() pl.GeneratorFunc[int] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]int{})
//		}, wantPanic: true},
//		{name: "empty with ignore empty slice", getGenFn: func() pl.GeneratorFunc[int] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]int{}, confGenFunc.WithIgnoreEmptySlice(true))
//		}, wantPanic: false, want: []int{0}},
//		{name: "oneInt", getGenFn: func() pl.GeneratorFunc[int] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]int{1})
//		}, want: []int{1}},
//		{name: "oneInt WithIgnoreEmptySlice", getGenFn: func() pl.GeneratorFunc[int] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]int{1}, confGenFunc.WithIgnoreEmptySlice(true))
//		}, want: []int{1}},
//		{name: "twoInt", getGenFn: func() pl.GeneratorFunc[int] {
//			s := []int{1, -1}
//			return pl.SliceGeneratorFuncFactory(s)
//		}, want: []int{-1, 1}},
//		{name: "tenInt", getGenFn: func() pl.GeneratorFunc[int] {
//			s := []int{1, -1, 23, 100, 26, 33, 78, 33, 221, 0}
//			return pl.SliceGeneratorFuncFactory(s)
//		}, want: []int{-1, 0, 1, 23, 26, 33, 78, 100, 221}},
//		{name: "tenInt WithIgnoreEmptySlice", getGenFn: func() pl.GeneratorFunc[int] {
//			s := []int{1, -1, 23, 100, 26, 33, 78, 33, 221, 0}
//			return pl.SliceGeneratorFuncFactory(s, confGenFunc.WithIgnoreEmptySlice(true))
//		}, want: []int{-1, 0, 1, 23, 26, 33, 78, 100, 221}},
//	}
//
//	for _, tt := range tests {
//		t.GenerateToStream(tt.name, func(t *testing.T) {
//
//			defer func() {
//				r := recover()
//				if (r != nil) != tt.wantPanic {
//					t.Errorf("recover = %v, wantPanic = %v", r, tt.wantPanic)
//				}
//			}()
//
//			ctxWithCancel, cancelFn := context.WithTimeout(context.Background(), time.Microsecond*1000)
//			defer cancelFn()
//			generatorHandler := pl.GeneratorHandlerFactory[int](tt.getGenFn())
//			var got []int
//			valStream := pl.GeneratorStageFactory[int](generatorHandler).GenerateToStream(ctxWithCancel)
//
//			for num := range valStream {
//				got = append(got, num)
//			}
//
//			slices.Sort(got)
//			res := slices.Compact(got)
//
//			assert.Equal(t, tt.want, res)
//		})
//	}
//}
//
//func Test_NToStreamGenerator(t *testing.T) {
//	tests := []struct {
//		name            string
//		getGenFn        func() pl.GeneratorFunc[string]
//		timesToGenerate uint
//		want            []string
//		wantPanic       bool
//	}{
//		{name: "nil", getGenFn: func() pl.GeneratorFunc[string] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]string(nil))
//		}, wantPanic: true},
//		{name: "nil with ignore empty slice", getGenFn: func() pl.GeneratorFunc[string] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]string(nil), confGenFunc.WithIgnoreEmptySlice(true))
//		}, wantPanic: false, want: []string{""}, timesToGenerate: 10},
//		{name: "empty", getGenFn: func() pl.GeneratorFunc[string] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]string{})
//		}, wantPanic: true},
//		{name: "empty with ignore empty slice", getGenFn: func() pl.GeneratorFunc[string] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]string{}, confGenFunc.WithIgnoreEmptySlice(true))
//		}, wantPanic: false, want: []string{""}, timesToGenerate: 10},
//		{name: "oneStr", getGenFn: func() pl.GeneratorFunc[string] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]string{"one"})
//		}, want: []string{"one"}, timesToGenerate: 10},
//		{name: "twoStr 1", getGenFn: func() pl.GeneratorFunc[string] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]string{"one", "two"})
//		}, want: []string{"one", "two"}, timesToGenerate: 10},
//		{name: "twoStr 2", getGenFn: func() pl.GeneratorFunc[string] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]string{"one", "two"})
//		}, want: []string{"one"}, timesToGenerate: 1},
//		{name: "twoStr 3", getGenFn: func() pl.GeneratorFunc[string] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]string{"one", "two"})
//		}, want: []string{"one", "two"}, timesToGenerate: 2},
//		{name: "multiStr 1", getGenFn: func() pl.GeneratorFunc[string] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]string{"1 one", "2 two", "3 three", "4 four", "5 five", "6 six"})
//		}, want: []string{"1 one", "2 two", "3 three", "4 four", "5 five", "6 six"}, timesToGenerate: 100},
//		{name: "multiStr 2", getGenFn: func() pl.GeneratorFunc[string] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]string{"1 one", "2 two", "3 three", "4 four", "5 five", "6 six"})
//		}, want: []string{"1 one", "2 two", "3 three", "4 four"}, timesToGenerate: 4},
//		{name: "multiStr 2", getGenFn: func() pl.GeneratorFunc[string] {
//			time.Sleep(time.Microsecond * 10)
//			return pl.SliceGeneratorFuncFactory([]string{"1 one", "2 two", "3 three", "4 four", "5 five", "6 six"})
//		}, want: []string{"1 one"}, timesToGenerate: 1},
//	}
//
//	for _, tt := range tests {
//		t.GenerateToStream(tt.name, func(t *testing.T) {
//
//			defer func() {
//				r := recover()
//				if (r != nil) != tt.wantPanic {
//					t.Errorf("recover = %v, wantPanic = %v", r, tt.wantPanic)
//				}
//			}()
//
//			ctxWithCancel, cancelFn := context.WithCancel(context.Background())
//			defer cancelFn()
//			generatorHandler := pl.GeneratorHandlerFactory[string](tt.getGenFn(), confGenHandler.WithTimesToGenerate(tt.timesToGenerate))
//
//			var got []string
//			valStream := pl.GeneratorStageFactory[string](generatorHandler).GenerateToStream(ctxWithCancel)
//
//			genCount := uint(0)
//			for s := range valStream {
//				genCount += 1
//				got = append(got, s)
//			}
//
//			t.Log("Generated times:", genCount)
//			slices.Sort(got)
//			res := slices.Compact(got)
//
//			assert.Equal(t, tt.want, res)
//			assert.Equal(t, tt.timesToGenerate, genCount)
//		})
//	}
//}
