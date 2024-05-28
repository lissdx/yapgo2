package logger

import "fmt"

func LoggerFactory(option ...Option) ILogger {
	lcgf := loggerConfig{}

	for _, opt := range option {
		opt.apply(&lcgf)
	}

	if lcgf.logImplementer == "" {
		lcgf.logImplementer = defaultLogImplementer
	}

	switch lcgf.logImplementer {
	case noopLogImplementor:
		return NewNoopLogger()
	case zapLogImplementor:
		if iLogger, err := NewWrappedZapLogger(lcgf); err != nil {
			panic(err.Error())
		} else {
			return iLogger
		}
	default:
		panic(fmt.Errorf("unknown log implementer %s", lcgf.logImplementer))
	}
}
