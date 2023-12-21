package proxmox

import (
	"testing"
)

func TestParamsTo(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]interface{}
		output []string
	}{{
		name: "basic_values",
		input: map[string]interface{}{
			"poolid": "test",
		},
		output: []string{"poolid=test"},
	}, {
		name: "multiple_values",
		input: map[string]interface{}{
			"poolid":  "test",
			"comment": "comment",
		},
		output: []string{"poolid=test&comment=comment", "comment=comment&poolid=test"},
	}, {
		name: "empty_values_are_removed",
		input: map[string]interface{}{
			"poolid":  "test",
			"comment": "",
		},
		output: []string{"poolid=test"},
	}, {
		name: "array",
		input: map[string]interface{}{
			"command": []string{"bash", "-c", "echo test"},
		},
		output: []string{"command=bash&command=-c&command=echo+test"},
	}}

	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			output := string(ParamsToBody(test.input))
			if !inArray(test.output, output) {
				t.Errorf("%s: expected `%+v`, got `%+v`",
					test.name, test.output, output)
			}
		})
	}
}

func TestParamsToWithEmpty(t *testing.T) {
	tests := []struct {
		name         string
		input        map[string]interface{}
		allowedEmpty []string
		output       []string
	}{{
		name: "basic_values",
		input: map[string]interface{}{
			"poolid": "test",
		},
		output: []string{"poolid=test"},
	}, {
		name: "multiple_values",
		input: map[string]interface{}{
			"poolid":  "test",
			"comment": "comment",
		},
		output: []string{"poolid=test&comment=comment", "comment=comment&poolid=test"},
	}, {
		name: "empty_values_are_removed",
		input: map[string]interface{}{
			"poolid":  "test",
			"comment": "",
		},
		output: []string{"poolid=test"},
	}, {
		name: "explicit_empty_values_are_not_removed",
		input: map[string]interface{}{
			"poolid":  "test",
			"comment": "",
		},
		allowedEmpty: []string{"comment"},
		output:       []string{"poolid=test&comment=", "comment=&poolid=test"},
	}}

	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			output := string(ParamsToBodyWithEmpty(test.input, test.allowedEmpty))
			if !inArray(test.output, output) {
				t.Errorf("%s: expected `%+v`, got `%+v`",
					test.name, test.output, output)
			}
		})
	}
}
