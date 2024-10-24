package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Vlan_String(t *testing.T) {
	tests := []struct {
		name   string
		input  Vlan
		output string
	}{
		{name: `0`,
			input:  0,
			output: "0"},
		{name: `4095`,
			input:  4095,
			output: "4095"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.String())
		})
	}
}

func Test_Vlan_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  Vlan
		output error
	}{
		{name: `Valid Maximum`,
			input: 4095},
		{name: `Valid Minimum`,
			input: 0},
		{name: `Invalid`,
			input:  4096,
			output: errors.New(Vlan_Error_Invalid),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_Vlans_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  Vlans
		output error
	}{
		{name: `Valid`,
			input: Vlans{0, 4095}},
		{name: `Valid Empty`,
			input: Vlans{}},
		{name: `Valid Duplicate`,
			input: Vlans{23, 78, 23, 90, 78},
		},
		{name: `Invalid`,
			input:  Vlans{0, 4095, 4096},
			output: errors.New(Vlan_Error_Invalid),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
