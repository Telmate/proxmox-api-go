package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func testDataEfiDiskMapToAPI() qemuTestsAPI {
	return qemuTestsAPI{category: `EfiDisk`,
		create: []qemuTestCaseAPI{
			{name: `PreEnrolledKeys True`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					PreEnrolledKeys: util.Pointer(true)}},
				body: map[string]string{"efidisk0": "%3A1%2Cpre-enrolled-keys%3D1"}}, // :1,pre-enrolled-keys=1
			{name: `PreEnrolledKeys False`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					PreEnrolledKeys: util.Pointer(false)}},
				body: map[string]string{"efidisk0": "%3A1"}}, // :1
			{name: `Type 2M`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					Type: util.Pointer(EfiDiskType2M)}},
				body: map[string]string{"efidisk0": "%3A1%2Cefitype%3D2m"}}, // :1,efitype=2m
			{name: `Type 4M`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					Type: util.Pointer(EfiDiskType4M)}},
				body: map[string]string{"efidisk0": "%3A1%2Cefitype%3D4m"}}, // :1,efitype=4m
			{name: `minimal`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					Format:  util.Pointer(QemuDiskFormat_Raw),
					Storage: util.Pointer(StorageName("test"))}},
				body: map[string]string{"efidisk0": "test%3A1%2Cformat%3Draw"}}, // test:1,format=raw
			{name: `full`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					Format:          util.Pointer(QemuDiskFormat_Qcow2),
					PreEnrolledKeys: util.Pointer(true),
					Storage:         util.Pointer(StorageName("test")),
					Type:            util.Pointer(EfiDiskType4M)}},
				body: map[string]string{"efidisk0": "test%3A1%2Cformat%3Dqcow2%2Cpre-enrolled-keys%3D1%2Cefitype%3D4m"}}}, // test:1,format=qcow2,pre-enrolled-keys=1,efitype=4m
		createUpdate: []qemuTestCaseAPI{
			{name: `Delete no effect`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{Delete: true}}}},
		update: []qemuTestCaseAPI{
			{name: `Delete`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{Delete: true}},
				currentUpdate: configQemuUpdate{
					efiDisk: &EfiDisk{
						PreEnrolledKeys: util.Pointer(false),
						Storage:         util.Pointer(StorageName("local-zfs"))}},
				body: map[string]string{"delete": "efidisk0"}},
			{name: `PreEnrolledKeys False`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					PreEnrolledKeys: util.Pointer(false)}},
				currentUpdate: configQemuUpdate{
					efiDisk: &EfiDisk{
						PreEnrolledKeys: util.Pointer(true),
						Storage:         util.Pointer(StorageName("local-lvm"))}},
				body: map[string]string{"efidisk0": "local-lvm%3A1"}}, // local-lvm:1
			{name: `PreEnrolledKeys False same`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					PreEnrolledKeys: util.Pointer(false)}},
				currentUpdate: configQemuUpdate{
					efiDisk: &EfiDisk{
						PreEnrolledKeys: util.Pointer(false),
						Storage:         util.Pointer(StorageName("local-zfs"))}}},
			{name: `PreEnrolledKeys True`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					PreEnrolledKeys: util.Pointer(true)}},
				currentUpdate: configQemuUpdate{
					efiDisk: &EfiDisk{
						PreEnrolledKeys: util.Pointer(false),
						Storage:         util.Pointer(StorageName("local-zfs"))}},
				body: map[string]string{"efidisk0": "local-zfs%3A1%2Cpre-enrolled-keys%3D1"}}, // local-zfs:1,pre-enrolled-keys=1
			{name: `PreEnrolledKeys True same`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					PreEnrolledKeys: util.Pointer(true)}},
				currentUpdate: configQemuUpdate{
					efiDisk: &EfiDisk{
						PreEnrolledKeys: util.Pointer(true),
						Storage:         util.Pointer(StorageName("local-zfs"))}}},
			{name: `Type 2M`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					Type: util.Pointer(EfiDiskType2M)}},
				currentUpdate: configQemuUpdate{
					efiDisk: &EfiDisk{
						Type:    util.Pointer(EfiDiskType4M),
						Storage: util.Pointer(StorageName("local-zfs"))}},
				body: map[string]string{"efidisk0": "local-zfs%3A1%2Cefitype%3D2m"}}, // local-zfs:1,efitype=2m
			{name: `Type 2M same`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					Type: util.Pointer(EfiDiskType2M)}},
				currentUpdate: configQemuUpdate{
					efiDisk: &EfiDisk{
						Type:    util.Pointer(EfiDiskType2M),
						Storage: util.Pointer(StorageName("local-zfs"))}}},
			{name: `Type 2M unset`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					Type: util.Pointer(EfiDiskTypeUnset)}},
				currentUpdate: configQemuUpdate{
					efiDisk: &EfiDisk{
						Type:    util.Pointer(EfiDiskType2M),
						Storage: util.Pointer(StorageName("local-zfs"))}},
				body: map[string]string{"efidisk0": "local-zfs%3A1"}}, // local-zfs:1
			{name: `Type 4M`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					Type: util.Pointer(EfiDiskType4M)}},
				currentUpdate: configQemuUpdate{
					efiDisk: &EfiDisk{
						Type:    util.Pointer(EfiDiskType2M),
						Format:  util.Pointer(QemuDiskFormat_Qcow2),
						Storage: util.Pointer(StorageName("local-zfs"))}},
				body: map[string]string{"efidisk0": "local-zfs%3A1%2Cformat%3Dqcow2%2Cefitype%3D4m"}}, // local-zfs:1,format=qcow2,efitype=4m
			{name: `Type 4M same`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					Type: util.Pointer(EfiDiskType4M)}},
				currentUpdate: configQemuUpdate{
					efiDisk: &EfiDisk{
						Type:    util.Pointer(EfiDiskType4M),
						Storage: util.Pointer(StorageName("local-zfs"))}}},
			{name: `Type 4M unset`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					Type: util.Pointer(EfiDiskTypeUnset)}},
				currentUpdate: configQemuUpdate{
					efiDisk: &EfiDisk{
						Type:    util.Pointer(EfiDiskType4M),
						Format:  util.Pointer(QemuDiskFormat_Raw),
						Storage: util.Pointer(StorageName("local-zfs"))}},
				body: map[string]string{"efidisk0": "local-zfs%3A1%2Cformat%3Draw"}}, // local-zfs:1,format=raw
			{name: `Format no inherit`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					Storage: util.Pointer(StorageName("local-lvm")),
					Type:    util.Pointer(EfiDiskTypeUnset)}},
				currentUpdate: configQemuUpdate{
					efiDisk: &EfiDisk{
						Format:          util.Pointer(QemuDiskFormat_Qcow2),
						Type:            util.Pointer(EfiDiskType4M),
						PreEnrolledKeys: util.Pointer(true),
						Storage:         util.Pointer(StorageName("local-zfs"))}},
				body: map[string]string{"efidisk0": "local-lvm%3A1%2Cpre-enrolled-keys%3D1"}}, // local-lvm:1,pre-enrolled-keys=1
			{name: `Format change`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					Storage:         util.Pointer(StorageName("local-lvm")),
					Format:          util.Pointer(QemuDiskFormat_Raw),
					PreEnrolledKeys: util.Pointer(false)}},
				currentUpdate: configQemuUpdate{
					efiDisk: &EfiDisk{
						Format:          util.Pointer(QemuDiskFormat_Qcow2),
						Type:            util.Pointer(EfiDiskType2M),
						PreEnrolledKeys: util.Pointer(true),
						Storage:         util.Pointer(StorageName("local-zfs"))}},
				body: map[string]string{"efidisk0": "local-lvm%3A1%2Cformat%3Draw%2Cefitype%3D2m"}}, // local-lvm:1,format=raw,efitype=2m
			{name: `Format, Storage change`,
				config: &ConfigQemu{EfiDisk: &EfiDisk{
					Format:  util.Pointer(QemuDiskFormat_Vmdk),
					Storage: util.Pointer(StorageName("local-zfs"))}},
				currentUpdate: configQemuUpdate{
					efiDisk: &EfiDisk{
						Format:  util.Pointer(QemuDiskFormat_Qcow2),
						Storage: util.Pointer(StorageName("local-lvm"))}}},
		}}
}

