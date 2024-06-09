package config_stage_not_in_use

import (
	"fmt"
	"github.com/lissdx/yapgo2/pkg/logger"
	"github.com/lissdx/yapgo2/pkg/utils"
	"strings"
)

const defaultPrefStageName = "GenStage"
const defaultRandomStrLength = 5

type generatorStageConfig struct {
	name   string
	logger logger.ILogger
}

type Option interface {
	apply(*generatorStageConfig)
}

type OptionFn func(*generatorStageConfig)

func (fn OptionFn) apply(config *generatorStageConfig) {
	fn(config)
}

func WithName(name string) OptionFn {
	return func(config *generatorStageConfig) {
		if strings.TrimSpace(name) == "" {
			config.name = fmt.Sprintf("%s_%s", defaultPrefStageName, utils.RandomStr(defaultRandomStrLength))
		}
	}
}

func WithLogger(logger logger.ILogger) OptionFn {
	return func(config *generatorStageConfig) {
		config.logger = logger
	}
}

func (g *generatorStageConfig) Name() string {
	return g.name
}

func (g *generatorStageConfig) Logger() logger.ILogger {
	return g.logger
}

func NewGeneratorStageConfig(options ...Option) *generatorStageConfig {
	config := generatorStageConfig{}
	for _, option := range options {
		option.apply(&config)
	}

	if config.Logger() == nil {
		config.logger = logger.NewNoopLogger()
	}

	return &config
}
