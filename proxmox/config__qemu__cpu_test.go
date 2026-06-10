package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func testData_ConfigQemu_CPU_Get() []qemuTestCaseGet {
	baseCpu := func(cpu QemuCPU) *QemuCPU {
		if cpu.Cores == nil {
			cpu.Cores = new(QemuCpuCores(1))
		}
		if cpu.Numa == nil {
			cpu.Numa = new(false)
		}
		if cpu.Sockets == nil {
			cpu.Sockets = new(QemuCpuSockets(1))
		}
		return &cpu
	}
	return []qemuTestCaseGet{
		{name: `all`,
			input: map[string]any{
				"cores":    float64(10),
				"cpulimit": float64(35),
				"cpuunits": float64(1234),
				"numa":     float64(0),
				"sockets":  float64(4),
				"vcpus":    float64(40),
				"cpu":      string("host,flags=-aes;+amd-no-ssb;-amd-ssbd;+hv-evmcs;-hv-tlbflush;+ibpb;+md-clear;-pcid;-pdpe1gb;-ssbd;+spec-ctrl;+virt-ssbd")},
			output: testQemuBaseConfig_get(ConfigQemu{
				CPU: baseCpu(QemuCPU{
					Cores: new(QemuCpuCores(10)),
					Flags: &CpuFlags{
						AES:        new(TriBoolFalse),
						AmdNoSSB:   new(TriBoolTrue),
						AmdSSBD:    new(TriBoolFalse),
						HvEvmcs:    new(TriBoolTrue),
						HvTlbFlush: new(TriBoolFalse),
						Ibpb:       new(TriBoolTrue),
						MdClear:    new(TriBoolTrue),
						PCID:       new(TriBoolFalse),
						Pdpe1GB:    new(TriBoolFalse),
						SSBD:       new(TriBoolFalse),
						SpecCtrl:   new(TriBoolTrue),
						VirtSSBD:   new(TriBoolTrue)},
					Limit:        new(CpuLimit(35)),
					Numa:         new(false),
					Sockets:      new(QemuCpuSockets(4)),
					Type:         new(CpuType_Host),
					Units:        new(QemuCpuUnits(1234)),
					VirtualCores: new(CpuVirtualCores(40))})})},
		{name: `affinity consecutive`,
			input:  map[string]any{"affinity": "2-4"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{Affinity: new([]uint{2, 3, 4})})})},
		{name: `affinity empty`,
			input:  map[string]any{"affinity": ""},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{Affinity: new([]uint{})})})},
		{name: `affinity mixed`,
			input:  map[string]any{"affinity": "2,4-6,8,10,12-15"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{Affinity: new([]uint{2, 4, 5, 6, 8, 10, 12, 13, 14, 15})})})},
		{name: `affinity singular`,
			input:  map[string]any{"affinity": "2"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{Affinity: new([]uint{2})})})},
		{name: `cores`,
			input:  map[string]any{"cores": float64(1)},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{Cores: new(QemuCpuCores(1))})})},
		{name: `cpu flag aes`,
			input: map[string]any{"cpu": ",flags=+aes"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{
				Flags: &CpuFlags{AES: new(TriBoolTrue)},
				Type:  new(CpuType(""))})})},
		{name: `cpu flag amd-no-ssb`,
			input: map[string]any{"cpu": ",flags=-amd-no-ssb"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{
				Flags: &CpuFlags{AmdNoSSB: new(TriBoolFalse)},
				Type:  new(CpuType(""))})})},
		{name: `cpu flag amd-ssbd`,
			input: map[string]any{"cpu": ",flags=+amd-ssbd"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{
				Flags: &CpuFlags{AmdSSBD: new(TriBoolTrue)},
				Type:  new(CpuType(""))})})},
		{name: `cpu flag hv-evmcs`,
			input: map[string]any{"cpu": ",flags=-hv-evmcs"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{
				Flags: &CpuFlags{HvEvmcs: new(TriBoolFalse)},
				Type:  new(CpuType(""))})})},
		{name: `cpu flag hv-tlbflush`,
			input: map[string]any{"cpu": ",flags=+hv-tlbflush"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{
				Flags: &CpuFlags{HvTlbFlush: new(TriBoolTrue)},
				Type:  new(CpuType(""))})})},
		{name: `cpu flag ibpb`,
			input: map[string]any{"cpu": ",flags=-ibpb"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{
				Flags: &CpuFlags{Ibpb: new(TriBoolFalse)},
				Type:  new(CpuType(""))})})},
		{name: `cpu flag md-clear`,
			input: map[string]any{"cpu": ",flags=+md-clear"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{
				Flags: &CpuFlags{MdClear: new(TriBoolTrue)},
				Type:  new(CpuType(""))})})},
		{name: `cpu flag pcid`,
			input: map[string]any{"cpu": ",flags=-pcid"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{
				Flags: &CpuFlags{PCID: new(TriBoolFalse)},
				Type:  new(CpuType(""))})})},
		{name: `cpu flag pdpe1gb`,
			input: map[string]any{"cpu": ",flags=+pdpe1gb"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{
				Flags: &CpuFlags{Pdpe1GB: new(TriBoolTrue)},
				Type:  new(CpuType(""))})})},
		{name: `cpu flag ssbd`,
			input: map[string]any{"cpu": ",flags=-ssbd"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{
				Flags: &CpuFlags{SSBD: new(TriBoolFalse)},
				Type:  new(CpuType(""))})})},
		{name: `cpu flag spec-ctrl`,
			input: map[string]any{"cpu": ",flags=+spec-ctrl"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{
				Flags: &CpuFlags{SpecCtrl: new(TriBoolTrue)},
				Type:  new(CpuType(""))})})},
		{name: `cpu flag virt-ssbd`,
			input: map[string]any{"cpu": ",flags=-virt-ssbd"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{
				Flags: &CpuFlags{VirtSSBD: new(TriBoolFalse)},
				Type:  new(CpuType(""))})})},
		{name: `cpu flags multiple`,
			input: map[string]any{"cpu": ",flags=-aes;+amd-no-ssb;-amd-ssbd;-hv-evmcs;-hv-tlbflush;+ibpb;+md-clear;+pcid;-virt-ssbd"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{
				Flags: &CpuFlags{
					AES:        new(TriBoolFalse),
					AmdNoSSB:   new(TriBoolTrue),
					AmdSSBD:    new(TriBoolFalse),
					HvEvmcs:    new(TriBoolFalse),
					HvTlbFlush: new(TriBoolFalse),
					Ibpb:       new(TriBoolTrue),
					MdClear:    new(TriBoolTrue),
					PCID:       new(TriBoolTrue),
					VirtSSBD:   new(TriBoolFalse)},
				Type: new(CpuType(""))})})},
		{name: `cpu model only, no flags`,
			input:  map[string]any{"cpu": string(CpuType_X86_64_v2_AES)},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{Type: new(CpuType("x86-64-v2-AES"))})})},
		{name: `cpu with flags`,
			input: map[string]any{"cpu": "x86-64-v2-AES,flags=+spec-ctrl;-md-clear"},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{
				Flags: &CpuFlags{
					MdClear:  new(TriBoolFalse),
					SpecCtrl: new(TriBoolTrue)},
				Type: new(CpuType_X86_64_v2_AES)})})},
		{name: `cpulimit float64`,
			input:  map[string]any{"cpulimit": float64(10)},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{Limit: new(CpuLimit(10))})})},
		{name: `cpulimit string`,
			input:  map[string]any{"cpulimit": string("25")},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{Limit: new(CpuLimit(25))})})},
		{name: `cpuunits`,
			input:  map[string]any{"cpuunits": float64(1000)},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{Units: new(QemuCpuUnits(1000))})})},
		{name: `numa true`,
			input:  map[string]any{"numa": float64(1)},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{Numa: new(true)})})},
		{name: `numa false`,
			input:  map[string]any{"numa": float64(0)},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{Numa: new(false)})})},
		{name: `sockets`,
			input:  map[string]any{"sockets": float64(1)},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{Sockets: new(QemuCpuSockets(1))})})},
		{name: `vcpus`,
			input:  map[string]any{"vcpus": float64(1)},
			output: testQemuBaseConfig_get(ConfigQemu{CPU: baseCpu(QemuCPU{VirtualCores: new(CpuVirtualCores(1))})})}}
}

