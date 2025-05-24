package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_ConfigLXC_mapToAPI(t *testing.T) {
	type test struct {
		name          string
		config        ConfigLXC
		currentConfig ConfigLXC
		output        map[string]any
		pool          PoolName
	}
	tests := []struct {
		category     string
		create       []test
		createUpdate []test // value of currentConfig wil be used for update and ignored for create
		update       []test
	}{
		{category: `BootMount`,
			create: []test{
				{name: `minimum 1G`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
						Storage:         util.Pointer("local-ext4")}},
					output: map[string]any{"rootfs": "local-ext4:1"}},
				{name: `minimum <1G`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(917504)),
						Storage:         util.Pointer("local-zfs")}},
					output: map[string]any{"rootfs": "local-zfs:0.875"}},
				{name: `minimum >1G`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(1179648)),
						Storage:         util.Pointer("local-lvm")}},
					output: map[string]any{"rootfs": "local-lvm:1"}}},
			createUpdate: []test{
				{name: `ACL true`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolTrue)}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolFalse)}},
					output: map[string]any{"rootfs": ",acl=1"}},
				{name: `ACL false`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolFalse)}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolTrue)}},
					output: map[string]any{"rootfs": ",acl=0"}},
				{name: `ACL none`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolNone)}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolTrue)}},
					output: map[string]any{"rootfs": ""}},
				{name: `Options Discard true`,
					config: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						Discard: util.Pointer(true)}}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						Discard: util.Pointer(false)}}},
					output: map[string]any{"rootfs": ",mountoptions=discard"}},
				{name: `Options Discard false`,
					config: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						Discard: util.Pointer(false)}}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						Discard: util.Pointer(true)}}},
					output: map[string]any{"rootfs": ""}},
				{name: `Options LazyTime true`,
					config: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						LazyTime: util.Pointer(true)}}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						LazyTime: util.Pointer(false)}}},
					output: map[string]any{"rootfs": ",mountoptions=lazytime"}},
				{name: `Options LazyTime false`,
					config: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						LazyTime: util.Pointer(false)}}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						LazyTime: util.Pointer(true)}}},
					output: map[string]any{"rootfs": ""}},
				{name: `Options NoATime true`,
					config: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						NoATime: util.Pointer(true)}}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						NoATime: util.Pointer(false)}}},
					output: map[string]any{"rootfs": ",mountoptions=noatime"}},
				{name: `Options NoATime false`,
					config: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						NoATime: util.Pointer(false)}}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						NoATime: util.Pointer(true)}}},
					output: map[string]any{"rootfs": ""}},
				{name: `Options NoSuid true`,
					config: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						NoSuid: util.Pointer(true)}}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						NoSuid: util.Pointer(false)}}},
					output: map[string]any{"rootfs": ",mountoptions=nosuid"}},
				{name: `Options NoSuid false`,
					config: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						NoSuid: util.Pointer(false)}}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						NoSuid: util.Pointer(true)}}},
					output: map[string]any{"rootfs": ""}},
				{name: `Replication false`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						Replication: util.Pointer(false)}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						Replication: util.Pointer(true)}},
					output: map[string]any{"rootfs": ",replicate=0"}},
				{name: `Replication true`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						Replication: util.Pointer(true)}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						Replication: util.Pointer(false)}},
					output: map[string]any{"rootfs": ""}}},
			update: []test{
				{name: `Options Discard in-place true`,
					config: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						Discard: util.Pointer(true)}}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						Discard:  util.Pointer(false),
						LazyTime: util.Pointer(true),
						NoATime:  util.Pointer(false),
						NoSuid:   util.Pointer(true)}}},
					output: map[string]any{"rootfs": ",mountoptions=discard;lazytime;nosuid"}},
				{name: `Options LazyTime in-place true`,
					config: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						LazyTime: util.Pointer(true)}}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						Discard:  util.Pointer(true),
						LazyTime: util.Pointer(false),
						NoATime:  util.Pointer(true),
						NoSuid:   util.Pointer(false)}}},
					output: map[string]any{"rootfs": ",mountoptions=discard;lazytime;noatime"}},
				{name: `Options NoATime in-place true`,
					config: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						NoATime: util.Pointer(true)}}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						Discard:  util.Pointer(true),
						LazyTime: util.Pointer(false),
						NoATime:  util.Pointer(false),
						NoSuid:   util.Pointer(true)}}},
					output: map[string]any{"rootfs": ",mountoptions=discard;noatime;nosuid"}},
				{name: `Options NoSuid in-place true`,
					config: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						NoSuid: util.Pointer(true)}}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{Options: &LxcBootMountOptions{
						Discard:  util.Pointer(true),
						LazyTime: util.Pointer(true),
						NoATime:  util.Pointer(true),
						NoSuid:   util.Pointer(false)}}},
					output: map[string]any{"rootfs": ",mountoptions=discard;lazytime;noatime;nosuid"}},
				{name: `Storage & size`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
						Storage:         util.Pointer("local-ext4")}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(2097152)),
						Storage:         util.Pointer("local-zfs"),
						rawDisk:         "subvol-101-disk-0"}},
					output: map[string]any{"rootfs": "local-ext4:subvol-101-disk-0"}},
				{name: `no change`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolTrue),
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(true),
							LazyTime: util.Pointer(true),
							NoATime:  util.Pointer(true),
							NoSuid:   util.Pointer(true)},
						Replication: util.Pointer(true),
						Storage:     util.Pointer("local-zfs"),
						rawDisk:     "subvol-101-disk-0"}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolTrue),
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(true),
							LazyTime: util.Pointer(true),
							NoATime:  util.Pointer(true),
							NoSuid:   util.Pointer(true)},
						Replication: util.Pointer(true),
						Storage:     util.Pointer("local-zfs")}},
					output: map[string]any{}}}},
		{category: `Description`,
			createUpdate: []test{
				{name: `set`,
					config:        ConfigLXC{Description: util.Pointer("test")},
					currentConfig: ConfigLXC{Description: util.Pointer("text")},
					output:        map[string]any{"description": "test"}},
				{name: `delete no effect`,
					config:        ConfigLXC{Description: util.Pointer("")},
					currentConfig: ConfigLXC{Description: util.Pointer("")},
					output:        map[string]any{}}},
			update: []test{
				{name: `delete`,
					config:        ConfigLXC{Description: util.Pointer("")},
					currentConfig: ConfigLXC{Description: util.Pointer("test")},
					output:        map[string]any{"delete": "description"}},
				{name: `same`,
					config:        ConfigLXC{Description: util.Pointer("test")},
					currentConfig: ConfigLXC{Description: util.Pointer("test")},
					output:        map[string]any{}}}},
		{category: `ID`,
			create: []test{
				{name: `set`,
					config: ConfigLXC{ID: util.Pointer(GuestID(15))},
					output: map[string]any{"vmid": GuestID(15)}}},
			update: []test{
				{name: `do nothing`,
					config:        ConfigLXC{ID: util.Pointer(GuestID(15))},
					currentConfig: ConfigLXC{ID: util.Pointer(GuestID(0))},
					output:        map[string]any{}}}},
		{category: `Memory`,
			createUpdate: []test{
				{name: `set`,
					config:        ConfigLXC{Memory: util.Pointer(LxcMemory(512))},
					currentConfig: ConfigLXC{Memory: util.Pointer(LxcMemory(256))},
					output:        map[string]any{"memory": LxcMemory(512)}}},
			update: []test{
				{name: `same`,
					config:        ConfigLXC{Memory: util.Pointer(LxcMemory(512))},
					currentConfig: ConfigLXC{Memory: util.Pointer(LxcMemory(512))},
					output:        map[string]any{}}}},
		{category: `Name`,
			createUpdate: []test{
				{name: `set`,
					config:        ConfigLXC{Name: util.Pointer(GuestName("test"))},
					currentConfig: ConfigLXC{Name: util.Pointer(GuestName("text"))},
					output:        map[string]any{"name": string("test")}}},
			update: []test{
				{name: `do nothing`,
					config:        ConfigLXC{Name: util.Pointer(GuestName("test"))},
					currentConfig: ConfigLXC{Name: util.Pointer(GuestName("test"))},
					output:        map[string]any{}}}},
		{category: `Node`,
			createUpdate: []test{
				{name: `do nothing`,
					config:        ConfigLXC{Node: util.Pointer(NodeName("test"))},
					currentConfig: ConfigLXC{Node: util.Pointer(NodeName("text"))},
					output:        map[string]any{}}}},
		{category: `OperatingSystem`,
			createUpdate: []test{
				{name: `do nothing`,
					config:        ConfigLXC{OperatingSystem: "test"},
					currentConfig: ConfigLXC{OperatingSystem: "text"},
					output:        map[string]any{}}}},
		{category: `Pool`,
			create: []test{
				{name: `set`,
					config: ConfigLXC{Pool: util.Pointer(PoolName("test"))},
					output: map[string]any{
						"pool": "test"},
					pool: "test"}},
			update: []test{
				{name: `do nothing`,
					config:        ConfigLXC{Pool: util.Pointer(PoolName("test"))},
					currentConfig: ConfigLXC{Pool: util.Pointer(PoolName("text"))},
					output:        map[string]any{}}}},
		{category: `Privileged`,
			create: []test{
				{name: `true`,
					config: ConfigLXC{Privileged: util.Pointer(true)},
					output: map[string]any{}},
				{name: `false`,
					config: ConfigLXC{Privileged: util.Pointer(false)},
					output: map[string]any{"unprivileged": int(1)}}},
			update: []test{
				{name: `true no effect`,
					config:        ConfigLXC{Privileged: util.Pointer(true)},
					currentConfig: ConfigLXC{Privileged: util.Pointer(false)},
					output:        map[string]any{}},
				{name: `false no effect`,
					config:        ConfigLXC{Privileged: util.Pointer(false)},
					currentConfig: ConfigLXC{Privileged: util.Pointer(true)},
					output:        map[string]any{}}}},
		{category: `Tags`,
			createUpdate: []test{
				{name: `set`,
					config:        ConfigLXC{Tags: &Tags{"test"}},
					currentConfig: ConfigLXC{Tags: &Tags{"text"}},
					output:        map[string]any{"tags": "test"}}},
			update: []test{
				{name: `do nothing`,
					config:        ConfigLXC{Tags: &Tags{"bbb", "aaa", "ccc"}},
					currentConfig: ConfigLXC{Tags: &Tags{"aaa", "ccc", "bbb"}},
					output:        map[string]any{}}}},
	}
	for _, test := range tests {
		for _, subTest := range append(test.create, test.createUpdate...) {
			name := test.category + "/Create/" + subTest.name
			t.Run(name, func(*testing.T) {
				tmpParams, pool := subTest.config.mapToApiCreate()
				require.Equal(t, subTest.output, tmpParams, name)
				require.Equal(t, subTest.pool, pool, name)
			})
		}
		for _, subTest := range append(test.update, test.createUpdate...) {
			name := test.category + "/Update/" + subTest.name
			t.Run(name, func(*testing.T) {
				tmpParams := subTest.config.mapToApiUpdate(subTest.currentConfig)
				require.Equal(t, subTest.output, tmpParams, name)
			})
		}
	}
}

