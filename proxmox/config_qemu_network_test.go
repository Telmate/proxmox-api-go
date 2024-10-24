package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_QemuMTU_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  QemuMTU
		output error
	}{
		{name: `Valid inherit`,
			input: QemuMTU{Inherit: true}},
		{name: `Valid value`,
			input: QemuMTU{Value: 1500}},
		{name: `Valid empty`},
		{name: `Invalid mutually exclusive`,
			input:  QemuMTU{Inherit: true, Value: 1500},
			output: errors.New(QemuMTU_Error_Invalid)},
		{name: `Invalid too small`,
			input:  QemuMTU{Value: 575},
			output: errors.New(MTU_Error_Invalid)},
		{name: `Invalid too large`,
			input:  QemuMTU{Value: 65521},
			output: errors.New(MTU_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuNetworkInterface_Validate(t *testing.T) {
	type testInput struct {
		config  QemuNetworkInterface
		current *QemuNetworkInterface
	}
	tests := []struct {
		name   string
		input  testInput
		output error
	}{
		{name: `Valid Delete`,
			input: testInput{
				config: QemuNetworkInterface{Delete: true}}},
		{name: `Valid MTU inherit`,
			input: testInput{
				config: QemuNetworkInterface{
					Model: util.Pointer(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Inherit: true}},
				current: &QemuNetworkInterface{}}},
		{name: `Valid MTU value`,
			input: testInput{
				config: QemuNetworkInterface{
					Model: util.Pointer(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Value: 1500}},
				current: &QemuNetworkInterface{}}},
		{name: `Valid MTU empty`,
			input: testInput{
				config:  QemuNetworkInterface{MTU: &QemuMTU{}},
				current: &QemuNetworkInterface{}}},
		{name: `Valid Model`,
			input: testInput{
				config:  QemuNetworkInterface{Model: util.Pointer(QemuNetworkModel("virtio"))},
				current: &QemuNetworkInterface{}}},
		{name: `Valid MultiQueue`,
			input: testInput{
				config:  QemuNetworkInterface{MultiQueue: util.Pointer(QemuNetworkQueue(64))},
				current: &QemuNetworkInterface{}}},
		{name: `Valid RateLimitKBps`,
			input: testInput{
				config:  QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(10240000))},
				current: &QemuNetworkInterface{}}},
		{name: `Valid NativeVlan`,
			input: testInput{
				config:  QemuNetworkInterface{NativeVlan: util.Pointer(Vlan(5))},
				current: &QemuNetworkInterface{}}},
		{name: `Valid TaggedVlans`,
			input: testInput{
				config:  QemuNetworkInterface{TaggedVlans: util.Pointer(Vlans{0, 45, 12, 4095, 12, 45})},
				current: &QemuNetworkInterface{}},
		},
		// Invalid
		{name: `Invalid errors.New(QemuNetworkInterface_Error_BridgeRequired)`,
			input:  testInput{config: QemuNetworkInterface{}},
			output: errors.New(QemuNetworkInterface_Error_BridgeRequired)},
		{name: `Invalid errors.New(QemuNetworkInterface_Error_ModelRequired)`,
			input:  testInput{config: QemuNetworkInterface{Bridge: util.Pointer("vmbr0")}},
			output: errors.New(QemuNetworkInterface_Error_ModelRequired)},
		{name: `Invalid errors.New(QemuMTU_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterface{
					Model: util.Pointer(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Inherit: true, Value: 1500}},
				current: &QemuNetworkInterface{}},
			output: errors.New(QemuMTU_Error_Invalid)},
		{name: `Invalid errors.New(MTU_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterface{
					Model: util.Pointer(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Value: 575}},
				current: &QemuNetworkInterface{}},
			output: errors.New(MTU_Error_Invalid)},

		{name: `Invalid Model`,
			input: testInput{
				config:  QemuNetworkInterface{Model: util.Pointer(QemuNetworkModel("invalid"))},
				current: &QemuNetworkInterface{}},
			output: QemuNetworkModel("").Error()},
		{name: `Invalid errors.New(QemuNetworkQueue_Error_Invalid)`,
			input: testInput{
				config:  QemuNetworkInterface{MultiQueue: util.Pointer(QemuNetworkQueue(65))},
				current: &QemuNetworkInterface{}},
			output: errors.New(QemuNetworkQueue_Error_Invalid)},
		{name: `Invalid errors.New(QemuNetworkRate_Error_Invalid)`,
			input: testInput{
				config:  QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(10240001))},
				current: &QemuNetworkInterface{}},
			output: errors.New(QemuNetworkRate_Error_Invalid)},
		{name: `Invalid NativeVlan errors.New(Vlan_Error_Invalid)`,
			input: testInput{
				config:  QemuNetworkInterface{NativeVlan: util.Pointer(Vlan(4096))},
				current: &QemuNetworkInterface{}},
			output: errors.New(Vlan_Error_Invalid)},
		{name: `Invalid TaggedVlans errors.New(Vlan_Error_Invalid)`,
			input: testInput{
				config:  QemuNetworkInterface{TaggedVlans: util.Pointer(Vlans{0, 45, 12, 4095, 12, 45, 4096})},
				current: &QemuNetworkInterface{}},
			output: errors.New(Vlan_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.config.Validate(test.input.current))
		})
	}
}

