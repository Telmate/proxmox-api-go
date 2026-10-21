package proxmox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_RawGuestStatus_Get(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawGuestStatus
		output GuestStatus
	}{
		{name: `HaState`,
			input: rawGuestStatus{a: map[string]any{"ha": map[string]any{
				"state": "running"}}},
			output: GuestStatus{HaState: new("running")}},
		{name: `Name`,
			input:  rawGuestStatus{a: map[string]any{"name": "test"}},
			output: GuestStatus{Name: "test"}},
		{name: `State`,
			input:  rawGuestStatus{a: map[string]any{"status": "running"}},
			output: GuestStatus{State: PowerStateRunning}},
		{name: `Uptime`,
			input:  rawGuestStatus{a: map[string]any{"uptime": float64(12345)}},
			output: GuestStatus{Uptime: time.Duration(12345) * time.Second}},
		{name: `All`,
			input: rawGuestStatus{a: map[string]any{
				"name":   "guest100",
				"status": "stopped",
				"uptime": float64(95673),
				"ha": map[string]any{
					"state": "stopped"}}},
			output: GuestStatus{
				HaState: new("stopped"),
				Name:    "guest100",
				State:   PowerStateStopped,
				Uptime:  time.Duration(95673) * time.Second}},
		{name: `Empty`,
			input:  rawGuestStatus{},
			output: GuestStatus{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, RawGuestStatus(&test.input).Get(), test.name)
		})
	}
}
