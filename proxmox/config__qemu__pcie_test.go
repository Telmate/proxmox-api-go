package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func testData_RawConfigQemu_PciDevices_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `Mapping all`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci0": "mapping=abc,device-id=0xa97f,pcie=1,x-vga=1,rombar=0,sub-device-id=0x61a4,sub-vendor-id=0x98f1,vendor-id=0x4003,mdev=vendor-test"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID0: QemuPci{
						Mapping: &QemuPciMapping{
							DeviceID:    new(PciDeviceID("0xa97f")),
							ID:          new(ResourceMappingPciID("abc")),
							MDev:        new(PciMediatedDevice("vendor-test")),
							PCIe:        new(true),
							PrimaryGPU:  new(true),
							ROMbar:      new(false),
							SubDeviceID: new(PciSubDeviceID("0x61a4")),
							SubVendorID: new(PciSubVendorID("0x98f1")),
							VendorID:    new(PciVendorID("0x4003"))}}}})},
			{name: `Mapping.DeviceID`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci1": "mapping=abc,device-id=0xa97f"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID1: QemuPci{
						Mapping: &QemuPciMapping{
							DeviceID:   new(PciDeviceID("0xa97f")),
							ID:         new(ResourceMappingPciID("abc")),
							PCIe:       new(false),
							PrimaryGPU: new(false),
							ROMbar:     new(true)}}}})},
			{name: `Mapping.ID`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci2": "mapping=xyz"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID2: QemuPci{
						Mapping: &QemuPciMapping{
							ID:         new(ResourceMappingPciID("xyz")),
							PCIe:       new(false),
							PrimaryGPU: new(false),
							ROMbar:     new(true)}}}})},
			{name: `Mapping.MDev`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci3": "mapping=abc,mdev=test"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID3: QemuPci{
						Mapping: &QemuPciMapping{
							ID:         new(ResourceMappingPciID("abc")),
							MDev:       new(PciMediatedDevice("test")),
							PCIe:       new(false),
							PrimaryGPU: new(false),
							ROMbar:     new(true)}}}})},
			{name: `Mapping.Pci`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci3": "mapping=abc,pcie=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID3: QemuPci{
						Mapping: &QemuPciMapping{
							ID:         new(ResourceMappingPciID("abc")),
							PCIe:       new(true),
							PrimaryGPU: new(false),
							ROMbar:     new(true)}}}})},
			{name: `Mapping.PrimaryGPU`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci4": "mapping=abc,x-vga=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID4: QemuPci{
						Mapping: &QemuPciMapping{
							ID:         new(ResourceMappingPciID("abc")),
							PCIe:       new(false),
							PrimaryGPU: new(true),
							ROMbar:     new(true)}}}})},
			{name: `Mapping.ROMbar`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci5": "mapping=abc,rombar=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID5: QemuPci{
						Mapping: &QemuPciMapping{
							ID:         new(ResourceMappingPciID("abc")),
							PCIe:       new(false),
							PrimaryGPU: new(false),
							ROMbar:     new(false)}}}})},
			{name: `Mapping.SubDeviceID`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci6": "mapping=abc,sub-device-id=0x61a4"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID6: QemuPci{
						Mapping: &QemuPciMapping{
							ID:          new(ResourceMappingPciID("abc")),
							PCIe:        new(false),
							PrimaryGPU:  new(false),
							ROMbar:      new(true),
							SubDeviceID: new(PciSubDeviceID("0x61a4"))}}}})},
			{name: `Mapping.SubVendorID`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci7": "mapping=abc,sub-vendor-id=0x98f1"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID7: QemuPci{
						Mapping: &QemuPciMapping{
							ID:          new(ResourceMappingPciID("abc")),
							PCIe:        new(false),
							PrimaryGPU:  new(false),
							ROMbar:      new(true),
							SubVendorID: new(PciSubVendorID("0x98f1"))}}}})},
			{name: `Mapping.VendorID`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci8": "vendor-id=0x4003,mapping=abc"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID8: QemuPci{
						Mapping: &QemuPciMapping{
							ID:         new(ResourceMappingPciID("abc")),
							PCIe:       new(false),
							PrimaryGPU: new(false),
							ROMbar:     new(true),
							VendorID:   new(PciVendorID("0x4003"))}}}})},
			{name: `Raw all`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci15": "0000:02:05.7,device-id=0xa97f,pcie=1,x-vga=1,rombar=0,sub-device-id=0x61a4,sub-vendor-id=0x98f1,vendor-id=0x4003,mdev=vendor-test"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID15: QemuPci{
						Raw: &QemuPciRaw{
							DeviceID:    new(PciDeviceID("0xa97f")),
							ID:          new(PciID("0000:02:05.7")),
							MDev:        new(PciMediatedDevice("vendor-test")),
							PCIe:        new(true),
							PrimaryGPU:  new(true),
							ROMbar:      new(false),
							SubDeviceID: new(PciSubDeviceID("0x61a4")),
							SubVendorID: new(PciSubVendorID("0x98f1")),
							VendorID:    new(PciVendorID("0x4003"))}}}})},
			{name: `Raw.DeviceID`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci14": "0000:02:05.7,device-id=0xa97f"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID14: QemuPci{
						Raw: &QemuPciRaw{
							DeviceID:   new(PciDeviceID("0xa97f")),
							ID:         new(PciID("0000:02:05.7")),
							PCIe:       new(false),
							PrimaryGPU: new(false),
							ROMbar:     new(true)}}}})},
			{name: `Raw.ID`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci13": "0001:43:86.5"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID13: QemuPci{
						Raw: &QemuPciRaw{
							ID:         new(PciID("0001:43:86.5")),
							PCIe:       new(false),
							PrimaryGPU: new(false),
							ROMbar:     new(true)}}}})},
			{name: `Raw.MDev`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci12": "0000:02:05.7,mdev=test"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID12: QemuPci{
						Raw: &QemuPciRaw{
							ID:         new(PciID("0000:02:05.7")),
							MDev:       new(PciMediatedDevice("test")),
							PCIe:       new(false),
							PrimaryGPU: new(false),
							ROMbar:     new(true)}}}})},
			{name: `Raw.Pci`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci12": "0000:02:05.7,pcie=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID12: QemuPci{
						Raw: &QemuPciRaw{
							ID:         new(PciID("0000:02:05.7")),
							PCIe:       new(true),
							PrimaryGPU: new(false),
							ROMbar:     new(true)}}}})},
			{name: `Raw.PrimaryGPU`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci11": "0000:02:05.7,x-vga=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID11: QemuPci{
						Raw: &QemuPciRaw{
							ID:         new(PciID("0000:02:05.7")),
							PCIe:       new(false),
							PrimaryGPU: new(true),
							ROMbar:     new(true)}}}})},
			{name: `Raw.ROMbar`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci10": "0000:02:05.7,rombar=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID10: QemuPci{
						Raw: &QemuPciRaw{
							ID:         new(PciID("0000:02:05.7")),
							PCIe:       new(false),
							PrimaryGPU: new(false),
							ROMbar:     new(false)}}}})},
			{name: `Raw.SubDeviceID`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci9": "0000:02:05.7,sub-device-id=0x61a4"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID9: QemuPci{
						Raw: &QemuPciRaw{
							ID:          new(PciID("0000:02:05.7")),
							PCIe:        new(false),
							PrimaryGPU:  new(false),
							ROMbar:      new(true),
							SubDeviceID: new(PciSubDeviceID("0x61a4"))}}}})},
			{name: `Raw.SubVendorID`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci8": "0000:02:05.7,sub-vendor-id=0x98f1"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID8: QemuPci{
						Raw: &QemuPciRaw{
							ID:          new(PciID("0000:02:05.7")),
							PCIe:        new(false),
							PrimaryGPU:  new(false),
							ROMbar:      new(true),
							SubVendorID: new(PciSubVendorID("0x98f1"))}}}})},
			{name: `Raw.VendorID`,
				input: rawConfigQemu{a: map[string]any{
					"hostpci7": "0000:02:05.7,vendor-id=0x4003"}},
				output: testQemuBaseConfig_get(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID7: QemuPci{
						Raw: &QemuPciRaw{
							ID:         new(PciID("0000:02:05.7")),
							PCIe:       new(false),
							PrimaryGPU: new(false),
							ROMbar:     new(true),
							VendorID:   new(PciVendorID("0x4003"))}}}})}}
	}
}