func testDataEfiDiskGet() struct {
	category string
	tests    []qemuTestCaseGet
} {
	return struct {
		category string
		tests    []qemuTestCaseGet
	}{
		category: `EfiDisk`,
		tests: []qemuTestCaseGet{
			{name: `all`,
				input: map[string]any{"efidisk0": "test:104/vm-104-disk-0.qcow2,efitype=2m,size=4M,pre-enrolled-keys=1,ms-cert=2011"},
				output: testQemuBaseConfig_get(ConfigQemu{EfiDisk: &EfiDisk{
					Format:          util.Pointer(QemuDiskFormat_Qcow2),
					MsCertType:      util.Pointer(EfiMsCertType2011),
					PreEnrolledKeys: util.Pointer(true),
					Size:            4096,
					Storage:         util.Pointer(StorageName("test")),
					Type:            util.Pointer(EfiDiskType2M)}})},
			{name: `minimal`,
				input: map[string]any{"efidisk0": "local-lvm:vm-104-disk-0,size=1M"},
				output: testQemuBaseConfig_get(ConfigQemu{EfiDisk: &EfiDisk{
					Format:          util.Pointer(QemuDiskFormat_Raw),
					PreEnrolledKeys: util.Pointer(false),
					Size:            1024,
					Storage:         util.Pointer(StorageName("local-lvm"))}})},
			{name: `Type 2M`,
				input: map[string]any{"efidisk0": "local-zfs:vm-1020-disk-0,size=1M,efitype=2m"},
				output: testQemuBaseConfig_get(ConfigQemu{EfiDisk: &EfiDisk{
					Format:          util.Pointer(QemuDiskFormat_Raw),
					PreEnrolledKeys: util.Pointer(false),
					Size:            1024,
					Storage:         util.Pointer(StorageName("local-zfs")),
					Type:            util.Pointer(EfiDiskType2M)}})},
			{name: `Type 4M`,
				input: map[string]any{"efidisk0": "local-zfs:vm-1020-disk-0,size=1M,efitype=4m"},
				output: testQemuBaseConfig_get(ConfigQemu{EfiDisk: &EfiDisk{
					Format:          util.Pointer(QemuDiskFormat_Raw),
					PreEnrolledKeys: util.Pointer(false),
					Size:            1024,
					Storage:         util.Pointer(StorageName("local-zfs")),
					Type:            util.Pointer(EfiDiskType4M)}})},
			{name: `MsCertType 2011`,
				input: map[string]any{"efidisk0": "local-zfs:vm-1020-disk-0,size=1M,ms-cert=2011"},
				output: testQemuBaseConfig_get(ConfigQemu{EfiDisk: &EfiDisk{
					Format:          util.Pointer(QemuDiskFormat_Raw),
					MsCertType:      util.Pointer(EfiMsCertType2011),
					PreEnrolledKeys: util.Pointer(false),
					Size:            1024,
					Storage:         util.Pointer(StorageName("local-zfs"))}})},
			{name: `MsCertType 2023`,
				input: map[string]any{"efidisk0": "local-zfs:vm-1020-disk-0,size=1M,ms-cert=2023"},
				output: testQemuBaseConfig_get(ConfigQemu{EfiDisk: &EfiDisk{
					Format:          util.Pointer(QemuDiskFormat_Raw),
					MsCertType:      util.Pointer(EfiMsCertType2023),
					PreEnrolledKeys: util.Pointer(false),
					Size:            1024,
					Storage:         util.Pointer(StorageName("local-zfs"))}})},
			{name: `PreEnrolledKeys True`,
				input: map[string]any{"efidisk0": "local-zfs:vm-1020-disk-0,size=1M,pre-enrolled-keys=1"},
				output: testQemuBaseConfig_get(ConfigQemu{EfiDisk: &EfiDisk{
					Format:          util.Pointer(QemuDiskFormat_Raw),
					PreEnrolledKeys: util.Pointer(true),
					Size:            1024,
					Storage:         util.Pointer(StorageName("local-zfs"))}})},
			{name: `PreEnrolledKeys False`,
				input: map[string]any{"efidisk0": "local-zfs:vm-1020-disk-0,size=512K,pre-enrolled-keys=0"},
				output: testQemuBaseConfig_get(ConfigQemu{EfiDisk: &EfiDisk{
					Format:          util.Pointer(QemuDiskFormat_Raw),
					PreEnrolledKeys: util.Pointer(false),
					Size:            512,
					Storage:         util.Pointer(StorageName("local-zfs"))}})}}}
}

