package config_process_handler

type ErrorStrategy int

const defaultPrefProcessName = "Process"
const defaultRandomStrLength = 5

const (
	OnErrorRemove = iota
	OnErrorContinue
)

type processHandlerConfig struct {
	onErrorStrategy ErrorStrategy
}

type Option interface {
	apply(*processHandlerConfig)
}

type OptionFn func(*processHandlerConfig)

func (fn OptionFn) apply(config *processHandlerConfig) {
	fn(config)
}

// WithOnErrorContinueStrategy if OnErrorContinue is true
// the process stage will ignore errors on ProcessFn
// and will pass the result to the next stage in a pipeline
// WithOnErrorContinueStrategy has no effect on FilterFn
// in case of Filter Processing this option will be ignored
// the default strategy is OnErrorRemove
func WithOnErrorContinueStrategy() OptionFn {
	return func(config *processHandlerConfig) {
		config.onErrorStrategy = OnErrorContinue
	}
}

// WithOnErrorRemoveStrategy is opposite to WithOnErrorContinueStrategy
// as well has no effect on a Filter Processing
// it's the default strategy
func WithOnErrorRemoveStrategy[T, RS any]() OptionFn {
	return func(config *processHandlerConfig) {
		config.onErrorStrategy = OnErrorRemove
	}
}

func (pc *processHandlerConfig) IsOnErrorContinueStrategy() bool {
	return pc.onErrorStrategy == OnErrorContinue
}

func NewProcessHandlerConfig(options ...Option) *processHandlerConfig {
	config := processHandlerConfig{}
	for _, option := range options {
		option.apply(&config)
	}

	return &config
}
