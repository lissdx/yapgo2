package config_producer

import (
	"fmt"
	"github.com/lissdx/yapgo2/pkg/logger"
	"github.com/lissdx/yapgo2/pkg/utils"
)

const produceInfinitely = 0
const defaultPrefProducerName = "GenProducer"
const defaultRandomStrLength = 5

type generatorProducerConfig struct {
	name            string
	timesToGenerate uint
	logger          logger.ILogger
}

type Option interface {
	apply(*generatorProducerConfig)
}

type OptionFn func(*generatorProducerConfig)

func (fn OptionFn) apply(config *generatorProducerConfig) {
	fn(config)
}

func WithTimesToGenerate(timesToGenerate uint) OptionFn {
	return func(config *generatorProducerConfig) {
		if timesToGenerate > 0 {
			config.timesToGenerate = timesToGenerate
		}
	}
}

func WithName(name string) OptionFn {
	return func(config *generatorProducerConfig) {
		config.name = name
		//if strings.TrimSpace(name) == "" {
		//	config.name = fmt.Sprintf("%s_%s", defaultPrefProducerName, randomStr(defaultRandomStrLength))
		//}
	}
}

func WithLogger(logger logger.ILogger) OptionFn {
	return func(config *generatorProducerConfig) {
		config.logger = logger
	}
}

func (g *generatorProducerConfig) GenerateInfinitely() bool {
	return g.timesToGenerate == produceInfinitely
}

func (g *generatorProducerConfig) TimesToGenerate() uint {
	return g.timesToGenerate
}

func (g *generatorProducerConfig) Name() string {
	return g.name
}

func (g *generatorProducerConfig) Logger() logger.ILogger {
	return g.logger
}

func NewGeneratorProducerConfig(options ...Option) *generatorProducerConfig {
	config := generatorProducerConfig{}
	for _, option := range options {
		option.apply(&config)
	}

	if config.name == "" {
		config.name = fmt.Sprintf("%s_%s", defaultPrefProducerName, utils.RandomStr(defaultRandomStrLength))
	}

	if config.logger == nil {
		config.logger = logger.NewNoopLogger()
	}

	return &config
}