func testDataEfiDiskValidate() struct {
	category string
	valid    qemuTestTypeValidate
	invalid  qemuTestTypeValidate
} {
	return struct {
		category string
		valid    qemuTestTypeValidate
		invalid  qemuTestTypeValidate
	}{
		category: `EfiDisk`,
		valid: qemuTestTypeValidate{
			createUpdate: []qemuTestCaseValidate{
				{name: `full`,
					input: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Format:          util.Pointer(QemuDiskFormat("raw")),
						PreEnrolledKeys: util.Pointer(true),
						Storage:         util.Pointer(StorageName("local-zfs")),
						Type:            util.Pointer(EfiDiskType("4m"))}}),
					current: &ConfigQemu{}},
				{name: `minimal`,
					input: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Storage: util.Pointer(StorageName("local-zfs"))}}),
					current: &ConfigQemu{}},
				{name: `delete full`,
					input: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Delete:          true,
						Format:          util.Pointer(QemuDiskFormat("raw")),
						PreEnrolledKeys: util.Pointer(true),
						Storage:         util.Pointer(StorageName("local-zfs")),
						Type:            util.Pointer(EfiDiskType("4m"))}}),
					current: &ConfigQemu{}},
				{name: `delete minimal`,
					input: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Delete: true}}),
					current: &ConfigQemu{}},
				{name: `unset`,
					input: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Type:    util.Pointer(EfiDiskType("")),
						Storage: util.Pointer(StorageName("local-zfs"))}}),
					current: &ConfigQemu{}}},
			update: []qemuTestCaseValidate{
				{name: `delete existing`,
					input: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{Delete: true}}),
					current: &ConfigQemu{EfiDisk: &EfiDisk{
						Storage: util.Pointer(StorageName("local-zfs"))}}},
				{name: `change all`,
					input: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Format:          util.Pointer(QemuDiskFormat("qcow2")),
						PreEnrolledKeys: util.Pointer(true),
						Storage:         util.Pointer(StorageName("local-lvm")),
						Type:            util.Pointer(EfiDiskType("4m"))}}),
					current: &ConfigQemu{EfiDisk: &EfiDisk{
						Format:  util.Pointer(QemuDiskFormat("raw")),
						Storage: util.Pointer(StorageName("local-zfs")),
						Type:    util.Pointer(EfiDiskType("2m"))}}}}},
		invalid: qemuTestTypeValidate{
			create: []qemuTestCaseValidate{
				{name: `errors.New(EFiDisk_Error_StorageRequired)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{}}),
					err:   errors.New(EFiDisk_Error_StorageRequired)},
				{name: `QemuDiskFormat("").Error()`,
					input: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Storage: util.Pointer(StorageName("local-zfs")),
						Format:  util.Pointer(QemuDiskFormat("invalid")),
					}}),
					err: QemuDiskFormat("").Error()},
				{name: `errors.New(EfiDiskType_Error)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Storage: util.Pointer(StorageName("local-zfs")),
						Format:  util.Pointer(QemuDiskFormat("raw")),
						Type:    util.Pointer(EfiDiskType("invalid")),
					}}),
					err: errors.New(EfiDiskType_Error)}}},
	}
}

