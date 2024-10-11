package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_QemuNetworkInterfaceID_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  QemuNetworkInterfaceID
		output error
	}{
		{name: "Valid",
			input: QemuNetworkInterfaceID0},
		{name: "Invalid",
			input:  32,
			output: errors.New(QemuNetworkInterfaceID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
