package confGeneratorFunc

type generatorFunctionConfig struct {
	ignoreEmptySliceLength bool
}

type Option interface {
	apply(*generatorFunctionConfig)
}

type OptionFn func(*generatorFunctionConfig)

func (fn OptionFn) apply(config *generatorFunctionConfig) {
	fn(config)
}

func WithIgnoreEmptySlice(ignoreEmptySlice bool) OptionFn {
	return func(config *generatorFunctionConfig) {
		config.ignoreEmptySliceLength = ignoreEmptySlice
	}
}

func (gfc *generatorFunctionConfig) IgnoreEmptySliceLength() bool {
	return gfc.ignoreEmptySliceLength
}

func NewGeneratorFunctionConfig(options ...Option) *generatorFunctionConfig {
	config := generatorFunctionConfig{}
	for _, option := range options {
		option.apply(&config)
	}

	return &config
}
