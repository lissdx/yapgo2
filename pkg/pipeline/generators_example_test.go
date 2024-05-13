package pipeline

import (
	"context"
	"math/rand"
	"testing"
	"time"
)

/**
 * Custom Generators Examples
 */

// GeneratorFnSliceString simple generation from
// given slice of strings
func GeneratorFnSliceString(s []string) GeneratorFn[string] {
	currentIndx := 0
	return func() string {
		defer func() {
			currentIndx = (currentIndx + 1) % len(s)
		}()

		return s[currentIndx]
	}
}

// Random int generation example
func TestExample_GeneratorFn(t *testing.T) {
	timesToGenerate := 10
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dataStream := GeneratorFnToStreamGeneratorFactory[GeneratorFn[int]](WithTimesToGenerate(uint(timesToGenerate))).
		Run(ctx, func() int {
			return rand.Intn(100)
		})

	dataCounter := 0
	for v := range dataStream {
		dataCounter += 1
		t.Log(v)
	}

	t.Log("Data was generated:", dataCounter, "times")
}

// TestExample_SliceOfStringsGeneratorFn1 "Infinitely" generates data
// from the given slice
func TestExample_SliceOfStringsGeneratorFn1(t *testing.T) {
	dataSlice := []string{"one", "two", "three", "four", "five"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()

	dataStream := GeneratorFnToStreamGeneratorFactory[GeneratorFn[string]]().
		Run(ctx, GeneratorFnSliceString(dataSlice))

	dataCounter := 0
	for v := range dataStream {
		dataCounter += 1
		t.Log(v)
	}

	t.Log("Data was generated:", dataCounter, "times")
}

// TestExample_SliceOfStringsGeneratorFn2 generates data
// from the given slice. Now we try to generate data by the len.
// of given slice
func TestExample_SliceOfStringsGeneratorFn2(t *testing.T) {
	dataSlice := []string{"one", "two", "three", "four", "five"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dataStream := GeneratorFnToStreamGeneratorFactory[GeneratorFn[string]](WithTimesToGenerate(uint(len(dataSlice)))).
		Run(ctx, GeneratorFnSliceString(dataSlice))

	dataCounter := 0
	for v := range dataStream {
		dataCounter += 1
		t.Log(v)
	}

	t.Log("Data was generated:", dataCounter, "times")
}

// TestExample_SliceOfStringsGeneratorFn3 generates data
// from the given slice. Now we try to generate
// 7 times
func TestExample_SliceOfStringsGeneratorFn3(t *testing.T) {
	dataSlice := []string{"one", "two", "three", "four", "five"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dataStream := GeneratorFnToStreamGeneratorFactory[GeneratorFn[string]](WithTimesToGenerate(7)).
		Run(ctx, GeneratorFnSliceString(dataSlice))

	dataCounter := 0
	for v := range dataStream {
		dataCounter += 1
		t.Log(v)
	}

	t.Log("Data was generated:", dataCounter, "times")
}

// TestExample_SliceToStreamOnePassGeneratorFactory1 generates data
// from the given slice of strings.
func TestExample_SliceToStreamOnePassGeneratorFactory1(t *testing.T) {
	dataSlice := []string{"one", "two", "three", "four", "five", "six", "seven", "eight"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dataStream := SliceToStreamOnePassGeneratorFactory[[]string]().Run(ctx, dataSlice)

	dataCounter := 0
	for v := range dataStream {
		dataCounter += 1
		t.Log(v)
	}

	t.Log("Data was generated:", dataCounter, "times")
}

// TestExample_SliceToStreamOnePassGeneratorFactory2 generates data
// from the given slice of int.
func TestExample_SliceToStreamOnePassGeneratorFactory2(t *testing.T) {
	dataSlice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dataStream := SliceToStreamOnePassGeneratorFactory[[]int]().Run(ctx, dataSlice)

	dataCounter := 0
	for v := range dataStream {
		dataCounter += 1
		t.Log(v)
	}

	t.Log("Data was generated:", dataCounter, "times")
}

// TestExample_SliceToStreamInfinitelyGeneratorFactory1 generates data
// from the given slice of int Infinitely.
func TestExample_SliceToStreamInfinitelyGeneratorFactory1(t *testing.T) {
	dataSlice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()

	dataStream := SliceToStreamInfinitelyGeneratorFactory[[]int]().Run(ctx, dataSlice)

	dataCounter := 0
	for v := range dataStream {
		dataCounter += 1
		t.Log(v)
	}

	t.Log("Data was generated:", dataCounter, "times")
}

// TestExample_SliceToStreamNGeneratorFactory generates data
// from the given slice of int N times.
func TestExample_SliceToStreamNGeneratorFactory(t *testing.T) {
	timesToGenerate := 7
	dataSlice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dataStream := SliceToStreamNGeneratorFactory[[]int](uint(timesToGenerate)).Run(ctx, dataSlice)

	dataCounter := 0
	for v := range dataStream {
		dataCounter += 1
		t.Log(v)
	}

	t.Log("Data was generated:", dataCounter, "times")
}
