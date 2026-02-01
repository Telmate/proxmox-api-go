package proxmox

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_PowerState_combine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   *PowerState
		current PowerState
		output  PowerState
	}{
		{name: `nil input`,
			input:   nil,
			current: PowerStateRunning,
			output:  PowerStateRunning},
		{name: `running input`,
			input:   util.Pointer(PowerStateRunning),
			current: PowerStateStopped,
			output:  PowerStateRunning},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.combine(test.current))
		})
	}
}

func Test_PowerState_parse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  string
		output PowerState
	}{
		{name: `running`,
			input:  "running",
			output: PowerStateRunning},
		{name: `stopped`,
			input:  "stopped",
			output: PowerStateStopped},
		{name: `unknown`,
			input:  "unknown_Fallback_value",
			output: PowerStateUnknown},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, PowerState(0).parse(test.input))
		})
	}
}

func Test_PowerState_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  PowerState
		output string
	}{
		{name: `running`,
			input:  PowerStateRunning,
			output: "running"},
		{name: `stopped`,
			input:  PowerStateStopped,
			output: "stopped"},
		{name: `unknown`,
			input:  PowerStateUnknown,
			output: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.String())
		})
	}
}
