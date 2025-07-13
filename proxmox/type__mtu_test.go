package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_mtu"
	"github.com/stretchr/testify/require"
)

func Test_MTU_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  MTU
		output error
	}{
		{name: `MTU Valid maximum`,
			input: MTU(test_data_mtu.MTU_Max_Legal())},
		{name: `MTU Valid minimum`,
			input: MTU(test_data_mtu.MTU_Min_Legal())},
		{name: `MTU Invalid maximum`,
			input:  MTU(test_data_mtu.MTU_Max_Illegal()),
			output: errors.New(MTU_Error_Invalid)},
		{name: `MTU Invalid minimum`,
			input:  MTU(test_data_mtu.MTU_Min_Illegal()),
			output: errors.New(MTU_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
