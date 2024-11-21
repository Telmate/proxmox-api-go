package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_QemuPciDevices_Validate(t *testing.T) {
	type testInput struct {
		config  QemuPciDevices
		current QemuPciDevices
	}
	tests := []struct {
		name   string
		input  testInput
		output error
	}{
		{name: `Valid Delete`,
			input: testInput{config: QemuPciDevices{
				QemuPciID0: QemuPci{Delete: true}}}},
		{name: `Valid Mapping update`,
			input: testInput{
				config: QemuPciDevices{
					QemuPciID1: QemuPci{
						Mapping: &QemuPciMapping{}}},
				current: QemuPciDevices{
					QemuPciID1: QemuPci{
						Mapping: &QemuPciMapping{
							ID: util.Pointer(ResourceMappingPciID("aaa"))}}}}},
		{name: `Valid Raw update`,
			input: testInput{
				config: QemuPciDevices{
					QemuPciID2: QemuPci{
						Raw: &QemuPciRaw{}}},
				current: QemuPciDevices{
					QemuPciID2: QemuPci{
						Raw: &QemuPciRaw{
							ID: util.Pointer(PciID("0000:00:00"))}}}}},
		{name: `Invalid update errors.New(QemuPci_Error_MutualExclusive)`,
			input: testInput{
				config: QemuPciDevices{
					QemuPciID3: QemuPci{
						Mapping: &QemuPciMapping{ID: util.Pointer(ResourceMappingPciID("aaa"))},
						Raw:     &QemuPciRaw{ID: util.Pointer(PciID("0000:00:00"))}}},
				current: QemuPciDevices{
					QemuPciID3: QemuPci{}}},
			output: errors.New(QemuPci_Error_MutualExclusive)},
		{name: `Invalid errors.New(QemuPciID_Error_Invalid)`,
			input: testInput{config: QemuPciDevices{
				16: QemuPci{}}},
			output: errors.New(QemuPciID_Error_Invalid)},
		{name: `Invalid errors.New(QemuPci_Error_MutualExclusive)`,
			input: testInput{config: QemuPciDevices{
				QemuPciID4: QemuPci{
					Mapping: &QemuPciMapping{
						ID: util.Pointer(ResourceMappingPciID("aaa"))},
					Raw: &QemuPciRaw{
						ID: util.Pointer(PciID("0000:00:00"))}}}},
			output: errors.New(QemuPci_Error_MutualExclusive)},
		{name: `Invalid errors.New(QemuPci_Error_MappedID)`,
			input: testInput{config: QemuPciDevices{
				QemuPciID5: QemuPci{
					Mapping: &QemuPciMapping{}}}},
			output: errors.New(QemuPci_Error_MappedID)},
		{name: `Invalid errors.New(QemuPci_Error_RawID)`,
			input: testInput{config: QemuPciDevices{
				QemuPciID6: QemuPci{
					Raw: &QemuPciRaw{}}}},
			output: errors.New(QemuPci_Error_RawID)},
		{name: `Invalid errors.New(ResourceMappingPciID_Error_Invalid)`,
			input: testInput{config: QemuPciDevices{
				QemuPciID7: QemuPci{
					Mapping: &QemuPciMapping{
						ID: util.Pointer(ResourceMappingPciID("a0%^#"))}}}},
			output: errors.New(ResourceMappingPciID_Error_Invalid)},
		{name: `Invalid Mapping errors.New(PciDeviceID_Error_Invalid)`,
			input: testInput{config: QemuPciDevices{
				QemuPciID8: QemuPci{
					Mapping: &QemuPciMapping{
						ID:       util.Pointer(ResourceMappingPciID("aaa")),
						DeviceID: util.Pointer(PciDeviceID("a0%^#"))}}}},
			output: errors.New(PciDeviceID_Error_Invalid)},
		{name: `Invalid Mapping errors.New(PciSubDeviceID_Error_Invalid)`,
			input: testInput{config: QemuPciDevices{
				QemuPciID9: QemuPci{
					Mapping: &QemuPciMapping{
						ID:          util.Pointer(ResourceMappingPciID("aaa")),
						SubDeviceID: util.Pointer(PciSubDeviceID("a0%^#"))}}}},
			output: errors.New(PciSubDeviceID_Error_Invalid)},
		{name: `Invalid Mapping errors.New(PciSubVendorID_Error_Invalid)`,
			input: testInput{config: QemuPciDevices{
				QemuPciID10: QemuPci{
					Mapping: &QemuPciMapping{
						ID:          util.Pointer(ResourceMappingPciID("aaa")),
						SubVendorID: util.Pointer(PciSubVendorID("a0%^#"))}}}},
			output: errors.New(PciSubVendorID_Error_Invalid)},
		{name: `Invalid Mapping errors.New(PciVendorID_Error_Invalid)`,
			input: testInput{config: QemuPciDevices{
				QemuPciID11: QemuPci{
					Mapping: &QemuPciMapping{
						ID:       util.Pointer(ResourceMappingPciID("aaa")),
						VendorID: util.Pointer(PciVendorID("a0%^#"))}}}},
			output: errors.New(PciVendorID_Error_Invalid)},
		{name: `Invalid errors.New(PciID_Error_MaximumFunction)`,
			input: testInput{config: QemuPciDevices{
				QemuPciID12: QemuPci{
					Raw: &QemuPciRaw{ID: util.Pointer(PciID("0000:00:00.8"))}}}},
			output: errors.New(PciID_Error_MaximumFunction)},
		{name: `Invalid Raw errors.New(PciDeviceID_Error_Invalid)`,
			input: testInput{config: QemuPciDevices{
				QemuPciID13: QemuPci{
					Raw: &QemuPciRaw{
						ID:       util.Pointer(PciID("0000:00:00")),
						DeviceID: util.Pointer(PciDeviceID("a0%^#"))}}}},
			output: errors.New(PciDeviceID_Error_Invalid)},
		{name: `Invalid Raw errors.New(PciSubDeviceID_Error_Invalid)`,
			input: testInput{config: QemuPciDevices{
				QemuPciID14: QemuPci{
					Raw: &QemuPciRaw{
						ID:          util.Pointer(PciID("0000:00:00")),
						SubDeviceID: util.Pointer(PciSubDeviceID("a0%^#"))}}}},
			output: errors.New(PciSubDeviceID_Error_Invalid)},
		{name: `Invalid Raw errors.New(PciSubVendorID_Error_Invalid)`,
			input: testInput{config: QemuPciDevices{
				QemuPciID15: QemuPci{
					Raw: &QemuPciRaw{
						ID:          util.Pointer(PciID("0000:00:00")),
						SubVendorID: util.Pointer(PciSubVendorID("a0%^#"))}}}},
			output: errors.New(PciSubVendorID_Error_Invalid)},
		{name: `Invalid Raw errors.New(PciVendorID_Error_Invalid)`,
			input: testInput{config: QemuPciDevices{
				QemuPciID0: QemuPci{
					Raw: &QemuPciRaw{
						ID:       util.Pointer(PciID("0000:00:00")),
						VendorID: util.Pointer(PciVendorID("a0%^#"))}}}},
			output: errors.New(PciVendorID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.config.Validate(test.input.current))
		})
	}
}

