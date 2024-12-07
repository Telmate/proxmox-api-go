package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
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

func Test_nodeFromUpID_Unsafe(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output task
	}{
		{name: "1",
			input: "UPID:pve-test:002860A9:051E01C1:67536165:qmmove:102:root@pam:",
			output: task{
				id:            "UPID:pve-test:002860A9:051E01C1:67536165:qmmove:102:root@pam:",
				node:          "pve-test",
				operationType: "qmmove",
				user: UserID{
					Name:  "root",
					Realm: "pam"}}},
		{name: "2",
			input: "UPID:pve:002860A9:051E01C1:67536165:qmshutdown:102:test-user@realm:",
			output: task{
				id:            "UPID:pve:002860A9:051E01C1:67536165:qmshutdown:102:test-user@realm:",
				node:          "pve",
				operationType: "qmshutdown",
				user: UserID{
					Name:  "test-user",
					Realm: "realm"}}},
	}
	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			tmpTask := &task{}
			tmpTask.mapToSDK_Unsafe(tests[i].input)
			require.Equal(t, tests[i].output.id, tmpTask.id)
			require.Equal(t, tests[i].output.node, tmpTask.node)
			require.Equal(t, tests[i].output.operationType, tmpTask.operationType)
			require.Equal(t, tests[i].output.user, tmpTask.user)
		})
	}
}
