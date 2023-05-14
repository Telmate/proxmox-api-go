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
