package pipeline

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewProcessResult(t *testing.T) {
	type testCase struct {
		name                 string
		want                 ProcessResultCarrier[int]
		newProcessResultFunc func() ProcessResultCarrier[int]
		isOmitted            bool
	}
	tests := []testCase{
		{
			name: "OmittedProcessResult",
			newProcessResultFunc: func() ProcessResultCarrier[int] {
				return NewOmittedProcessResult[int]()
			},
			isOmitted: true,
			want:      &processResult[int]{omitted: true},
		},
		{
			name: "not OmittedProcessResult",
			newProcessResultFunc: func() ProcessResultCarrier[int] {
				return NewProcessResult[int](1, false)
			},
			isOmitted: false,
			want:      &processResult[int]{data: 1, omitted: false},
		},
		{
			name: "OmittedProcessResult 2",
			newProcessResultFunc: func() ProcessResultCarrier[int] {
				return NewProcessResult[int](1, true)
			},
			isOmitted: true,
			want:      &processResult[int]{data: 1, omitted: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processResult := tt.newProcessResultFunc()
			assert.Equal(t, tt.isOmitted, processResult.IsOmitted())
			assert.Equalf(t, tt.want, tt.newProcessResultFunc(), "TestNewProcessResult()")
		})
	}
}
