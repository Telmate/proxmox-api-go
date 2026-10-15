package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func testData_RawConfigQemu_Boot_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `empty`,
				input:  map[string]any{"boot": " "},
				output: testQemuBaseConfig_get(ConfigQemu{})},
			{name: `single`,
				input: map[string]any{"boot": "order=ide1"},
				output: testQemuBaseConfig_get(ConfigQemu{Boot: &BootOrder{
					"ide1",
				}})},
			{name: `double`,
				input: map[string]any{"boot": "order=scsi3;usb1"},
				output: testQemuBaseConfig_get(ConfigQemu{Boot: &BootOrder{
					"scsi3",
					"usb1",
				}})},
			{name: `multiple`,
				input: map[string]any{"boot": "order=net17;sata3;hostpci1"},
				output: testQemuBaseConfig_get(ConfigQemu{Boot: &BootOrder{
					"net17",
					"sata3",
					"hostpci1",
				}})},
		}
	}
}

func testData_ConfigQemu_BootOrder_Validate() qemuTestTypeValidateFunc {
	return qemuTestTypeValidateFunc(func() (qemuTestTypeInvalid, qemuTestTypeValid) {
		invalid := qemuTestTypeInvalid{
			createUpdate: []qemuTestCaseInvalid{
				{name: `errors.New(BootOrder_Error_Duplicate) 1`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"sata1",
						"sata1"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootOrder_Error_Duplicate)},
				{name: `errors.New(BootOrder_Error_Duplicate) 2`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"sata1",
						"ide2",
						"sata1"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootOrder_Error_Duplicate)},
				{name: `errors.New(BootOrder_Error_Duplicate) 3`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"ide2",
						"sata1",
						"sata1",
						"scsi6"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootOrder_Error_Duplicate)},
				{name: `errors.New(BootOrder_Error_Duplicate) 4`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"ide2",
						"sata1",
						"usb3",
						"sata1",
						"scsi6"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootOrder_Error_Duplicate)},
			}}
		valid := qemuTestTypeValid{
			createUpdate: []qemuTestCaseValid{
				{name: `empty`,
					input:   testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `minimum id`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"hostpci0",
						"ide0",
						"net0",
						"sata0",
						"scsi0",
						"usb0",
						"virtio0"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `maximum id`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"hostpci15",
						"ide3",
						"net31",
						"sata5",
						"scsi30",
						"usb4",
						"virtio15"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
			}}
		return invalid, valid
	})
}

func testData_ConfigQemu_BootDevice_Validate() qemuTestTypeValidateFunc {
	return qemuTestTypeValidateFunc(func() (qemuTestTypeInvalid, qemuTestTypeValid) {
		invalid := qemuTestTypeInvalid{
			createUpdate: []qemuTestCaseInvalid{
				{name: `maximum hostpci16`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"hostpci16"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `maximum ide4`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"ide4"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `maximum net32`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"net32"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `maximum sata6`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"sata6"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `maximum scsi31`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"scsi31"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `maximum usb5`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"usb5"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `maximum virtio16`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"virtio16"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `no number hostpciN`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"hostpciN"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `no number hostpciNN`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"hostpciNN"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `no number ideN`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"ideN"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `no number netN`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"netN"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `no number netNN`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"netNN"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `no number sataN`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"sataN"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `no number scsiN`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"scsiN"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `no number scsiNN`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"scsiNN"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `no number usbN`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"usbN"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `no number virtioN`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"virtioN"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
				{name: `no number virtioNN`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"virtioNN"}}),
					current: &ConfigQemu{Boot: new(BootOrder)},
					err:     errors.New(BootDevice_Error_Invalid)},
			}}
		valid := qemuTestTypeValid{
			createUpdate: []qemuTestCaseValid{
				{name: `maximum hostpci15`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"hostpci15"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `maximum ide3`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"ide3"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `maximum net31`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"net31"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `maximum sata5`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"sata5"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `maximum scsi30`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"scsi30"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `maximum usb4`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"usb4"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `maximum virtio15`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"virtio15"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `minimum hostpci0`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"hostpci0"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `minimum ide0`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"ide0"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `minimum net0`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"net0"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `minimum sata0`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"sata0"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `minimum scsi0`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"scsi0"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `minimum usb0`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"usb0"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
				{name: `minimum virtio0`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Boot: &BootOrder{
						"virtio0"}}),
					current: &ConfigQemu{Boot: new(BootOrder)}},
			}}
		return invalid, valid
	})
}