func Test_QemuPciID_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  QemuPciID
		output error
	}{
		{name: `Valid`,
			input: QemuPciIDMaximum},
		{name: `Invalid`,
			input:  QemuPciIDMaximum + 1,
			output: errors.New(QemuPciID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuPci_Validate(t *testing.T) {
	type testInput struct {
		config  QemuPci
		current QemuPci
	}
	tests := []struct {
		name   string
		input  testInput
		output error
	}{
		{name: `Valid Delete`,
			input: testInput{config: QemuPci{
				Delete: true}}},
		{name: `Valid Mapping update`,
			input: testInput{
				config: QemuPci{
					Mapping: &QemuPciMapping{}},
				current: QemuPci{
					Mapping: &QemuPciMapping{
						ID: util.Pointer(ResourceMappingPciID("aaa"))}}}},
		{name: `Valid Raw update`,
			input: testInput{
				config: QemuPci{
					Raw: &QemuPciRaw{}},
				current: QemuPci{
					Raw: &QemuPciRaw{
						ID: util.Pointer(PciID("0000:00:00"))}}}},
		{name: `Invalid errors.New(QemuPci_Error_MutualExclusive)`,
			input: testInput{config: QemuPci{
				Mapping: &QemuPciMapping{
					ID: util.Pointer(ResourceMappingPciID("aaa"))},
				Raw: &QemuPciRaw{
					ID: util.Pointer(PciID("0000:00:00"))}}},
			output: errors.New(QemuPci_Error_MutualExclusive)},
		{name: `Invalid errors.New(QemuPci_Error_MappedID)`,
			input: testInput{config: QemuPci{
				Mapping: &QemuPciMapping{}}},
			output: errors.New(QemuPci_Error_MappedID)},
		{name: `Invalid errors.New(QemuPci_Error_RawID)`,
			input: testInput{config: QemuPci{
				Raw: &QemuPciRaw{}}},
			output: errors.New(QemuPci_Error_RawID)},
		{name: `Invalid errors.New(ResourceMappingPciID_Error_Invalid)`,
			input: testInput{config: QemuPci{
				Mapping: &QemuPciMapping{
					ID: util.Pointer(ResourceMappingPciID("a0%^#"))}}},
			output: errors.New(ResourceMappingPciID_Error_Invalid)},
		{name: `Invalid Mapping errors.New(PciDeviceID_Error_Invalid)`,
			input: testInput{config: QemuPci{
				Mapping: &QemuPciMapping{
					ID:       util.Pointer(ResourceMappingPciID("aaa")),
					DeviceID: util.Pointer(PciDeviceID("a0%^#"))}}},
			output: errors.New(PciDeviceID_Error_Invalid)},
		{name: `Invalid Mapping errors.New(PciSubDeviceID_Error_Invalid)`,
			input: testInput{config: QemuPci{
				Mapping: &QemuPciMapping{
					ID:          util.Pointer(ResourceMappingPciID("aaa")),
					SubDeviceID: util.Pointer(PciSubDeviceID("a0%^#"))}}},
			output: errors.New(PciSubDeviceID_Error_Invalid)},
		{name: `Invalid Mapping errors.New(PciSubVendorID_Error_Invalid)`,
			input: testInput{config: QemuPci{
				Mapping: &QemuPciMapping{
					ID:          util.Pointer(ResourceMappingPciID("aaa")),
					SubVendorID: util.Pointer(PciSubVendorID("a0%^#"))}}},
			output: errors.New(PciSubVendorID_Error_Invalid)},
		{name: `Invalid Mapping errors.New(PciVendorID_Error_Invalid)`,
			input: testInput{config: QemuPci{
				Mapping: &QemuPciMapping{
					ID:       util.Pointer(ResourceMappingPciID("aaa")),
					VendorID: util.Pointer(PciVendorID("a0%^#"))}}},
			output: errors.New(PciVendorID_Error_Invalid)},
		{name: `Invalid errors.New(PciID_Error_MaximumFunction)`,
			input: testInput{config: QemuPci{
				Raw: &QemuPciRaw{ID: util.Pointer(PciID("0000:00:00.8"))}}},
			output: errors.New(PciID_Error_MaximumFunction)},
		{name: `Invalid Raw errors.New(PciDeviceID_Error_Invalid)`,
			input: testInput{config: QemuPci{
				Raw: &QemuPciRaw{
					ID:       util.Pointer(PciID("0000:00:00")),
					DeviceID: util.Pointer(PciDeviceID("a0%^#"))}}},
			output: errors.New(PciDeviceID_Error_Invalid)},
		{name: `Invalid Raw errors.New(PciSubDeviceID_Error_Invalid)`,
			input: testInput{config: QemuPci{
				Raw: &QemuPciRaw{
					ID:          util.Pointer(PciID("0000:00:00")),
					SubDeviceID: util.Pointer(PciSubDeviceID("a0%^#"))}}},
			output: errors.New(PciSubDeviceID_Error_Invalid)},
		{name: `Invalid Raw errors.New(PciSubVendorID_Error_Invalid)`,
			input: testInput{config: QemuPci{
				Raw: &QemuPciRaw{
					ID:          util.Pointer(PciID("0000:00:00")),
					SubVendorID: util.Pointer(PciSubVendorID("a0%^#"))}}},
			output: errors.New(PciSubVendorID_Error_Invalid)},
		{name: `Invalid Raw errors.New(PciVendorID_Error_Invalid)`,
			input: testInput{config: QemuPci{
				Raw: &QemuPciRaw{
					ID:       util.Pointer(PciID("0000:00:00")),
					VendorID: util.Pointer(PciVendorID("a0%^#"))}}},
			output: errors.New(PciVendorID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.config.Validate(test.input.current))
		})
	}
}

func Test_PciDeviceID_String(t *testing.T) {
	tests := []struct {
		name   string
		input  PciDeviceID
		output string
	}{
		{name: `No prefix`,
			input:  "ffff",
			output: "0xffff"},
		{name: `With prefix`,
			input:  "0x0000",
			output: "0x0000"},
		{name: `Empty`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.String())
		})
	}
}

