package proxmox

import (
	"errors"
	"slices"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_LxcBootMount_markMountChanges_Unsafe(t *testing.T) {
	tests := []struct {
		name    string
		input   LxcBootMount
		current *LxcBootMount
		output  lxcUpdateChanges
	}{
		{name: `resize`,
			input: LxcBootMount{
				SizeInKibibytes: util.Pointer(LxcMountSize(1051648))},
			current: &LxcBootMount{
				SizeInKibibytes: util.Pointer(LxcMountSize(1048576))},
			output: lxcUpdateChanges{
				resize: []lxcMountResize{{
					sizeInKibibytes: LxcMountSize(1051648),
					id:              "rootfs"}}}},
		{name: `move`,
			input: LxcBootMount{
				Storage: util.Pointer("local-zfs")},
			current: &LxcBootMount{
				Storage: util.Pointer("local-ext")},
			output: lxcUpdateChanges{
				move: []lxcMountMove{{
					storage: "local-zfs",
					id:      "rootfs"}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.markMountChanges_Unsafe(test.current))
		})
	}
}

func Test_LxcBindMount_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   LxcBindMount
		current *LxcBindMount
		output  error
	}{
		{name: `valid update HostPath`,
			input: LxcBindMount{
				HostPath: util.Pointer(LxcHostPath("/test/new"))},
			current: &LxcBindMount{
				HostPath:  util.Pointer(LxcHostPath("/test/old")),
				GuestPath: util.Pointer(LxcMountPath("/test/guest"))}},
		{name: `valid update LxcMountPath`,
			input: LxcBindMount{
				GuestPath: util.Pointer(LxcMountPath("/test/new"))},
			current: &LxcBindMount{
				HostPath:  util.Pointer(LxcHostPath("/test/old")),
				GuestPath: util.Pointer(LxcMountPath("/test/guest"))}},
		{name: `invalid create`,
			input:  LxcBindMount{},
			output: errors.New(LxcBindMountErrorHostPathRequired)},
		{name: `invalid update`,
			input: LxcBindMount{
				HostPath: util.Pointer(LxcHostPath("./test"))},
			current: &LxcBindMount{
				HostPath:  util.Pointer(LxcHostPath("/test/old")),
				GuestPath: util.Pointer(LxcMountPath("/test/guest"))},
			output: errors.New(LxcHostPathErrorRelative)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate(test.current))
		})
	}
}

func Test_LxcDataMount_Validate(t *testing.T) {
	tests := []struct {
		name       string
		input      LxcDataMount
		current    *LxcDataMount
		privileged bool
		output     error
	}{
		{name: `invalid create`,
			input:  LxcDataMount{},
			output: errors.New(LxcDataMountErrorPathRequired)},
		{name: `invalid update`,
			input: LxcDataMount{
				SizeInKibibytes: util.Pointer(LxcMountSize(131070))},
			current: &LxcDataMount{
				Storage:         util.Pointer("test"),
				SizeInKibibytes: util.Pointer(LxcMountSize(131072))},
			output: errors.New(LxcMountSizeErrorMinimum)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate(test.current, test.privileged))
		})
	}
}

func Test_LxcHostPath_String(t *testing.T) {
	require.Equal(t, string("/mnt/test"), LxcHostPath("/mnt/test").String())
}

func Test_LxcMounts_Validate(t *testing.T) {
	tests := []struct {
		name       string
		input      LxcMounts
		current    LxcMounts
		privileged bool
		output     error
	}{
		{name: `valid create during update`,
			input: LxcMounts{
				LxcMountID100: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  util.Pointer(LxcHostPath("/host/path")),
						GuestPath: util.Pointer(LxcMountPath("/guest/path"))}}},
			current: LxcMounts{
				LxcMountID200: LxcMount{
					BindMount: &LxcBindMount{}}}},
		{name: `valid create`,
			input: LxcMounts{
				LxcMountID10: LxcMount{
					DataMount: &LxcDataMount{
						Path:            util.Pointer(LxcMountPath("/opt/test")),
						SizeInKibibytes: util.Pointer(LxcMountSize(68343802)),
						Storage:         util.Pointer("test")}}}},
		{name: `valid update`,
			input: LxcMounts{
				LxcMountID113: LxcMount{
					DataMount: &LxcDataMount{}}},
			current: LxcMounts{
				LxcMountID113: LxcMount{
					DataMount: &LxcDataMount{}}}},
		{name: `invalid create during update`,
			input: LxcMounts{
				LxcMountID120: LxcMount{
					BindMount: &LxcBindMount{},
					DataMount: &LxcDataMount{}}},
			current: LxcMounts{
				LxcMountID210: LxcMount{
					BindMount: &LxcBindMount{}}},
			output: errors.New(LxcMountErrorMutuallyExclusive)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate(test.current, test.privileged))
		})
	}
}

