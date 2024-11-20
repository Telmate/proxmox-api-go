package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ensurePrefix(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		prefix string
		output string
	}{
		{name: "only prefix",
			input:  "",
			prefix: "prefix",
			output: "prefix"},
		{name: "prefix and text",
			input:  "text",
			prefix: "prefix",
			output: "prefixtext"},
		{name: "prefix already in text",
			input:  "prefixtext",
			prefix: "prefix",
			output: "prefixtext"},
		{name: "prefix is text",
			input:  "prefix",
			prefix: "prefix",
			output: "prefix"},
		{name: "no prefix",
			input:  "text",
			prefix: "",
			output: "text"},
		{name: "no text or prefix",
			input:  "",
			prefix: "",
			output: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, ensurePrefix(test.prefix, test.input), test.name)
		})
	}
}

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
		Output map[string]string
	}{
		{
			Input: "setting=a,thing=b,randomString,doubleTest=value=equals,object=test",
			Output: map[string]string{
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

func Test_subtractArray(t *testing.T) {
	type testInput struct {
		A []string
		B []string
	}
	tests := []struct {
		name   string
		input  testInput
		output []string
	}{
		{name: "A and B different",
			input:  testInput{A: []string{"a", "b", "c"}, B: []string{"a", "b"}},
			output: []string{"c"}},
		{name: "A and B empty"},
		{name: "A and B same",
			input: testInput{A: []string{"a", "b", "c"}, B: []string{"a", "b", "c"}}},
		{name: "A empty",
			input: testInput{B: []string{"a", "b", "c"}}},
		{name: "B empty",
			input:  testInput{A: []string{"a", "b", "c"}},
			output: []string{"a", "b", "c"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, subtractArray(test.input.A, test.input.B))
		})
	}
}