func Test_PciDeviceID_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  PciDeviceID
		output error
	}{
		{name: `Valid Maximum`,
			input: "0xffff"},
		{name: `Valid Minimum`,
			input: "0x0000"},
		{name: `Valid no prefix`,
			input: "8086"},
		{name: `Valid empty`,
			input: ""},
		{name: `Invalid not hex`,
			input:  "0xg000",
			output: errors.New(PciDeviceID_Error_Invalid)},
		{name: `Invalid Maximum`,
			input:  "0x1ffff",
			output: errors.New(PciDeviceID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_PciSubDeviceID_String(t *testing.T) {
	tests := []struct {
		name   string
		input  PciSubDeviceID
		output string
	}{
		{name: `No prefix`,
			input:  "ffff",
			output: "0xffff"},
		{name: `With prefix`,
			input:  "0x0000",
			output: "0x0000"},
		{name: `Empty`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.String())
		})
	}
}

func Test_PciSubDeviceID_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  PciSubDeviceID
		output error
	}{
		{name: `Valid Maximum`,
			input: "0xffff"},
		{name: `Valid Minimum`,
			input: "0x0000"},
		{name: `Valid no prefix`,
			input: "8086"},
		{name: `Valid empty`,
			input: ""},
		{name: `Invalid not hex`,
			input:  "0xg000",
			output: errors.New(PciSubDeviceID_Error_Invalid)},
		{name: `Invalid Maximum`,
			input:  "0x1ffff",
			output: errors.New(PciSubDeviceID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_PciSubVendorID_String(t *testing.T) {
	tests := []struct {
		name   string
		input  PciSubVendorID
		output string
	}{
		{name: `No prefix`,
			input:  "ffff",
			output: "0xffff"},
		{name: `With prefix`,
			input:  "0x0000",
			output: "0x0000"},
		{name: `Empty`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.String())
		})
	}
}

func Test_PciSubVendorID_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  PciSubVendorID
		output error
	}{
		{name: `Valid Maximum`,
			input: "0xffff"},
		{name: `Valid Minimum`,
			input: "0x0000"},
		{name: `Valid no prefix`,
			input: "8086"},
		{name: `Valid empty`,
			input: ""},
		{name: `Invalid not hex`,
			input:  "0xg000",
			output: errors.New(PciSubVendorID_Error_Invalid)},
		{name: `Invalid Maximum`,
			input:  "0x1ffff",
			output: errors.New(PciSubVendorID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_PciVendorID_String(t *testing.T) {
	tests := []struct {
		name   string
		input  PciVendorID
		output string
	}{
		{name: `No prefix`,
			input:  "ffff",
			output: "0xffff"},
		{name: `With prefix`,
			input:  "0x0000",
			output: "0x0000"},
		{name: `Empty`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.String())
		})
	}
}

func Test_PciVendorID_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  PciVendorID
		output error
	}{
		{name: `Valid Maximum`,
			input: "0xffff"},
		{name: `Valid Minimum`,
			input: "0x0000"},
		{name: `Valid no prefix`,
			input: "8086"},
		{name: `Valid empty`,
			input: ""},
		{name: `Invalid not hex`,
			input:  "0xg000",
			output: errors.New(PciVendorID_Error_Invalid)},
		{name: `Invalid Maximum`,
			input:  "0x1ffff",
			output: errors.New(PciVendorID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_PciID_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  []PciID
		output error
	}{
		{name: `Valid`,
			input: []PciID{"1234:56:78", "0000:00:00.0"}},
		{name: `Invalid errors.New(PciID_Error_MissingBus)`,
			input:  []PciID{"0000"},
			output: errors.New(PciID_Error_MissingBus)},
		{name: `Invalid errors.New(PciID_Error_MissingDevice)`,
			input:  []PciID{"0000:00"},
			output: errors.New(PciID_Error_MissingDevice)},
		{name: `Invalid errors.New(PciID_Error_LengthDomain)`,
			input:  []PciID{"0:00:00", "0:00:00.0", "00:00:00", "00:00:00.0", "000:00:00", "000:00:00.0", "00000:00:00", "00000:00:00.0"},
			output: errors.New(PciID_Error_LengthDomain)},
		{name: `Invalid errors.New(PciID_Error_InvalidDomain)`,
			input:  []PciID{"gggg:00:00", "gggg:00:00.0"},
			output: errors.New(PciID_Error_InvalidDomain)},
		{name: `Invalid errors.New(PciID_Error_LengthBus)`,
			input:  []PciID{"0000:0:00", "0000:0:00.0", "0000:000:00", "0000:000:00.0"},
			output: errors.New(PciID_Error_LengthBus)},
		{name: `Invalid errors.New(PciID_Error_InvalidBus)`,
			input:  []PciID{"0000:gg:00", "0000:gg:00.0"},
			output: errors.New(PciID_Error_InvalidBus)},
		{name: `Invalid errors.New(PciID_Error_LengthDevice)`,
			input:  []PciID{"0000:00:0", "0000:00:0.0", "0000:00:000", "0000:00:000.0"},
			output: errors.New(PciID_Error_LengthDevice)},
		{name: `Invalid errors.New(PciID_Error_InvalidDevice)`,
			input:  []PciID{"0000:00:gg", "0000:00:gg.0"},
			output: errors.New(PciID_Error_InvalidDevice)},
		{name: `Invalid errors.New(PciID_Error_InvalidFunction)`,
			input:  []PciID{"0000:00:00.", "0000:00:00.a"},
			output: errors.New(PciID_Error_InvalidFunction)},
		{name: `Invalid errors.New(PciID_Error_MaximumFunction)`,
			input:  []PciID{"0000:00:00.8", "0000:00:00.76"},
			output: errors.New(PciID_Error_MaximumFunction)},
	}
	for _, test := range tests {
		for _, item := range test.input {
			t.Run(test.name+" :"+item.String(), func(t *testing.T) {
				require.Equal(t, test.output, item.Validate())
			})
		}
	}
}
