package logger

// ILogger Logger interface
type ILogger interface {
	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Panic(args ...interface{})
	Fatal(args ...interface{})
	TraceIsEnabled() bool
	DebugIsEnabled() bool
	InfoIsEnabled() bool
	WarnIsEnabled() bool
	ErrorIsEnabled() bool
	PanicIsEnabled() bool
	FatalIsEnabled() bool
}
