package config_generator_handler

const generateInfinitely = 0

type generatorHandlerConfig struct {
	timesToGenerate uint
}

type Option interface {
	apply(*generatorHandlerConfig)
}

type OptionFn func(*generatorHandlerConfig)

func (fn OptionFn) apply(config *generatorHandlerConfig) {
	fn(config)
}

func WithTimesToGenerate(timesToGenerate uint) OptionFn {
	return func(config *generatorHandlerConfig) {
		if timesToGenerate > 0 {
			config.timesToGenerate = timesToGenerate
		}
	}
}

func (g *generatorHandlerConfig) GenerateInfinitely() bool {
	return g.timesToGenerate == generateInfinitely
}

func (g *generatorHandlerConfig) TimesToGenerate() uint {
	return g.timesToGenerate
}

func NewGeneratorHandlerConfig(options ...Option) *generatorHandlerConfig {
	config := generatorHandlerConfig{}
	for _, option := range options {
		option.apply(&config)
	}

	return &config
}