func Test_ConfigQemu_PciDevices_MapToApi(t *testing.T) {
	t.Parallel()
	tests := qemuTestsApiFunc(func() qemuTestsAPI {
		return qemuTestsAPI{
			create: []qemuTestCaseAPI{
				{name: `Delete`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID5: QemuPci{Delete: true}}}}},
			createUpdate: []qemuTestCaseAPI{
				{name: `Mapping.DeviceID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID15: QemuPci{Mapping: &QemuPciMapping{
							DeviceID: new(PciDeviceID("8086"))}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID15: QemuPci{Mapping: &QemuPciMapping{
							DeviceID: new(PciDeviceID("0x8000"))}}}},
					body: map[string]string{"hostpci15": "mapping%3D%2Crombar%3D0%2Cdevice-id%3D0x8086"}}, // "mapping=,rombar=0,device-id=0x8086"
				{name: `Mapping.ID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID14: QemuPci{Mapping: &QemuPciMapping{
							ID: new(ResourceMappingPciID("aaaaa"))}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID14: QemuPci{Mapping: &QemuPciMapping{
							ID: new(ResourceMappingPciID("bbbbb"))}}}},
					body: map[string]string{"hostpci14": "mapping%3Daaaaa%2Crombar%3D0"}}, // "mapping=aaaaa,rombar=0"
				{name: `MApping.MDev`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID14: QemuPci{Mapping: &QemuPciMapping{
							MDev: new(PciMediatedDevice("vendor-665"))}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID14: QemuPci{Mapping: &QemuPciMapping{
							MDev: new(PciMediatedDevice(PciMediatedDevice("vendor-000")))}}}},
					body: map[string]string{"hostpci14": "mapping%3D%2Crombar%3D0%2Cmdev%3Dvendor-665"}}, // "mapping=,rombar=0,mdev=vendor-665"
				{name: `Mapping.Pci`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID13: QemuPci{Mapping: &QemuPciMapping{
							PCIe: new(true)}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID13: QemuPci{Mapping: &QemuPciMapping{
							PCIe: new(false)}}}},
					body: map[string]string{"hostpci13": "mapping%3D%2Cpcie%3D1%2Crombar%3D0"}}, // "mapping=,pcie=1,rombar=0"
				{name: `Mapping.PrimaryGPU`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID12: QemuPci{Mapping: &QemuPciMapping{
							PrimaryGPU: new(true)}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID12: QemuPci{Mapping: &QemuPciMapping{
							PrimaryGPU: new(false)}}}},
					body: map[string]string{"hostpci12": "mapping%3D%2Cx-vga%3D1%2Crombar%3D0"}}, // "mapping=,x-vga=1,rombar=0"
				{name: `Mapping.ROMbar`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID11: QemuPci{Mapping: &QemuPciMapping{
							ROMbar: new(true)}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID11: QemuPci{Mapping: &QemuPciMapping{
							ROMbar: new(false)}}}},
					body: map[string]string{"hostpci11": "mapping%3D"}}, // "mapping="
				{name: `Mapping.SubDeviceID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID10: QemuPci{Mapping: &QemuPciMapping{
							SubDeviceID: new(PciSubDeviceID("8086"))}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID10: QemuPci{Mapping: &QemuPciMapping{
							SubDeviceID: new(PciSubDeviceID("0x8000"))}}}},
					body: map[string]string{"hostpci10": "mapping%3D%2Crombar%3D0%2Csub-device-id%3D0x8086"}}, // "mapping=,rombar=0,sub-device-id=0x8086"
				{name: `Mapping.SubVendorID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID9: QemuPci{Mapping: &QemuPciMapping{
							SubVendorID: new(PciSubVendorID("8086"))}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID9: QemuPci{Mapping: &QemuPciMapping{
							SubVendorID: new(PciSubVendorID("0x8000"))}}}},
					body: map[string]string{"hostpci9": "mapping%3D%2Crombar%3D0%2Csub-vendor-id%3D0x8086"}}, // "mapping=,rombar=0,sub-vendor-id=0x8086"
				{name: `Mapping.VendorID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID8: QemuPci{Mapping: &QemuPciMapping{
							VendorID: new(PciVendorID("8086"))}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID8: QemuPci{Mapping: &QemuPciMapping{
							VendorID: new(PciVendorID("0x8000"))}}}},
					body: map[string]string{"hostpci8": "mapping%3D%2Crombar%3D0%2Cvendor-id%3D0x8086"}}, // "mapping=,rombar=0,vendor-id=0x8086"
				{name: `Raw.DeviceID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID0: QemuPci{Raw: &QemuPciRaw{
							DeviceID: new(PciDeviceID("8086"))}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID0: QemuPci{Raw: &QemuPciRaw{
							DeviceID: new(PciDeviceID("0x8000"))}}}},
					body: map[string]string{"hostpci0": "%2Crombar%3D0%2Cdevice-id%3D0x8086"}}, // ",rombar=0,device-id=0x8086"
				{name: `Raw.ID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID1: QemuPci{Raw: &QemuPciRaw{
							ID: new(PciID("0000:00:00.0"))}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID1: QemuPci{Raw: &QemuPciRaw{
							ID: new(PciID("0000:00:00.1"))}}}},
					body: map[string]string{"hostpci1": "0000%3A00%3A00.0%2Crombar%3D0"}}, // "0000:00:00.0,rombar=0"
				{name: `Raw.MDev`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID2: QemuPci{Raw: &QemuPciRaw{
							MDev: new(PciMediatedDevice("vendor-665"))}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID2: QemuPci{Raw: &QemuPciRaw{
							MDev: new(PciMediatedDevice("vendor-000"))}}}},
					body: map[string]string{"hostpci2": "%2Crombar%3D0%2Cmdev%3Dvendor-665"}}, // ",rombar=0,mdev=vendor-665"
				{name: `Raw.Pci`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID2: QemuPci{Raw: &QemuPciRaw{
							PCIe: new(true)}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID2: QemuPci{Raw: &QemuPciRaw{
							PCIe: new(false)}}}},
					body: map[string]string{"hostpci2": "%2Cpcie%3D1%2Crombar%3D0"}}, // ",pcie=1,rombar=0"
				{name: `Raw.PrimaryGPU`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID3: QemuPci{Raw: &QemuPciRaw{
							PrimaryGPU: new(true)}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID3: QemuPci{Raw: &QemuPciRaw{
							PrimaryGPU: new(false)}}}},
					body: map[string]string{"hostpci3": "%2Cx-vga%3D1%2Crombar%3D0"}}, // ",x-vga=1,rombar=0"
				{name: `Raw.ROMbar`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID4: QemuPci{Raw: &QemuPciRaw{
							ROMbar: new(true)}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID4: QemuPci{Raw: &QemuPciRaw{
							ROMbar: new(false)}}}},
					body: map[string]string{"hostpci4": ""}},
				{name: `Raw.SubDeviceID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID5: QemuPci{Raw: &QemuPciRaw{
							SubDeviceID: new(PciSubDeviceID("8086"))}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID5: QemuPci{Raw: &QemuPciRaw{
							SubDeviceID: new(PciSubDeviceID("0x8000"))}}}},
					body: map[string]string{"hostpci5": "%2Crombar%3D0%2Csub-device-id%3D0x8086"}}, // ",rombar=0,sub-device-id=0x8086"
				{name: `Raw.SubVendorID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID6: QemuPci{Raw: &QemuPciRaw{
							SubVendorID: new(PciSubVendorID("8086"))}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID6: QemuPci{Raw: &QemuPciRaw{
							SubVendorID: new(PciSubVendorID("0x8000"))}}}},
					body: map[string]string{"hostpci6": "%2Crombar%3D0%2Csub-vendor-id%3D0x8086"}}, // ",rombar=0,sub-vendor-id=0x8086"
				{name: `Raw.VendorID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID7: QemuPci{Raw: &QemuPciRaw{
							VendorID: new(PciVendorID("8086"))}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID7: QemuPci{Raw: &QemuPciRaw{
							VendorID: new(PciVendorID("0x8000"))}}}},
					body: map[string]string{"hostpci7": "%2Crombar%3D0%2Cvendor-id%3D0x8086"}}}, // ",rombar=0,vendor-id=0x8086"
			update: []qemuTestCaseAPI{
				{name: `Delete`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID5: QemuPci{Delete: true}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID5: QemuPci{}}},
					body: map[string]string{"delete": "hostpci5"}},
				{name: `Delete create no effect`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID5: QemuPci{Delete: true}}}},
				{name: `Delete no effect`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID2: QemuPci{Delete: true}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID7: QemuPci{}}}},
				{name: `Mapping.DeviceID create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID15: QemuPci{Mapping: &QemuPciMapping{
							DeviceID: new(PciDeviceID("8086"))}}}},
					body: map[string]string{"hostpci15": "mapping%3D%2Crombar%3D0%2Cdevice-id%3D0x8086"}}, // "mapping=,rombar=0,device-id=0x8086"
				{name: `Mapping.ID create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID14: QemuPci{Mapping: &QemuPciMapping{
							ID: new(ResourceMappingPciID("aaaaa"))}}}},
					body: map[string]string{"hostpci14": "mapping%3Daaaaa%2Crombar%3D0"}}, // "mapping=aaaaa,rombar=0"
				{name: `MApping.MDev create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID14: QemuPci{Mapping: &QemuPciMapping{
							MDev: new(PciMediatedDevice("vendor-665"))}}}},
					body: map[string]string{"hostpci14": "mapping%3D%2Crombar%3D0%2Cmdev%3Dvendor-665"}}, // "mapping=,rombar=0,mdev=vendor-665"
				{name: `Mapping.Pci create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID13: QemuPci{Mapping: &QemuPciMapping{
							PCIe: new(true)}}}},
					body: map[string]string{"hostpci13": "mapping%3D%2Cpcie%3D1%2Crombar%3D0"}}, // "mapping=,pcie=1,rombar=0"
				{name: `Mapping.PrimaryGPU create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID12: QemuPci{Mapping: &QemuPciMapping{
							PrimaryGPU: new(true)}}}},
					body: map[string]string{"hostpci12": "mapping%3D%2Cx-vga%3D1%2Crombar%3D0"}}, // "mapping=,x-vga=1,rombar=0"
				{name: `Mapping.ROMbar create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID11: QemuPci{Mapping: &QemuPciMapping{
							ROMbar: new(true)}}}},
					body: map[string]string{"hostpci11": "mapping%3D"}}, // "mapping="
				{name: `Mapping.SubDeviceID create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID10: QemuPci{Mapping: &QemuPciMapping{
							SubDeviceID: new(PciSubDeviceID("8086"))}}}},
					body: map[string]string{"hostpci10": "mapping%3D%2Crombar%3D0%2Csub-device-id%3D0x8086"}}, // "mapping=,rombar=0,sub-device-id=0x8086"
				{name: `Mapping.SubVendorID create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID9: QemuPci{Mapping: &QemuPciMapping{
							SubVendorID: new(PciSubVendorID("8086"))}}}},
					body: map[string]string{"hostpci9": "mapping%3D%2Crombar%3D0%2Csub-vendor-id%3D0x8086"}}, // "mapping=,rombar=0,sub-vendor-id=0x8086"
				{name: `Mapping.VendorID create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID8: QemuPci{Mapping: &QemuPciMapping{
							VendorID: new(PciVendorID("8086"))}}}},
					body: map[string]string{"hostpci8": "mapping%3D%2Crombar%3D0%2Cvendor-id%3D0x8086"}}, // "mapping=,rombar=0,vendor-id=0x8086"
				{name: `Raw.DeviceID create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID0: QemuPci{Raw: &QemuPciRaw{
							DeviceID: new(PciDeviceID("8086"))}}}},
					body: map[string]string{"hostpci0": "%2Crombar%3D0%2Cdevice-id%3D0x8086"}}, // ",rombar=0,device-id=0x8086"
				{name: `Raw.ID create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID1: QemuPci{Raw: &QemuPciRaw{
							ID: new(PciID("0000:00:00.0"))}}}},
					body: map[string]string{"hostpci1": "0000%3A00%3A00.0%2Crombar%3D0"}}, // "0000:00:00.0,rombar=0"
				{name: `Raw.MDev create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID2: QemuPci{Raw: &QemuPciRaw{
							MDev: new(PciMediatedDevice("vendor-665"))}}}},
					body: map[string]string{"hostpci2": "%2Crombar%3D0%2Cmdev%3Dvendor-665"}}, // ",rombar=0,mdev=vendor-665"
				{name: `Raw.Pci create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID2: QemuPci{Raw: &QemuPciRaw{
							PCIe: new(true)}}}},
					body: map[string]string{"hostpci2": "%2Cpcie%3D1%2Crombar%3D0"}}, // ",pcie=1,rombar=0"
				{name: `Raw.PrimaryGPU create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID3: QemuPci{Raw: &QemuPciRaw{
							PrimaryGPU: new(true)}}}},
					body: map[string]string{"hostpci3": "%2Cx-vga%3D1%2Crombar%3D0"}}, // ",x-vga=1,rombar=0"
				{name: `Raw.ROMbar create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID4: QemuPci{Raw: &QemuPciRaw{
							ROMbar: new(true)}}}},
					body: map[string]string{"hostpci4": ""}},
				{name: `Raw.SubDeviceID create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID5: QemuPci{Raw: &QemuPciRaw{
							SubDeviceID: new(PciSubDeviceID("8086"))}}}},
					body: map[string]string{"hostpci5": "%2Crombar%3D0%2Csub-device-id%3D0x8086"}}, // ",rombar=0,sub-device-id=0x8086"
				{name: `Raw.SubVendorID create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID6: QemuPci{Raw: &QemuPciRaw{
							SubVendorID: new(PciSubVendorID("8086"))}}}},
					body: map[string]string{"hostpci6": "%2Crombar%3D0%2Csub-vendor-id%3D0x8086"}}, // ",rombar=0,sub-vendor-id=0x8086"
				{name: `Raw.VendorID create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID7: QemuPci{Raw: &QemuPciRaw{
							VendorID: new(PciVendorID("8086"))}}}},
					body: map[string]string{"hostpci7": "%2Crombar%3D0%2Cvendor-id%3D0x8086"}},
				{name: `Raw.VendorID update create`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID7: QemuPci{Raw: &QemuPciRaw{
							VendorID: new(PciVendorID("8086"))}}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{QemuPciID10: QemuPci{}}},
					body:          map[string]string{"hostpci7": "%2Crombar%3D0%2Cvendor-id%3D0x8086"}},
				{name: `same`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID4: QemuPci{
							Mapping: &QemuPciMapping{
								DeviceID:    new(PciDeviceID("8086")),
								ID:          new(ResourceMappingPciID("aaaaa")),
								MDev:        new(PciMediatedDevice("vendor-665")),
								PCIe:        new(false),
								PrimaryGPU:  new(true),
								ROMbar:      new(false),
								SubDeviceID: new(PciSubDeviceID("4522")),
								SubVendorID: new(PciSubVendorID("74526")),
								VendorID:    new(PciVendorID("2321"))},
						}}},
					currentLegacy: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID4: QemuPci{
							Mapping: &QemuPciMapping{
								DeviceID:    new(PciDeviceID("8086")),
								ID:          new(ResourceMappingPciID("aaaaa")),
								MDev:        new(PciMediatedDevice("vendor-665")),
								PCIe:        new(false),
								PrimaryGPU:  new(true),
								ROMbar:      new(false),
								SubDeviceID: new(PciSubDeviceID("4522")),
								SubVendorID: new(PciSubVendorID("74526")),
								VendorID:    new(PciVendorID("2321"))},
						}}},
				},
			}}
	})
	tests.Test(t)
}

