package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_QemuCpuCores_Validate(t *testing.T) {
	testData := []struct {
		name   string
		input  QemuCpuCores
		output error
	}{
		// Invalid
		{name: `Invalid errors.New(QemuCpuCores_Error_LowerBound)`,
			input:  0,
			output: errors.New(QemuCpuCores_Error_LowerBound)},
		{name: `Invalid errors.New(QemuCpuCores_Error_UpperBound)`,
			input:  129,
			output: errors.New(QemuCpuCores_Error_UpperBound)},
		// Valid
		{name: `Valid LowerBound`,
			input: 1},
		{name: `Valid UpperBound`,
			input: 128},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.input.Validate(), test.output, test.name)
		})
	}
}
