package logger

var _ ILogger = (*noopLogger)(nil)

type noopLogger struct {
}

func (nl *noopLogger) TraceIsEnabled() bool {
	return false
}

func (nl *noopLogger) DebugIsEnabled() bool {
	return false
}

func (nl *noopLogger) InfoIsEnabled() bool {
	return false
}

func (nl *noopLogger) WarnIsEnabled() bool {
	return false
}

func (nl *noopLogger) ErrorIsEnabled() bool {
	return false
}

func (nl *noopLogger) PanicIsEnabled() bool {
	return false
}

func (nl *noopLogger) FatalIsEnabled() bool {
	return false
}

func (nl *noopLogger) Trace(args ...interface{}) {
}

func (nl *noopLogger) Debug(args ...interface{}) {
}

func (nl *noopLogger) Info(args ...interface{}) {
}

func (nl *noopLogger) Warn(args ...interface{}) {
}

func (nl *noopLogger) Error(args ...interface{}) {
}

func (nl *noopLogger) Panic(args ...interface{}) {
}

func (nl *noopLogger) Fatal(args ...interface{}) {
}

func NewNoopLogger() ILogger {
	return &noopLogger{}
}