func Test_EfiDisk_markChanges(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   EfiDisk
		current *EfiDisk
		output  *qemuDiskMove
	}{
		{name: `explicit delete`,
			input:   EfiDisk{Delete: true},
			current: &EfiDisk{Storage: util.Pointer(StorageName("local-zfs"))}},
		{name: `change Storage`,
			input: EfiDisk{Storage: util.Pointer(StorageName("local-lvm"))},
			current: &EfiDisk{
				Format:  util.Pointer(QemuDiskFormat_Raw),
				Storage: util.Pointer(StorageName("local-zfs"))},
			output: &qemuDiskMove{Storage: "local-lvm", Id: QemuDiskId(qemuApiKeyEfiDisk)}},
		{name: `change Format only`,
			input: EfiDisk{Format: util.Pointer(QemuDiskFormat_Qcow2)},
			current: &EfiDisk{
				Format:  util.Pointer(QemuDiskFormat_Raw),
				Storage: util.Pointer(StorageName("local-lvm"))},
			output: &qemuDiskMove{
				Format:  util.Pointer(QemuDiskFormat_Qcow2),
				Storage: "local-lvm",
				Id:      QemuDiskId(qemuApiKeyEfiDisk)}},
		{name: `change Storage and Format`,
			input: EfiDisk{
				Format:  util.Pointer(QemuDiskFormat_Qcow2),
				Storage: util.Pointer(StorageName("local-zfs"))},
			current: &EfiDisk{
				Format:  util.Pointer(QemuDiskFormat_Raw),
				Storage: util.Pointer(StorageName("local-lvm"))},
			output: &qemuDiskMove{
				Format:  util.Pointer(QemuDiskFormat_Qcow2),
				Storage: "local-zfs",
				Id:      QemuDiskId(qemuApiKeyEfiDisk)}},
		{name: `change Storage and Format, replaced`,
			input: EfiDisk{
				Type:    util.Pointer(EfiDiskType4M),
				Format:  util.Pointer(QemuDiskFormat_Qcow2),
				Storage: util.Pointer(StorageName("local-zfs"))},
			current: &EfiDisk{
				Type:    util.Pointer(EfiDiskType2M),
				Format:  util.Pointer(QemuDiskFormat_Raw),
				Storage: util.Pointer(StorageName("local-lvm"))}},
		{name: `no change`,
			input: EfiDisk{
				Format:  util.Pointer(QemuDiskFormat_Raw),
				Storage: util.Pointer(StorageName("local-lvm"))},
			current: &EfiDisk{
				Format:  util.Pointer(QemuDiskFormat_Raw),
				Storage: util.Pointer(StorageName("local-lvm"))},
			output: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.markChangesUnsafe(test.current))
		})
	}
}

