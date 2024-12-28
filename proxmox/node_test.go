package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_node"
	"github.com/stretchr/testify/require"
)

func Test_NodeName_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		output error
	}{
		{name: `Valid NodeName`,
			input: test_data_node.NodeName_Legals()},
		{name: `Invalid Empty`,
			input:  []string{""},
			output: errors.New(NodeName_Error_Empty)},
		{name: `Invalid Length`,
			input:  []string{test_data_node.NodeName_Max_Illegal()},
			output: errors.New(NodeName_Error_Length)},
		{name: `Invalid Start Hyphen`,
			input:  test_data_node.NodeName_StartHyphens(),
			output: errors.New(NodeName_Error_HyphenStart)},
		{name: `Invalid End Hyphen`,
			input:  test_data_node.NodeName_EndHyphens(),
			output: errors.New(NodeName_Error_HyphenEnd)},
		{name: `Invalid Alphabetical`,
			input:  test_data_node.NodeName_Numeric_Illegal(),
			output: errors.New(NodeName_Error_Alphabetical)},
		{name: `Invalid Characters`,
			input:  test_data_node.NodeName_Error_Characters(),
			output: errors.New(NodeName_Error_Illegal)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, input := range test.input {
				require.Equal(t, test.output, NodeName(input).Validate())
			}
		})
	}
}

func Test_NodeName_String(t *testing.T) {
	tests := []struct {
		name   string
		input  NodeName
		output string
	}{
		{name: `Empty`,
			input:  "",
			output: ""},
		{name: `Valid`,
			input:  "node1",
			output: "node1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.String())
		})
	}
}
