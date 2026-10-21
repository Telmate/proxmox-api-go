package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func test_Data_guestRef_pendingConfig() []struct {
	name    string
	input   []any
	pending bool
	output  map[string]any
} {
	return []struct {
		name    string
		input   []any
		pending bool
		output  map[string]any
	}{
		{name: `No pending`,
			input: []any{
				map[string]any{"key": string("test"), "value": string("value")},
				map[string]any{"key": string("cpu"), "value": float64(2)},
				map[string]any{"key": string("disks"), "value": string("sata0")}},
			output: map[string]any{
				"cpu":   float64(2),
				"disks": string("sata0"),
				"test":  string("value")}},
		{name: `Pending`,
			input: []any{
				map[string]any{"key": string("test"), "value": string("value")},
				map[string]any{"key": string("cores"), "value": float64(2), "pending": float64(3)},
				map[string]any{"key": string("disks"), "value": string("sata0")}},
			output: map[string]any{
				"cores": float64(2),
				"disks": string("sata0"),
				"test":  string("value")},
			pending: true},
		{name: `Delete`,
			input: []any{
				map[string]any{"key": string("test"), "value": string("value")},
				map[string]any{"key": string("cores"), "value": float64(2)},
				map[string]any{"key": string("tpmstate0"), "value": string("local-zfs:vm-1001-disk-2,size=4M,version=v1.2"), "delete": float64(1)}},
			output: map[string]any{
				"cores":     float64(2),
				"test":      string("value"),
				"tpmstate0": string("local-zfs:vm-1001-disk-2,size=4M,version=v1.2")},
			pending: true},
		{name: `Missing Value Pending (Real-World Case)`,
			input: []any{
				map[string]any{"key": string("bios"), "pending": string("ovmf")},
				map[string]any{"key": string("cores"), "value": float64(2)}},
			output:  map[string]any{"cores": float64(2)},
			pending: true},
	}
}

func Test_guestRef_pendingChanges(t *testing.T) {
	t.Parallel()
	tests := test_Data_guestRef_pendingConfig()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pending := parsePendingChanges(test.input)
			require.Equal(t, test.pending, pending)
		})
	}
}

func Benchmark_guestRef_pendingChanges(b *testing.B) {
	tests := test_Data_guestRef_pendingConfig()
	for i := range tests {
		b.Run(tests[i].name, func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				parsePendingChanges(tests[i].input)
			}
		})
	}
}

func Test_guestRef_pendingConfig(t *testing.T) {
	t.Parallel()
	tests := test_Data_guestRef_pendingConfig()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, pending := parsePendingConfig(test.input)
			require.Equal(t, test.pending, pending)
			require.Equal(t, test.output, output)
		})
	}
}

func Benchmark_guestRef_pendingConfig(b *testing.B) {
	tests := test_Data_guestRef_pendingConfig()
	for i := range tests {
		b.Run(tests[i].name, func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				parsePendingConfig(tests[i].input)
			}
		})
	}
}