func Test_EfiDisk_Validate(t *testing.T) {
	t.Parallel()
	tests := testDataEfiDiskValidate()
	validateCreate := func(prefix string, v qemuTestCaseValidate) {
		t.Run(prefix+"/Create/"+v.name, func(t *testing.T) {
			require.Equal(t, v.err, v.input.EfiDisk.Validate(nil))
		})
	}
	validateUpdate := func(prefix string, v qemuTestCaseValidate) {
		t.Run(prefix+"/Update/"+v.name, func(t *testing.T) {
			if v.current == nil {
				require.Equal(t, v.err, v.input.EfiDisk.Validate(nil))
			} else {
				require.Equal(t, v.err, v.input.EfiDisk.Validate(v.current.EfiDisk))
			}
		})
	}
	for _, test := range append(tests.valid.create, tests.valid.createUpdate...) {
		validateCreate("/Valid", test)
	}
	for _, test := range append(tests.valid.update, tests.valid.createUpdate...) {
		validateUpdate("/Valid", test)
	}
	for _, test := range append(tests.invalid.create, tests.invalid.createUpdate...) {
		validateCreate("/Invalid", test)
	}
	for _, test := range append(tests.invalid.update, tests.invalid.createUpdate...) {
		validateUpdate("/Invalid", test)
	}
}
