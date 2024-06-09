package logger

import (
	"fmt"
	"testing"
)

func Test_getArgs(t *testing.T) {
	const (
		strFormat = iota
		errorFormat
	)
	type args struct {
		args []interface{}
	}
	tests := []struct {
		name       string
		args       args
		formatType int
		wantResStr string
	}{
		{name: "Str Simple OK",
			args:       args{args: []interface{}{"test1"}},
			wantResStr: "test1",
		},
		{name: "Str Params OK",
			args:       args{args: []interface{}{"test1 %s, %d", "test2", 22}},
			wantResStr: "test1 test2, 22",
		},
		{name: "Str EXTRA param",
			args:       args{args: []interface{}{"test1", "test2", "test3"}},
			wantResStr: "test1%!(EXTRA string=test2, string=test3)",
		},
		{name: "Error simple",
			args:       args{args: []interface{}{fmt.Errorf("simple error")}},
			wantResStr: "simple error",
			formatType: errorFormat,
		},
		{name: "Error format as string",
			args:       args{args: []interface{}{"simple error: %w", fmt.Errorf("nested error")}},
			wantResStr: "simple error: nested error",
			formatType: errorFormat,
		},
		{name: "Error mixed format as string",
			args:       args{args: []interface{}{"%s, %d, simple error: %w", "the error", 22, fmt.Errorf("nested error")}},
			wantResStr: "the error, 22, simple error: nested error",
			formatType: errorFormat,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotResArgs := unsafeGetArgs(tt.args.args...)
			gotResStr := ""

			switch tt.formatType {
			case strFormat:
				gotResStr = fmt.Sprintf(gotStr, gotResArgs...)
			case errorFormat:
				gotResStr = fmt.Errorf(gotStr, gotResArgs...).Error()
			}

			if gotResStr != tt.wantResStr {
				t.Errorf("unsafeGetArgs() gotResStr = %v, want %v", gotResStr, tt.wantResStr)
			}

		})
	}
}
