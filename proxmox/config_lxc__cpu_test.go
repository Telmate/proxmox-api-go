package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_LxcCPU_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  LxcCPU
		output error
	}{
		{name: `Valid nil`,
			input: LxcCPU{}},
		{name: `Valid all`,
			input: LxcCPU{
				Cores: util.Pointer(LxcCpuCores(0)),
				Limit: util.Pointer(LxcCpuLimit(0)),
				Units: util.Pointer(LxcCpuUnits(0))}},
		{name: `Invalid Cores`,
			input:  LxcCPU{Cores: util.Pointer(LxcCpuCores(8193))},
			output: errors.New(LxcCpuCores_Error_Invalid)},
		{name: `Invalid Limit`,
			input:  LxcCPU{Limit: util.Pointer(LxcCpuLimit(8193))},
			output: errors.New(LxcCpuLimit_Error_Invalid)},
		{name: `Invalid Units`,
			input:  LxcCPU{Units: util.Pointer(LxcCpuUnits(100001))},
			output: errors.New(LxcCpuUnits_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_LxcCpuCores_String(t *testing.T) {
	require.Equal(t, string("10"), LxcCpuCores(10).String())
}

func Test_LxcCpuCores_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  LxcCpuCores
		output error
	}{
		{name: `Valid minimum`,
			input: 0},
		{name: `Valid maximum`,
			input: 8192},
		{name: `Invalid 8193`,
			input:  8193,
			output: errors.New(LxcCpuCores_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_LxcCpuLimit_String(t *testing.T) {
	require.Equal(t, string("10"), LxcCpuLimit(10).String())
}

func Test_LxcCpuLimit_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  LxcCpuLimit
		output error
	}{
		{name: `Valid minimum`,
			input: 0},
		{name: `Valid maximum`,
			input: 8192},
		{name: `Invalid 8193`,
			input:  8193,
			output: errors.New(LxcCpuLimit_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_LxcCpuUnits_String(t *testing.T) {
	require.Equal(t, string("10"), LxcCpuUnits(10).String())
}

func Test_LxcCpuUnits_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  LxcCpuUnits
		output error
	}{
		{name: `Valid minimum`,
			input: 0},
		{name: `Valid maximum`,
			input: 100000},
		{name: `Invalid 8193`,
			input:  100001,
			output: errors.New(LxcCpuUnits_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
