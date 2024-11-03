package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_resourcemapping"
	"github.com/stretchr/testify/require"
)

func Test_QemuUsbID_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  QemuUsbID
		output error
	}{
		{name: "Valid",
			input: 0},
		{name: "Valid max",
			input: 4},
		// Invalid
		{name: "QemuUsbID_Error_Invalid",
			input:  5,
			output: errors.New(QemuUsbID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuUSB_Validate(t *testing.T) {
	type testInput struct {
		config  QemuUSB
		current *QemuUSB
	}
	tests := []struct {
		name   string
		input  testInput
		output error
	}{
		{name: "Valid delete",
			input: testInput{config: QemuUSB{Delete: true}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.config.Validate(test.input.current))
		})
	}
}

func Test_QemuUsbDevice_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  QemuUsbDevice
		output error
	}{
		{name: "Valid set",
			input: QemuUsbDevice{
				ID: util.Pointer(UsbDeviceID("1234:5678"))}},
		{name: "Valid nil"},
		// Invalid
		{name: "UsbDeviceID_Error_Invalid",
			input: QemuUsbDevice{
				ID: util.Pointer(UsbDeviceID("162E"))},
			output: errors.New(UsbDeviceID_Error_Invalid)},
		{name: "UsbDeviceID_Error_VendorID",
			input: QemuUsbDevice{
				ID: util.Pointer(UsbDeviceID("7P03:162E"))},
			output: errors.New(UsbDeviceID_Error_VendorID)},
		{name: "UsbDeviceID_Error_ProductID",
			input: QemuUsbDevice{
				ID: util.Pointer(UsbDeviceID("162e:7P03"))},
			output: errors.New(UsbDeviceID_Error_ProductID)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuUsbMapped_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  QemuUsbMapping
		output error
	}{
		{name: "Valid set",
			input: QemuUsbMapping{
				ID: util.Pointer(ResourceMappingUsbID(test_data_resourcemapping.ResourceMappingUsbID_Legal()[0]))}},
		{name: "Valid nil"},
		// Invalid
		{name: "ResourceMappingUsbID_Error_MinLength",
			input: QemuUsbMapping{
				ID: util.Pointer(ResourceMappingUsbID(test_data_resourcemapping.ResourceMappingUsbID_Min_Illegal()[0]))},
			output: errors.New(ResourceMappingUsbID_Error_MinLength)},
		{name: "ResourceMappingUsbID_Error_MaxLength",
			input: QemuUsbMapping{
				ID: util.Pointer(ResourceMappingUsbID(test_data_resourcemapping.ResourceMappingUsbID_Max_Illegal()))},
			output: errors.New(ResourceMappingUsbID_Error_MaxLength)},
		{name: "ResourceMappingUsbID_Error_Start",
			input: QemuUsbMapping{
				ID: util.Pointer(ResourceMappingUsbID(test_data_resourcemapping.ResourceMappingUsbID_Start_Illegal()[0]))},
			output: errors.New(ResourceMappingUsbID_Error_Start)},
		{name: "ResourceMappingUsbID_Error_Invalid",
			input: QemuUsbMapping{
				ID: util.Pointer(ResourceMappingUsbID(test_data_resourcemapping.ResourceMappingUsbID_Character_Illegal()[0]))},
			output: errors.New(ResourceMappingUsbID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuUsbPort_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  QemuUsbPort
		output error
	}{
		{name: "Valid set",
			input: QemuUsbPort{ID: util.Pointer(UsbPortID("1-3"))}},
		{name: "Valid nil"},
		// Invalid
		{name: "UsbDeviceID_Error_Invalid",
			input:  QemuUsbPort{ID: util.Pointer(UsbPortID("2-4-5"))},
			output: errors.New(UsbPortID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_UsbDeviceID_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  UsbDeviceID
		output error
	}{
		{name: "Valid",
			input: "1234:5678"},
		// Invalid
		{name: "UsbDeviceID_Error_Invalid",
			input:  "162E",
			output: errors.New(UsbDeviceID_Error_Invalid)},
		{name: "UsbDeviceID_Error_VendorID",
			input:  "7P03:162E",
			output: errors.New(UsbDeviceID_Error_VendorID)},
		{name: "UsbDeviceID_Error_ProductID",
			input:  "162e:7P03",
			output: errors.New(UsbDeviceID_Error_ProductID)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_UsbPortID_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  UsbPortID
		output error
	}{
		{name: "Valid",
			input: "2-4"},
		// Invalid
		{name: "UsbPortID_Error_Invalid",
			input:  "2-4-5",
			output: errors.New(UsbPortID_Error_Invalid)},
		{name: "UsbPortID_Error_Invalid",
			input:  "a-2",
			output: errors.New(UsbPortID_Error_Invalid)},
		{name: "UsbPortID_Error_Invalid",
			input:  "2-b",
			output: errors.New(UsbPortID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