func Test_ConfigQemu_Boot_MapToApi(t *testing.T) {
	t.Parallel()
	qemuTestsApiFunc(func() qemuTestsAPI {
		return qemuTestsAPI{
			create: []qemuTestCaseAPI{
				{name: `empty`,
					config: &ConfigQemu{Boot: &BootOrder{}},
					body:   map[string]string{"boot": ""}}},
			createUpdate: []qemuTestCaseAPI{
				{name: `normal`,
					config: &ConfigQemu{Boot: &BootOrder{
						QemuIdeID2.BootDevice(),
						QemuSataID1.BootDevice(),
						QemuScsiID12.BootDevice(),
						QemuVirtIoID14.BootDevice(),
						QemuPciID2.BootDevice(),
						QemuUsbID0.BootDevice(),
						QemuNetworkInterfaceID15.BootDevice()}},
					currentLegacy: ConfigQemu{Boot: &BootOrder{
						QemuNetworkInterfaceID3.BootDevice(),
						QemuScsiID15.BootDevice(),
						QemuUsbID0.BootDevice(),
						QemuVirtIoID1.BootDevice()}},
					body: map[string]string{"boot": "order%3Dide2%3Bsata1%3Bscsi12%3Bvirtio14%3Bhostpci2%3Busb0%3Bnet15"}}}, // "order=ide2;sata1;scsi12;virtio14;hostpci2;usb0;net15"
			update: []qemuTestCaseAPI{
				{name: `same do nothing empty`,
					config:        &ConfigQemu{Boot: &BootOrder{}},
					currentLegacy: ConfigQemu{Boot: new(BootOrder)}},
				{name: `same do nothing set`,
					config: &ConfigQemu{Boot: &BootOrder{
						QemuNetworkInterfaceID3.BootDevice(),
						QemuPciID5.BootDevice(),
						QemuSataID1.BootDevice(),
						QemuIdeID2.BootDevice(),
						QemuScsiID15.BootDevice(),
						QemuUsbID0.BootDevice(),
						QemuVirtIoID1.BootDevice()}},
					currentLegacy: ConfigQemu{Boot: &BootOrder{
						QemuNetworkInterfaceID3.BootDevice(),
						QemuPciID5.BootDevice(),
						QemuSataID1.BootDevice(),
						QemuIdeID2.BootDevice(),
						QemuScsiID15.BootDevice(),
						QemuUsbID0.BootDevice(),
						QemuVirtIoID1.BootDevice()}}},
				{name: `clear`,
					config: &ConfigQemu{Boot: &BootOrder{}},
					currentLegacy: ConfigQemu{Boot: &BootOrder{
						QemuNetworkInterfaceID3.BootDevice(),
						QemuScsiID15.BootDevice(),
						QemuUsbID0.BootDevice(),
						QemuVirtIoID1.BootDevice()}},
					body: map[string]string{"boot": ""}}}}
	}).Test(t)
}

func Test_RawConfigQemu_GetBoot(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Boot_Get().Inject(t,
		func(t *testing.T, raw RawConfigQemu, output *ConfigQemu, _ error) {
			require.Equal(t, *output.Boot, raw.GetBoot())
		})
}

func Test_RawConfigQemu_Boot_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Boot_Get().Test(t)

}

func Test_ConfigQemu_BootOrder_Validate(t *testing.T) {
	t.Parallel()
	testData_ConfigQemu_BootOrder_Validate().Test(t)
	testData_ConfigQemu_BootDevice_Validate().Test(t)
}

func Test_BootOrder_Validate(t *testing.T) {
	t.Parallel()
	validate := func(t *testing.T, config ConfigQemu, current *ConfigQemu, version Version, expectedErr error, valid bool) {
		t.Helper()
		err := config.Boot.Validate()
		if valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			if expectedErr != nil {
				require.Equal(t, expectedErr, err)
			}
		}
	}
	testData_ConfigQemu_BootOrder_Validate().Inject(t, validate)
	testData_ConfigQemu_BootDevice_Validate().Inject(t, validate)
}

func Test_BootDevice_Validate(t *testing.T) {
	t.Parallel()
	validate := func(t *testing.T, config ConfigQemu, current *ConfigQemu, version Version, expectedErr error, valid bool) {
		t.Helper()
		err := (*config.Boot)[0].Validate()
		if valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			if expectedErr != nil {
				require.Equal(t, expectedErr, err)
			}
		}
	}
	testData_ConfigQemu_BootDevice_Validate().Inject(t, validate)
}
