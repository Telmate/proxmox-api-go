package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_CpuType_Error(t *testing.T) {
	testData := []struct {
		name    string
		input   Version
		compare error
	}{
		{name: `v8 > v7`,
			input:   Version{Major: 8},
			compare: CpuType("").Error(Version{Major: 7, Minor: 255, Patch: 255})},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			require.Greater(t, len(CpuType("").Error(test.input).Error()), len(test.compare.Error()), test.name)
		})
	}
}

func Test_CpuType_Validate(t *testing.T) {
	type testInput struct {
		config  CpuType
		version Version
	}
	testData := []struct {
		name   string
		input  testInput
		output error
	}{
		// Invalid
		{name: `Invalid`,
			input: testInput{
				config:  CpuType("gibbers"),
				version: Version{}.max()},
			output: CpuType("").Error(Version{}.max())},
		{name: `Invalid V7`,
			input: testInput{
				config:  CpuType_AmdEPYCRomeV2,
				version: Version{Major: 7}.max()},
			output: CpuType("").Error(Version{Major: 7}.max())},
		// Valid
		{name: `Valid empty`,
			input: testInput{
				config:  CpuType(""),
				version: Version{}.max()}},
		{name: `Valid normal`,
			input: testInput{
				config:  CpuType("Skylake-Server-noTSX-IBRS"),
				version: Version{}.max()}},
		{name: `Valid lowercase`,
			input: testInput{
				config:  CpuType("skylakeclientnotsxibrs"),
				version: Version{}.max()}},
		{name: `Valid weird`,
			input: testInput{config: CpuType("S-k__-Yl_-A--k-e__-Se-R-v-__Er--n-OTs_X---I-_br-S"),
				version: Version{}.max()}},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.input.config.Validate(test.input.version), test.output, test.name)
		})
	}
}

func Test_CpuVirtualCores_Validate(t *testing.T) {
	type testInput struct {
		virtualCores CpuVirtualCores
		cores        *QemuCpuCores
		sockets      *QemuCpuSockets
		current      *QemuCPU
	}
	testData := []struct {
		name   string
		input  testInput
		output error
	}{
		// Invalid
		{name: `Invalid Create`,
			input: testInput{
				virtualCores: 5,
				cores:        util.Pointer(QemuCpuCores(2)),
				sockets:      util.Pointer(QemuCpuSockets(2))},
			output: CpuVirtualCores(4).Error()},
		{name: `Invalid Update Cores`,
			input: testInput{
				virtualCores: 8,
				cores:        util.Pointer(QemuCpuCores(1)),
				current: &QemuCPU{
					Cores:   util.Pointer(QemuCpuCores(3)),
					Sockets: util.Pointer(QemuCpuSockets(2))}},
			output: CpuVirtualCores(2).Error()},
		{name: `Invalid Update Sockets`,
			input: testInput{
				virtualCores: 10,
				sockets:      util.Pointer(QemuCpuSockets(2)),
				current: &QemuCPU{
					Cores:   util.Pointer(QemuCpuCores(4)),
					Sockets: util.Pointer(QemuCpuSockets(3))}},
			output: CpuVirtualCores(8).Error()},
		{name: `Invalid Update`,
			input: testInput{
				virtualCores: 16,
				current: &QemuCPU{
					Cores:   util.Pointer(QemuCpuCores(4)),
					Sockets: util.Pointer(QemuCpuSockets(3))}},
			output: CpuVirtualCores(12).Error()},
		// Valid
		{name: `Valid Create`,
			input: testInput{
				virtualCores: 1,
				cores:        util.Pointer(QemuCpuCores(1)),
				sockets:      util.Pointer(QemuCpuSockets(1))}},
		{name: `Valid Update Cores`,
			input: testInput{
				virtualCores: 2,
				cores:        util.Pointer(QemuCpuCores(2)),
				current: &QemuCPU{
					Cores:   util.Pointer(QemuCpuCores(1)),
					Sockets: util.Pointer(QemuCpuSockets(1))}}},
		{name: `Valid Update Sockets`,
			input: testInput{
				virtualCores: 3,
				sockets:      util.Pointer(QemuCpuSockets(3)),
				current: &QemuCPU{
					Cores:   util.Pointer(QemuCpuCores(1)),
					Sockets: util.Pointer(QemuCpuSockets(4))}}},
		{name: `Valid Update`,
			input: testInput{
				virtualCores: 4,
				current: &QemuCPU{
					Cores:   util.Pointer(QemuCpuCores(2)),
					Sockets: util.Pointer(QemuCpuSockets(2))}}},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.input.virtualCores.Validate(test.input.cores, test.input.sockets, test.input.current), test.output, test.name)
		})
	}
}

