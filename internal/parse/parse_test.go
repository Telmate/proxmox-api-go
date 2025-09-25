package parse

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Bool(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		output bool
		err    error
	}{
		{name: `float64 1`,
			input:  float64(1),
			output: true},
		{name: `float64 0`,
			input:  float64(0),
			output: false},
		{name: `string empty`,
			input: "",
			err:   errors.New(Empty)},
		{name: `string 1`,
			input:  "1",
			output: true},
		{name: `string 1 suffixed`,
			input:  "1;suffixed",
			output: true},
		{name: `string 0`,
			input:  "0",
			output: false},
		{name: `string 0 suffixed`,
			input:  "0;suffixed",
			output: false},
		{name: `invalid type`,
			input: any(nil),
			err:   errors.New(InvalidType)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpOutput, tmpErr := Bool(test.input)
			require.Equal(t, test.err, tmpErr)
			require.Equal(t, test.output, tmpOutput)
		})
	}
}

func Test_Int(t *testing.T) {
	tests := []struct {
		name   string
		input  any
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
			input: any(nil),
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
		input  any
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
			input: any(nil),
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
