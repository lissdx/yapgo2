package logger

import "testing"

func TestLevelString_String(t *testing.T) {
	tests := []struct {
		name string
		ls   LevelString
		want string
	}{
		{
			name: "trace", ls: LevelString("trace"), want: "UNKNOWN",
		},
		{
			name: "TRACE", ls: LevelString("TRACE"), want: "TRACE",
		},
		{
			name: "debug", ls: LevelString("debug"), want: "UNKNOWN",
		},
		{
			name: "DEBUG", ls: LevelString("DEBUG"), want: "DEBUG",
		},
		{
			name: "info", ls: LevelString("info"), want: "UNKNOWN",
		},
		{
			name: "INFO", ls: LevelString("INFO"), want: "INFO",
		},
		{
			name: "INFO empty", ls: LevelString(""), want: "INFO",
		},
		{
			name: "warn", ls: LevelString("warn"), want: "UNKNOWN",
		},
		{
			name: "WARN", ls: LevelString("WARN"), want: "WARN",
		},
		{
			name: "error", ls: LevelString("error"), want: "UNKNOWN",
		},
		{
			name: "ERROR", ls: LevelString("ERROR"), want: "ERROR",
		},
		{
			name: "panic", ls: LevelString("panic"), want: "UNKNOWN",
		},
		{
			name: "PANIC", ls: LevelString("PANIC"), want: "PANIC",
		},
		{
			name: "fatal", ls: LevelString("fatal"), want: "UNKNOWN",
		},
		{
			name: "FATAL", ls: LevelString("FATAL"), want: "FATAL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ls.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name    string
		args    args
		want    Level
		wantErr bool
	}{
		{name: "TRACE", args: args{text: "TRACE"}, want: TraceLevel, wantErr: false},
		{name: "trace", args: args{text: "trace"}, want: TraceLevel, wantErr: false},
		{name: "DEBUG", args: args{text: "DEBUG"}, want: DebugLevel, wantErr: false},
		{name: "debug", args: args{text: "debug"}, want: DebugLevel, wantErr: false},
		{name: "INFO", args: args{text: "INFO"}, want: InfoLevel, wantErr: false},
		{name: "info", args: args{text: "info"}, want: InfoLevel, wantErr: false},
		{name: "INFO empty", args: args{text: ""}, want: InfoLevel, wantErr: false},
		{name: "WARN", args: args{text: "WARN"}, want: WarnLevel, wantErr: false},
		{name: "warn", args: args{text: "warn"}, want: WarnLevel, wantErr: false},
		{name: "error", args: args{text: "error"}, want: ErrorLevel, wantErr: false},
		{name: "ERROR", args: args{text: "ERROR"}, want: ErrorLevel, wantErr: false},
		{name: "panic", args: args{text: "panic"}, want: PanicLevel, wantErr: false},
		{name: "PANIC", args: args{text: "PANIC"}, want: PanicLevel, wantErr: false},
		{name: "fatal", args: args{text: "fatal"}, want: FatalLevel, wantErr: false},
		{name: "FATAL", args: args{text: "FATAL"}, want: FatalLevel, wantErr: false},
		{name: "FATAL_DEBUG", args: args{text: "FATAL_DEBUG"}, wantErr: true},
		{name: "unknown", args: args{text: "unknown"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLevel(tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseLevel() got = %v, want %v", got, tt.want)
			}
		})
	}
}
