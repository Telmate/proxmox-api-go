package proxmox

import (
	"crypto"
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_qemu"
	"github.com/stretchr/testify/require"
)

func Test_sshKeyUrlDecode(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output []crypto.PublicKey
	}{
		{name: "Decode",
			input:  test_data_qemu.PublicKey_Encoded_Input(),
			output: test_data_qemu.PublicKey_Decoded_Output()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, sshKeyUrlDecode(test.input))
		})
	}
}

// Test the encoding logic to encode the ssh keys
func Test_sshKeyUrlEncode(t *testing.T) {
	tests := []struct {
		name   string
		input  []crypto.PublicKey
		output string
	}{
		{name: "Encode",
			input:  test_data_qemu.PublicKey_Decoded_Input(),
			output: test_data_qemu.PublicKey_Encoded_Output()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, sshKeyUrlEncode(test.input))
		})
	}
}
