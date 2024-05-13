package pipeline

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestExamplePipeline(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	dataStream := SliceToStreamInfinitelyGeneratorFactory[[]int]().Run(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})

	stg1 := ProcessStageFactory[int, string](func(i int) (string, error) {
		return fmt.Sprintf("%d", i), nil
	}, func(err error) {
		t.Error(err)
	})

	stg2 := ProcessStageFactory[string, string](func(s string) (string, error) {
		return fmt.Sprintf("stage 2: %s", s), nil
	}, func(err error) {
		t.Error(err)
	})

	stg3 := ProcessStageFactory[string, string](func(s string) (string, error) {
		return fmt.Sprintf("stage 3 -> %s", s), nil
	}, func(err error) {
		t.Error(err)
	})

	stg1Stream := stg1.Run(ctx, dataStream)
	stg2Stream := stg2.Run(ctx, stg1Stream)
	resStream := stg3.Run(ctx, stg2Stream)

	evenCount := 0
	for v := range OrDoneFnFactory[string]().Run(ctx, resStream) {
		t.Log(v)
		evenCount++
	}

	t.Log("Total events: ", evenCount)
}
