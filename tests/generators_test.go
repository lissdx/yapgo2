package tests

import (
	"context"
	"github.com/lissdx/yapgo2/pkg/pipeline"
	"go.uber.org/goleak"
	"slices"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func Test_GeneratorFnToStreamGenerator(t *testing.T) {
	tests := []struct {
		name      string
		arg       func() pipeline.GeneratorFn[int]
		want      []int
		wantPanic bool
	}{
		{name: "nil", arg: func() pipeline.GeneratorFn[int] {
			time.Sleep(time.Microsecond * 10)
			return pipeline.GeneratorFnFromSliceFactory([]int(nil))
		}, wantPanic: true},
		{name: "empty", arg: func() pipeline.GeneratorFn[int] {
			time.Sleep(time.Microsecond * 10)
			return pipeline.GeneratorFnFromSliceFactory([]int{})
		}, wantPanic: true},
		{name: "oneInt", arg: func() pipeline.GeneratorFn[int] {
			time.Sleep(time.Microsecond * 10)
			return pipeline.GeneratorFnFromSliceFactory([]int{1})
		}, want: []int{1}},
		{name: "twoInt", arg: func() pipeline.GeneratorFn[int] {
			s := []int{1, -1}
			return pipeline.GeneratorFnFromSliceFactory(s)
		}, want: []int{1, -1}},
		{name: "tenInt", arg: func() pipeline.GeneratorFn[int] {
			s := []int{1, -1, 23, 100, 26, 33, 78, 33, 221, 0}
			return pipeline.GeneratorFnFromSliceFactory(s)
		}, want: []int{1, -1, 23, 100, 26, 33, 78, 33, 221, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("recover = %v, wantPanic = %v", r, tt.wantPanic)
				}
			}()

			ctxWithCancel, cancelFn := context.WithTimeout(context.Background(), time.Microsecond*1000)
			defer cancelFn()

			var got []int
			valStream := pipeline.GeneratorFnToStreamGeneratorFactory[pipeline.GeneratorFn[int]]().Run(ctxWithCancel, tt.arg())

			for num := range valStream {
				got = append(got, num)
			}

			t.Logf("%s got: %v", t.Name(), got)
			if len(got) < len(tt.want) {
				t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
			}

			for i, wantVal := range tt.want {
				if got[i] != wantVal {
					t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
				}
			}
		})
	}
}

