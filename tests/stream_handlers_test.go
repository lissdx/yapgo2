package tests

import (
	"context"
	pl "github.com/lissdx/yapgo2/pkg/pipeline"
	genHandlerConf "github.com/lissdx/yapgo2/pkg/pipeline/config/generator/config_generator_handler"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"slices"
	"testing"
	"time"
)

func TestOrDone1(t *testing.T) {
	defer goleak.VerifyNone(t)

	tests := []struct {
		name     string
		dataList []int
		want     []int
	}{
		{name: "oneInt", dataList: []int{1}, want: []int{1}},
		{name: "twoInt", dataList: []int{-1, 0}, want: []int{-1, 0}},
		{name: "multiInt", dataList: []int{777, 555, 999, 999, 999, 999, 999, 999}, want: []int{555, 777, 999}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := context.WithCancel(context.Background())
			defer cancelFn()

			genHandler := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList), genHandlerConf.WithTimesToGenerate(100))
			dataStream := pl.GeneratorStageFactory(genHandler).Run(ctx)

			var got []int
			for num := range pl.OrDoneFnFactory[int]().Run(ctx, dataStream) {
				got = append(got, num)
			}

			t.Logf("%s got: %v", t.Name(), got)
			slices.Sort(got)
			res := slices.Compact(got)

			assert.Equal(t, tt.want, res)
		})
	}
}

func TestOrDoneContextWithTimeout(t *testing.T) {
	tests := []struct {
		name     string
		dataList []int
		want     []int
	}{
		{name: "oneInt", dataList: []int{1}, want: []int{1}},
		{name: "twoInt", dataList: []int{-1, 0}, want: []int{-1, 0}},
		{name: "multiInt", dataList: []int{777, 555, 999, 999, 999, 999, 999, 999}, want: []int{555, 777, 999}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := context.WithTimeout(context.Background(), time.Microsecond*1000)
			defer cancelFn()

			genHandler := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList))
			dataStream := pl.GeneratorStageFactory(genHandler).Run(ctx)

			var got []int
			for num := range pl.OrDoneFnFactory[int]().Run(ctx, dataStream) {
				got = append(got, num)
			}

			t.Logf("%s got: %v", t.Name(), got)
			slices.Sort(got)
			res := slices.Compact(got)

			assert.Equal(t, tt.want, res)
		})
	}
}

// Close channel
func TestOrDoneReadFromClosedChannel(t *testing.T) {
	tests := []struct {
		name     string
		dataList []int
		want     []int
	}{
		{name: "oneInt", dataList: []int{1}, want: []int{}},
		{name: "twoInt", dataList: []int{-1, 0}, want: []int{}},
		{name: "multiInt", dataList: []int{777, 555, 999, 999, 999, 999, 999, 999}, want: []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := context.WithTimeout(context.Background(), time.Microsecond*1000)
			defer cancelFn()

			genHandler := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList))
			dataStream := pl.GeneratorStageFactory(genHandler).Run(ctx)

			// drain and wait for
			// channel closing
			for range dataStream {
				// do nothing
			}

			orDoneCtx, orDoneCancel := context.WithCancel(context.Background())
			defer orDoneCancel()
			var got = make([]int, 0)
			for num := range pl.OrDoneFnFactory[int]().Run(orDoneCtx, dataStream) {
				got = append(got, num)
			}

			t.Logf("%s got: %v", t.Name(), got)
			slices.Sort(got)
			res := slices.Compact(got)

			assert.Equal(t, tt.want, res)
		})
	}
}

func TestMergeFnFactory1(t *testing.T) {
	tests := []struct {
		name     string
		dataList [][]int
		want     []int
	}{
		{name: "oneInt", dataList: [][]int{{1}}, want: []int{1}},
		{name: "twoInt", dataList: [][]int{{-1, 0}}, want: []int{-1, 0}},
		{name: "multiInt", dataList: [][]int{{1}, {-1, 0}, {777, 555, 999, 999, 999, 999, 999, 999}}, want: []int{-1, 0, 1, 555, 777, 999}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := context.WithTimeout(context.Background(), time.Microsecond*1000)
			defer cancelFn()

			dataStreams := make([]pl.ReadOnlyStream[int], 0, len(tt.dataList))
			for i := 0; i < len(tt.dataList); i++ {
				genHandler := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList[i]))
				dataStream := pl.GeneratorStageFactory(genHandler).Run(ctx)
				dataStreams = append(dataStreams, dataStream)
			}

			resultDataStream := pl.MergeFnFactory[[]pl.ReadOnlyStream[int]]().Run(ctx, dataStreams)

			var got []int
			for num := range pl.OrDoneFnFactory[int]().Run(ctx, resultDataStream) {
				got = append(got, num)
			}

			t.Logf("%s got: %v", t.Name(), got)
			slices.Sort(got)
			res := slices.Compact(got)

			assert.Equal(t, tt.want, res)
		})
	}
}

