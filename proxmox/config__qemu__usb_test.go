package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_resourcemapping"
	"github.com/stretchr/testify/require"
)

func testData_ConfigQemu_USB_Get() []qemuTestCaseGet {
	return []qemuTestCaseGet{
		{name: `Device`,
			input: map[string]any{
				"usb0": "host=1234:5678"},
			output: testQemuBaseConfig_get(ConfigQemu{USBs: QemuUSBs{
				QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
					ID:   new(UsbDeviceID("1234:5678")),
					USB3: new(false)}}}})},
		{name: `Device usb3`,
			input: map[string]any{
				"usb1": "host=1234:5678,usb3=1"},
			output: testQemuBaseConfig_get(ConfigQemu{USBs: QemuUSBs{
				QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
					ID:   new(UsbDeviceID("1234:5678")),
					USB3: new(true)}}}})},
		{name: `Port`,
			input: map[string]any{"usb2": "host=1-2"},
			output: testQemuBaseConfig_get(ConfigQemu{USBs: QemuUSBs{
				QemuUsbID2: QemuUSB{Port: &QemuUsbPort{
					ID:   new(UsbPortID("1-2")),
					USB3: new(false)}}}})},
		{name: `Port usb3`,
			input: map[string]any{"usb3": "host=2-4,usb3=1"},
			output: testQemuBaseConfig_get(ConfigQemu{USBs: QemuUSBs{
				QemuUsbID3: QemuUSB{Port: &QemuUsbPort{
					ID:   new(UsbPortID("2-4")),
					USB3: new(true)}}}})},
		{name: `mapping`,
			input: map[string]any{"usb4": "mapping=abcde"},
			output: testQemuBaseConfig_get(ConfigQemu{USBs: QemuUSBs{
				QemuUsbID4: QemuUSB{Mapping: &QemuUsbMapping{
					ID:   new(ResourceMappingUsbID("abcde")),
					USB3: new(false)}}}})},
		{name: `mapping usb3`,
			input: map[string]any{"usb0": "mapping=testmapping,usb3=1"},
			output: testQemuBaseConfig_get(ConfigQemu{USBs: QemuUSBs{
				QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
					ID:   new(ResourceMappingUsbID("testmapping")),
					USB3: new(true)}}}})},
		{name: `spice`,
			input: map[string]any{"usb1": "spice"},
			output: testQemuBaseConfig_get(ConfigQemu{USBs: QemuUSBs{
				QemuUsbID1: QemuUSB{Spice: &QemuUsbSpice{USB3: false}}}})},
		{name: `spice usb3`,
			input: map[string]any{"usb2": "spice,usb3=1"},
			output: testQemuBaseConfig_get(ConfigQemu{USBs: QemuUSBs{
				QemuUsbID2: QemuUSB{Spice: &QemuUsbSpice{USB3: true}}}})},
		{name: `code coverage`,
			input:  map[string]any{"usb3": ""},
			output: testQemuBaseConfig_get(ConfigQemu{USBs: QemuUSBs{QemuUsbID3: QemuUSB{}}})}}
}

