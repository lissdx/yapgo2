package pipeline

type ErrorStrategy int

const (
	OnErrorRemove = iota
	OnErrorContinue
)

type yapgo2Config struct {
	timesToGenerate        uint
	ignoreEmptySliceLength bool
	onErrorStrategy        ErrorStrategy
}

type Option interface {
	apply(*yapgo2Config)
}

type OptionFn func(*yapgo2Config)

func (fn OptionFn) apply(config *yapgo2Config) {
	fn(config)
}

func WithTimesToGenerate(timesToGenerate uint) OptionFn {
	return func(config *yapgo2Config) {
		if timesToGenerate > 0 {
			config.timesToGenerate = timesToGenerate
		}
	}
}

func WithIgnoreEmptySlice(ignoreEmptySlice bool) OptionFn {
	return func(config *yapgo2Config) {
		config.ignoreEmptySliceLength = ignoreEmptySlice
	}
}

func WithOnErrorContinueStrategy() OptionFn {
	return func(config *yapgo2Config) {
		config.onErrorStrategy = OnErrorContinue
	}
}

func WithOnErrorRemoveStrategy[T, RS any]() OptionFn {
	return func(config *yapgo2Config) {
		config.onErrorStrategy = OnErrorRemove
	}
}
