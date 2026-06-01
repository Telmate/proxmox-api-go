package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func testData_ConfigQemu_EfiDisk() qemuTests {
	return qemuTests{
		Invalid: qemuInvalid{
			CreateUpdate: []qemuInvalidUpdate{
				{name: `errors.New(EFiDisk_Error_StorageRequired)`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{}}),
					output: map[string]string{"efidisk0": "%3A1"}, // ":1"
					err:    errors.New(EFiDisk_Error_StorageRequired)},
				{name: `QemuDiskFormat("").Error()`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Format:  new(QemuDiskFormat("invalid")),
						Storage: new(StorageName("local-zfs"))}}),
					output: map[string]string{"efidisk0": "local-zfs%3A1%2Cformat%3Dinvalid"}, // "local-zfs:1,format=invalid"
					err:    QemuDiskFormat("").Error()},
				{name: `errors.New(EfiDiskType_Error)`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Format:  new(QemuDiskFormat_Raw),
						Storage: new(StorageName("local-zfs")),
						Type:    new(EfiDiskType("invalid"))}}),
					output: map[string]string{"efidisk0": "local-zfs%3A1%2Cformat%3Draw%2Cefitype%3Dinvalid"}, // "local-zfs:1,format=raw,efitype=invalid"
					err:    errors.New(EfiDiskType_Error)},
				{name: `PreEnrolledKeys True`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						PreEnrolledKeys: new(true)}}),
					output: map[string]string{"efidisk0": "%3A1%2Cpre-enrolled-keys%3D1"}}, // :1,pre-enrolled-keys=1
				{name: `PreEnrolledKeys False`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						PreEnrolledKeys: new(false)}}),
					output: map[string]string{"efidisk0": "%3A1"}}, // ":1"
				{name: `Type 2M`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Type: new(EfiDiskType2M)}}),
					output: map[string]string{"efidisk0": "%3A1%2Cefitype%3D2m"}}, // ":1,efitype=2m"
				{name: `Type 4M`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Type: new(EfiDiskType4M)}}),
					output: map[string]string{"efidisk0": "%3A1%2Cefitype%3D4m"}}, // ":1,efitype=4m"
			},
		},
		Valid: qemuValid{
			CreateUpdate: []qemuValidUpdate{
				{name: `full`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Format:          new(QemuDiskFormat_Raw),
						PreEnrolledKeys: new(true),
						Storage:         new(StorageName("local-zfs")),
						Type:            new(EfiDiskType4M)}}),
					output: map[string]string{"efidisk0": "local-zfs%3A1%2Cformat%3Draw%2Cpre-enrolled-keys%3D1%2Cefitype%3D4m"}}, // "local-zfs:1,format=raw,pre-enrolled-keys=1,efitype=4m"
				{name: `minimal`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Storage: new(StorageName("local-zfs"))}}),
					output: map[string]string{"efidisk0": "local-zfs%3A1"}}, // "local-zfs:1"
				{name: `delete full, no effect`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Delete:          true,
						Format:          new(QemuDiskFormat_Raw),
						PreEnrolledKeys: new(true),
						Storage:         new(StorageName("local-zfs")),
						Type:            new(EfiDiskType4M)}})},
				{name: `delete minimal, no effect`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{Delete: true}})},
				{name: `unset`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Type:    new(EfiDiskTypeUnset),
						Storage: new(StorageName("local-zfs"))}}),
					output: map[string]string{"efidisk0": "local-zfs%3A1"}}, // "local-zfs:1"
			},
			Update: []qemuValidUpdate{
				{name: `delete existing`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{Delete: true}}),
					current: ConfigQemu{EfiDisk: &EfiDisk{
						Storage: new(StorageName("local-zfs"))}},
					current2: configQemuUpdate{efiDisk: &EfiDisk{
						Storage: new(StorageName("local-zfs"))}},
					output: map[string]string{"delete": "efidisk0"}},
				{name: `change all`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Format:          new(QemuDiskFormat_Qcow2),
						PreEnrolledKeys: new(true),
						Storage:         new(StorageName("local-lvm")),
						Type:            new(EfiDiskType4M)}}),
					current: ConfigQemu{
						EfiDisk: &EfiDisk{
							Format:          new(QemuDiskFormat_Raw),
							PreEnrolledKeys: new(false),
							Storage:         new(StorageName("local-zfs")),
							Type:            new(EfiDiskType2M)}},
					current2: configQemuUpdate{
						efiDisk: &EfiDisk{
							Format:          new(QemuDiskFormat_Raw),
							PreEnrolledKeys: new(false),
							Storage:         new(StorageName("local-zfs")),
							Type:            new(EfiDiskType2M)}},
					output: map[string]string{"efidisk0": "local-lvm%3A1%2Cformat%3Dqcow2%2Cpre-enrolled-keys%3D1%2Cefitype%3D4m"}}, // "local-lvm:1,format=qcow2,pre-enrolled-keys=1,efitype=4m"
				{name: `PreEnrolledKeys False`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						PreEnrolledKeys: new(false)}}),
					current: ConfigQemu{
						EfiDisk: &EfiDisk{
							PreEnrolledKeys: new(false),
							Storage:         new(StorageName("local-lvm"))}},
					current2: configQemuUpdate{
						efiDisk: &EfiDisk{
							PreEnrolledKeys: new(true),
							Storage:         new(StorageName("local-lvm"))}},
					output: map[string]string{"efidisk0": "local-lvm%3A1"}}, // "local-lvm:1"
				{name: `PreEnrolledKeys False same`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						PreEnrolledKeys: new(false)}}),
					current: ConfigQemu{
						EfiDisk: &EfiDisk{
							PreEnrolledKeys: new(false),
							Storage:         new(StorageName("local-zfs"))}},
					current2: configQemuUpdate{
						efiDisk: &EfiDisk{
							PreEnrolledKeys: new(false),
							Storage:         new(StorageName("local-zfs"))}}},
				{name: `PreEnrolledKeys True`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						PreEnrolledKeys: new(true)}}),
					current: ConfigQemu{
						EfiDisk: &EfiDisk{
							PreEnrolledKeys: new(false),
							Storage:         new(StorageName("local-zfs"))}},
					current2: configQemuUpdate{
						efiDisk: &EfiDisk{
							PreEnrolledKeys: new(false),
							Storage:         new(StorageName("local-zfs"))}},
					output: map[string]string{"efidisk0": "local-zfs%3A1%2Cpre-enrolled-keys%3D1"}}, // "local-zfs:1,pre-enrolled-keys=1"
				{name: `PreEnrolledKeys True same`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						PreEnrolledKeys: new(true)}}),
					current: ConfigQemu{
						EfiDisk: &EfiDisk{
							PreEnrolledKeys: new(true),
							Storage:         new(StorageName("local-zfs"))}},
					current2: configQemuUpdate{
						efiDisk: &EfiDisk{
							PreEnrolledKeys: new(true),
							Storage:         new(StorageName("local-zfs"))}}},
				{name: `Type 2M`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Type: new(EfiDiskType2M)}}),
					current: ConfigQemu{
						EfiDisk: &EfiDisk{
							Type:    new(EfiDiskType4M),
							Storage: new(StorageName("local-zfs"))}},
					current2: configQemuUpdate{
						efiDisk: &EfiDisk{
							Type:    new(EfiDiskType4M),
							Storage: new(StorageName("local-zfs"))}},
					output: map[string]string{"efidisk0": "local-zfs%3A1%2Cefitype%3D2m"}}, // "local-zfs:1,efitype=2m"
				{name: `Type 2M same`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Type: new(EfiDiskType2M)}}),
					current: ConfigQemu{
						EfiDisk: &EfiDisk{
							Type:    new(EfiDiskType2M),
							Storage: new(StorageName("local-zfs"))}},
					current2: configQemuUpdate{
						efiDisk: &EfiDisk{
							Type:    new(EfiDiskType2M),
							Storage: new(StorageName("local-zfs"))}}},
				{name: `Type 2M unset`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Type: new(EfiDiskTypeUnset)}}),
					current: ConfigQemu{
						EfiDisk: &EfiDisk{
							Type:    new(EfiDiskType2M),
							Storage: new(StorageName("local-zfs"))}},
					current2: configQemuUpdate{
						efiDisk: &EfiDisk{
							Type:    new(EfiDiskType2M),
							Storage: new(StorageName("local-zfs"))}},
					output: map[string]string{"efidisk0": "local-zfs%3A1"}}, // "local-zfs:1"
				{name: `Type 4M`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Type: new(EfiDiskType4M)}}),
					current: ConfigQemu{
						EfiDisk: &EfiDisk{
							Type:    new(EfiDiskType2M),
							Format:  new(QemuDiskFormat_Qcow2),
							Storage: new(StorageName("local-zfs"))}},
					current2: configQemuUpdate{
						efiDisk: &EfiDisk{
							Type:    new(EfiDiskType2M),
							Format:  new(QemuDiskFormat_Qcow2),
							Storage: new(StorageName("local-zfs"))}},
					output: map[string]string{"efidisk0": "local-zfs%3A1%2Cformat%3Dqcow2%2Cefitype%3D4m"}}, // "local-zfs:1,format=qcow2,efitype=4m"
				{name: `Type 4M same`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Type: new(EfiDiskType4M)}}),
					current: ConfigQemu{
						EfiDisk: &EfiDisk{
							Type:    new(EfiDiskType4M),
							Storage: new(StorageName("local-zfs"))}},
					current2: configQemuUpdate{
						efiDisk: &EfiDisk{
							Type:    new(EfiDiskType4M),
							Storage: new(StorageName("local-zfs"))}}},
				{name: `Type 4M unset`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Type: new(EfiDiskTypeUnset)}}),
					current: ConfigQemu{
						EfiDisk: &EfiDisk{
							Type:    new(EfiDiskType4M),
							Format:  new(QemuDiskFormat_Raw),
							Storage: new(StorageName("local-zfs"))}},
					current2: configQemuUpdate{
						efiDisk: &EfiDisk{
							Type:    new(EfiDiskType4M),
							Format:  new(QemuDiskFormat_Raw),
							Storage: new(StorageName("local-zfs"))}},
					output: map[string]string{"efidisk0": "local-zfs%3A1%2Cformat%3Draw"}}, // "local-zfs:1,format=raw"
				{name: `Format no inherit`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Storage: new(StorageName("local-lvm")),
						Type:    new(EfiDiskTypeUnset)}}),
					current: ConfigQemu{
						EfiDisk: &EfiDisk{
							Format:          new(QemuDiskFormat_Qcow2),
							Type:            new(EfiDiskType4M),
							PreEnrolledKeys: new(true),
							Storage:         new(StorageName("local-zfs"))}},
					current2: configQemuUpdate{
						efiDisk: &EfiDisk{
							Format:          new(QemuDiskFormat_Qcow2),
							Type:            new(EfiDiskType4M),
							PreEnrolledKeys: new(true),
							Storage:         new(StorageName("local-zfs"))}},
					output: map[string]string{"efidisk0": "local-lvm%3A1%2Cpre-enrolled-keys%3D1"}}, // "local-lvm:1,pre-enrolled-keys=1"
				{name: `Format change`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Storage:         new(StorageName("local-lvm")),
						Format:          new(QemuDiskFormat_Raw),
						PreEnrolledKeys: new(false)}}),
					current: ConfigQemu{
						EfiDisk: &EfiDisk{
							Format:          new(QemuDiskFormat_Qcow2),
							Type:            new(EfiDiskType2M),
							PreEnrolledKeys: new(true),
							Storage:         new(StorageName("local-zfs"))}},
					current2: configQemuUpdate{
						efiDisk: &EfiDisk{
							Format:          new(QemuDiskFormat_Qcow2),
							Type:            new(EfiDiskType2M),
							PreEnrolledKeys: new(true),
							Storage:         new(StorageName("local-zfs"))}},
					output: map[string]string{"efidisk0": "local-lvm%3A1%2Cformat%3Draw%2Cefitype%3D2m"}}, // "local-lvm:1,format=raw,efitype=2m"
				{name: `Format, Storage change`,
					config: testQemuBaseConfig_Validate(ConfigQemu{EfiDisk: &EfiDisk{
						Format:  new(QemuDiskFormat_Vmdk),
						Storage: new(StorageName("local-zfs"))}}),
					current: ConfigQemu{
						EfiDisk: &EfiDisk{
							Format:  new(QemuDiskFormat_Qcow2),
							Storage: new(StorageName("local-lvm"))}},
					current2: configQemuUpdate{
						efiDisk: &EfiDisk{
							Format:  new(QemuDiskFormat_Qcow2),
							Storage: new(StorageName("local-lvm"))}}},
			},
		},
	}
}

func Test_ConfigQemu_EfiDisk_set(t *testing.T) {
	t.Parallel()
	qemuTestHelper(t, testData_ConfigQemu_EfiDisk)
}

func Test_ConfigQemu_EfiDisk_Validate(t *testing.T) {
	t.Parallel()
	test := func(t *testing.T, config ConfigQemu, current *ConfigQemu, version Version, output map[string]string, expectedErr error, valid bool) {
		var currentEfiDisk *EfiDisk
		if current != nil {
			currentEfiDisk = current.EfiDisk
		}
		err := config.EfiDisk.Validate(currentEfiDisk)
		if valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			if expectedErr != nil {
				require.Equal(t, expectedErr, err)
			}
		}
	}
	qemuTestInjected(t, testData_ConfigQemu_EfiDisk, test)
}

func testDataEfiDiskGet() []qemuTestCaseGet {
	return []qemuTestCaseGet{
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
				Storage:         util.Pointer(StorageName("local-zfs"))}})}}
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
