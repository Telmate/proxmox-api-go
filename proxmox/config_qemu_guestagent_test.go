package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_QemuGuestAgent_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  QemuGuestAgent
		output error
	}{
		{name: "Valid Type",
			input: QemuGuestAgent{Type: util.Pointer(QemuGuestAgentType("isa"))}},
		{name: "Invalid Type",
			input:  QemuGuestAgent{Type: util.Pointer(QemuGuestAgentType("invalid"))},
			output: errors.New(QemuGuestAgentType_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuGuestAgentType_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  QemuGuestAgentType
		output error
	}{
		{name: `Valid ""`,
			input: QemuGuestAgentType("")},
		{name: "Valid lowercase",
			input: QemuGuestAgentType("virtio")},
		{name: "Valid UpperCase",
			input: QemuGuestAgentType("VirtIO")},
		{name: `Invalid`,
			input:  QemuGuestAgentType("invalid"),
			output: errors.New(QemuGuestAgentType_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