func testData_ConfigQemu_PciDevices_Validate() qemuTestTypeValidateFunc {
	return qemuTestTypeValidateFunc(func() (qemuTestTypeInvalid, qemuTestTypeValid) {
		invalid := qemuTestTypeInvalid{
			createUpdate: []qemuTestCaseInvalid{
				{name: `errors.New(QemuPci_Error_MutualExclusive)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID12: QemuPci{
							Mapping: &QemuPciMapping{
								ID: new(ResourceMappingPciID("aaa"))},
							Raw: &QemuPciRaw{
								ID: new(PciID("0000:00:00"))}}}}),
					current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID12: QemuPci{}}},
					err:     errors.New(QemuPci_Error_MutualExclusive)},
				{name: `errors.New(QemuPci_Error_RawID)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID10: QemuPci{
							Raw: &QemuPciRaw{}}}}),
					current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID10: QemuPci{}}},
					err:     errors.New(QemuPciRaw_Error_RequiredID)},
				{name: `errors.New(PciID_Error_MaximumFunction)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID8: QemuPci{
							Raw: &QemuPciRaw{ID: new(PciID("0000:00:00.8"))}}}}),
					current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID8: QemuPci{}}},
					err:     errors.New(PciID_Error_MaximumFunction)}},
			update: []qemuTestCaseInvalid{
				{name: `create errors.New(QemuPci_Error_MutualExclusive)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID12: QemuPci{
							Mapping: &QemuPciMapping{
								ID: new(ResourceMappingPciID("aaa"))},
							Raw: &QemuPciRaw{
								ID: new(PciID("0000:00:00"))}}}}),
					current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID1: QemuPci{}}},
					err:     errors.New(QemuPci_Error_MutualExclusive)}}}
		valid := qemuTestTypeValid{
			createUpdate: []qemuTestCaseValid{
				{name: `Delete`,
					input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID15: QemuPci{Delete: true}}}),
					current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID0: QemuPci{}}}}}}
		invalidMapping, validMapping := testData_ConfigQemu_PciDevices_Mapping_Validate()
		invalidRaw, validRaw := testData_ConfigQemu_PciDevices_Raw_Validate()
		return invalid.append(invalidMapping).append(invalidRaw), valid.append(validMapping).append(validRaw)
	})
}

