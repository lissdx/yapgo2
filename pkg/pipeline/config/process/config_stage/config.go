package config_stage

import (
	"fmt"
	"github.com/lissdx/yapgo2/pkg/logger"
	"github.com/lissdx/yapgo2/pkg/utils"
)

type ErrorStrategy int

const defaultPrefProcessStageName = "ProcessStage"
const defaultRandomStrLength = 5

type processStageConfig struct {
	name   string
	logger logger.ILogger
}

type Option interface {
	apply(*processStageConfig)
}

type OptionFn func(*processStageConfig)

func (fn OptionFn) apply(config *processStageConfig) {
	fn(config)
}

func WithName(name string) OptionFn {
	return func(config *processStageConfig) {
		config.name = name
	}
}

func WithLogger(logger logger.ILogger) OptionFn {
	return func(config *processStageConfig) {
		config.logger = logger
	}
}

func (pc *processStageConfig) Name() string {
	return pc.name
}

func (pc *processStageConfig) Logger() logger.ILogger {
	return pc.logger
}

func NewProcessStageConfig(options ...Option) *processStageConfig {
	config := processStageConfig{}
	for _, option := range options {
		option.apply(&config)
	}

	if config.name == "" {
		config.name = fmt.Sprintf("%s_%s", defaultPrefProcessStageName, utils.RandomStr(defaultRandomStrLength))
	}

	if config.logger == nil {
		config.logger = logger.NewNoopLogger()
	}

	return &config
}
