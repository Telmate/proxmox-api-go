package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