func Test_QemuCPU_Validate(t *testing.T) {
	baseConfig := func(config QemuCPU) QemuCPU {
		if config.Cores == nil {
			config.Cores = util.Pointer(QemuCpuCores(1))
		}
		return config
	}
	type testInput struct {
		config  QemuCPU
		current *QemuCPU
		version Version
	}
	testData := []struct {
		name   string
		input  testInput
		output error
	}{
		// Invalid
		{name: `Invalid errors.New(QemuCpuCores_Error_LowerBound)`,
			input:  testInput{config: QemuCPU{Cores: util.Pointer(QemuCpuCores(0))}},
			output: errors.New(QemuCpuCores_Error_LowerBound)},
		{name: `Invalid errors.New(QemuCPU_Error_CoresRequired)`,
			input:  testInput{config: QemuCPU{}},
			output: errors.New(QemuCPU_Error_CoresRequired)},
		{name: `Invalid errors.New(QemuCpuSockets_Error_LowerBound)`,
			input:  testInput{config: baseConfig(QemuCPU{Sockets: util.Pointer(QemuCpuSockets(0))})},
			output: errors.New(QemuCpuSockets_Error_LowerBound)},
		{name: `Invalid CpuVirtualCores(1).Error() 1 1`,
			input: testInput{config: QemuCPU{
				Cores:        util.Pointer(QemuCpuCores(1)),
				Sockets:      util.Pointer(QemuCpuSockets(1)),
				VirtualCores: util.Pointer(CpuVirtualCores(2))}},
			output: CpuVirtualCores(1).Error()},
		{name: `Invalid Type`,
			input: testInput{
				config:  baseConfig(QemuCPU{Type: util.Pointer(CpuType("gibbers"))}),
				version: Version{}.max()},
			output: CpuType("").Error(Version{}.max())},
		// Valid
		{name: `Valid Maximum`,
			input: testInput{
				config: QemuCPU{
					Cores:        util.Pointer(QemuCpuCores(128)),
					Sockets:      util.Pointer(QemuCpuSockets(4)),
					Type:         util.Pointer(CpuType(cpuType_AmdEPYCRomeV2_Lower)),
					VirtualCores: util.Pointer(CpuVirtualCores(512))},
				version: Version{}.max()}},
		{name: `Valid Minimum`,
			input: testInput{config: QemuCPU{
				Cores:        util.Pointer(QemuCpuCores(128)),
				Sockets:      util.Pointer(QemuCpuSockets(4)),
				Type:         util.Pointer(CpuType("")),
				VirtualCores: util.Pointer(CpuVirtualCores(0))},
				version: Version{}.max()}},
		{name: `Valid Update`,
			input: testInput{
				config:  QemuCPU{},
				current: &QemuCPU{}}},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.input.config.Validate(test.input.current, test.input.version), test.output, test.name)
		})
	}
}

func Test_QemuCpuCores_Validate(t *testing.T) {
	testData := []struct {
		name   string
		input  QemuCpuCores
		output error
	}{
		// Invalid
		{name: `Invalid errors.New(QemuCpuCores_Error_LowerBound)`,
			input:  0,
			output: errors.New(QemuCpuCores_Error_LowerBound)},
		{name: `Invalid errors.New(QemuCpuCores_Error_UpperBound)`,
			input:  129,
			output: errors.New(QemuCpuCores_Error_UpperBound)},
		// Valid
		{name: `Valid LowerBound`,
			input: 1},
		{name: `Valid UpperBound`,
			input: 128},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.input.Validate(), test.output, test.name)
		})
	}
}

func Test_QemuCpuSockets_Validate(t *testing.T) {
	testData := []struct {
		name   string
		input  QemuCpuSockets
		output error
	}{
		// Invalid
		{name: "Invalid errors.New(CpuSockets_Error_LowerBound)",
			input:  0,
			output: errors.New(QemuCpuSockets_Error_LowerBound)},
		{name: "Invalid errors.New(CpuSockets_Error_UpperBound)",
			input:  5,
			output: errors.New(QemuCpuSockets_Error_UpperBound)},
		// Valid
		{name: "Valid LowerBound",
			input: 1},
		{name: "Valid UpperBound",
			input: 4},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.input.Validate(), test.output, test.name)
		})
	}
}