func Test_ConfigLXC_Validate(t *testing.T) {
	var baseConfig = func(config ConfigLXC) ConfigLXC {
		if config.BootMount == nil {
			config.BootMount = &LxcBootMount{
				Storage: util.Pointer("local-lvm")}
		}
		return config
	}
	type test struct {
		name    string
		input   ConfigLXC
		current *ConfigLXC
		err     error
	}
	type testType struct {
		create       []test
		createUpdate []test // value of currentConfig wil be used for update and ignored for create
		update       []test
	}
	tests := []struct {
		category string
		valid    testType
		invalid  testType
	}{
		{category: `BootMount`,
			valid: testType{
				createUpdate: []test{
					{name: `set`,
						input: baseConfig(ConfigLXC{BootMount: &LxcBootMount{
							SizeInKibibytes: util.Pointer(LxcMountSize(150000)),
							Storage:         util.Pointer("test")}}),
						current: &ConfigLXC{BootMount: &LxcBootMount{Storage: util.Pointer("text")}}}}},
			invalid: testType{
				create: []test{
					{name: `errors.New(ConfigLXC_Error_BootMountMissing)`,
						input: ConfigLXC{},
						err:   errors.New(ConfigLXC_Error_BootMountMissing)},
					{name: `errors.New(LxcBootMount_Error_NoStorageDuringCreation)`,
						input: ConfigLXC{BootMount: &LxcBootMount{}},
						err:   errors.New(LxcBootMount_Error_NoStorageDuringCreation)}},
				createUpdate: []test{
					{name: `errors.New(TriBool_Error_Invalid)`,
						input: baseConfig(ConfigLXC{BootMount: &LxcBootMount{
							ACL: util.Pointer(TriBool(34))}}),
						current: &ConfigLXC{BootMount: &LxcBootMount{
							ACL: util.Pointer(TriBoolNone)}},
						err: errors.New(TriBool_Error_Invalid)},
					{name: `errors.New(LxcMountSize_Error_Minimum)`,
						input: baseConfig(ConfigLXC{BootMount: &LxcBootMount{
							Storage:         util.Pointer("local-lvm"),
							SizeInKibibytes: util.Pointer(lxcMountSize_Minimum - 1)}}),
						current: &ConfigLXC{BootMount: &LxcBootMount{
							SizeInKibibytes: util.Pointer(LxcMountSize(131071))}},
						err: errors.New(LxcMountSize_Error_Minimum)}}}},
		{category: `ID`,
			valid: testType{
				createUpdate: []test{
					{name: `set`,
						input:   baseConfig(ConfigLXC{ID: util.Pointer(GuestID(150))}),
						current: &ConfigLXC{ID: util.Pointer(GuestID(0))}}}},
			invalid: testType{
				createUpdate: []test{
					{name: `empty`,
						input:   baseConfig(ConfigLXC{ID: util.Pointer(GuestID(0))}),
						current: &ConfigLXC{ID: util.Pointer(GuestID(0))},
						err:     errors.New(GuestID_Error_Minimum)},
					{name: `minimum`,
						input:   baseConfig(ConfigLXC{ID: util.Pointer(GuestIdMinimum - 1)}),
						current: &ConfigLXC{ID: util.Pointer(GuestID(0))},
						err:     errors.New(GuestID_Error_Minimum)},
					{name: `maximum`,
						input:   baseConfig(ConfigLXC{ID: util.Pointer(GuestIdMaximum + 1)}),
						current: &ConfigLXC{ID: util.Pointer(GuestID(0))},
						err:     errors.New(GuestID_Error_Maximum)}}}},
		{category: `Memory`,
			valid: testType{
				createUpdate: []test{
					{name: `set`,
						input:   baseConfig(ConfigLXC{Memory: util.Pointer(LxcMemory(512))}),
						current: &ConfigLXC{Memory: util.Pointer(LxcMemory(256))}}}},
			invalid: testType{
				createUpdate: []test{
					{name: `minimum`,
						input:   baseConfig(ConfigLXC{Memory: util.Pointer(LxcMemory(LxcMemoryMinimum - 1))}),
						current: &ConfigLXC{Memory: util.Pointer(LxcMemory(256))},
						err:     errors.New(LxcMemory_Error_Minimum)}}}},
		{category: `Name`,
			valid: testType{
				createUpdate: []test{
					{name: `set`,
						input:   baseConfig(ConfigLXC{Name: util.Pointer(GuestName("test"))}),
						current: &ConfigLXC{Name: util.Pointer(GuestName("text"))}}}},
			invalid: testType{
				createUpdate: []test{
					{name: `empty`,
						input:   baseConfig(ConfigLXC{Name: util.Pointer(GuestName(""))}),
						current: &ConfigLXC{Name: util.Pointer(GuestName("text"))},
						err:     errors.New(GuestName_Error_Empty)}}}},
		{category: `Node`,
			valid: testType{
				createUpdate: []test{
					{name: `set`,
						input:   baseConfig(ConfigLXC{Node: util.Pointer(NodeName("test"))}),
						current: &ConfigLXC{Node: util.Pointer(NodeName("text"))}}}},
			invalid: testType{
				createUpdate: []test{
					{name: `empty`,
						input:   baseConfig(ConfigLXC{Node: util.Pointer(NodeName(""))}),
						current: &ConfigLXC{Node: util.Pointer(NodeName("text"))},
						err:     errors.New(NodeName_Error_Empty)}}}},
		{category: `Pool`,
			valid: testType{
				createUpdate: []test{
					{name: `set`,
						input:   baseConfig(ConfigLXC{Pool: util.Pointer(PoolName("test"))}),
						current: &ConfigLXC{Pool: util.Pointer(PoolName("text"))}}}},
			invalid: testType{
				createUpdate: []test{
					{name: `empty`,
						input:   baseConfig(ConfigLXC{Pool: util.Pointer(PoolName(""))}),
						current: &ConfigLXC{Pool: util.Pointer(PoolName("text"))},
						err:     errors.New(PoolName_Error_Empty)}}}},
		{category: `Tags`,
			valid: testType{
				createUpdate: []test{
					{name: `set`,
						input:   baseConfig(ConfigLXC{Tags: &Tags{"test"}}),
						current: &ConfigLXC{Tags: &Tags{"text"}}}}},
			invalid: testType{
				createUpdate: []test{
					{name: `empty`,
						input:   baseConfig(ConfigLXC{Tags: &Tags{""}}),
						current: &ConfigLXC{Tags: &Tags{"text"}},
						err:     errors.New(Tag_Error_Empty)}}}},
	}
	for _, test := range tests {
		for _, subTest := range append(test.valid.create, test.valid.createUpdate...) {
			name := test.category + "/Valid/Create"
			if len(test.valid.create)+len(test.valid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.Equal(t, subTest.err, subTest.input.Validate(nil), name)
			})
		}
		for _, subTest := range append(test.valid.update, test.valid.createUpdate...) {
			name := test.category + "/Valid/Update"
			if len(test.valid.update)+len(test.valid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.NotNil(t, subTest.current)
				require.Equal(t, subTest.err, subTest.input.Validate(subTest.current), name)
			})
		}
		for _, subTest := range append(test.invalid.create, test.invalid.createUpdate...) {
			name := test.category + "/Invalid/Create"
			if len(test.invalid.create)+len(test.invalid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.Equal(t, subTest.err, subTest.input.Validate(nil), name)
			})
		}
		for _, subTest := range append(test.invalid.update, test.invalid.createUpdate...) {
			name := test.category + "/Invalid/Update"
			if len(test.invalid.update)+len(test.invalid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.NotNil(t, subTest.current)
				require.Equal(t, subTest.err, subTest.input.Validate(subTest.current), name)
			})
		}
	}
}

