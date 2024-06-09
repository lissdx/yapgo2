package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ ILogger = (*zapAdapter)(nil)

type zapAdapter struct {
	logger        *zap.Logger
	loggerSugared *zap.SugaredLogger
	level         Level
}

func (z *zapAdapter) TraceIsEnabled() bool {
	return z.level.Enabled(TraceLevel)
}

func (z *zapAdapter) DebugIsEnabled() bool {
	return z.level.Enabled(DebugLevel)
}

func (z *zapAdapter) InfoIsEnabled() bool {
	return z.level.Enabled(InfoLevel)
}

func (z *zapAdapter) WarnIsEnabled() bool {
	return z.level.Enabled(WarnLevel)
}

func (z *zapAdapter) ErrorIsEnabled() bool {
	return z.level.Enabled(ErrorLevel)
}

func (z *zapAdapter) PanicIsEnabled() bool {
	return z.level.Enabled(PanicLevel)
}

func (z *zapAdapter) FatalIsEnabled() bool {
	return z.level.Enabled(FatalLevel)
}

func newZapAdapter(logger *zap.Logger, level Level) ILogger {
	return &zapAdapter{logger: logger, level: level, loggerSugared: logger.Sugar()}
}

func (z *zapAdapter) Trace(args ...interface{}) {
	if z.level.Enabled(TraceLevel) {
		logStr, a := unsafeGetArgs(args...)
		z.logger.Debug(fmt.Sprintf(logStr, a...), zap.String("TRACE", "ON"))
	}
}

func (z *zapAdapter) Debug(args ...interface{}) {
	if z.level.Enabled(DebugLevel) {
		logStr, a := unsafeGetArgs(args...)
		z.loggerSugared.Debugf(logStr, a...)
	}
}

func (z *zapAdapter) Info(args ...interface{}) {
	if z.level.Enabled(InfoLevel) {
		logStr, a := unsafeGetArgs(args...)
		z.logger.Info(fmt.Sprintf(logStr, a...))
	}
}

func (z *zapAdapter) Warn(args ...interface{}) {
	if z.level.Enabled(WarnLevel) {
		logStr, a := unsafeGetArgs(args...)
		z.loggerSugared.Warnf(logStr, a...)
	}
}

func (z *zapAdapter) Error(args ...interface{}) {
	logStr, a := unsafeGetArgs(args...)
	z.logger.Error(fmt.Errorf(logStr, a...).Error())
}

func (z *zapAdapter) Panic(args ...interface{}) {
	logStr, a := unsafeGetArgs(args...)
	z.loggerSugared.Panicf(logStr, a...)
}

func (z *zapAdapter) Fatal(args ...interface{}) {
	logStr, a := unsafeGetArgs(args...)
	z.loggerSugared.Fatalf(logStr, a...)
}

func NewWrappedZapLogger(lcfg loggerConfig) (ILogger, error) {
	var zapLoggerConfig zap.Config

	lcfgLogLevel, err := ParseLevel(lcfg.loggerLevelString)
	if err != nil {
		return nil, err
	}

	var zapLogLevel zapcore.Level
	switch lcfgLogLevel {
	// shadow TraceLevel for zap logger
	// (trace level currently is not supported by zap logger)
	case TraceLevel:
		if err = zapLogLevel.UnmarshalText([]byte(DebugLevelString)); err != nil {
			return nil, err
		}
	default:
		if err = zapLogLevel.UnmarshalText([]byte(lcfgLogLevel.String())); err != nil {
			return nil, err
		}
	}

	zapLoggerConfig = zap.Config{
		Development:      zapLogLevel.Enabled(zapcore.DebugLevel),
		Encoding:         string(lcfg.zapEncoding),
		Level:            zap.NewAtomicLevelAt(zapLogLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:     "time",
			EncodeTime:  zapcore.ISO8601TimeEncoder,
			MessageKey:  "message",
			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,
		},
	}

	if lcfg.isColored && zapLoggerConfig.Encoding == zapConsoleEncoding {
		zapLoggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	res, err := zapLoggerConfig.Build()

	if err != nil {
		return nil, err
	}

	return newZapAdapter(res, lcfgLogLevel), nil
}