func testData_ConfigQemu_PciDevices_Mapping_Validate_func() qemuTestTypeValidateFunc {
	return qemuTestTypeValidateFunc(func() (qemuTestTypeInvalid, qemuTestTypeValid) {
		return testData_ConfigQemu_PciDevices_Mapping_Validate()
	})
}

func testData_ConfigQemu_PciDevices_Mapping_Validate() (qemuTestTypeInvalid, qemuTestTypeValid) {
	invalid := qemuTestTypeInvalid{
		createUpdate: []qemuTestCaseInvalid{
			{name: `errors.New(QemuPci_Error_MappedID)`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID11: QemuPci{
						Mapping: &QemuPciMapping{}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID11: QemuPci{}}},
				err:     errors.New(QemuPciMapping_Error_RequiredID)},
			{name: `errors.New(ResourceMappingPciID_Error_Invalid`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID9: QemuPci{
						Mapping: &QemuPciMapping{
							ID: new(ResourceMappingPciID("a0%^#"))}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID9: QemuPci{}}},
				err:     errors.New(ResourceMappingPciID_Error_Invalid)},
			{name: `Mapping errors.New(PciMediatedDevice_Error_Invalid)`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID7: QemuPci{
						Mapping: &QemuPciMapping{
							ID:   new(ResourceMappingPciID("aaa")),
							MDev: new(PciMediatedDevice(","))}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID7: QemuPci{}}},
				err:     errors.New(PciMediatedDevice_Error_Invalid)},
			{name: `Mapping errors.New(PciDeviceID_Error_Invalid)`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID7: QemuPci{
						Mapping: &QemuPciMapping{
							ID:       new(ResourceMappingPciID("aaa")),
							DeviceID: new(PciDeviceID("a0%^#"))}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID7: QemuPci{}}},
				err:     errors.New(PciDeviceID_Error_Invalid)},
			{name: `Mapping errors.New(PciSubDeviceID_Error_Invalid)`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID6: QemuPci{
						Mapping: &QemuPciMapping{
							ID:          new(ResourceMappingPciID("aaa")),
							SubDeviceID: new(PciSubDeviceID("a0%^#"))}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID6: QemuPci{}}},
				err:     errors.New(PciSubDeviceID_Error_Invalid)},
			{name: `Mapping errors.New(PciSubVendorID_Error_Invalid)`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID5: QemuPci{
						Mapping: &QemuPciMapping{
							ID:          new(ResourceMappingPciID("aaa")),
							SubVendorID: new(PciSubVendorID("a0%^#"))}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID5: QemuPci{}}},
				err:     errors.New(PciSubVendorID_Error_Invalid)},
			{name: `Mapping errors.New(PciVendorID_Error_Invalid)`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID4: QemuPci{
						Mapping: &QemuPciMapping{
							ID:       new(ResourceMappingPciID("aaa")),
							VendorID: new(PciVendorID("a0%^#"))}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID4: QemuPci{}}},
				err:     errors.New(PciVendorID_Error_Invalid)}}}
	valid := qemuTestTypeValid{
		createUpdate: []qemuTestCaseValid{
			{name: `Mapping normal`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID15: QemuPci{
						Mapping: &QemuPciMapping{
							ID: new(ResourceMappingPciID("test")),
						}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID0: QemuPci{}}}}},
		update: []qemuTestCaseValid{
			{name: `Mapping`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID14: QemuPci{
						Mapping: &QemuPciMapping{}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID14: QemuPci{
						Mapping: &QemuPciMapping{
							ID: new(ResourceMappingPciID("aaa"))}}}}}}}
	return invalid, valid
}

