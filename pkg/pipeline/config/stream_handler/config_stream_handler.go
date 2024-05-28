package config_stream_handler

import (
	"fmt"
	"github.com/lissdx/yapgo2/pkg/logger"
	"github.com/lissdx/yapgo2/pkg/utils"
)

const defaultPrefStreamHandlerName = "StreamHandler"
const defaultRandomStrLength = 5

type streamHandlerConfig struct {
	name   string
	logger logger.ILogger
}

type Option interface {
	apply(*streamHandlerConfig)
}

type OptionFn func(*streamHandlerConfig)

func (fn OptionFn) apply(config *streamHandlerConfig) {
	fn(config)
}

func WithName(name string) OptionFn {
	return func(config *streamHandlerConfig) {
		config.name = name
	}
}

func WithLogger(logger logger.ILogger) OptionFn {
	return func(config *streamHandlerConfig) {
		config.logger = logger
	}
}

func (pc *streamHandlerConfig) Name() string {
	return pc.name
}
func (pc *streamHandlerConfig) Logger() logger.ILogger {
	return pc.logger
}

func NewStreamHandlerConfig(options ...Option) *streamHandlerConfig {
	config := streamHandlerConfig{}
	for _, option := range options {
		option.apply(&config)
	}

	if config.name == "" {
		config.name = fmt.Sprintf("%s_%s", defaultPrefStreamHandlerName, utils.RandomStr(defaultRandomStrLength))
	}

	if config.logger == nil {
		config.logger = logger.NewNoopLogger()
	}

	return &config
}