func Test_QemuNetworkInterfaceID_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  QemuNetworkInterfaceID
		output error
	}{
		{name: "Valid",
			input: QemuNetworkInterfaceID0},
		{name: "Invalid",
			input:  32,
			output: errors.New(QemuNetworkInterfaceID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuNetworkInterfaces_Validate(t *testing.T) {
	type testInput struct {
		config  QemuNetworkInterfaces
		current QemuNetworkInterfaces
	}
	tests := []struct {
		name   string
		input  testInput
		output error
	}{
		{name: `Valid Delete`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{Delete: true}}}},
		{name: `Valid MTU inherit`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{
					Model: util.Pointer(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Inherit: true}}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{}}}},
		{name: `Valid MTU value`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID1: QemuNetworkInterface{
					Model: util.Pointer(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Value: 1500}}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID1: QemuNetworkInterface{}}}},
		{name: `Valid MTU empty`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID2: QemuNetworkInterface{
					MTU: &QemuMTU{}}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID2: QemuNetworkInterface{}}}},
		{name: `Valid Model`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID3: QemuNetworkInterface{
					Model: util.Pointer(QemuNetworkModel("virtio"))}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID3: QemuNetworkInterface{}}}},
		{name: `Valid MultiQueue`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID4: QemuNetworkInterface{
					MultiQueue: util.Pointer(QemuNetworkQueue(64))}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID4: QemuNetworkInterface{}}}},
		{name: `Valid RateLimitKBps`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID5: QemuNetworkInterface{
					RateLimitKBps: util.Pointer(QemuNetworkRate(10240000))}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID5: QemuNetworkInterface{}}}},
		{name: `Valid NativeVlan`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID6: QemuNetworkInterface{
					NativeVlan: util.Pointer(Vlan(5))}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID6: QemuNetworkInterface{}}}},
		{name: `Valid TaggedVlans`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{
					TaggedVlans: util.Pointer(Vlans{0, 45, 12, 4095, 12, 45})}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{}}}},
		// Invalid
		{name: `Invalid errors.New(QemuNetworkInterfaceID_Error_Invalid)`,
			input:  testInput{config: QemuNetworkInterfaces{32: QemuNetworkInterface{}}},
			output: errors.New(QemuNetworkInterfaceID_Error_Invalid)},
		{name: `Invalid errors.New(QemuNetworkInterface_Error_BridgeRequired)`,
			input:  testInput{config: QemuNetworkInterfaces{QemuNetworkInterfaceID8: QemuNetworkInterface{}}},
			output: errors.New(QemuNetworkInterface_Error_BridgeRequired)},
		{name: `Invalid errors.New(QemuNetworkInterface_Error_ModelRequired)`,
			input: testInput{config: QemuNetworkInterfaces{QemuNetworkInterfaceID8: QemuNetworkInterface{
				Bridge: util.Pointer("vmbr0")}}},
			output: errors.New(QemuNetworkInterface_Error_ModelRequired)},
		{name: `Invalid errors.New(MTU_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID9: QemuNetworkInterface{
					Model: util.Pointer(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Value: 575}}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID9: QemuNetworkInterface{}}},
			output: errors.New(MTU_Error_Invalid)},
		{name: `Invalid errors.New(QemuMTU_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID10: QemuNetworkInterface{
					Model: util.Pointer(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Inherit: true, Value: 1500}}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID10: QemuNetworkInterface{}}},
			output: errors.New(QemuMTU_Error_Invalid)},
		{name: `Invalid Model`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID11: QemuNetworkInterface{
					Model: util.Pointer(QemuNetworkModel("invalid"))}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID11: QemuNetworkInterface{}}},
			output: QemuNetworkModel("").Error()},
		{name: `Invalid errors.New(QemuNetworkQueue_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID12: QemuNetworkInterface{
					MultiQueue: util.Pointer(QemuNetworkQueueMaximum + 1)}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID12: QemuNetworkInterface{}}},
			output: errors.New(QemuNetworkQueue_Error_Invalid)},
		{name: `Invalid errors.New(QemuNetworkRate_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{
					RateLimitKBps: util.Pointer(QemuNetworkRate(10240001))}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{}}},
			output: errors.New(QemuNetworkRate_Error_Invalid)},
		{name: `Invalid NativeVlan errors.New(Vlan_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID14: QemuNetworkInterface{
					NativeVlan: util.Pointer(Vlan(4096))}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID14: QemuNetworkInterface{}}},
			output: errors.New(Vlan_Error_Invalid)},
		{name: `Invalid TaggedVlans errors.New(Vlan_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID15: QemuNetworkInterface{
					TaggedVlans: util.Pointer(Vlans{0, 45, 12, 4095, 12, 45, 4096})}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID15: QemuNetworkInterface{}}},
			output: errors.New(Vlan_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.config.Validate(test.input.current))
		})
	}
}

func Test_QemuNetworkModel_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  QemuNetworkModel
		output error
	}{
		{name: `Valid weird`,
			input: "E__1--0__-__00-8__2--__--545_Em__"},
		{name: `Valid normal`,
			input: QemuNetworkModelE100082544gc},
		{name: `Invalid`,
			input:  "invalid",
			output: QemuNetworkModel("").Error()},
		{name: `Invalid empty`,
			input:  "",
			output: QemuNetworkModel("").Error()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuNetworkQueue_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  QemuNetworkQueue
		output error
	}{
		{name: `Valid Minimum`,
			input: 0},
		{name: `Valid Maximum`,
			input: 64},
		{name: `Invalid`,
			input:  65,
			output: errors.New(QemuNetworkQueue_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuNetworkRate_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  QemuNetworkRate
		output error
	}{
		{name: `Valid maximum`,
			input: 10240000},
		{name: `Valid minimum`,
			input: 0},
		{name: `Invalid`,
			input:  10240001,
			output: errors.New(QemuNetworkRate_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
