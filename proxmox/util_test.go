package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_keyExists(t *testing.T) {
	tests := []struct {
		name   string
		input  []interface{}
		key    string
		output bool
	}{
		{name: "key empty",
			input: []interface{}{
				map[string]interface{}{"aaa": "", "bbb": "", "ccc": ""},
				map[string]interface{}{"aab": "", "bba": "", "cca": ""},
				map[string]interface{}{"aac": "", "bbc": "", "ccb": ""},
			},
		},
		{name: "Key in map",
			input: []interface{}{
				map[string]interface{}{"aaa": "", "bbb": "", "ccc": ""},
				map[string]interface{}{"aab": "", "bba": "", "cca": ""},
				map[string]interface{}{"aac": "", "bbc": "", "ccb": ""},
			},
			key:    "bba",
			output: true,
		},
		{name: "Key not in map",
			input: []interface{}{
				map[string]interface{}{"aaa": "", "bbb": "", "ccc": ""},
				map[string]interface{}{"aab": "", "bba": "", "cca": ""},
				map[string]interface{}{"aac": "", "bbc": "", "ccb": ""},
			},
			key: "ddd",
		},
		{name: "no array",
			key: "aaa",
		},
		{name: "no keys",
			input: []interface{}{map[string]interface{}{}},
			key:   "aaa",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, keyExists(test.input, test.key), test.name)
		})
	}
}

func Test_floatToTrimmedString(t *testing.T) {
	type testStruct struct {
		number   float64
		decimals uint8
	}
	tests := []struct {
		name   string
		input  testStruct
		output string
	}{
		{name: "float64",
			input:  testStruct{number: 1.23456789, decimals: 8},
			output: "1.23456789",
		},
		{name: "float32",
			input:  testStruct{number: float64(float32(1.23456789)), decimals: 8},
			output: "1.23456788",
		},
		{name: "no decimal trimmed",
			input:  testStruct{number: 1.0000000, decimals: 10},
			output: "1",
		},
		{name: "one decimal trimmed",
			input:  testStruct{number: 10.3000000, decimals: 10},
			output: "10.3",
		},
		{name: "tree decimal trimmed",
			input:  testStruct{number: 45.73300000, decimals: 10},
			output: "45.733",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, floatToTrimmedString(test.input.number, test.input.decimals), test.name)
		})
	}
}

func Test_splitStringOfSettings(t *testing.T) {
	testData := []struct {
		Input  string
		Output map[string]interface{}
	}{
		{
			Input: "setting=a,thing=b,randomString,doubleTest=value=equals,object=test",
			Output: map[string]interface{}{
				"setting":      "a",
				"thing":        "b",
				"randomString": "",
				"doubleTest":   "value=equals",
				"object":       "test",
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.Output, splitStringOfSettings(e.Input))
	}
}