func Test_LxcMount_Validate(t *testing.T) {
	tests := []struct {
		name       string
		input      LxcMount
		current    *LxcMount
		privileged bool
		output     error
	}{
		{name: `valid create nothing`,
			input: LxcMount{}},
		{name: `valid update nothing`,
			input:   LxcMount{},
			current: &LxcMount{DataMount: &LxcDataMount{}}},
		{name: `valid detach`,
			input: LxcMount{
				Detach: true}},
		{name: `valid create`,
			input: LxcMount{
				DataMount: &LxcDataMount{
					ACL:             util.Pointer(TriBoolFalse),
					Backup:          util.Pointer(true),
					SizeInKibibytes: util.Pointer(LxcMountSize(53763756)),
					Storage:         util.Pointer("test"),
					Path:            util.Pointer(LxcMountPath("/mnt/opt"))}}},
		{name: `valid update`,
			privileged: true,
			input: LxcMount{
				DataMount: &LxcDataMount{
					ACL:       util.Pointer(TriBoolFalse),
					Replicate: util.Pointer(false),
					Storage:   util.Pointer("test"),
					Quota:     util.Pointer(false),
					ReadOnly:  util.Pointer(false)}},
			current: &LxcMount{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate(test.current, test.privileged))
		})
	}
}

