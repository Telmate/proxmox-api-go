package proxmox

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_TriBool_MarshalJSON(t *testing.T) {
	type testData struct {
		TriBool TriBool `json:"triBool"`
	}
	tests := []struct {
		name   string
		input  testData
		output []byte
		err    error
	}{
		{name: `True`,
			input:  testData{TriBool: TriBoolTrue},
			output: []byte(`{"triBool":"true"}`)},
		{name: `False`,
			input:  testData{TriBool: TriBoolFalse},
			output: []byte(`{"triBool":"false"}`)},
		{name: `None`,
			input:  testData{TriBool: TriBoolNone},
			output: []byte(`{"triBool":"none"}`)},
		{name: `Invalid`,
			input: testData{TriBool: TriBool(2)},
			err:   errors.New(TriBool_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := json.Marshal(test.input)
			require.Equal(t, test.output, output)
			if test.err == nil {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, test.err.Error())
			}
		})
	}
}

func Test_TriBool_UnmarshalJSON(t *testing.T) {
	type testData struct {
		TriBool TriBool `json:"triBool"`
	}
	tests := []struct {
		name   string
		input  string
		output testData
		err    error
	}{
		{name: `True`,
			input:  `{"triBool":"true"}`,
			output: testData{TriBool: TriBoolTrue}},
		{name: `Yes`,
			input:  `{"triBool":"yes"}`,
			output: testData{TriBool: TriBoolTrue}},
		{name: `On`,
			input:  `{"triBool":"on"}`,
			output: testData{TriBool: TriBoolTrue}},
		{name: `False`,
			input:  `{"triBool":"false"}`,
			output: testData{TriBool: TriBoolFalse}},
		{name: `No`,
			input:  `{"triBool":"no"}`,
			output: testData{TriBool: TriBoolFalse}},
		{name: `Off`,
			input:  `{"triBool":"off"}`,
			output: testData{TriBool: TriBoolFalse}},
		{name: `None`,
			input:  `{"triBool":"none"}`,
			output: testData{TriBool: TriBoolNone}},
		{name: `""`,
			input:  `{"triBool":""}`,
			output: testData{TriBool: TriBoolNone}},
		{name: `Invalid`,
			input: `{"triBool":"invalid"}`,
			err:   errors.New(TriBool_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output testData
			err := json.Unmarshal([]byte(test.input), &output)
			require.Equal(t, test.output, output)
			require.Equal(t, test.err, err)
		})
	}
}

func Test_TriBool_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  TriBool
		output error
	}{
		{name: `Valid True`,
			input: 1},
		{name: `Valid False`,
			input: -1},
		{name: `Valid None`},
		{name: `Invalid upperBound`,
			input:  2,
			output: errors.New(TriBool_Error_Invalid)},
		{name: `Invalid lowerBound`,
			input:  -2,
			output: errors.New(TriBool_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