func testData_ConfigQemu_PciDevices_Raw_Validate_func() qemuTestTypeValidateFunc {
	return qemuTestTypeValidateFunc(func() (qemuTestTypeInvalid, qemuTestTypeValid) {
		return testData_ConfigQemu_PciDevices_Raw_Validate()
	})
}

func testData_ConfigQemu_PciDevices_Raw_Validate() (qemuTestTypeInvalid, qemuTestTypeValid) {
	invalid := qemuTestTypeInvalid{
		createUpdate: []qemuTestCaseInvalid{
			{name: `Raw errors.New(PciMediatedDevice_Error_Invalid)`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID3: QemuPci{
						Raw: &QemuPciRaw{
							ID:   new(PciID("0000:00:00")),
							MDev: new(PciMediatedDevice(","))}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID3: QemuPci{}}},
				err:     errors.New(PciMediatedDevice_Error_Invalid)},
			{name: `Raw errors.New(PciDeviceID_Error_Invalid)`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID3: QemuPci{
						Raw: &QemuPciRaw{
							ID:       new(PciID("0000:00:00")),
							DeviceID: new(PciDeviceID("a0%^#"))}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID3: QemuPci{}}},
				err:     errors.New(PciDeviceID_Error_Invalid)},
			{name: `Raw errors.New(PciSubDeviceID_Error_Invalid)`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID2: QemuPci{
						Raw: &QemuPciRaw{
							ID:          new(PciID("0000:00:00")),
							SubDeviceID: new(PciSubDeviceID("a0%^#"))}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID2: QemuPci{}}},
				err:     errors.New(PciSubDeviceID_Error_Invalid)},
			{name: `Raw errors.New(PciSubVendorID_Error_Invalid)`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID1: QemuPci{
						Raw: &QemuPciRaw{
							ID:          new(PciID("0000:00:00")),
							SubVendorID: new(PciSubVendorID("a0%^#"))}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID1: QemuPci{}}},
				err:     errors.New(PciSubVendorID_Error_Invalid)},
			{name: `Raw errors.New(PciVendorID_Error_Invalid)`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID0: QemuPci{
						Raw: &QemuPciRaw{
							ID:       new(PciID("0000:00:00")),
							VendorID: new(PciVendorID("a0%^#"))}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID0: QemuPci{}}},
				err:     errors.New(PciVendorID_Error_Invalid)}}}
	valid := qemuTestTypeValid{
		createUpdate: []qemuTestCaseValid{
			{name: `Mapping normal`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID15: QemuPci{
						Raw: &QemuPciRaw{
							ID: new(PciID("0000:00:00.1")),
						}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID0: QemuPci{}}}}},
		update: []qemuTestCaseValid{
			{name: `Raw`,
				input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID13: QemuPci{
						Raw: &QemuPciRaw{}}}}),
				current: &ConfigQemu{PciDevices: QemuPciDevices{
					QemuPciID13: QemuPci{
						Raw: &QemuPciRaw{
							ID: new(PciID("0000:00:00"))}}}}}}}
	return invalid, valid
}