func Test_LxcMountPath_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  LxcMountPath
		output error
	}{
		{name: `valid`,
			input: "/mnt/test"},
		{name: `invalid empty`,
			input:  "",
			output: errors.New(LxcMountPathErrorInvalid)},
		{name: `invalid relative`,
			input:  "./test",
			output: errors.New(LxcMountPathErrorRelative)},
		{name: `invalid contains ,`,
			input:  "/mnt/path/,/test",
			output: errors.New(LxcMountPathErrorInvalidCharacter)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_LxcMounts_markMountChanges(t *testing.T) {
	baseMount := func() LxcMount {
		return LxcMount{
			DataMount: &LxcDataMount{
				SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
				Storage:         util.Pointer("local-ext4")}}
	}
	tests := []struct {
		name    string
		input   LxcMounts
		current LxcMounts
		output  lxcUpdateChanges
	}{
		{name: `create BindMount current empty`,
			input: LxcMounts{
				LxcMountID100: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  util.Pointer(LxcHostPath("/host/path")),
						GuestPath: util.Pointer(LxcMountPath("/guest/path"))}}},
			current: LxcMounts{}},
		{name: `create BindMount current not empty`,
			input: LxcMounts{
				LxcMountID100: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  util.Pointer(LxcHostPath("/host/path")),
						GuestPath: util.Pointer(LxcMountPath("/guest/path"))}}},
			current: LxcMounts{
				LxcMountID100: LxcMount{}}},
		{name: `create DataMount current empty`,
			input: LxcMounts{
				LxcMountID101: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
						Storage:         util.Pointer("local-ext4")}}},
			current: LxcMounts{}},
		{name: `create DataMount current not empty`,
			input: LxcMounts{
				LxcMountID101: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
						Storage:         util.Pointer("local-ext4")}}},
			current: LxcMounts{LxcMountID101: LxcMount{}}},
		{name: `detach BindMount`,
			input: LxcMounts{
				LxcMountID100: LxcMount{
					Detach: true}},
			current: LxcMounts{
				LxcMountID100: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  util.Pointer(LxcHostPath("/host/path")),
						GuestPath: util.Pointer(LxcMountPath("/guest/path"))}}},
			output: lxcUpdateChanges{
				offState: true}},
		{name: `detach DataMount`,
			input: LxcMounts{
				LxcMountID101: LxcMount{
					Detach: true}},
			current: LxcMounts{
				LxcMountID101: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
						Storage:         util.Pointer("local-ext4")}}},
			output: lxcUpdateChanges{
				offState: true}},
		{name: `recreate BindMount due to host path`,
			input: LxcMounts{
				LxcMountID100: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  util.Pointer(LxcHostPath("/new/host/path")),
						GuestPath: util.Pointer(LxcMountPath("/guest/path"))}}},
			current: LxcMounts{
				LxcMountID100: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  util.Pointer(LxcHostPath("/host/path")),
						GuestPath: util.Pointer(LxcMountPath("/guest/path"))}}},
			output: lxcUpdateChanges{
				offState: true}},
		{name: `recreate BindMount due to guest path`,
			input: LxcMounts{
				LxcMountID100: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  util.Pointer(LxcHostPath("/host/path")),
						GuestPath: util.Pointer(LxcMountPath("/new/guest/path"))}}},
			current: LxcMounts{
				LxcMountID100: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  util.Pointer(LxcHostPath("/host/path")),
						GuestPath: util.Pointer(LxcMountPath("/guest/path"))}}},
			output: lxcUpdateChanges{
				offState: true}},
		{name: `recreate DataMount due to resize`,
			input: LxcMounts{
				LxcMountID200: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(1048575)),
						Storage:         util.Pointer("local-zfs")}}},
			current: LxcMounts{
				LxcMountID200: baseMount()},
			output: lxcUpdateChanges{
				offState: true}},

		{name: `replace BindMount with DataMount`,
			input: LxcMounts{
				LxcMountID105: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(1051648)),
						Storage:         util.Pointer("local-zfs")}}},
			current: LxcMounts{
				LxcMountID105: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  util.Pointer(LxcHostPath("/host/path")),
						GuestPath: util.Pointer(LxcMountPath("/guest/path"))}}},
			output: lxcUpdateChanges{
				offState: true}},
		{name: `replace DataMount with BindMount`,
			input: LxcMounts{
				LxcMountID105: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  util.Pointer(LxcHostPath("/host/path")),
						GuestPath: util.Pointer(LxcMountPath("/guest/path"))}}},
			current: LxcMounts{
				LxcMountID105: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(1051648)),
						Storage:         util.Pointer("local-zfs")}}},
			output: lxcUpdateChanges{
				offState: true}},
		{name: `resize`,
			input: LxcMounts{
				LxcMountID105: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(1051648))}}},
			current: LxcMounts{
				LxcMountID105: baseMount()},
			output: lxcUpdateChanges{
				resize: []lxcMountResize{{
					sizeInKibibytes: LxcMountSize(1051648),
					id:              "mp105"}}}},
		{name: `move`,
			input: LxcMounts{
				LxcMountID150: LxcMount{
					DataMount: &LxcDataMount{
						Storage: util.Pointer("local-zfs")}}},
			current: LxcMounts{
				LxcMountID150: baseMount()},
			output: lxcUpdateChanges{
				move: []lxcMountMove{{
					storage: "local-zfs",
					id:      "mp150"}},
				offState: true}},
		{name: `resize and move`,
			input: LxcMounts{
				LxcMountID12: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(1051648)),
						Storage:         util.Pointer("local-zfs")}}},
			current: LxcMounts{
				LxcMountID12: baseMount()},
			output: lxcUpdateChanges{
				resize: []lxcMountResize{{
					sizeInKibibytes: LxcMountSize(1051648),
					id:              "mp12"}},
				move: []lxcMountMove{{
					storage: "local-zfs",
					id:      "mp12"}},
				offState: true}},
		{name: `resize and move multiple`,
			input: LxcMounts{
				LxcMountID60: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(1051648)),
						Storage:         util.Pointer("local-zfs")}},
				LxcMountID80: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(1051648)),
						Storage:         util.Pointer("local-zfs")}}},
			current: LxcMounts{
				LxcMountID60: baseMount(),
				LxcMountID80: baseMount()},
			output: lxcUpdateChanges{
				resize: []lxcMountResize{
					{id: "mp60",
						sizeInKibibytes: LxcMountSize(1051648)},
					{id: "mp80",
						sizeInKibibytes: LxcMountSize(1051648)}},
				move: []lxcMountMove{
					{id: "mp60",
						storage: "local-zfs"},
					{id: "mp80",
						storage: "local-zfs"}},
				offState: true}},
		{name: `no changes / no settings`,
			input: LxcMounts{
				LxcMountID80: LxcMount{}},
			current: LxcMounts{
				LxcMountID80: baseMount()}},
		{name: `no changes / no current`,
			input: LxcMounts{
				LxcMountID60: LxcMount{}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpOutput := test.input.markMountChanges(test.current)
			slices.SortFunc(tmpOutput.move, func(a, b lxcMountMove) int {
				if a.id < b.id {
					return -1
				} else if a.id > b.id {
					return 1
				}
				return 0
			})
			slices.SortFunc(tmpOutput.resize, func(a, b lxcMountResize) int {
				if a.id < b.id {
					return -1
				} else if a.id > b.id {
					return 1
				}
				return 0
			})
			require.Equal(t, test.output, tmpOutput)
		})
	}
}