func testData_ConfigQemu_USB_Validate() qemuTestTypeValidateFunc {
	return qemuTestTypeValidateFunc(func() (qemuTestTypeInvalid, qemuTestTypeValid) {
		invalid := qemuTestTypeInvalid{
			createUpdate: []qemuTestCaseInvalid{
				{name: `errors.New(QemuUsbID_Error_Invalid)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						20: QemuUSB{}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{}}}},
					err: errors.New(QemuUsbID_Error_Invalid)},
				{name: `errors.New(QemuUSB_Error_MutualExclusive)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{
							Device:  &QemuUsbDevice{ID: new(UsbDeviceID("1234:5678"))},
							Mapping: &QemuUsbMapping{ID: new(ResourceMappingUsbID("test"))}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{}}},
					err: errors.New(QemuUSB_Error_MutualExclusive)},
				{name: `errors.New(QemuUSB_Error_DeviceID)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{}}}},
					err: errors.New(QemuUSB_Error_DeviceID)},
				{name: `errors.New(QemuUSB_Error_MappedID)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Port: &QemuUsbPort{}}}},
					err: errors.New(QemuUSB_Error_MappingID)},
				{name: `errors.New(QemuUSB_Error_PortID)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Port: &QemuUsbPort{}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{}}}},
					err: errors.New(QemuUSB_Error_PortID)},
				{name: `errors.New(UsbDeviceID_Error_Invalid)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
							ID: new(UsbDeviceID("1234"))}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{}}}},
					err: errors.New(UsbDeviceID_Error_Invalid)},
				{name: `errors.New(ResourceMappingUsbID_Error_Invalid)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
							ID: new(ResourceMappingUsbID("Invalid%"))}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{}}}},
					err: errors.New(ResourceMappingUsbID_Error_Invalid)},
				{name: `errors.New(UsbPortID_Error_Invalid)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
							ID: new(UsbPortID("2-3-4"))}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Port: &QemuUsbPort{}}}},
					err: errors.New(UsbPortID_Error_Invalid)}},
			update: []qemuTestCaseInvalid{
				{name: `create errors.New(QemuUSB_Error_MutualExclusive)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{
							Device:  &QemuUsbDevice{ID: new(UsbDeviceID("1234:5678"))},
							Mapping: &QemuUsbMapping{ID: new(ResourceMappingUsbID("test"))}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{}}},
					err: errors.New(QemuUSB_Error_MutualExclusive)},
				{name: `create errors.New(QemuUSB_Error_DeviceID)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{}}},
					err: errors.New(QemuUSB_Error_DeviceID)},
				{name: `create errors.New(QemuUSB_Error_MappedID)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{}}},
					err: errors.New(QemuUSB_Error_MappingID)},
				{name: `create errors.New(QemuUSB_Error_PortID)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Port: &QemuUsbPort{}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{}}},
					err: errors.New(QemuUSB_Error_PortID)},
				{name: `create errors.New(UsbDeviceID_Error_Invalid)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
							ID: new(UsbDeviceID("1234"))}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{}}},
					err: errors.New(UsbDeviceID_Error_Invalid)},
				{name: `create errors.New(ResourceMappingUsbID_Error_Invalid)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
							ID: new(ResourceMappingUsbID("Invalid%"))}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{}}},
					err: errors.New(ResourceMappingUsbID_Error_Invalid)},
				{name: `create errors.New(UsbPortID_Error_Invalid)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
							ID: new(UsbPortID("2-3-4"))}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{}}},
					err: errors.New(UsbPortID_Error_Invalid)}}}
		valid := qemuTestTypeValid{
			createUpdate: []qemuTestCaseValid{
				{name: `delete`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Delete: true}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
							ID: new(UsbDeviceID("1234:5678"))}}}}},
				{name: `USBs.Device.ID set/update`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
							ID: new(UsbDeviceID("5678:1234"))}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
							ID:   new(UsbDeviceID("1234:5678")),
							USB3: new(true)}}}}},
				{name: `USBs.Mapped.ID set/update`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
							ID: new(ResourceMappingUsbID("valid"))}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   new(ResourceMappingUsbID("test")),
							USB3: new(true)}}}}},
				{name: `USBs.Port.ID set/update`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
							ID: new(UsbPortID("6-4"))}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
							ID:   new(UsbPortID("1-5")),
							USB3: new(true)}}}}},
				{name: `USBs.Spice.USB3 set/update`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Spice: &QemuUsbSpice{
							USB3: true}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Spice: &QemuUsbSpice{
							USB3: false}}}}}},
			update: []qemuTestCaseValid{
				{name: `USBs.Device to USBs.Mapped`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
							ID: new(UsbDeviceID("1234:5678"))}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
							ID: new(ResourceMappingUsbID("test"))}}}}},
				{name: `USBs.Device.USB3 update`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
							USB3: new(true)}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
							ID:   new(UsbDeviceID("1234:5678")),
							USB3: new(false)}}}}},
				{name: `USBs.Mapped to USBs.Port`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
							ID: new(ResourceMappingUsbID("test"))}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
							ID: new(UsbPortID("3-5"))}}}}},
				{name: `USBs.Mapped.USB3 update`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
							USB3: new(true)}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   new(ResourceMappingUsbID("test")),
							USB3: new(false)}}}}},
				{name: `USBs.Port to USBs.Spice`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
							ID: new(UsbPortID("2-6"))}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Spice: &QemuUsbSpice{}}}}},
				{name: `USBs.Port.USB3 update`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
							USB3: new(true)}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
							ID:   new(UsbPortID("2-6")),
							USB3: new(false)}}}}},
				{name: `USBs.Spice to USBs.Device`,
					input: testQemuBaseConfig_Validate(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Spice: &QemuUsbSpice{}}}}),
					current: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
							ID: new(UsbDeviceID("5678:1234"))}}}}}}}
		return invalid, valid
	})
}

func Test_ConfigQemu_USB_Validate(t *testing.T) {
	t.Parallel()
	testData_ConfigQemu_USB_Validate().Test(t)
}

func Test_QemuUSBs_Validate(t *testing.T) {
	t.Parallel()
	validate := func(t *testing.T, config ConfigQemu, current *ConfigQemu, version Version, expectedErr error, valid bool) {
		t.Helper()
		var currentUSBs QemuUSBs
		if current != nil {
			currentUSBs = current.USBs
		}
		err := config.USBs.Validate(currentUSBs)
		if valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			if expectedErr != nil {
				require.Equal(t, expectedErr, err)
			}
		}
	}
	testData_ConfigQemu_USB_Validate().Inject(t, validate)
}

func Test_QemuUsbID_Validate(t *testing.T) {
	t.Parallel()
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

func Test_ConfigQemu_USB_Api(t *testing.T) {
	t.Parallel()
	tests := qemuTestsApiFunc(func() qemuTestsAPI {
		return qemuTestsAPI{
			createUpdate: []qemuTestCaseAPI{
				{name: `delete no effect`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Delete: true}}}},
				{name: `Device all`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
							ID:   new(UsbDeviceID("1234:5678")),
							USB3: new(true)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{}}}},
					body: map[string]string{"usb0": "host%3D1234%3A5678%2Cusb3%3D1"}}, // "host=1234:5678,usb3=1"
				{name: `Device.USB3 false`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID:   new(UsbDeviceID("abcd:35fe")),
							USB3: new(false)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{}}}},
					body: map[string]string{"usb1": "host%3Dabcd%3A35fe"}}, // "host=abcd:35fe"
				{name: `Device.USB3 nil`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID: new(UsbDeviceID("8235:95af"))}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{}}}},
					body: map[string]string{"usb1": "host%3D8235%3A95af"}}, // "host=8235:95af"
				{name: `Mapping all`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   new(ResourceMappingUsbID("test")),
							USB3: new(true)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{}}}},
					body: map[string]string{"usb1": "mapping%3Dtest%2Cusb3%3D1"}}, // "mapping=test,usb3=1"
				{name: `Mapping.USB3 false`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   new(ResourceMappingUsbID("test")),
							USB3: new(false)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{}}}},
					body: map[string]string{"usb1": "mapping%3Dtest"}}, // "mapping=test"
				{name: `Mapping.USB3 nil`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID: new(ResourceMappingUsbID("test"))}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{}}}},
					body: map[string]string{"usb1": "mapping%3Dtest"}}, // "mapping=test"
				{name: `Port all`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID2: QemuUSB{Port: &QemuUsbPort{
							ID:   new(UsbPortID("1-2")),
							USB3: new(true)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID2: QemuUSB{Spice: &QemuUsbSpice{}}}},
					body: map[string]string{"usb2": "host%3D1-2%2Cusb3%3D1"}}, // "host=1-2,usb3=1"
				{name: `Port.USB3 false`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID2: QemuUSB{Port: &QemuUsbPort{
							ID:   new(UsbPortID("1-2")),
							USB3: new(false)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID2: QemuUSB{Spice: &QemuUsbSpice{}}}},
					body: map[string]string{"usb2": "host%3D1-2"}}, // "host=1-2"
				{name: `Port.USB3 nil`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID2: QemuUSB{Port: &QemuUsbPort{
							ID: new(UsbPortID("1-2"))}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID2: QemuUSB{Spice: &QemuUsbSpice{}}}},
					body: map[string]string{"usb2": "host%3D1-2"}}, // "host=1-2"
				{name: `Spice all`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID3: QemuUSB{Spice: &QemuUsbSpice{
							USB3: true}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID3: QemuUSB{Device: &QemuUsbDevice{}}}},
					body: map[string]string{"usb3": "spice%2Cusb3%3D1"}}, // "spice,usb3=1"
				{name: `Spice.USB3 false`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID3: QemuUSB{Spice: &QemuUsbSpice{
							USB3: false}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID3: QemuUSB{Device: &QemuUsbDevice{}}}},
					body: map[string]string{"usb3": "spice"}}}, // "spice"
			update: []qemuTestCaseAPI{
				{name: `delete`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Delete: true}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{}}}},
					body: map[string]string{"delete": "usb0"}}, // "usb0"
				{name: `create`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
							ID:   new(UsbDeviceID("1234:5678")),
							USB3: new(true)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{}}},
					body: map[string]string{"usb0": "host%3D1234%3A5678%2Cusb3%3D1"}}, // "host=1234:5678,usb3=1"
				{name: `no change`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID:   new(UsbDeviceID("abcd:35fe")),
							USB3: new(false)}}}}},
				{name: `Device same`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID:   new(UsbDeviceID("abcd:35fe")),
							USB3: new(false)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID:   new(UsbDeviceID("abcd:35fe")),
							USB3: new(false)}}}}},
				{name: `Device.ID update`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID: new(UsbDeviceID("1234:5678"))}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID:   new(UsbDeviceID("abcd:35fe")),
							USB3: new(true)}}}},
					body: map[string]string{"usb1": "host%3D1234%3A5678%2Cusb3%3D1"}}, // "host=1234:5678,usb3=1"
				{name: `Device.ID same`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID: new(UsbDeviceID("abcd:35fe"))}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID:   new(UsbDeviceID("abcd:35fe")),
							USB3: new(true)}}}}},
				{name: `Device.USB3 update`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							USB3: new(true)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID:   new(UsbDeviceID("abcd:35fe")),
							USB3: new(false)}}}},
					body: map[string]string{"usb1": "host%3Dabcd%3A35fe%2Cusb3%3D1"}}, // "host=abcd:35fe,usb3=1"
				{name: `Device.USB3 same`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							USB3: new(false)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID:   new(UsbDeviceID("abcd:35fe")),
							USB3: new(false)}}}}},
				{name: `Mapping same`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   new(ResourceMappingUsbID("test2")),
							USB3: new(true)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   new(ResourceMappingUsbID("test2")),
							USB3: new(true)}}}}},
				{name: `Mapping.ID update`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID: new(ResourceMappingUsbID("test"))}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   new(ResourceMappingUsbID("test2")),
							USB3: new(true)}}}},
					body: map[string]string{"usb1": "mapping%3Dtest%2Cusb3%3D1"}}, // "mapping=test,usb3=1"
				{name: `Mapping.ID same`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID: new(ResourceMappingUsbID("test2"))}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   new(ResourceMappingUsbID("test2")),
							USB3: new(true)}}}}},
				{name: `Mapping.USB3 update`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							USB3: new(true)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   new(ResourceMappingUsbID("test2")),
							USB3: new(false)}}}},
					body: map[string]string{"usb1": "mapping%3Dtest2%2Cusb3%3D1"}}, // "mapping=test2,usb3=1"
				{name: `Mapping.USB3 same`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							USB3: new(true)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   new(ResourceMappingUsbID("test2")),
							USB3: new(true)}}}}},
				{name: `Port same`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{
							ID:   new(UsbPortID("2-3")),
							USB3: new(true)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{
							ID:   new(UsbPortID("2-3")),
							USB3: new(true)}}}}},
				{name: `Port.ID update`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{
							ID: new(UsbPortID("1-2"))}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{
							ID:   new(UsbPortID("2-3")),
							USB3: new(true)}}}},
					body: map[string]string{"usb1": "host%3D1-2%2Cusb3%3D1"}}, // "host=1-2,usb3=1"
				{name: `Port.ID same`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{
							ID: new(UsbPortID("2-3"))}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{
							ID:   new(UsbPortID("2-3")),
							USB3: new(true)}}}}},
				{name: `Port.USB3 update`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{
							USB3: new(true)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{
							ID:   new(UsbPortID("2-3")),
							USB3: new(false)}}}},
					body: map[string]string{"usb1": "host%3D2-3%2Cusb3%3D1"}}, // "host=2-3,usb3=1"
				{name: `Port.USB3 same`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{
							USB3: new(false)}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{
							ID:   new(UsbPortID("2-3")),
							USB3: new(false)}}}}},
				{name: `Spice`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Spice: &QemuUsbSpice{
							USB3: false}}}},
					currentLegacy: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{}}}},
					body: map[string]string{"usb1": "spice"}}}} // "spice"
	})
	tests.Test(t)
}

func Test_QemuUSB_Validate(t *testing.T) {
	t.Parallel()
	type testInput struct {
		config  QemuUSB
		current *QemuUSB
	}
	tests := []struct {
		name   string
		input  testInput
		output error
	}{
		{name: "create Valid delete",
			input: testInput{config: QemuUSB{Delete: true}}},
		{name: "update Valid delete",
			input: testInput{
				config:  QemuUSB{Delete: true},
				current: &QemuUSB{}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.config.Validate(test.input.current))
		})
	}
}

func Test_QemuUsbDevice_Validate(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	tests := []struct {
		name   string
		input  UsbPortID
		output error
	}{
		{name: "Valid",
			input: "2-4"},
		{name: "Valid",
			input: "2-4.1"},
		{name: "Valid",
			input: "3-1.2.3.4.5.6.7.8.9"},
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
