package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func testData_RawConfigLXC_CPU_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{
			{name: `Cores`,
				input:  rawConfigLXC{a: map[string]any{"cores": float64(1)}},
				output: lxcGetBaseConfig(ConfigLXC{CPU: &LxcCPU{Cores: new(LxcCpuCores(1))}})},
			{name: `Limit`,
				input:  rawConfigLXC{a: map[string]any{"cpulimit": string("2")}},
				output: lxcGetBaseConfig(ConfigLXC{CPU: &LxcCPU{Limit: new(LxcCpuLimit(2))}})},
			{name: `Units`,
				input:  rawConfigLXC{a: map[string]any{"cpuunits": float64(3)}},
				output: lxcGetBaseConfig(ConfigLXC{CPU: &LxcCPU{Units: new(LxcCpuUnits(3))}})}}
	}
}

func Test_LxcCPU_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  LxcCPU
		output error
	}{
		{name: `Valid nil`,
			input: LxcCPU{}},
		{name: `Valid all`,
			input: LxcCPU{
				Cores: new(LxcCpuCores(0)),
				Limit: new(LxcCpuLimit(0)),
				Units: new(LxcCpuUnits(0))}},
		{name: `Invalid Cores`,
			input:  LxcCPU{Cores: new(LxcCpuCores(8193))},
			output: errors.New(LxcCpuCores_Error_Invalid)},
		{name: `Invalid Limit`,
			input:  LxcCPU{Limit: new(LxcCpuLimit(8193))},
			output: errors.New(LxcCpuLimit_Error_Invalid)},
		{name: `Invalid Units`,
			input:  LxcCPU{Units: new(LxcCpuUnits(100001))},
			output: errors.New(LxcCpuUnits_Error_Maximum)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_LxcCpuCores_String(t *testing.T) {
	t.Parallel()
	require.Equal(t, string("10"), LxcCpuCores(10).String())
}

func Test_LxcCpuCores_Validate(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	require.Equal(t, string("10"), LxcCpuLimit(10).String())
}

func Test_LxcCpuLimit_Validate(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	require.Equal(t, string("10"), LxcCpuUnits(10).String())
}

func Test_LxcCpuUnits_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  LxcCpuUnits
		output error
	}{
		{name: `Valid default`,
			input: 0},
		{name: `Valid minimum`,
			input: 8},
		{name: `Valid maximum`,
			input: 100000},
		{name: `Invalid 7`,
			input:  7,
			output: errors.New(LxcCpuUnits_Error_Minimum)},
		{name: `Invalid 8193`,
			input:  100001,
			output: errors.New(LxcCpuUnits_Error_Maximum)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_RawConfigLXC_GetCPU(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_CPU_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, output.CPU, raw.GetCPU())
		})
}

func Test_RawConfigLXC_CPU_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_CPU_Get().Test(t)
}