func Test_LxcMountSize_String(t *testing.T) {
	tests := []struct {
		name   string
		input  LxcMountSize
		output string
	}{
		{name: "Kibibyte",
			input:  kibiByte,
			output: "1K"},
		{name: "Mebibyte",
			input:  mebiByte,
			output: "1M"},
		{name: "Gibibyte",
			input:  gibiByte,
			output: "1G"},
		{name: "Tebibyte",
			input:  tebiByte,
			output: "1T"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.String())
		})
	}
}

func Test_lxcMountMove_mapToAPI(t *testing.T) {
	tests := []struct {
		name   string
		input  lxcMountMove
		delete bool
		output map[string]any
	}{
		{name: `delete false`,
			input: lxcMountMove{
				id:      "100",
				storage: "local-ext4"},
			output: map[string]any{
				"storage": string("local-ext4"),
				"volume":  string("100")}},
		{name: `delete true`,
			input: lxcMountMove{
				id:      "100",
				storage: "local-ext4"},
			delete: true,
			output: map[string]any{
				"storage": string("local-ext4"),
				"volume":  string("100"),
				"delete":  string("1")}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.mapToAPI(test.delete))
		})
	}
}

func Test_RawConfigLXC_BootMount(t *testing.T) {
	require.Equal(t,
		&LxcBootMount{
			ACL:   util.Pointer(TriBoolTrue),
			Quota: util.Pointer(true),
			Options: &LxcBootMountOptions{
				Discard:  util.Pointer(true),
				LazyTime: util.Pointer(true),
				NoATime:  util.Pointer(true),
				NoSuid:   util.Pointer(true)},
			Replicate:       util.Pointer(true),
			Storage:         util.Pointer("local-ext4"),
			SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
			rawDisk:         "local-ext4:subvol-101-disk-0"},
		RawConfigLXC{"rootfs": "local-ext4:subvol-101-disk-0,acl=1,mountoptions=discard;lazytime;noatime;nosuid,size=1G,quota=1,replicate=1"}.BootMount())
}

func Test_RawConfigLXC_Mounts(t *testing.T) {
	require.Equal(t,
		LxcMounts{
			LxcMountID150: LxcMount{DataMount: &LxcDataMount{
				ACL:    util.Pointer(TriBoolFalse),
				Backup: util.Pointer(true),
				Options: &LxcMountOptions{
					Discard:  util.Pointer(true),
					LazyTime: util.Pointer(true),
					NoATime:  util.Pointer(false),
					NoDevice: util.Pointer(false),
					NoExec:   util.Pointer(true),
					NoSuid:   util.Pointer(false)},
				Path:            util.Pointer(LxcMountPath("/opt/test")),
				Quota:           util.Pointer(true),
				ReadOnly:        util.Pointer(true),
				Replicate:       util.Pointer(true),
				SizeInKibibytes: util.Pointer(LxcMountSize(18)),
				Storage:         util.Pointer("local-zfs"),
				rawDisk:         "local-zfs:subvol-100-disk-1"}}},
		RawConfigLXC{"mp150": "local-zfs:subvol-100-disk-1,size=18K,acl=0,backup=1,quota=1,mountoptions=lazytime;noexec;discard,mp=/opt/test,replicate=1,ro=1"}.Mounts())
}
