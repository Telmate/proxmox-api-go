package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func testData_RawConfigQemu_Agent_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `ALL`,
				input: rawConfigQemu{a: map[string]any{
					"agent": string("1,freeze-fs-on-backup=1,fstrim_cloned_disks=1,type=virtio")}},
				output: testQemuBaseConfig_get(ConfigQemu{Agent: &QemuGuestAgent{
					Enable: new(true),
					Freeze: new(true),
					FsTrim: new(true),
					Type:   new(QemuGuestAgentType_VirtIO)}})},
			{name: `Enabled`,
				input: rawConfigQemu{a: map[string]any{
					"agent": string("1")}},
				output: testQemuBaseConfig_get(ConfigQemu{Agent: &QemuGuestAgent{Enable: new(true)}})},
			{name: `Freeze`,
				input: rawConfigQemu{a: map[string]any{
					"agent": string("0,freeze-fs-on-backup=1")}},
				output: testQemuBaseConfig_get(ConfigQemu{Agent: &QemuGuestAgent{Enable: new(false), Freeze: new(true)}})},
			{name: `FsTrim`,
				input: rawConfigQemu{a: map[string]any{
					"agent": string("0,fstrim_cloned_disks=1")}},
				output: testQemuBaseConfig_get(ConfigQemu{Agent: &QemuGuestAgent{Enable: new(false), FsTrim: new(true)}})},
			{name: `Type`,
				input: rawConfigQemu{a: map[string]any{
					"agent": string("1,type=virtio")}},
				output: testQemuBaseConfig_get(ConfigQemu{Agent: &QemuGuestAgent{Enable: new(true), Type: new(QemuGuestAgentType_VirtIO)}})}}
	}
}

func Test_QemuGuestAgent_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  QemuGuestAgent
		output error
	}{
		{name: "Valid Type",
			input: QemuGuestAgent{Type: new(QemuGuestAgentType("isa"))}},
		{name: "Invalid Type",
			input:  QemuGuestAgent{Type: new(QemuGuestAgentType("invalid"))},
			output: errors.New(QemuGuestAgentType_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuGuestAgentType_Validate(t *testing.T) {
	t.Parallel()
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

func Test_RawConfigQemu_GetAgent(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Agent_Get().Inject(t,
		func(t *testing.T, raw RawConfigQemu, output *ConfigQemu, _ error) {
			require.Equal(t, output.Agent, raw.GetAgent())
		})
}

func Test_RawConfigQemu_Agent_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Agent_Get().Test(t)
}