func Test_RawConfigLXC_ALL(t *testing.T) {
	baseConfig := func(config ConfigLXC) *ConfigLXC {
		if config.ID == nil {
			config.ID = util.Pointer(GuestID(0))
		}
		if config.Node == nil {
			config.Node = util.Pointer(NodeName(""))
		}
		if config.Privileged == nil {
			config.Privileged = util.Pointer(false)
		}
		return &config
	}
	type test struct {
		name   string
		input  RawConfigLXC
		vmr    VmRef
		output *ConfigLXC
	}
	tests := []struct {
		category string
		tests    []test
	}{
		{category: `Architecture`,
			tests: []test{
				{name: `amd64`,
					input:  RawConfigLXC{"arch": "amd64"},
					output: baseConfig(ConfigLXC{Architecture: "amd64"})},
				{name: `""`,
					input:  RawConfigLXC{"arch": ""},
					output: baseConfig(ConfigLXC{Architecture: ""})}}},
		{category: `BootMount`,
			tests: []test{
				{name: `ACL true`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,acl=1"},
					output: baseConfig(ConfigLXC{BootMount: &LxcBootMount{
						ACL:         util.Pointer(TriBoolTrue),
						Replication: util.Pointer(true),
						Storage:     util.Pointer("local-zfs"),
						rawDisk:     "subvol-101-disk-0"}})},
				{name: `ACL false`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,acl=0"},
					output: baseConfig(ConfigLXC{BootMount: &LxcBootMount{
						ACL:         util.Pointer(TriBoolFalse),
						Replication: util.Pointer(true),
						Storage:     util.Pointer("local-zfs"),
						rawDisk:     "subvol-101-disk-0"}})},
				{name: `Options Discard true`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=discard"},
					output: baseConfig(ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolNone),
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(true),
							LazyTime: util.Pointer(false),
							NoATime:  util.Pointer(false),
							NoSuid:   util.Pointer(false)},
						Replication: util.Pointer(true),
						Storage:     util.Pointer("local-zfs"),
						rawDisk:     "subvol-101-disk-0"}})},
				{name: `Options LazyTime true`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=lazytime"},
					output: baseConfig(ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolNone),
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(false),
							LazyTime: util.Pointer(true),
							NoATime:  util.Pointer(false),
							NoSuid:   util.Pointer(false)},
						Replication: util.Pointer(true),
						Storage:     util.Pointer("local-zfs"),
						rawDisk:     "subvol-101-disk-0"}})},
				{name: `Options NoATime true`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=noatime"},
					output: baseConfig(ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolNone),
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(false),
							LazyTime: util.Pointer(false),
							NoATime:  util.Pointer(true),
							NoSuid:   util.Pointer(false)},
						Replication: util.Pointer(true),
						Storage:     util.Pointer("local-zfs"),
						rawDisk:     "subvol-101-disk-0"}})},
				{name: `Options NoSuid true`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=nosuid"},
					output: baseConfig(ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolNone),
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(false),
							LazyTime: util.Pointer(false),
							NoATime:  util.Pointer(false),
							NoSuid:   util.Pointer(true)},
						Replication: util.Pointer(true),
						Storage:     util.Pointer("local-zfs"),
						rawDisk:     "subvol-101-disk-0"}})},
				{name: `Replication false`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,replicate=0"},
					output: baseConfig(ConfigLXC{BootMount: &LxcBootMount{
						ACL:         util.Pointer(TriBoolNone),
						Replication: util.Pointer(false),
						Storage:     util.Pointer("local-zfs"),
						rawDisk:     "subvol-101-disk-0"}})},
				{name: `Replication true`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,replicate=1"},
					output: baseConfig(ConfigLXC{BootMount: &LxcBootMount{
						ACL:         util.Pointer(TriBoolNone),
						Replication: util.Pointer(true),
						Storage:     util.Pointer("local-zfs"),
						rawDisk:     "subvol-101-disk-0"}})},
				{name: `SizeInKibibytes`,
					input: RawConfigLXC{"rootfs": "local-ext4:subvol-101-disk-0,size=999M"},
					output: baseConfig(ConfigLXC{BootMount: &LxcBootMount{
						ACL:             util.Pointer(TriBoolNone),
						Replication:     util.Pointer(true),
						Storage:         util.Pointer("local-ext4"),
						SizeInKibibytes: util.Pointer(LxcMountSize(1022976)),
						rawDisk:         "subvol-101-disk-0"}})},
				{name: `all`,
					input: RawConfigLXC{"rootfs": "local-ext4:subvol-101-disk-0,acl=1,mountoptions=discard;lazytime;noatime;nosuid,size=1G"},
					output: baseConfig(ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolTrue),
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(true),
							LazyTime: util.Pointer(true),
							NoATime:  util.Pointer(true),
							NoSuid:   util.Pointer(true)},
						Replication:     util.Pointer(true),
						Storage:         util.Pointer("local-ext4"),
						SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
						rawDisk:         "subvol-101-disk-0"}})}}},
		{category: `Description`,
			tests: []test{
				{name: `test`,
					input:  RawConfigLXC{"description": "test"},
					output: baseConfig(ConfigLXC{Description: util.Pointer("test")})},
				{name: `""`,
					input:  RawConfigLXC{"description": ""},
					output: baseConfig(ConfigLXC{Description: util.Pointer("")})}}},
		{category: `ID`,
			tests: []test{
				{name: `set`,
					vmr:    VmRef{vmId: 15},
					output: baseConfig(ConfigLXC{ID: util.Pointer(GuestID(15))})}}},
		{category: `Memory`,
			tests: []test{
				{name: `set`,
					input:  RawConfigLXC{"memory": float64(512)},
					output: baseConfig(ConfigLXC{Memory: util.Pointer(LxcMemory(512))})}}},
		{category: `Node`,
			tests: []test{
				{name: `set`,
					vmr:    VmRef{node: "test"},
					output: baseConfig(ConfigLXC{Node: util.Pointer(NodeName("test"))})}}},
		{category: `Name`,
			tests: []test{
				{name: `set`,
					input:  RawConfigLXC{"name": "test"},
					output: baseConfig(ConfigLXC{Name: util.Pointer(GuestName("test"))})}}},
		{category: `OperatingSystem`,
			tests: []test{
				{name: `set`,
					input:  RawConfigLXC{"ostype": "test"},
					output: baseConfig(ConfigLXC{OperatingSystem: "test"})}}},
		{category: `Pool`,
			tests: []test{
				{name: `set`,
					vmr:    VmRef{pool: "test"},
					output: baseConfig(ConfigLXC{Pool: util.Pointer(PoolName("test"))})}}},
		{category: `Privileged`,
			tests: []test{
				{name: `true`,
					input:  RawConfigLXC{"unprivileged": float64(0)},
					output: baseConfig(ConfigLXC{Privileged: util.Pointer(true)})},
				{name: `false`,
					input:  RawConfigLXC{"unprivileged": float64(1)},
					output: baseConfig(ConfigLXC{Privileged: util.Pointer(false)})},
				{name: `default false`,
					input:  RawConfigLXC{},
					output: baseConfig(ConfigLXC{Privileged: util.Pointer(false)})}}},
		{category: `Tags`,
			tests: []test{
				{name: `set`,
					input:  RawConfigLXC{"tags": "test"},
					output: baseConfig(ConfigLXC{Tags: &Tags{"test"}})}}},
	}
	for _, test := range tests {
		for _, subTest := range test.tests {
			name := test.category
			if len(test.tests) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.Equal(t, subTest.output, subTest.input.ALL(subTest.vmr), name)
			})
		}
	}
}