func testData_ConfigQemu_CPU_Validate_1() qemuTestTypeValidateFunc {
	return qemuTestTypeValidateFunc(func() (qemuTestTypeInvalid, qemuTestTypeValid) {
		invalid := qemuTestTypeInvalid{
			create: []qemuTestCaseInvalid{
				{name: `errors.New(QemuCPU_Error_CoresRequired)`,
					input: ConfigQemu{
						CPU:    &QemuCPU{},
						ID:     new(GuestID(111)),
						Memory: &QemuMemory{CapacityMiB: new(QemuMemoryCapacity(16))},
						Node:   new(NodeName("test"))},
					err: errors.New(QemuCPU_Error_CoresRequired)}},
			createUpdate: []qemuTestCaseInvalid{
				{name: `errors.New(CpuLimit_Error_Maximum)`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Limit: new(CpuLimit(129))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(CpuLimit_Error_Maximum)},
				{name: `errors.New(CpuUnits_Error_Maximum)`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Units: new(QemuCpuUnits(262145))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(CpuUnits_Error_Maximum)},
				{name: `errors.New(QemuCpuCores_Error_LowerBound)`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Cores: new(QemuCpuCores(0))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(QemuCpuCores_Error_LowerBound)},
				{name: `errors.New(QemuCpuSockets_Error_LowerBound)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{
						Cores:   new(QemuCpuCores(1)),
						Sockets: new(QemuCpuSockets(0))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(QemuCpuSockets_Error_LowerBound)},
				{name: `CpuVirtualCores(1).Error() 1 1`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{
						Cores:        new(QemuCpuCores(1)),
						Sockets:      new(QemuCpuSockets(1)),
						VirtualCores: new(CpuVirtualCores(2))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     CpuVirtualCores(1).Error()},
				{name: `Invalid AES`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AES: new(TriBool(-2))}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(TriBool_Error_Invalid)},
				{name: `Invalid AmdNoSSB`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AmdNoSSB: new(TriBool(2))}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(TriBool_Error_Invalid)},
				{name: `Invalid AmdSSBD`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AmdSSBD: new(TriBool(-27))}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(TriBool_Error_Invalid)},
				{name: `Invalid HvEvmcs`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						HvEvmcs: new(TriBool(32))}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(TriBool_Error_Invalid)},
				{name: `Invalid HvTlbFlush`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						HvTlbFlush: new(TriBool(-2))}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(TriBool_Error_Invalid)},
				{name: `Invalid Ibpb`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						Ibpb: new(TriBool(52))}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(TriBool_Error_Invalid)},
				{name: `Invalid MdClear`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						MdClear: new(TriBool(-52))}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(TriBool_Error_Invalid)},
				{name: `Invalid PCID`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						PCID: new(TriBool(82))}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(TriBool_Error_Invalid)},
				{name: `Invalid Pdpe1GB`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						Pdpe1GB: new(TriBool(-2))}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(TriBool_Error_Invalid)},
				{name: `Invalid SSBD`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						SSBD: new(TriBool(3))}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(TriBool_Error_Invalid)},
				{name: `Invalid SpecCtrl`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						SpecCtrl: new(TriBool(-2))}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(TriBool_Error_Invalid)},
				{name: `Invalid VirtSSBD`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						VirtSSBD: new(TriBool(2))}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					err:     errors.New(TriBool_Error_Invalid)},
				{name: `Type`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Type: new(CpuType("invalid"))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					version: Version{}.max(),
					err:     CpuType("").Error(Version{}.max())}}}
		valid := qemuTestTypeValid{
			createUpdate: []qemuTestCaseValid{
				{name: `Cores`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Cores: new(QemuCpuCores(1))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Maximum`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{
						Cores: new(QemuCpuCores(128)),
						Flags: new(CpuFlags{
							AES:        new(TriBoolTrue),
							AmdNoSSB:   new(TriBoolFalse),
							AmdSSBD:    new(TriBoolNone),
							HvEvmcs:    new(TriBoolTrue),
							HvTlbFlush: new(TriBoolFalse),
							Ibpb:       new(TriBoolNone),
							MdClear:    new(TriBoolTrue),
							PCID:       new(TriBoolFalse),
							Pdpe1GB:    new(TriBoolNone),
							SSBD:       new(TriBoolTrue),
							SpecCtrl:   new(TriBoolFalse),
							VirtSSBD:   new(TriBoolNone)}),
						Limit:        new(CpuLimit(128)),
						Sockets:      new(QemuCpuSockets(4)),
						Type:         new(CpuType(cpuType_AmdEPYCRomeV2_Lower)),
						Units:        new(QemuCpuUnits(262144)),
						VirtualCores: new(CpuVirtualCores(512))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					version: Version{}.max()},
				{name: `Minimum`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{
						Cores:        new(QemuCpuCores(128)),
						Flags:        new(CpuFlags{}),
						Limit:        new(CpuLimit(0)),
						Sockets:      new(QemuCpuSockets(4)),
						Type:         new(CpuType("")),
						Units:        new(QemuCpuUnits(0)),
						VirtualCores: new(CpuVirtualCores(0))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}},
					version: Version{}.max()},
				{name: `Flags all set`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AES:        new(TriBoolFalse),
						AmdNoSSB:   new(TriBoolNone),
						AmdSSBD:    new(TriBoolTrue),
						HvEvmcs:    new(TriBoolFalse),
						HvTlbFlush: new(TriBoolNone),
						Ibpb:       new(TriBoolTrue),
						MdClear:    new(TriBoolFalse),
						PCID:       new(TriBoolNone),
						Pdpe1GB:    new(TriBoolTrue),
						SSBD:       new(TriBoolFalse),
						SpecCtrl:   new(TriBoolNone),
						VirtSSBD:   new(TriBoolTrue)}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Flags all nil`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Flags mixed`,
					input: testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AmdNoSSB:   new(TriBoolTrue),
						AmdSSBD:    new(TriBoolFalse),
						HvTlbFlush: new(TriBoolTrue),
						Ibpb:       new(TriBoolFalse),
						MdClear:    new(TriBoolNone),
						PCID:       new(TriBoolTrue),
						Pdpe1GB:    new(TriBoolFalse),
						SpecCtrl:   new(TriBoolTrue)}}}),
					current: &ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Limit maximum`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Limit: new(CpuLimit(128))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Limit minimum`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Limit: new(CpuLimit(0))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Sockets`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Sockets: new(QemuCpuSockets(1))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Type empty`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Type: new(CpuType(""))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Type host`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Type: new(CpuType_Host)}}),
					current: &ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Units Minimum`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Units: new(QemuCpuUnits(0))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Units Maximum`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{Units: new(QemuCpuUnits(262144))}}),
					current: &ConfigQemu{CPU: &QemuCPU{}}}},
			update: []qemuTestCaseValid{
				{name: `nothing`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{CPU: &QemuCPU{}}),
					current: &ConfigQemu{CPU: &QemuCPU{}}}}}
		return invalid, valid
	})
}

func testData_ConfigQemu_CPU_Validate_2() qemuTestTypeValidateFunc {
	return qemuTestTypeValidateFunc(func() (qemuTestTypeInvalid, qemuTestTypeValid) {
		invalid := qemuTestTypeInvalid{
			create: []qemuTestCaseInvalid{
				{name: `erross.New(ConfigQemu_Error_CpuRequired)`,
					input: ConfigQemu{
						ID:     new(GuestID(111)),
						Memory: &QemuMemory{CapacityMiB: new(QemuMemoryCapacity(16))},
						Node:   new(NodeName("test"))},
					err: errors.New(ConfigQemu_Error_CpuRequired)}}}
		return invalid, qemuTestTypeValid{}
	})
}

func Test_ConfigQemu_CPU_MapToApi(t *testing.T) {
	t.Parallel()
	tests := qemuTestsApiFunc(func() qemuTestsAPI {
		return qemuTestsAPI{
			create: []qemuTestCaseAPI{
				{name: `Flags AES`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{AES: new(TriBoolTrue)}}},
					body:   map[string]string{"cpu": "%2Cflags%3D%2Baes"}}, // ",flags=+aes"
				{name: `Flags AmdNoSSB`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{AmdNoSSB: new(TriBoolFalse)}}},
					body:   map[string]string{"cpu": "%2Cflags%3D-amd-no-ssb"}}, // ",flags=-amd-no-ssb"
				{name: `Flags AmdSSBD`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{AmdSSBD: new(TriBoolTrue)}}},
					body:   map[string]string{"cpu": "%2Cflags%3D%2Bamd-ssbd"}}, // ",flags=+amd-ssbd"
				{name: `Flags HvEvmcs`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{HvEvmcs: new(TriBoolFalse)}}},
					body:   map[string]string{"cpu": "%2Cflags%3D-hv-evmcs"}}, // ",flags=-hv-evmcs"
				{name: `Flags HvTlbFlush`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{HvTlbFlush: new(TriBoolTrue)}}},
					body:   map[string]string{"cpu": "%2Cflags%3D%2Bhv-tlbflush"}}, // ",flags=+hv-tlbflush"
				{name: `Flags Ibpb`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{Ibpb: new(TriBoolFalse)}}},
					body:   map[string]string{"cpu": "%2Cflags%3D-ibpb"}}, // ",flags=-ibpb"
				{name: `Flags MdClear`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{MdClear: new(TriBoolTrue)}}},
					body:   map[string]string{"cpu": "%2Cflags%3D%2Bmd-clear"}}, // ",flags=+md-clear"
				{name: `Flags PCID`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{PCID: new(TriBoolFalse)}}},
					body:   map[string]string{"cpu": "%2Cflags%3D-pcid"}}, // ",flags=-pcid"
				{name: `Flags Pdpe1GB`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{Pdpe1GB: new(TriBoolTrue)}}},
					body:   map[string]string{"cpu": "%2Cflags%3D%2Bpdpe1gb"}}, // ",flags=+pdpe1gb"
				{name: `Flags SSBD`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{SSBD: new(TriBoolFalse)}}},
					body:   map[string]string{"cpu": "%2Cflags%3D-ssbd"}}, // ",flags=-ssbd"
				{name: `Flags SpecCtrl`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{SpecCtrl: new(TriBoolTrue)}}},
					body:   map[string]string{"cpu": "%2Cflags%3D%2Bspec-ctrl"}}, // ",flags=+spec-ctrl"
				{name: `Flags VirtSSBD`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{VirtSSBD: new(TriBoolFalse)}}},
					body:   map[string]string{"cpu": "%2Cflags%3D-virt-ssbd"}}, // ",flags=-virt-ssbd"
				{name: `Flags mixed`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AES:        new(TriBoolTrue),
						AmdNoSSB:   new(TriBoolFalse),
						AmdSSBD:    new(TriBoolTrue),
						HvEvmcs:    new(TriBoolNone),
						HvTlbFlush: new(TriBoolTrue),
						MdClear:    new(TriBoolTrue),
						PCID:       new(TriBoolFalse),
						Pdpe1GB:    new(TriBoolNone)}}},
					body: map[string]string{"cpu": "%2Cflags%3D%2Baes%3B-amd-no-ssb%3B%2Bamd-ssbd%3B%2Bhv-tlbflush%3B%2Bmd-clear%3B-pcid"}}, // ",flags=+aes;-amd-no-ssb;+amd-ssbd;+hv-tlbflush;+md-clear;-pcid"
				{name: `Flags all nil`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{}}}},
				{name: `Flags all none`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AES:        new(TriBoolNone),
						AmdNoSSB:   new(TriBoolNone),
						AmdSSBD:    new(TriBoolNone),
						HvEvmcs:    new(TriBoolNone),
						HvTlbFlush: new(TriBoolNone),
						MdClear:    new(TriBoolNone),
						PCID:       new(TriBoolNone),
						Pdpe1GB:    new(TriBoolNone),
						SSBD:       new(TriBoolNone),
						SpecCtrl:   new(TriBoolNone),
						VirtSSBD:   new(TriBoolNone)}}}},
				{name: `Flags all none & Type ""`,
					config: &ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{
							AES:        new(TriBoolNone),
							AmdNoSSB:   new(TriBoolNone),
							AmdSSBD:    new(TriBoolNone),
							HvEvmcs:    new(TriBoolNone),
							HvTlbFlush: new(TriBoolNone),
							MdClear:    new(TriBoolNone),
							PCID:       new(TriBoolNone),
							Pdpe1GB:    new(TriBoolNone),
							SSBD:       new(TriBoolNone),
							SpecCtrl:   new(TriBoolNone),
							VirtSSBD:   new(TriBoolNone)},
						Type: new(CpuType(""))}}}},
			createUpdate: []qemuTestCaseAPI{
				{name: `Affinity empty no effect`,
					config:        &ConfigQemu{CPU: &QemuCPU{Affinity: new([]uint{})}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Affinity consecutive`,
					config:        &ConfigQemu{CPU: &QemuCPU{Affinity: new([]uint{0, 0, 1, 2, 2, 3})}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Affinity: new([]uint{0, 1, 2})}},
					body:          map[string]string{"affinity": "0-3"}},
				{name: `Affinity singular`,
					config:        &ConfigQemu{CPU: &QemuCPU{Affinity: new([]uint{2})}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Affinity: new([]uint{0, 1, 2})}},
					body:          map[string]string{"affinity": "2"}},
				{name: `Affinity mixed`,
					config:        &ConfigQemu{CPU: &QemuCPU{Affinity: new([]uint{5, 0, 4, 2, 9, 3, 2, 11, 7, 2, 12, 4, 13})}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Affinity: new([]uint{0, 1, 2})}},
					body:          map[string]string{"affinity": "0%2C2-5%2C7%2C9%2C11-13"}}, // "0,2-5,7,9,11-13"
				{name: `Cores`,
					config:        &ConfigQemu{CPU: &QemuCPU{Cores: new(QemuCpuCores(1))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Cores: new(QemuCpuCores(2))}},
					body:          map[string]string{"cores": "1"}},
				{name: `Limit`,
					config:        &ConfigQemu{CPU: &QemuCPU{Limit: new(CpuLimit(50))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Limit: new(CpuLimit(100))}},
					body:          map[string]string{"cpulimit": "50"}},
				{name: `Limit 0 no effect`,
					config:        &ConfigQemu{CPU: &QemuCPU{Limit: new(CpuLimit(0))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Numa`,
					config:        &ConfigQemu{CPU: &QemuCPU{Numa: new(true)}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Numa: new(false)}},
					body:          map[string]string{"numa": "1"}},
				{name: `Sockets`,
					config:        &ConfigQemu{CPU: &QemuCPU{Sockets: new(QemuCpuSockets(3))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Sockets: new(QemuCpuSockets(2))}},
					body:          map[string]string{"sockets": "3"}},
				{name: `Type clear no effect`,
					config:        &ConfigQemu{CPU: &QemuCPU{Type: new(CpuType(""))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Type custom`,
					config:        &ConfigQemu{CPU: &QemuCPU{Type: new(CpuType("custom-TeSt"))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Type: new(CpuType_Host)}},
					body:          map[string]string{"cpu": "custom-TeSt"}},
				{name: `Type lower`,
					config:        &ConfigQemu{CPU: &QemuCPU{Type: new(cpuType_X86_64_v2_AES_Lower)}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Type: new(CpuType_Host)}},
					version:       Version{}.max(),
					body:          map[string]string{"cpu": "x86-64-v2-AES"}},
				{name: `Type normal`,
					config:        &ConfigQemu{CPU: &QemuCPU{Type: new(CpuType_X86_64_v2_AES)}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Type: new(CpuType_Host)}},
					version:       Version{}.max(),
					body:          map[string]string{"cpu": "x86-64-v2-AES"}},
				{name: `Type weird`,
					config:        &ConfigQemu{CPU: &QemuCPU{Type: new(CpuType("X_-8-_6_-6-4---V_-2-aE--s__"))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Type: new(CpuType_Host)}},
					version:       Version{}.max(),
					body:          map[string]string{"cpu": "x86-64-v2-AES"}},
				{name: `Units`,
					config:        &ConfigQemu{CPU: &QemuCPU{Units: new(QemuCpuUnits(100))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Units: new(QemuCpuUnits(200))}},
					body:          map[string]string{"cpuunits": "100"}},
				{name: `Units 0 no effect`,
					config:        &ConfigQemu{CPU: &QemuCPU{Units: new(QemuCpuUnits(0))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{}}},
				{name: `VirtualCores`,
					config:        &ConfigQemu{CPU: &QemuCPU{VirtualCores: new(CpuVirtualCores(4))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{VirtualCores: new(CpuVirtualCores(12))}},
					body:          map[string]string{"vcpus": "4"}},
				{name: `VirtualCores 0 no effect`,
					config:        &ConfigQemu{CPU: &QemuCPU{VirtualCores: new(CpuVirtualCores(0))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{}}}},
			update: []qemuTestCaseAPI{
				{name: `Affinity create`,
					config:        &ConfigQemu{CPU: &QemuCPU{Affinity: new([]uint{5, 0, 4, 2, 9, 3, 2, 11, 7, 2, 12, 4, 13})}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{}},
					body:          map[string]string{"affinity": "0%2C2-5%2C7%2C9%2C11-13"}}, // "0,2-5,7,9,11-13"
				{name: `Affinity empty`,
					config:        &ConfigQemu{CPU: &QemuCPU{Affinity: new([]uint{})}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Affinity: new([]uint{0, 1, 2})}},
					body:          map[string]string{"delete": "affinity"}},
				{name: `Affinity empty no current`,
					config:        &ConfigQemu{CPU: &QemuCPU{Affinity: new([]uint{})}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Type create`,
					config:        &ConfigQemu{CPU: &QemuCPU{Type: new(CpuType_X86_64_v2_AES)}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{}},
					version:       Version{}.max(),
					body:          map[string]string{"cpu": "x86-64-v2-AES"}},
				{name: `Flags nil`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{}}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AES:  new(TriBoolTrue),
						PCID: new(TriBoolFalse)}}}},
				{name: `Flags`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AES:        new(TriBoolTrue),
						AmdNoSSB:   new(TriBoolNone),
						HvTlbFlush: new(TriBoolTrue),
						Ibpb:       new(TriBoolNone),
						MdClear:    new(TriBoolFalse),
						PCID:       new(TriBoolTrue),
						SpecCtrl:   new(TriBoolFalse),
						VirtSSBD:   new(TriBoolFalse)}}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AmdNoSSB:   new(TriBoolTrue),
						HvEvmcs:    new(TriBoolFalse),
						HvTlbFlush: new(TriBoolFalse),
						Ibpb:       new(TriBoolTrue),
						MdClear:    new(TriBoolTrue),
						SpecCtrl:   new(TriBoolFalse)}}},
					body: map[string]string{"cpu": "%2Cflags%3D%2Baes%3B-hv-evmcs%3B%2Bhv-tlbflush%3B-md-clear%3B%2Bpcid%3B-spec-ctrl%3B-virt-ssbd"}}, // ",flags=+aes;-hv-evmcs;+hv-tlbflush;-md-clear;+pcid;-spec-ctrl;-virt-ssbd"
				{name: `Flags all none`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AES:        new(TriBoolNone),
						AmdNoSSB:   new(TriBoolNone),
						AmdSSBD:    new(TriBoolNone),
						HvEvmcs:    new(TriBoolNone),
						HvTlbFlush: new(TriBoolNone),
						MdClear:    new(TriBoolNone),
						PCID:       new(TriBoolNone),
						Pdpe1GB:    new(TriBoolNone),
						SSBD:       new(TriBoolNone),
						SpecCtrl:   new(TriBoolNone),
						VirtSSBD:   new(TriBoolNone)}}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{
							AES:        new(TriBoolTrue),
							AmdNoSSB:   new(TriBoolTrue),
							AmdSSBD:    new(TriBoolTrue),
							HvEvmcs:    new(TriBoolTrue),
							HvTlbFlush: new(TriBoolTrue),
							MdClear:    new(TriBoolTrue),
							PCID:       new(TriBoolTrue),
							Pdpe1GB:    new(TriBoolTrue),
							SSBD:       new(TriBoolTrue),
							SpecCtrl:   new(TriBoolTrue),
							VirtSSBD:   new(TriBoolTrue)},
						Type: new(CpuType_Host)}},
					body: map[string]string{"cpu": "host"}},
				{name: `Flags create`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AES:        new(TriBoolTrue),
						AmdNoSSB:   new(TriBoolFalse),
						AmdSSBD:    new(TriBoolTrue),
						HvEvmcs:    new(TriBoolNone),
						HvTlbFlush: new(TriBoolTrue),
						MdClear:    new(TriBoolTrue),
						PCID:       new(TriBoolFalse),
						Pdpe1GB:    new(TriBoolNone)}}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{}},
					body:          map[string]string{"cpu": "%2Cflags%3D%2Baes%3B-amd-no-ssb%3B%2Bamd-ssbd%3B%2Bhv-tlbflush%3B%2Bmd-clear%3B-pcid"}}, // ",flags=+aes;-amd-no-ssb;+amd-ssbd;+hv-tlbflush;+md-clear;-pcid"
				{name: `Flags & Type, clear`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{AmdNoSSB: new(TriBoolNone)},
						Type: new(CpuType(""))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{AmdNoSSB: new(TriBoolFalse)},
						Type:  new(CpuType_Host)}},
					body: map[string]string{"delete": "cpu"}},
				{name: `Flags & Type, update Flags`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AmdNoSSB: new(TriBoolTrue)}}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{
							HvEvmcs:    new(TriBoolFalse),
							HvTlbFlush: new(TriBoolFalse),
							Ibpb:       new(TriBoolTrue),
							MdClear:    new(TriBoolTrue),
							SpecCtrl:   new(TriBoolFalse)},
						Type: new(CpuType_Host)}},
					body: map[string]string{"cpu": "host%2Cflags%3D%2Bamd-no-ssb%3B-hv-evmcs%3B-hv-tlbflush%3B%2Bibpb%3B%2Bmd-clear%3B-spec-ctrl"}}, // "host,flags=+amd-no-ssb;-hv-evmcs;-hv-tlbflush;+ibpb;+md-clear;-spec-ctrl"
				{name: `Flags & Type, update Type`,
					config: &ConfigQemu{CPU: &QemuCPU{Type: new(CpuType_X86_64_v2_AES)}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{
							HvEvmcs:    new(TriBoolFalse),
							HvTlbFlush: new(TriBoolFalse),
							Ibpb:       new(TriBoolTrue),
							MdClear:    new(TriBoolTrue),
							SpecCtrl:   new(TriBoolFalse)},
						Type: new(CpuType_Host)}},
					version: Version{}.max(),
					body:    map[string]string{"cpu": "x86-64-v2-AES%2Cflags%3D-hv-evmcs%3B-hv-tlbflush%3B%2Bibpb%3B%2Bmd-clear%3B-spec-ctrl"}}, // "x86-64-v2-AES,flags=-hv-evmcs;-hv-tlbflush;+ibpb;+md-clear;-spec-ctrl"
				{name: `Limit 0`,
					config:        &ConfigQemu{CPU: &QemuCPU{Limit: new(CpuLimit(0))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Limit: new(CpuLimit(100))}},
					body:          map[string]string{"delete": "cpulimit"}},
				{name: `Limit 0 no current`,
					config:        &ConfigQemu{CPU: &QemuCPU{Limit: new(CpuLimit(0))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{}}},
				{name: `Limit create`,
					config:        &ConfigQemu{CPU: &QemuCPU{Limit: new(CpuLimit(67))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{}},
					body:          map[string]string{"cpulimit": "67"}},
				{name: `Units 0`,
					config:        &ConfigQemu{CPU: &QemuCPU{Units: new(QemuCpuUnits(0))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{Units: new(QemuCpuUnits(100))}},
					body:          map[string]string{"delete": "cpuunits"}},
				{name: `Units create`,
					config:        &ConfigQemu{CPU: &QemuCPU{Units: new(QemuCpuUnits(169))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{}},
					body:          map[string]string{"cpuunits": "169"}},
				{name: `VirtualCores 0`,
					config:        &ConfigQemu{CPU: &QemuCPU{VirtualCores: new(CpuVirtualCores(0))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{VirtualCores: new(CpuVirtualCores(4))}},
					body:          map[string]string{"delete": "vcpus"}},
				{name: `VirtualCores create`,
					config:        &ConfigQemu{CPU: &QemuCPU{VirtualCores: new(CpuVirtualCores(7))}},
					currentLegacy: ConfigQemu{CPU: &QemuCPU{}},
					body:          map[string]string{"vcpus": "7"}}}}
	})
	tests.Test(t)
}

func Test_ConfigQemu_CPU_Validate(t *testing.T) {
	t.Parallel()
	testData_ConfigQemu_CPU_Validate_1().Test(t)
	testData_ConfigQemu_CPU_Validate_2().Test(t)
}

func Test_CpuFlags_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  CpuFlags
		output error
	}{
		{name: `Valid`,
			input: CpuFlags{
				AES:        new(TriBoolTrue),
				AmdNoSSB:   new(TriBoolFalse),
				AmdSSBD:    new(TriBoolNone),
				HvEvmcs:    new(TriBoolTrue),
				HvTlbFlush: new(TriBoolFalse),
				Ibpb:       new(TriBoolNone),
				MdClear:    new(TriBoolTrue),
				PCID:       new(TriBoolFalse),
				Pdpe1GB:    new(TriBoolNone),
				SSBD:       new(TriBoolTrue),
				SpecCtrl:   new(TriBoolFalse),
				VirtSSBD:   new(TriBoolNone)}},
		{name: `Invalid AES`,
			input: CpuFlags{
				AES: new(TriBool(2))},
			output: errors.New(TriBool_Error_Invalid)},
		{name: `Invalid AmdNoSSB`,
			input: CpuFlags{
				AmdNoSSB: new(TriBool(-2))},
			output: errors.New(TriBool_Error_Invalid)},
		{name: `Invalid AmdSSBD`,
			input: CpuFlags{
				AmdSSBD: new(TriBool(27))},
			output: errors.New(TriBool_Error_Invalid)},
		{name: `Invalid HvEvmcs`,
			input: CpuFlags{
				HvEvmcs: new(TriBool(-32))},
			output: errors.New(TriBool_Error_Invalid)},
		{name: `Invalid HvTlbFlush`,
			input: CpuFlags{
				HvTlbFlush: new(TriBool(2))},
			output: errors.New(TriBool_Error_Invalid)},
		{name: `Invalid Ibpb`,
			input: CpuFlags{
				Ibpb: new(TriBool(-52))},
			output: errors.New(TriBool_Error_Invalid)},
		{name: `Invalid MdClear`,
			input: CpuFlags{
				MdClear: new(TriBool(52))},
			output: errors.New(TriBool_Error_Invalid)},
		{name: `Invalid PCID`,
			input: CpuFlags{
				PCID: new(TriBool(-82))},
			output: errors.New(TriBool_Error_Invalid)},
		{name: `Invalid Pdpe1GB`,
			input: CpuFlags{
				Pdpe1GB: new(TriBool(2))},
			output: errors.New(TriBool_Error_Invalid)},
		{name: `Invalid SSBD`,
			input: CpuFlags{
				SSBD: new(TriBool(-3))},
			output: errors.New(TriBool_Error_Invalid)},
		{name: `Invalid SpecCtrl`,
			input: CpuFlags{
				SpecCtrl: new(TriBool(2))},
			output: errors.New(TriBool_Error_Invalid)},
		{name: `Invalid VirtSSBD`,
			input: CpuFlags{
				VirtSSBD: new(TriBool(-2))},
			output: errors.New(TriBool_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_CpuLimit_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  CpuLimit
		output error
	}{
		{name: "Valid minimum",
			input: 0},
		{name: "Valid maximum",
			input: 128},
		{name: "Invalid maximum",
			input:  129,
			output: errors.New(CpuLimit_Error_Maximum)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_CpuType_Error(t *testing.T) {
	t.Parallel()
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

func test_CpuTypeValidate_data() []struct {
	name    string
	config  CpuType
	version Version
	output  error
} {
	return []struct {
		name    string
		config  CpuType
		version Version
		output  error
	}{
		// Invalid
		{name: `Invalid`,
			config:  CpuType("gibbers"),
			version: Version{}.max(),
			output:  CpuType("").Error(Version{}.max())},
		{name: `Invalid V7`,
			config:  CpuType_AmdEPYCRomeV2,
			version: Version{Major: 7}.max(),
			output:  CpuType("").Error(Version{Major: 7}.max())},
		{name: `Invalid V7 EPYC-Genoa`,
			config:  CpuType_AmdEPYCGenoa,
			version: Version{Major: 7}.max(),
			output:  CpuType("").Error(Version{Major: 7}.max())},
		{name: `Invalid V8 EPYC-Turin`,
			config:  CpuType_AmdEPYCTurin,
			version: Version{Major: 8}.max(),
			output:  CpuType("").Error(Version{Major: 8}.max())},
		// Valid
		{name: `Valid custom`,
			config: CpuType("custom-TeSt")},
		{name: `Valid empty`,
			config:  CpuType(""),
			version: Version{}.max()},
		{name: `Valid normal`,
			config:  CpuType("Skylake-Server-noTSX-IBRS"),
			version: Version{}.max()},
		{name: `Valid lowercase`,
			config:  CpuType("skylakeclientnotsxibrs"),
			version: Version{}.max()},
		{name: `Valid weird`,
			config:  CpuType("S-k__-Yl_-A--k-e__-Se-R-v-__Er--n-OTs_X---I-_br-S"),
			version: Version{}.max()},
		{name: `Valid EPYC-Genoa`,
			config:  CpuType_AmdEPYCGenoa,
			version: Version{Major: 8}.max()},
		{name: `Valid EPYC-Genoa-v2`,
			config:  CpuType_AmdEPYCGenoaV2,
			version: Version{Major: 8}.max()},
		{name: `Valid EPYC-Turin`,
			config:  CpuType_AmdEPYCTurin,
			version: Version{Major: 9}.max()},
		{name: `Valid cortex-a57`,
			config:  CpuType_ArmCortexA57,
			version: Version{}.max()},
		{name: `Valid cortex-a57 hyphen-stripped lookup`,
			config:  CpuType("cortexa57"),
			version: Version{}.max()},
		{name: `Valid cortex-a72`,
			config:  CpuType_ArmCortexA72,
			version: Version{}.max()},
		{name: `Valid cortex-a72 hyphen-stripped lookup`,
			config:  CpuType("cortexa72"),
			version: Version{}.max()},
	}
}

func Test_CpuType_Validate(t *testing.T) {
	t.Parallel()
	for _, test := range test_CpuTypeValidate_data() {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.config.Validate(test.version), test.output, test.name)
		})
	}
}

func Benchmark_CpuType_Validate(b *testing.B) {
	// prevent compiler optimizations
	var result error
	tests := test_CpuTypeValidate_data()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			result = test.config.Validate(test.version)
		}
	}
	_ = result
}

func Test_CpuUnits_Validate(t *testing.T) {
	t.Parallel()
	testData := []struct {
		name   string
		input  QemuCpuUnits
		output error
	}{
		{name: `Invalid errors.New(CpuUnits_Error_Maximum)`,
			input:  262145,
			output: errors.New(CpuUnits_Error_Maximum)},
		{name: `Valid minimum`,
			input: 0},
		{name: `Valid maximum`,
			input: 262144},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.input.Validate(), test.output, test.name)
		})
	}
}

func Test_CpuVirtualCores_Validate(t *testing.T) {
	t.Parallel()
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
				cores:        new(QemuCpuCores(2)),
				sockets:      new(QemuCpuSockets(2))},
			output: CpuVirtualCores(4).Error()},
		{name: `Invalid Update Cores`,
			input: testInput{
				virtualCores: 8,
				cores:        new(QemuCpuCores(1)),
				current: &QemuCPU{
					Cores:   new(QemuCpuCores(3)),
					Sockets: new(QemuCpuSockets(2))}},
			output: CpuVirtualCores(2).Error()},
		{name: `Invalid Update Sockets`,
			input: testInput{
				virtualCores: 10,
				sockets:      new(QemuCpuSockets(2)),
				current: &QemuCPU{
					Cores:   new(QemuCpuCores(4)),
					Sockets: new(QemuCpuSockets(3))}},
			output: CpuVirtualCores(8).Error()},
		{name: `Invalid Update`,
			input: testInput{
				virtualCores: 16,
				current: &QemuCPU{
					Cores:   new(QemuCpuCores(4)),
					Sockets: new(QemuCpuSockets(3))}},
			output: CpuVirtualCores(12).Error()},
		// Valid
		{name: `Valid Create`,
			input: testInput{
				virtualCores: 1,
				cores:        new(QemuCpuCores(1)),
				sockets:      new(QemuCpuSockets(1))}},
		{name: `Valid Update Cores`,
			input: testInput{
				virtualCores: 2,
				cores:        new(QemuCpuCores(2)),
				current: &QemuCPU{
					Cores:   new(QemuCpuCores(1)),
					Sockets: new(QemuCpuSockets(1))}}},
		{name: `Valid Update Sockets`,
			input: testInput{
				virtualCores: 3,
				sockets:      new(QemuCpuSockets(3)),
				current: &QemuCPU{
					Cores:   new(QemuCpuCores(1)),
					Sockets: new(QemuCpuSockets(4))}}},
		{name: `Valid Update`,
			input: testInput{
				virtualCores: 4,
				current: &QemuCPU{
					Cores:   new(QemuCpuCores(2)),
					Sockets: new(QemuCpuSockets(2))}}},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.input.virtualCores.Validate(test.input.cores, test.input.sockets, test.input.current), test.output, test.name)
		})
	}
}

func Test_QemuCPU_Validate(t *testing.T) {
	t.Parallel()
	validate := func(t *testing.T, config ConfigQemu, current *ConfigQemu, version Version, expectedErr error, valid bool) {
		t.Helper()
		var currentCPU *QemuCPU
		if current != nil {
			currentCPU = current.CPU
		}
		err := config.CPU.Validate(currentCPU, version)
		if valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			if expectedErr != nil {
				require.Equal(t, expectedErr, err)
			}
		}
	}
	testData_ConfigQemu_CPU_Validate_1().Inject(t, validate)
}

func Test_QemuCpuCores_Validate(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