// with closed channels
func TestMergeFnFactory2(t *testing.T) {
	tests := []struct {
		name          string
		dataList      [][]int
		want          []int
		timesGenerate int
	}{
		{name: "one", dataList: [][]int{{1}}, want: []int{1}},
		{name: "one close", dataList: [][]int{{2}}, want: []int{}},
		{name: "two", dataList: [][]int{{1}, {-1, 0}}, want: []int{-1, 0, 1}},
		{name: "multi", timesGenerate: 4, dataList: [][]int{{1}, {-1, 0}, {777, 555, 888, 111, 33333, 999, 999, 999}}, want: []int{-1, 0, 111, 555, 777, 888}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx, cancelFn := context.WithTimeout(context.Background(), time.Millisecond*1000)
			defer cancelFn()

			dataStreams := make([]pl.ReadOnlyStream[int], 0, len(tt.dataList))

			switch tt.name {
			case "one":
				genHandler := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList[0]))
				dataStream := pl.GeneratorStageFactory(genHandler).Run(ctx)
				dataStreams = append(dataStreams, dataStream)
			case "one close":
				genHandler := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList[0]), genHandlerConf.WithTimesToGenerate(1000))
				dataStream := pl.GeneratorStageFactory(genHandler).Run(ctx)
				dataStreams = append(dataStreams, dataStream)

				waitFor1 := pl.NoOpSubFnFactory[int]()(ctx, dataStream)
				<-waitFor1.Done()
			case "two":
				for i := 0; i < len(tt.dataList); i++ {
					genHandler := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList[i]))
					dataStream := pl.GeneratorStageFactory(genHandler).Run(ctx)
					dataStreams = append(dataStreams, dataStream)
				}
			case "multi":
				mCtx1, mCancelFn := context.WithTimeout(context.Background(), time.Millisecond*10)
				defer mCancelFn()
				genHandler1 := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList[0]))
				dataStream1 := pl.GeneratorStageFactory(genHandler1).Run(mCtx1)
				dataStreams = append(dataStreams, dataStream1)
				// wait and close dataStream1
				waitFor1 := pl.NoOpSubFnFactory[int]()(ctx, dataStream1)
				select {
				case <-waitFor1.Done():
					t.Log("dataStream1 is closed")
				}

				// dataStream2 will be closed after 1 time generation of all elements
				genHandler2 := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList[1]),
					genHandlerConf.WithTimesToGenerate(1000))
				dataStream2 := pl.GeneratorStageFactory(genHandler2).Run(ctx)
				dataStreams = append(dataStreams, dataStream2)

				// dataStream2 will be closed after 1 time generation of all elements
				genHandler3 := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.dataList[2]),
					genHandlerConf.WithTimesToGenerate(4))
				dataStream3 := pl.GeneratorStageFactory(genHandler3).Run(ctx)
				dataStreams = append(dataStreams, dataStream3)
			}

			resultDataStream := pl.MergeFnFactory[[]pl.ReadOnlyStream[int]]().Run(ctx, dataStreams)

			t.Logf("%s data resiving", t.Name())
			var got = make([]int, 0, 100)
			for num := range pl.OrDoneFnFactory[int]().Run(ctx, resultDataStream) {
				got = append(got, num)
			}

			//log.Default().Println(got)
			slices.Sort(got)
			res := slices.Compact(got)

			assert.Equal(t, tt.want, res)
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

			ctx, cancelFn := context.WithTimeout(context.Background(), time.Millisecond*300)
			defer cancelFn()

			genHandler := pl.GeneratorHandlerFactory(pl.SliceGeneratorFuncFactory(tt.args))
			dataStream := pl.GeneratorStageFactory(genHandler).Run(ctx)
			resultStream := pl.FlatSlicesToStreamFnFactory[[]int]().Run(ctx, dataStream)

			var got []int
			for num := range pl.OrDoneFnFactory[int]().Run(ctx, resultStream) {
				got = append(got, num)
			}

			//log.Default().Println(got)
			slices.Sort(got)
			res := slices.Compact(got)

			assert.Equal(t, tt.want, res)
		})
	}
}
