package parse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Int(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		output int
		err    bool
	}{
		{name: `float64 negative`,
			input:  float64(-1),
			output: -1},
		{name: `float64 positive`,
			input:  float64(1),
			output: 1},
		{name: `string invalid`,
			input: "a",
			err:   true},
		{name: `string negative`,
			input:  "-1",
			output: -1},
		{name: `string positive`,
			input:  "1",
			output: 1},
		{name: `invalid type`,
			input: interface{}(nil),
			err:   true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpOutput, tmpErr := Int(test.input)
			if test.err {
				require.Error(t, tmpErr)
			} else {
				require.NoError(t, tmpErr)
				require.Equal(t, test.output, tmpOutput)
			}
		})
	}
}

func Test_Uint(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		output uint
		err    bool
	}{
		{name: `float64 negative`,
			input: float64(-1),
			err:   true},
		{name: `float64 positive`,
			input:  float64(1),
			output: 1},
		{name: `string negative`,
			input: "-1",
			err:   true},
		{name: `string invalid`,
			input: "a",
			err:   true},
		{name: `string positive`,
			input:  "1",
			output: 1},
		{name: `invalid type`,
			input: interface{}(nil),
			err:   true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpOutput, tmpErr := Uint(test.input)
			if test.err {
				require.Error(t, tmpErr)
			} else {
				require.NoError(t, tmpErr)
				require.Equal(t, test.output, tmpOutput)
			}
		})
	}
}
