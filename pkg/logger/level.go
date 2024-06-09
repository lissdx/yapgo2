package logger

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"strings"
)

// A Level is a logging priority. Higher levels are more important.
type Level int8

type LevelString string

var _ fmt.Stringer = (*Level)(nil)
var _ encoding.TextUnmarshaler = (*Level)(nil)
var _ encoding.TextMarshaler = (*Level)(nil)
var _ fmt.Stringer = LevelString("")

var _ LevelEnabler = (*Level)(nil)

const (
	TraceLevel Level = iota - 2
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	// PanicLevel logs a message, then panics.
	PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel

	_minLevel = TraceLevel
	_maxLevel = FatalLevel

	// InvalidLevel is an invalid value for Level.
	//
	// Core implementations may panic if they see messages of this level.
	InvalidLevel = _maxLevel + 1
)

const (
	TraceLevelString   LevelString = "TRACE"
	DebugLevelString   LevelString = "DEBUG"
	InfoLevelString    LevelString = "INFO"
	WarnLevelString    LevelString = "WARN"
	ErrorLevelString   LevelString = "ERROR"
	PanicLevelString   LevelString = "PANIC"
	FatalLevelString   LevelString = "FATAL"
	EmptyLevelString   LevelString = ""
	UnknownLevelString LevelString = "UNKNOWN"
)

var levelToStringMap = map[Level]string{
	TraceLevel: TraceLevelString.String(),
	DebugLevel: DebugLevelString.String(),
	InfoLevel:  InfoLevelString.String(),
	WarnLevel:  WarnLevelString.String(),
	ErrorLevel: ErrorLevelString.String(),
	PanicLevel: PanicLevelString.String(),
	FatalLevel: FatalLevelString.String(),
}

var errUnmarshalNilLevel = errors.New("can't unmarshal a nil *level")

// LevelEnabler decides whether a given logging level is enabled when logging a
// message.
//
// Enablers are intended to be used to implement deterministic filters;
// concerns like sampling are better implemented as a Core.
//
// Each concrete Level value implements a static LevelEnabler which returns
// true for itself and all higher logging levels. For example WarnLevel.Enabled()
// will return true for WarnLevel, ErrorLevel, DPanicLevel, PanicLevel, and
// FatalLevel, but return false for InfoLevel and DebugLevel.
type LevelEnabler interface {
	Enabled(Level) bool
}

// LevelString part
func (ls LevelString) String() string {
	switch ls {
	case TraceLevelString, DebugLevelString, InfoLevelString, WarnLevelString, ErrorLevelString, PanicLevelString, FatalLevelString:
		return string(ls)
	case EmptyLevelString:
		return string(InfoLevelString)
	default:
		return string(UnknownLevelString)
	}
}

func parseLevelString(text string) (levelStr LevelString, err error) {
	if levelStr = LevelString(strings.ToUpper(text)); levelStr == UnknownLevelString {
		err = fmt.Errorf("invalid logger level string: %s", text)
	}

	return
}

// Level part -------------------------------------
func (l *Level) String() string {
	if strLevel, ok := levelToStringMap[*l]; ok {
		return strLevel
	}
	return fmt.Sprintf("level(%d)", l)
}

// MarshalText marshals the Level to text. Note that the text representation
// drops the -level suffix (see example).
func (l *Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText unmarshals text to a level. Like MarshalText, UnmarshalText
// expects the text representation of a Level to drop the -level suffix (see
// example).
//
// In particular, this makes it easy to configure logging levels using YAML,
// TOML, or JSON files.
func (l *Level) UnmarshalText(text []byte) error {
	if l == nil {
		return errUnmarshalNilLevel
	}
	if !l.unmarshalText(text) && !l.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized level: %q", text)
	}
	return nil
}

func (l *Level) unmarshalText(text []byte) bool {
	levelString, err := parseLevelString(string(text))
	if err != nil {
		return false
	}

	switch levelString.String() {
	case string(TraceLevelString):
		*l = TraceLevel
	case string(DebugLevelString):
		*l = DebugLevel
	// Default level
	// make the zero value useful
	case string(InfoLevelString), string(EmptyLevelString):
		*l = InfoLevel
	case string(WarnLevelString):
		*l = WarnLevel
	case string(ErrorLevelString):
		*l = ErrorLevel
	case string(PanicLevelString):
		*l = PanicLevel
	case string(FatalLevelString):
		*l = FatalLevel
	default:
		return false
	}
	return true
}

func (l *Level) Enabled(level Level) bool {
	return level >= *l
}

// ParseLevel parses a level based on the lower-case or all-caps ASCII
// representation of the log level. If the provided ASCII representation is
// invalid an error is returned.
//
// This is particularly useful when dealing with text input to configure log
// levels.
func ParseLevel(text string) (Level, error) {
	var level Level
	err := level.UnmarshalText([]byte(text))
	return level, err
}
