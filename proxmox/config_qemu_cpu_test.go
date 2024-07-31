package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

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
	testData := []struct {
		name    string
		input   QemuCPU
		current *QemuCPU
		output  error
	}{
		// Invalid
		{name: `Invalid errors.New(QemuCpuCores_Error_LowerBound)`,
			input:  QemuCPU{Cores: util.Pointer(QemuCpuCores(0))},
			output: errors.New(QemuCpuCores_Error_LowerBound)},
		{name: `Invalid errors.New(QemuCPU_Error_CoresRequired)`,
			input:  QemuCPU{},
			output: errors.New(QemuCPU_Error_CoresRequired)},
		{name: `Invalid errors.New(QemuCpuSockets_Error_LowerBound)`,
			input:  baseConfig(QemuCPU{Sockets: util.Pointer(QemuCpuSockets(0))}),
			output: errors.New(QemuCpuSockets_Error_LowerBound)},
		{name: `Invalid CpuVirtualCores(1).Error() 1 1`,
			input: QemuCPU{
				Cores:        util.Pointer(QemuCpuCores(1)),
				Sockets:      util.Pointer(QemuCpuSockets(1)),
				VirtualCores: util.Pointer(CpuVirtualCores(2))},
			output: CpuVirtualCores(1).Error()},
		// Valid
		{name: `Valid Maximum`,
			input: QemuCPU{
				Cores:        util.Pointer(QemuCpuCores(128)),
				Sockets:      util.Pointer(QemuCpuSockets(4)),
				VirtualCores: util.Pointer(CpuVirtualCores(512))}},
		{name: `Valid Minimum`,
			input: QemuCPU{
				Cores:        util.Pointer(QemuCpuCores(128)),
				Sockets:      util.Pointer(QemuCpuSockets(4)),
				VirtualCores: util.Pointer(CpuVirtualCores(0))}},
		{name: `Valid Update`,
			input:   QemuCPU{},
			current: &QemuCPU{}},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.input.Validate(test.current), test.output, test.name)
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