func testData_ConfigQemu_PciDevices_Validate_ID() qemuTestTypeValidateFunc {
	return qemuTestTypeValidateFunc(func() (qemuTestTypeInvalid, qemuTestTypeValid) {
		invalid := qemuTestTypeInvalid{
			createUpdate: []qemuTestCaseInvalid{
				{name: `errors.New(QemuPciID_Error_Invalid)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{PciDevices: QemuPciDevices{
						20: QemuPci{}}}),
					current: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID4: QemuPci{}}},
					err: errors.New(QemuPciID_Error_Invalid)}}}
		return invalid, qemuTestTypeValid{}
	})
}

func Test_ConfigQemu_PciDevices_Validate(t *testing.T) {
	t.Parallel()
	testData_ConfigQemu_PciDevices_Validate().Test(t)
	testData_ConfigQemu_PciDevices_Validate_ID().Test(t)
}

func Test_QemuPciDevices_Validate(t *testing.T) {
	t.Parallel()
	validate := func(t *testing.T, config ConfigQemu, current *ConfigQemu, version Version, expectedErr error, valid bool) {
		t.Helper()
		var currentPciDevices QemuPciDevices
		if current != nil {
			currentPciDevices = current.PciDevices
		}
		err := config.PciDevices.Validate(currentPciDevices)
		if valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			if expectedErr != nil {
				require.Equal(t, expectedErr, err)
			}
		}
	}
	testData_ConfigQemu_PciDevices_Validate().Inject(t, validate)
	testData_ConfigQemu_PciDevices_Validate_ID().Inject(t, validate)
}

func Test_QemuPci_Validate(t *testing.T) {
	t.Parallel()
	validate := func(t *testing.T, config ConfigQemu, current *ConfigQemu, version Version, expectedErr error, valid bool) {
		t.Helper()
		var configQemuPci QemuPci
		var currentQemuPci *QemuPci
		for i, e := range config.PciDevices {
			configQemuPci = e
			if current != nil && len(current.PciDevices) > 0 {
				if v, ok := current.PciDevices[i]; ok {
					currentQemuPci = &v
				}
			}
		}
		err := configQemuPci.Validate(currentQemuPci)
		if valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			if expectedErr != nil {
				require.Equal(t, expectedErr, err)
			}
		}
	}
	testData_ConfigQemu_PciDevices_Validate().Inject(t, validate)
}

func Test_QemuPciMapping_Validate(t *testing.T) {
	t.Parallel()
	validate := func(t *testing.T, config ConfigQemu, current *ConfigQemu, version Version, expectedErr error, valid bool) {
		t.Helper()
		var err error
		for i, e := range config.PciDevices {
			var currentQemuPciMapping *QemuPciMapping
			if current != nil && current.PciDevices != nil {
				if v, ok := current.PciDevices[i]; ok {
					currentQemuPciMapping = v.Mapping
				}
			}
			err = e.Mapping.Validate(currentQemuPciMapping)
		}
		if valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			if expectedErr != nil {
				require.Equal(t, expectedErr, err)
			}
		}
	}
	testData_ConfigQemu_PciDevices_Mapping_Validate_func().Inject(t, validate)
}

func Test_QemuPciRaw_Validate(t *testing.T) {
	t.Parallel()
	validate := func(t *testing.T, config ConfigQemu, current *ConfigQemu, version Version, expectedErr error, valid bool) {
		t.Helper()
		var err error
		for i, e := range config.PciDevices {
			var currentQemuPciRaw *QemuPciRaw
			if current != nil && current.PciDevices != nil {
				if v, ok := current.PciDevices[i]; ok {
					currentQemuPciRaw = v.Raw
				}
			}
			err = e.Raw.Validate(currentQemuPciRaw)
		}
		if valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			if expectedErr != nil {
				require.Equal(t, expectedErr, err)
			}
		}
	}
	testData_ConfigQemu_PciDevices_Raw_Validate_func().Inject(t, validate)
}

func Test_QemuPciID_Validate(t *testing.T) {
	t.Parallel()
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

func Test_PciDeviceID_String(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

func Test_PciMediatedDevice_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  PciMediatedDevice
		output error
	}{
		{name: `Valid`,
			input: "vendor-643"},
		{name: `Invalid`,
			input:  "vendor,-643",
			output: errors.New(PciMediatedDevice_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_PciSubDeviceID_String(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

func Test_RawConfigQemu_GetPciDevices(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_PciDevices_Get().Inject(t,
		func(t *testing.T, raw RawConfigQemu, output *ConfigQemu, _ error) {
			require.Equal(t, output.PciDevices, raw.GetPciDevices())
		})
}

func Test_RawConfigQemu_PciDevices_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_PciDevices_Get().Test(t)
}
