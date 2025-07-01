package proxmox

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_LxcBootMount_markMountChanges_Unsafe(t *testing.T) {
	tests := []struct {
		name    string
		input   LxcBootMount
		current *LxcBootMount
		output  lxcUpdateChanges
	}{
		{name: `resize`,
			input: LxcBootMount{
				SizeInKibibytes: util.Pointer(LxcMountSize(1051648))},
			current: &LxcBootMount{
				SizeInKibibytes: util.Pointer(LxcMountSize(1048576))},
			output: lxcUpdateChanges{
				resize: []lxcMountResize{{
					sizeInKibibytes: LxcMountSize(1051648),
					id:              "rootfs"}}}},
		{name: `move`,
			input: LxcBootMount{
				Storage: util.Pointer("local-zfs")},
			current: &LxcBootMount{
				Storage: util.Pointer("local-ext")},
			output: lxcUpdateChanges{
				move: []lxcMountMove{{
					storage: "local-zfs",
					id:      "rootfs"}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.markMountChanges_Unsafe(test.current))
		})
	}
}

func Test_LxcMountSize_String(t *testing.T) {
	tests := []struct {
		name   string
		input  LxcMountSize
		output string
	}{
		{name: "Kibibyte",
			input:  kibiByte,
			output: "1K"},
		{name: "Mebibyte",
			input:  mebiByte,
			output: "1M"},
		{name: "Gibibyte",
			input:  gibiByte,
			output: "1G"},
		{name: "Tebibyte",
			input:  tebiByte,
			output: "1T"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.String())
		})
	}
}
