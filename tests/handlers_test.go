package tests

import (
	"context"
	"github.com/lissdx/yapgo2/pkg/pipeline"
	"slices"
	"testing"
	"time"
)

//func TestMain(m *testing.M) {
//	goleak.VerifyTestMain(m)
//}

func TestOrDone1(t *testing.T) {
	tests := []struct {
		name string
		arg  []int
		want []int
	}{
		{name: "oneInt", arg: []int{1}, want: []int{1}},
		{name: "twoInt", arg: []int{-1, 0}, want: []int{-1, 0}},
		{name: "multiInt", arg: []int{777, 555, 999, 999, 999, 999, 999, 999}, want: []int{777, 555, 999, 999, 999, 999, 999, 999}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := context.WithCancel(context.Background())
			defer cancelFn()

			dataStream := pipeline.SliceToStreamOnePassGeneratorFactory[[]int]().Run(ctx, tt.arg)

			var got []int
			for num := range pipeline.OrDoneFnFactory[int]().Run(ctx, dataStream) {
				got = append(got, num)
			}

			t.Logf("%s got: %v", t.Name(), got)
			if len(got) != len(tt.want) {
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

func TestOrDone2(t *testing.T) {
	tests := []struct {
		name string
		arg  []int
		want []int
	}{
		{name: "oneInt", arg: []int{1}, want: []int{1}},
		{name: "twoInt", arg: []int{-1, 0}, want: []int{-1, 0}},
		{name: "multiInt", arg: []int{777, 555, 999, 999, 999, 999, 999, 999}, want: []int{777, 555, 999, 999, 999, 999, 999, 999}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := context.WithTimeout(context.Background(), time.Microsecond*1000)
			defer cancelFn()

			dataStream := pipeline.SliceToStreamInfinitelyGeneratorFactory[[]int]().Run(ctx, tt.arg)

			var got []int
			for num := range pipeline.OrDoneFnFactory[int]().Run(ctx, dataStream) {
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

// Close channel
func TestOrDone3(t *testing.T) {
	tests := []struct {
		name string
		arg  []int
		want []int
	}{
		{name: "oneInt", arg: []int{1}, want: []int{}},
		{name: "twoInt", arg: []int{-1, 0}, want: []int{}},
		{name: "multiInt", arg: []int{777, 555, 999, 999, 999, 999, 999, 999}, want: []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := context.WithCancel(context.Background())
			defer cancelFn()

			dataStream := pipeline.SliceToStreamOnePassGeneratorFactory[[]int]().Run(ctx, tt.arg)

			// drain and close channel
			for range dataStream {
				// do nothing
			}

			var got []int
			for num := range pipeline.OrDoneFnFactory[int]().Run(ctx, dataStream) {
				got = append(got, num)
			}

			t.Logf("%s got: %v", t.Name(), got)
			if len(got) != len(tt.want) {
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

func TestMergeFnFactory1(t *testing.T) {
	tests := []struct {
		name string
		arg  [][]int
		want []int
	}{
		{name: "oneInt", arg: [][]int{{1}}, want: []int{1}},
		{name: "twoInt", arg: [][]int{{-1, 0}}, want: []int{-1, 0}},
		{name: "multiInt", arg: [][]int{{1}, {-1, 0}, {777, 555, 999, 999, 999, 999, 999, 999}}, want: []int{-1, 0, 1, 555, 777, 999}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := context.WithTimeout(context.Background(), time.Microsecond*1000)
			defer cancelFn()

			dataStreams := make([]pipeline.ReadOnlyStream[int], 0, len(tt.arg))
			for i := 0; i < len(tt.arg); i++ {
				dataStream := pipeline.SliceToStreamInfinitelyGeneratorFactory[[]int]().Run(ctx, tt.arg[i])
				dataStreams = append(dataStreams, dataStream)
			}

			resultDataStream := pipeline.MergeFnFactory[[]pipeline.ReadOnlyStream[int]]().Run(ctx, dataStreams)

			var got []int
			for num := range pipeline.OrDoneFnFactory[int]().Run(ctx, resultDataStream) {
				got = append(got, num)
			}

			t.Logf("%s got: %v", t.Name(), got)
			if len(got) < len(tt.want) {
				t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
			}

			slices.Sort(got)
			res := slices.Compact(got)
			for i, wantVal := range tt.want {
				if res[i] != wantVal {
					t.Errorf("%s = %v, want %v", t.Name(), res, tt.want)
				}
			}
		})
	}
}

// with closed channels
func TestMergeFnFactory2(t *testing.T) {
	tests := []struct {
		name          string
		arg           [][]int
		want          []int
		timesGenerate int
	}{
		{name: "one", arg: [][]int{{1}}, want: []int{1}},
		{name: "two", arg: [][]int{{1}, {-1, 0}}, want: []int{-1, 0, 1}},
		{name: "multi", timesGenerate: 4, arg: [][]int{{1}, {-1, 0}, {777, 555, 888, 111, 33333, 999, 999, 999}}, want: []int{-1, 0, 111, 555, 777, 888}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := context.WithTimeout(context.Background(), time.Millisecond*100)
			defer cancelFn()

			dataStreams := make([]pipeline.ReadOnlyStream[int], 0, len(tt.arg))

			if len(tt.arg) == 1 {
				dataStream := pipeline.SliceToStreamInfinitelyGeneratorFactory[[]int]().Run(ctx, tt.arg[0])
				dataStreams = append(dataStreams, dataStream)
			}

			if len(tt.arg) == 2 {
				dataStream1 := pipeline.SliceToStreamInfinitelyGeneratorFactory[[]int]().Run(ctx, tt.arg[0])
				dataStreams = append(dataStreams, dataStream1)
				dataStream2 := pipeline.SliceToStreamOnePassGeneratorFactory[[]int]().Run(ctx, tt.arg[1])
				dataStreams = append(dataStreams, dataStream2)
			}

			if len(tt.arg) == 3 {
				genCtx, localCancel := context.WithTimeout(ctx, time.Microsecond*100)
				dataStream1 := pipeline.SliceToStreamInfinitelyGeneratorFactory[[]int]().Run(genCtx, tt.arg[0])
				// drain channel until not closed
				pipeline.NoOpSubFnFactory[int]()(genCtx, dataStream1)
				localCancel()
				// add closed channel
				dataStreams = append(dataStreams, dataStream1)
				dataStream2 := pipeline.SliceToStreamInfinitelyGeneratorFactory[[]int]().Run(ctx, tt.arg[1])
				dataStreams = append(dataStreams, dataStream2)
				dataStream3 := pipeline.SliceToStreamNGeneratorFactory[[]int](uint(tt.timesGenerate)).Run(ctx, tt.arg[2])
				dataStreams = append(dataStreams, dataStream3)
			}

			resultDataStream := pipeline.MergeFnFactory[[]pipeline.ReadOnlyStream[int]]().Run(ctx, dataStreams)

			var got []int
			for num := range pipeline.OrDoneFnFactory[int]().Run(ctx, resultDataStream) {
				got = append(got, num)
			}

			//t.Logf("%s got: %v", t.Name(), got)
			//if len(got) < len(tt.want) {
			//	t.Errorf("%s = %v, want %v", t.Name(), got, tt.want)
			//}

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

func TestFlatSlicesToStreamFnFactory1(t *testing.T) {
	tests := []struct {
		name          string
		args          [][]int
		want          []int
		timesGenerate int
	}{
		{name: "one", args: [][]int{{1}}, want: []int{1}},
		{name: "two", args: [][]int{{1}, {-1, 0}}, want: []int{-1, 0, 1}},
		{name: "multi", args: [][]int{{1}, {-1, 0}, {777, 555, 888, 111, 33333, 999, 999, 999}}, want: []int{-1, 0, 1, 111, 555, 777, 888, 999, 33333}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := context.WithTimeout(context.Background(), time.Millisecond*100)
			defer cancelFn()

			dataStream := pipeline.SliceToStreamInfinitelyGeneratorFactory[[][]int]().Run(ctx, tt.args)
			resultDataStream := pipeline.FlatSlicesToStreamFnFactory[[]int]().Run(ctx, dataStream)

			var got []int
			for num := range pipeline.OrDoneFnFactory[int]().Run(ctx, resultDataStream) {
				got = append(got, num)
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

//func Test_SliceToStreamGenerator(t *testing.T) {
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
//			defer cancelFn()
//
//			var got []int
//			valStream := pipeline.SliceToStreamGeneratorFactory[[]int]().Run(ctxWithCancel, tt.args)
//
//			for num := range valStream {
//				got = append(got, num)
//			}
//
//			if len(got) != len(tt.want) {
//				t.Errorf("Generator = %v, want %v", got, tt.want)
//			}
//
//			for i, wantVal := range tt.want {
//				if got[i] != wantVal {
//					t.Errorf("Generator = %v, want %v", got, tt.want)
//				}
//			}
//		})
//	}
//}

//func Test_SliceToStreamInfinitelyRepeatGenerator(t *testing.T) {
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
//func Test_FuncRepeatGenerator(t *testing.T) {
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
//			ctxWithCancel, cancelFn := context.WithCancel(context.Background())
//
//			var got []int
//			valStream := pipeline.FnToStreamInfinitelyGenerateFnFactory[pipeline.GeneratorFn[int]]().Run(ctxWithCancel, tt.args)
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
//			valStream := pipeline.FnToStreamInfinitelyGenerateFnFactory[pipeline.GeneratorFn[int]]()(pCtx, tt.args)
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
