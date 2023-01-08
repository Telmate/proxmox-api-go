package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ConfigContent_Template_error(t *testing.T) {
	require.Equal(t, errors.New("the value of (Node) may not be empty"), ConfigContent_Template{}.error("Node"))
}

func Test_ConfigContent_Template_mapToApiValues(t *testing.T) {
	testData := []struct {
		input  ConfigContent_Template
		output map[string]interface{}
	}{
		{
			input: ConfigContent_Template{
				Storage:  "a",
				Template: "b",
			},
			output: map[string]interface{}{
				"storage":  "a",
				"template": "b",
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.input.mapToApiValues())
	}
}

func Test_ConfigContent_Template_Validate(t *testing.T) {
	testData := []struct {
		input  ConfigContent_Template
		output error
	}{
		{
			input:  ConfigContent_Template{},
			output: ConfigContent_Template{}.error("Node"),
		},
		{
			input:  ConfigContent_Template{Node: "notEmpty"},
			output: ConfigContent_Template{}.error("Storage"),
		},
		{
			input: ConfigContent_Template{
				Node:    "notEmpty",
				Storage: "notEmpty",
			},
			output: ConfigContent_Template{}.error("Template"),
		},
		{
			input: ConfigContent_Template{
				Node:     "notEmpty",
				Storage:  "notEmpty",
				Template: "notEmpty",
			},
			output: nil,
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.input.Validate())
	}
}