func Test_SliceToStreamOnePassGeneratorFactory(t *testing.T) {
	tests := []struct {
		name      string
		args      []int
		want      []int
		wantPanic bool
	}{
		{name: "nil", args: []int(nil), want: []int{}, wantPanic: true},
		{name: "empty", args: []int{}, want: []int{}, wantPanic: true},
		{name: "oneInt", args: []int{1}, want: []int{1}},
		{name: "twoInt", args: []int{1, -1}, want: []int{1, -1}},
		{name: "tenInt", args: []int{1, -1, 23, 100, 26, 33, 78, 33, 221, 0}, want: []int{1, -1, 23, 100, 26, 33, 78, 33, 221, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("recover = %v, wantPanic = %v", r, tt.wantPanic)
				}
			}()

			ctxWithCancel, cancelFn := context.WithCancel(context.Background())
			defer cancelFn()

			var got []int
			valStream := pipeline.SliceToStreamOnePassGeneratorFactory[[]int]().Run(ctxWithCancel, tt.args)

			for num := range valStream {
				got = append(got, num)
			}

			if len(got) != len(tt.want) {
				t.Errorf("Generator = %v, want %v", got, tt.want)
			}

			for i, wantVal := range tt.want {
				if got[i] != wantVal {
					t.Errorf("Generator = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// Test_SliceToStreamOnePassGeneratorFactory2
// generates slices []T from [][]T
func Test_SliceToStreamOnePassGeneratorFactory2(t *testing.T) {
	tests := []struct {
		name      string
		args      [][]int
		want      []int
		wantPanic bool
	}{
		{name: "nil", args: [][]int(nil), want: []int{}, wantPanic: true},
		{name: "empty", args: [][]int{}, want: []int{}, wantPanic: true},
		{name: "oneInt", args: [][]int{{1}}, want: []int{1}},
		{name: "twoInt", args: [][]int{{1, -1}, {7, 5}}, want: []int{-1, 1, 5, 7}},
		{name: "multi", args: [][]int{{1, -1}, {7, 5}, {3}, {0, 777, 999, 888, 77}}, want: []int{-1, 0, 1, 3, 5, 7, 77, 777, 888, 999}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("recover = %v, wantPanic = %v", r, tt.wantPanic)
				}
			}()

			ctxWithCancel, cancelFn := context.WithCancel(context.Background())
			defer cancelFn()

			var got []int
			valStream := pipeline.SliceToStreamOnePassGeneratorFactory[[][]int]().Run(ctxWithCancel, tt.args)

			for num := range valStream {
				got = append(got, num...)
			}

			slices.Sort(got)
			res := slices.Compact(got)

			if len(res) != len(tt.want) {
				t.Errorf("%s = %v, want %v", t.Name(), res, tt.want)
			}
			for i, wantVal := range tt.want {
				if res[i] != wantVal {
					t.Errorf("%s = %v, want %v", t.Name(), res, tt.want)
				}
			}
		})
	}
}

func Test_SliceToStreamInfinitelyGeneratorFactory(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		want      []string
		wantPanic bool
	}{
		{name: "nil", args: []string(nil), wantPanic: true},
		{name: "empty", args: []string{}, wantPanic: true},
		{name: "oneStr", args: []string{"one"}, want: []string{"one"}},
		{name: "twoStr", args: []string{"one", "two"}, want: []string{"one", "two"}},
		{name: "fiveStr", args: []string{"one", "two", "three", "four", "five"}, want: []string{"one", "two", "three", "four", "five"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("recover = %v, wantPanic = %v", r, tt.wantPanic)
				}
			}()

			ctxWithCancel, cancelFn := context.WithTimeout(context.Background(), time.Microsecond*1000)
			defer cancelFn()

			var got []string
			valStream := pipeline.SliceToStreamInfinitelyGeneratorFactory[[]string]().Run(ctxWithCancel, tt.args)

			for s := range valStream {
				got = append(got, s)
			}

			if len(got) < len(tt.want) {
				t.Errorf("Generator = %v, want %v", got, tt.want)
			}

			for i, wantVal := range tt.want {
				if got[i] != wantVal {
					t.Errorf("Generator = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_SliceToStreamNGeneratorFactory(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		nTimes    int
		want      []string
		wantPanic bool
	}{
		{name: "nil", args: []string(nil), wantPanic: true},
		{name: "empty", args: []string{}, wantPanic: true},
		{name: "oneStr", args: []string{"one"}, want: []string{"one", "one", "one", "one", "one"}, nTimes: 5},
		{name: "twoStr", args: []string{"one", "two"}, want: []string{"one", "two", "one", "two", "one"}, nTimes: 5},
		{name: "fiveStr", args: []string{"one", "two", "three", "four", "five"}, want: []string{"one"}, nTimes: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("recover = %v, wantPanic = %v", r, tt.wantPanic)
				}
			}()

			ctxWithCancel, cancelFn := context.WithTimeout(context.Background(), time.Microsecond*1000)
			defer cancelFn()

			var got []string
			valStream := pipeline.SliceToStreamNGeneratorFactory[[]string](uint(tt.nTimes)).Run(ctxWithCancel, tt.args)

			for s := range valStream {
				got = append(got, s)
			}

			if len(got) != len(tt.want) {
				t.Errorf("Generator = %v, want %v", got, tt.want)
			}

			for i, wantVal := range tt.want {
				if got[i] != wantVal {
					t.Errorf("Generator = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

//func Test_SliceToStreamRepeatGenerator(t *testing.T) {
//	tests := []struct {
//		name string
//		args []int
//		want []int
//	}{
//		{name: "empty", args: []int{}, want: []int{}},
//		{name: "oneInt", args: []int{1}, want: []int{1}},
//		{name: "twoInt", args: []int{1, -1}, want: []int{1, -1}},
//		{name: "tenInt", args: []int{1, -1, 23, 100, 26, 33, 78, 33, 221, 0}, want: []int{1, -1, 23, 100, 26, 33, 78, 33, 221, 0}},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//
//			ctxWithCancel, cancelFn := context.WithCancel(context.Background())
//
//			var got []int
//			valStream := pipeline.SliceToStreamRepeatGeneratorFactory[[]int]().Run(ctxWithCancel, tt.args)
//
//			go func() {
//				defer cancelFn()
//				time.Sleep(time.Microsecond * 1000)
//			}()
//
//			for num := range valStream {
//				got = append(got, num)
//			}
//
//			t.Logf("%s got: %v", t.Name(), got)
//			if len(got) < len(tt.want) {
//				t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
//			}
//
//			for i, wantVal := range tt.want {
//				if got[i] != wantVal {
//					t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
//				}
//			}
//		})
//	}
//}

//
//func TestOrDone1(t *testing.T) {
//	tests := []struct {
//		name string
//		args  [][]int
//		want []int
//	}{
//		{name: "empty", args: [][]int{}, want: []int{}},
//		{name: "oneInt", args: [][]int{{1}}, want: []int{1}},
//		{name: "twoInt", args: [][]int{{1}, {-1, 0}}, want: []int{1, -1, 0}},
//		{name: "multiInt", args: [][]int{{1}, {-1, 0}, {3, 4, 5}, {777, 555}, {999, 999, 999, 999, 999, 999}}, want: []int{1, -1, 0, 3, 4, 5, 777, 555, 999, 999, 999, 999, 999, 999}},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//
//			ctxP, cancelFn := context.WithCancel(context.Background())
//			defer cancelFn()
//
//			slicesStream := pipeline.SliceToStreamGeneratorFactory[[][]int]().Run(ctxP, tt.args)
//			ctxCh1, _ := context.WithCancel(ctxP)
//			valStream := pipeline.FlatSlicesToStreamFnFactory[[]int]().Run(ctxCh1, slicesStream)
//
//			ctxCh2, _ := context.WithCancel(ctxCh1)
//			var got []int
//			for num := range pipeline.OrDoneFnFactory[int]().Run(ctxCh2, valStream) {
//				got = append(got, num)
//			}
//
//			for num := range valStream {
//				got = append(got, num)
//			}
//
//			t.Logf("%s got: %v", t.Name(), got)
//			if len(got) < len(tt.want) {
//				t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
//			}
//
//			for i, wantVal := range tt.want {
//				if got[i] != wantVal {
//					t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
//				}
//			}
//		})
//	}
//}
//
//func TestOrDone2(t *testing.T) {
//	tests := []struct {
//		name string
//		args  pipeline.GeneratorFn[int]
//		want []int
//	}{
//		{name: "oneInt", args: func() int {
//			time.Sleep(time.Microsecond * 10)
//			return 1
//		}, want: []int{1}},
//		{name: "twoInt", args: func() func() int {
//			s := []int{1, -1}
//			indx := 0
//			return func() int {
//				res := s[indx%2]
//				indx += 1
//				return res
//			}
//		}(), want: []int{1, -1}},
//		{name: "tenInt", args: func() func() int {
//			s := []int{1, -1, 23, 100, 26, 33, 78, 33, 221, 0}
//			indx := 0
//			return func() int {
//				res := s[indx]
//				indx += 1
//				if indx >= len(s) {
//					indx = 0
//				}
//				return res
//			}
//		}(), want: []int{1, -1, 23, 100, 26, 33, 78, 33, 221, 0}},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//
//			pCtx, cancelFn := context.WithCancel(context.Background())
//
//			var got []int
//			valStream := pipeline.GeneratorFnToStreamGeneratorFactory[pipeline.GeneratorFn[int]]()(pCtx, tt.args)
//
//			go func() {
//				defer cancelFn()
//				time.Sleep(time.Microsecond * 1000)
//			}()
//
//			chCtx, _ := context.WithCancel(pCtx)
//			for num := range pipeline.OrDoneFnFactory[int]().Run(chCtx, valStream) {
//				got = append(got, num)
//			}
//
//			t.Logf("%s got: %v", t.Name(), got)
//			if len(got) < len(tt.want) {
//				t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
//			}
//
//			for i, wantVal := range tt.want {
//				if got[i] != wantVal {
//					t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
//				}
//			}
//		})
//	}
//}
//
//func TestOrDone3(t *testing.T) {
//	tests := []struct {
//		name string
//		args []int
//		want []int
//	}{
//		{name: "empty", args: []int{}, want: []int{}},
//		{name: "oneInt", args: []int{1}, want: []int{1}},
//		{name: "twoInt", args: []int{1, -1}, want: []int{1, -1}},
//		{name: "tenInt", args: []int{1, -1, 23, 100, 26, 33, 78, 33, 221, 0}, want: []int{1, -1, 23, 100, 26, 33, 78, 33, 221, 0}},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//
//			pCtx, cancelFn := context.WithCancel(context.Background())
//
//			var got []int
//			valStream := pipeline.SliceToStreamRepeatGeneratorFactory[[]int]()(pCtx, tt.args)
//
//			go func() {
//				defer cancelFn()
//				time.Sleep(time.Microsecond * 1000)
//			}()
//
//			chCtx, _ := context.WithCancel(pCtx)
//			for num := range pipeline.OrDoneFnFactory[int]().Run(chCtx, valStream) {
//				got = append(got, num)
//			}
//
//			t.Logf("%s got: %v", t.Name(), got)
//			if len(got) < len(tt.want) {
//				t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
//			}
//
//			for i, wantVal := range tt.want {
//				if got[i] != wantVal {
//					t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
//				}
//			}
//		})
//	}
//}
//
//func TestMergeGenerator(t *testing.T) {
//	tests := []struct {
//		name string
//		args [4][]int
//		want []int
//	}{
//		{name: "empty", args: [4][]int{{}, {}, {}, {}}, want: []int{}},
//		{name: "oneInt", args: [4][]int{{1}, {2}, {3}, {4}}, want: []int{1, 2, 3, 4}},
//		{name: "twoInt", args: [4][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}}, want: []int{1, 2, 3, 4, 5, 6, 7, 8}},
//		{name: "common", args: [4][]int{{1, 2}, {}, {5}, {7, 8, 9, 10}}, want: []int{1, 2, 5, 7, 8, 9, 10}},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//
//			ctxWithCancel, cancelFn := context.WithCancel(context.Background())
//
//			streams := make([]pipeline.ReadOnlyStream[int], 4)
//			var got []int
//			streams[0] = pipeline.SliceToStreamRepeatGeneratorFactory[[]int]()(ctxWithCancel, tt.args[0])
//			streams[1] = pipeline.SliceToStreamRepeatGeneratorFactory[[]int]()(ctxWithCancel, tt.args[1])
//			streams[2] = pipeline.SliceToStreamRepeatGeneratorFactory[[]int]()(ctxWithCancel, tt.args[2])
//			streams[3] = pipeline.SliceToStreamRepeatGeneratorFactory[[]int]()(ctxWithCancel, tt.args[3])
//
//			go func() {
//				defer cancelFn()
//				time.Sleep(time.Microsecond * 1000)
//			}()
//
//			chCtx, c := context.WithCancel(ctxWithCancel)
//			defer c()
//			for num := range pipeline.MergeFnFactory[[]pipeline.ReadOnlyStream[int]]()(chCtx, streams) {
//				got = append(got, num)
//			}
//
//			t.Logf("%s got: %v", t.Name(), got)
//			if len(got) < len(tt.want) {
//				t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
//			}
//
//			slices.Sort(got)
//			res := slices.Compact(got)
//			for i, wantVal := range tt.want {
//				if res[i] != wantVal {
//					t.Errorf("%s = %v, want %v", t.Name(), res, tt.want)
//				}
//			}
//		})
//	}
//}
