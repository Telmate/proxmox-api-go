package proxmox

import (
	"context"
	"crypto/sha1"
	"errors"
	"maps"
	"net/netip"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_guest"
	"github.com/stretchr/testify/require"
)

type lxcDefaults uint8

const (
	lxcDefaultsNone lxcDefaults = iota
	lxcDefaultsCreate
	lxcDefaultsUpdate
	lxcDefaultsAll
)

type (
	lxcTestCaseAPI struct {
		name          string
		config        ConfigLXC
		currentConfig ConfigLXC
		output        map[string]any
		body          map[string]string
		omitDefaults  lxcDefaults // opt out of default values in returned map
		pool          PoolName
	}
	lxcTestsAPI struct {
		category     string
		create       []lxcTestCaseAPI
		createUpdate []lxcTestCaseAPI // value of currentConfig wil be used for update and ignored for create
		update       []lxcTestCaseAPI
	}
)

func lxcGetBaseConfig(config ConfigLXC) *ConfigLXC {
	if config.ID == nil {
		config.ID = new(GuestID(0))
	}
	if config.Memory == nil {
		config.Memory = new(LxcMemory(0))
	}
	if config.Name == nil {
		config.Name = new(GuestName(""))
	}
	if config.Networks == nil {
		config.Networks = make(LxcNetworks)
	}
	if config.Node == nil {
		config.Node = new(NodeName(""))
	}
	if config.Privileged == nil {
		config.Privileged = new(true)
	}
	if config.Protection == nil {
		config.Protection = new(false)
	}
	if config.StartAtNodeBoot == nil {
		config.StartAtNodeBoot = new(false)
	}
	if config.Swap == nil {
		config.Swap = new(LxcSwap(0))
	}
	if config.Tags == nil {
		config.Tags = new(Tags)
	}
	return &config
}

func Test_CpuArchitecture_String(t *testing.T) {
	t.Parallel()
	require.Equal(t, "amd64", LxcCpuArchitecture("amd64").String())
}

func Test_ConfigLXC_mapToAPI(t *testing.T) {
	t.Parallel()
	baseDataMount := func(m LxcDataMount) *LxcDataMount {
		m.Storage = util.Pointer("local-zfs")
		m.SizeInKibibytes = util.Pointer(LxcMountSize(1048576))
		return &m
	}
	featuresPrivileged := func(value bool) *LxcFeatures {
		return &LxcFeatures{Privileged: &PrivilegedFeatures{
			CreateDeviceNodes: util.Pointer(value),
			FUSE:              util.Pointer(value),
			NFS:               util.Pointer(value),
			Nesting:           util.Pointer(value),
			SMB:               util.Pointer(value)}}
	}
	featuresUnprivileged := func(value bool) *LxcFeatures {
		return &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
			CreateDeviceNodes: util.Pointer(value),
			FUSE:              util.Pointer(value),
			KeyCtl:            util.Pointer(value),
			Nesting:           util.Pointer(value)}}
	}
	parseIP := func(rawIP string) netip.Addr {
		ip, err := netip.ParseAddr(rawIP)
		failPanic(err)
		return ip
	}
	bindMount := func() *LxcBindMount {
		return &LxcBindMount{
			HostPath:  util.Pointer(LxcHostPath("/mnt/data")),
			GuestPath: util.Pointer(LxcMountPath("/mnt/bind")),
			Options: &LxcMountOptions{
				NoATime:  util.Pointer(true),
				NoDevice: util.Pointer(true),
				NoExec:   util.Pointer(true),
				NoSuid:   util.Pointer(true)},
			ReadOnly:  util.Pointer(true),
			Replicate: util.Pointer(false)}
	}
	dataMount := func() *LxcDataMount {
		return &LxcDataMount{
			ACL:    util.Pointer(TriBoolFalse),
			Backup: util.Pointer(true),
			Options: &LxcMountOptions{
				Discard:  util.Pointer(true),
				LazyTime: util.Pointer(true),
				NoATime:  util.Pointer(true),
				NoSuid:   util.Pointer(true)},
			Path:            util.Pointer(LxcMountPath("/mnt/data")),
			ReadOnly:        util.Pointer(true),
			Replicate:       util.Pointer(false),
			SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
			Storage:         util.Pointer("local-zfs"),
			rawDisk:         "local-zfs:subvol-101-disk-0"}
	}
	publicKeys := func() []AuthorizedKey {
		data := test_data_guest.AuthorizedKey_Decoded_Input()
		keys := make([]AuthorizedKey, len(data))
		for i := range data {
			keys[i] = AuthorizedKey{Options: data[i].Options, PublicKey: data[i].PublicKey, Comment: data[i].Comment}
		}
		return keys
	}
	tests := []lxcTestsAPI{
		{category: `BootMount`,
			create: []lxcTestCaseAPI{
				{name: `all`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolTrue),
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(true),
							LazyTime: util.Pointer(true),
							NoATime:  util.Pointer(true),
							NoSuid:   util.Pointer(true)},
						Replicate:       util.Pointer(false),
						SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
						Storage:         util.Pointer("local-zfs")}},
					output: map[string]any{"rootfs": "local-zfs:1,acl=1,mountoptions=discard;lazytime;noatime;nosuid,replicate=0"}},
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
					output: map[string]any{"rootfs": "local-lvm:1"}},
				{name: `Quota false, Privileged false`,
					config: ConfigLXC{
						BootMount: &LxcBootMount{
							Quota: util.Pointer(false)},
						Privileged: util.Pointer(false)},
					output: map[string]any{"rootfs": ""}},
				{name: `Quota false, Privileged true`,
					config: ConfigLXC{
						BootMount: &LxcBootMount{
							Quota: util.Pointer(false)},
						Privileged: util.Pointer(true)},
					omitDefaults: lxcDefaultsAll,
					body:         map[string]string{"unprivileged": "0"},
					output:       map[string]any{"rootfs": ""}},
				{name: `Quota true, Privileged false`,
					config: ConfigLXC{
						BootMount: &LxcBootMount{
							Quota: util.Pointer(true)},
						Privileged: util.Pointer(false)},
					output: map[string]any{"rootfs": ""}},
				{name: `Quota true, Privileged true`,
					config: ConfigLXC{
						BootMount: &LxcBootMount{
							Quota: util.Pointer(true)},
						Privileged: util.Pointer(true)},
					omitDefaults: lxcDefaultsAll,
					body:         map[string]string{"unprivileged": "0"},
					output:       map[string]any{"rootfs": ",quota=1"}}},
			createUpdate: []lxcTestCaseAPI{
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
						Replicate: util.Pointer(false)}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						Replicate: util.Pointer(true)}},
					output: map[string]any{"rootfs": ",replicate=0"}},
				{name: `Replication true`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						Replicate: util.Pointer(true)}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						Replicate: util.Pointer(false)}},
					output: map[string]any{"rootfs": ""}}},
			update: []lxcTestCaseAPI{
				{name: `all storage change, linked clone`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						Storage:         util.Pointer("local-zfs"),
						SizeInKibibytes: util.Pointer(LxcMountSize(4 * gibibyte)),
						ACL:             util.Pointer(TriBoolTrue),
						Options: &LxcBootMountOptions{
							Discard: util.Pointer(true)},
						Replicate: util.Pointer(false)}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						Storage:         util.Pointer("local-zfs"),
						SizeInKibibytes: util.Pointer(LxcMountSize(4 * gibibyte)),
						ACL:             util.Pointer(TriBoolTrue),
						Replicate:       util.Pointer(false),
						rawDisk:         "local-zfs:basevol-9100-disk-0/subvol-105-disk-0"}},
					output: map[string]any{"rootfs": "local-zfs:basevol-9100-disk-0/subvol-105-disk-0,acl=1,mountoptions=discard,replicate=0,size=4G"}},
				{name: `all storage change, no api change`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						Storage: util.Pointer("local-zfs")}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						Storage:         util.Pointer("local-ext4"),
						SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
						ACL:             util.Pointer(TriBoolTrue),
						Quota:           util.Pointer(true),
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(true),
							LazyTime: util.Pointer(true),
							NoATime:  util.Pointer(true),
							NoSuid:   util.Pointer(true)},
						Replicate: util.Pointer(false),
						rawDisk:   "local-ext4:subvol-101-disk-0"}},
					omitDefaults: lxcDefaultsAll},
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
				{name: `Quota false, Privileged false`,
					config: ConfigLXC{
						BootMount: &LxcBootMount{
							Quota: util.Pointer(false)},
						Privileged: util.Pointer(false)},
					currentConfig: ConfigLXC{
						BootMount: &LxcBootMount{
							Quota: util.Pointer(true)},
						Privileged: util.Pointer(false)},
					omitDefaults: lxcDefaultsAll},
				{name: `Quota false, Privileged true`,
					config: ConfigLXC{
						BootMount: &LxcBootMount{
							Quota: util.Pointer(false)},
						Privileged: util.Pointer(true)},
					currentConfig: ConfigLXC{
						BootMount: &LxcBootMount{
							Quota: util.Pointer(true)},
						Privileged: util.Pointer(true)},
					output: map[string]any{"rootfs": ""}},
				{name: `Quota true, Privileged false`,
					config: ConfigLXC{
						BootMount: &LxcBootMount{
							Quota: util.Pointer(true)},
						Privileged: util.Pointer(false)},
					currentConfig: ConfigLXC{
						BootMount: &LxcBootMount{
							Quota: util.Pointer(false)},
						Privileged: util.Pointer(false)},
					omitDefaults: lxcDefaultsAll},
				{name: `Quota true, Privileged true`,
					config: ConfigLXC{
						BootMount: &LxcBootMount{
							Quota: util.Pointer(true)},
						Privileged: util.Pointer(true)},
					currentConfig: ConfigLXC{
						BootMount: &LxcBootMount{
							Quota: util.Pointer(false)},
						Privileged: util.Pointer(true)},
					output: map[string]any{"rootfs": ",quota=1"}},
				{name: `Storage & size, no api change`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(2621440)),
						Storage:         util.Pointer("local-ext4")}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(2097152)),
						Storage:         util.Pointer("local-zfs"),
						rawDisk:         "subvol-101-disk-0"}},
					omitDefaults: lxcDefaultsAll},
				{name: `no change`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolTrue),
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(true),
							LazyTime: util.Pointer(true),
							NoATime:  util.Pointer(true),
							NoSuid:   util.Pointer(true)},
						Replicate: util.Pointer(true),
						Storage:   util.Pointer("local-zfs"),
						rawDisk:   "subvol-101-disk-0"}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						ACL: util.Pointer(TriBoolTrue),
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(true),
							LazyTime: util.Pointer(true),
							NoATime:  util.Pointer(true),
							NoSuid:   util.Pointer(true)},
						Replicate: util.Pointer(true),
						Storage:   util.Pointer("local-zfs")}},
					omitDefaults: lxcDefaultsAll}}},
		{category: `CPU`,
			createUpdate: []lxcTestCaseAPI{
				{name: `Cores`,
					config:        ConfigLXC{CPU: &LxcCPU{Cores: util.Pointer(LxcCpuCores(1))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{Cores: util.Pointer(LxcCpuCores(2))}},
					output:        map[string]any{"cores": int(1)}},
				{name: `Limit`,
					config:        ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(2))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(3))}},
					output:        map[string]any{"cpulimit": int(2)}},
				{name: `Units`,
					config:        ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(3))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(4))}},
					output:        map[string]any{"cpuunits": int(3)}},
				{name: `Cores delete no effect`,
					config:        ConfigLXC{CPU: &LxcCPU{Cores: util.Pointer(LxcCpuCores(0))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{}},
					omitDefaults:  lxcDefaultsUpdate},
				{name: `Limit delete no effect`,
					config:        ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(0))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{}},
					omitDefaults:  lxcDefaultsUpdate},
				{name: `Units delete no effect`,
					config:        ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(0))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{}},
					omitDefaults:  lxcDefaultsUpdate}},
			update: []lxcTestCaseAPI{
				{name: `Cores delete`,
					config:        ConfigLXC{CPU: &LxcCPU{Cores: util.Pointer(LxcCpuCores(0))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{Cores: util.Pointer(LxcCpuCores(1))}},
					output:        map[string]any{"delete": "cores"}},
				{name: `Limit delete`,
					config:        ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(0))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(2))}},
					output:        map[string]any{"delete": "cpulimit"}},
				{name: `Units delete`,
					config:        ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(0))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(3))}},
					output:        map[string]any{"delete": "cpuunits"}},
				{name: `Cores same`,
					config:        ConfigLXC{CPU: &LxcCPU{Cores: util.Pointer(LxcCpuCores(1))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{Cores: util.Pointer(LxcCpuCores(1))}},
					omitDefaults:  lxcDefaultsAll},
				{name: `Limit same`,
					config:        ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(2))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(2))}},
					omitDefaults:  lxcDefaultsAll},
				{name: `Units same`,
					config:        ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(3))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(3))}},
					omitDefaults:  lxcDefaultsAll},
				{name: `Cores set`,
					config:        ConfigLXC{CPU: &LxcCPU{Cores: util.Pointer(LxcCpuCores(1))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{}},
					output:        map[string]any{"cores": int(1)}},
				{name: `Limit set`,
					config:        ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(2))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{}},
					output:        map[string]any{"cpulimit": int(2)}},
				{name: `Units set`,
					config:        ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(3))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{}},
					output:        map[string]any{"cpuunits": int(3)}},
				{name: `Cores delete no current`,
					config:        ConfigLXC{CPU: &LxcCPU{Cores: util.Pointer(LxcCpuCores(0))}},
					currentConfig: ConfigLXC{},
					omitDefaults:  lxcDefaultsAll},
				{name: `Limit delete no current`,
					config:        ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(0))}},
					currentConfig: ConfigLXC{},
					omitDefaults:  lxcDefaultsAll},
				{name: `Units delete no current`,
					config:        ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(0))}},
					currentConfig: ConfigLXC{},
					omitDefaults:  lxcDefaultsAll}}},
		{category: `CreateOptions`,
			create: []lxcTestCaseAPI{
				{name: `all`,
					config: ConfigLXC{CreateOptions: &LxcCreateOptions{
						OsTemplate: &LxcTemplate{
							Storage: "local",
							File:    "test-template"},
						UserPassword:  util.Pointer("myPassword!"),
						PublicSSHkeys: publicKeys()}},
					output: map[string]any{
						"ostemplate":      string("local:vztmpl/test-template"),
						"password":        string("myPassword!"),
						"ssh-public-keys": string(test_data_guest.AuthorizedKey_Trimmed_Output())}},
				{name: `OsTemplate`,
					config: ConfigLXC{CreateOptions: &LxcCreateOptions{
						OsTemplate: &LxcTemplate{
							Storage: "local",
							File:    "test-template"}}},
					output: map[string]any{
						"ostemplate": string("local:vztmpl/test-template")}},
				{name: `UserPassword`,
					config: ConfigLXC{CreateOptions: &LxcCreateOptions{
						UserPassword: util.Pointer("myPassword!")}},
					output: map[string]any{
						"password": string("myPassword!")}},
				{name: `UserPassword empty`,
					config: ConfigLXC{CreateOptions: &LxcCreateOptions{
						UserPassword: util.Pointer("")}},
					output: map[string]any{
						"password": string("")}},
				{name: `PublicSSHkeys`,
					config: ConfigLXC{CreateOptions: &LxcCreateOptions{
						PublicSSHkeys: publicKeys()}},
					output: map[string]any{
						"ssh-public-keys": string(test_data_guest.AuthorizedKey_Trimmed_Output())}},
				{name: `PublicSSHkeys empty`,
					config: ConfigLXC{CreateOptions: &LxcCreateOptions{
						PublicSSHkeys: []AuthorizedKey{}}}}},
			update: []lxcTestCaseAPI{
				{name: `all do nothing`,
					config: ConfigLXC{CreateOptions: &LxcCreateOptions{
						OsTemplate: &LxcTemplate{
							Storage: "local",
							File:    "test-template"},
						UserPassword:  util.Pointer("myPassword!"),
						PublicSSHkeys: publicKeys()}},
					omitDefaults: lxcDefaultsAll}}},
		{category: `Description`,
			createUpdate: []lxcTestCaseAPI{
				{name: `set`,
					config:        ConfigLXC{Description: util.Pointer("test")},
					currentConfig: ConfigLXC{Description: util.Pointer("text")},
					output:        map[string]any{"description": "test"}},
				{name: `delete no effect`,
					config:        ConfigLXC{Description: util.Pointer("")},
					currentConfig: ConfigLXC{Description: util.Pointer("")},
					omitDefaults:  lxcDefaultsUpdate}},
			update: []lxcTestCaseAPI{
				{name: `delete`,
					config:        ConfigLXC{Description: util.Pointer("")},
					currentConfig: ConfigLXC{Description: util.Pointer("test")},
					output:        map[string]any{"delete": "description"}},
				{name: `same`,
					config:        ConfigLXC{Description: util.Pointer("test")},
					currentConfig: ConfigLXC{Description: util.Pointer("test")},
					omitDefaults:  lxcDefaultsAll}}},
		{category: `Digest`,
			update: []lxcTestCaseAPI{
				{name: `not set when only setting`,
					config:        ConfigLXC{},
					currentConfig: ConfigLXC{rawDigest: "af064923bbf2301596aac4c273ba32178ebc4a96"},
					omitDefaults:  lxcDefaultsUpdate},
				{name: `set`,
					config:        ConfigLXC{Description: util.Pointer("test")},
					currentConfig: ConfigLXC{rawDigest: "af064923bbf2301596aac4c273ba32178ebc4a96"},
					omitDefaults:  lxcDefaultsUpdate,
					output: map[string]any{
						"digest":      "af064923bbf2301596aac4c273ba32178ebc4a96",
						"description": "test"}}}},
		{category: `DNS`,
			createUpdate: []lxcTestCaseAPI{
				{name: `all`,
					config: ConfigLXC{DNS: &GuestDNS{
						NameServers:  util.Pointer([]netip.Addr{parseIP("1.1.1.1"), parseIP("8.8.8.8")}),
						SearchDomain: util.Pointer("example.com")}},
					currentConfig: ConfigLXC{DNS: &GuestDNS{
						NameServers:  util.Pointer([]netip.Addr{parseIP("8.8.8.8"), parseIP("1.1.1.1")}),
						SearchDomain: util.Pointer("test.net")}},
					output: map[string]any{
						"nameserver":   string("1.1.1.1 8.8.8.8"),
						"searchdomain": string("example.com")}}},
			create: []lxcTestCaseAPI{
				{name: `do nothing`,
					config: ConfigLXC{DNS: &GuestDNS{
						NameServers:  util.Pointer([]netip.Addr{}),
						SearchDomain: util.Pointer("")}}}},
			update: []lxcTestCaseAPI{
				{name: `NameServers add`,
					config:        ConfigLXC{DNS: &GuestDNS{NameServers: util.Pointer([]netip.Addr{parseIP("1.1.1.1"), parseIP("9.9.9.9"), parseIP("8.8.8.8")})}},
					currentConfig: ConfigLXC{DNS: &GuestDNS{NameServers: util.Pointer([]netip.Addr{parseIP("1.1.1.1")})}},
					output:        map[string]any{"nameserver": string("1.1.1.1 9.9.9.9 8.8.8.8")}},
				{name: `NameServers delete`,
					config:        ConfigLXC{DNS: &GuestDNS{NameServers: util.Pointer([]netip.Addr{})}},
					currentConfig: ConfigLXC{DNS: &GuestDNS{NameServers: util.Pointer([]netip.Addr{parseIP("1.1.1.1")})}},
					output:        map[string]any{"delete": "nameserver"}},
				{name: `NameServers current empty`,
					config: ConfigLXC{DNS: &GuestDNS{
						NameServers: util.Pointer([]netip.Addr{parseIP("1.1.1.1")})}},
					currentConfig: ConfigLXC{},
					output:        map[string]any{"nameserver": string("1.1.1.1")}},
				{name: `SearchDomain delete`,
					config:        ConfigLXC{DNS: &GuestDNS{SearchDomain: util.Pointer("")}},
					currentConfig: ConfigLXC{DNS: &GuestDNS{SearchDomain: util.Pointer("example.com")}},
					output:        map[string]any{"delete": "searchdomain"}},
				{name: `SearchDomain current empty`,
					config: ConfigLXC{DNS: &GuestDNS{
						SearchDomain: util.Pointer("example.com")}},
					currentConfig: ConfigLXC{},
					output:        map[string]any{"searchdomain": string("example.com")}},
				{name: `do nothing`,
					config: ConfigLXC{DNS: &GuestDNS{
						NameServers:  util.Pointer([]netip.Addr{parseIP("1.1.1.1"), parseIP("8.8.8.8")}),
						SearchDomain: util.Pointer("example.com")}},
					currentConfig: ConfigLXC{DNS: &GuestDNS{
						NameServers:  util.Pointer([]netip.Addr{parseIP("1.1.1.1"), parseIP("8.8.8.8")}),
						SearchDomain: util.Pointer("example.com")}},
					omitDefaults: lxcDefaultsAll}}},
		{category: `Features`,
			create: []lxcTestCaseAPI{
				{name: `all false Privileged`,
					config: ConfigLXC{Features: featuresPrivileged(false)}},
				{name: `all false Unprivileged`,
					config: ConfigLXC{Features: featuresUnprivileged(false)}}},
			createUpdate: []lxcTestCaseAPI{
				{name: `CreateDeviceNodes Privileged`,
					config: ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						CreateDeviceNodes: util.Pointer(true)}}},
					output: map[string]any{"features": "mknod=1"}},
				{name: `CreateDeviceNodes Unprivileged`,
					config: ConfigLXC{Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
						CreateDeviceNodes: util.Pointer(true)}}},
					output: map[string]any{"features": "mknod=1"}},
				{name: `FUSE Privileged`,
					config: ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						FUSE: util.Pointer(true)}}},
					output: map[string]any{"features": "fuse=1"}},
				{name: `FUSE Unprivileged`,
					config: ConfigLXC{Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
						FUSE: util.Pointer(true)}}},
					output: map[string]any{"features": "fuse=1"}},
				{name: `KeyCtl Unprivileged`,
					config: ConfigLXC{Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
						KeyCtl: util.Pointer(true)}}},
					output: map[string]any{"features": "keyctl=1"}},
				{name: `NFS Privileged`,
					config: ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						NFS: util.Pointer(true)}}},
					output: map[string]any{"features": "mount=nfs"}},
				{name: `SMB Privileged`,
					config: ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						SMB: util.Pointer(true)}}},
					output: map[string]any{"features": "mount=cifs"}},
				{name: `Nesting Privileged`,
					config: ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						Nesting: util.Pointer(true)}}},
					output: map[string]any{"features": "nesting=1"}},
				{name: `Nesting Unprivileged`,
					config: ConfigLXC{Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
						Nesting: util.Pointer(true)}}},
					output: map[string]any{"features": "nesting=1"}},
				{name: `NFS and SMB Privileged`,
					config: ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						NFS: util.Pointer(true), SMB: util.Pointer(true)}}},
					output: map[string]any{"features": "mount=nfs;cifs"}},
				{name: `delete no effect false Privileged`,
					config:        ConfigLXC{Features: featuresPrivileged(false)},
					currentConfig: ConfigLXC{Features: featuresPrivileged(false)},
					omitDefaults:  lxcDefaultsUpdate},
				{name: `delete no effect false Unprivileged`,
					config:        ConfigLXC{Features: featuresUnprivileged(false)},
					currentConfig: ConfigLXC{Features: featuresUnprivileged(false)},
					omitDefaults:  lxcDefaultsUpdate},
				{name: `only top-level set, no effect`,
					config:        ConfigLXC{Features: &LxcFeatures{}},
					currentConfig: ConfigLXC{Features: &LxcFeatures{}},
					omitDefaults:  lxcDefaultsUpdate}},
			update: []lxcTestCaseAPI{
				{name: `CreateDeviceNodes false Privileged`,
					config: ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						CreateDeviceNodes: util.Pointer(false)}}},
					currentConfig: ConfigLXC{Features: featuresPrivileged(true)},
					output:        map[string]any{"features": "fuse=1,mount=nfs;cifs,nesting=1"}},
				{name: `CreateDeviceNodes false Unprivileged`,
					config: ConfigLXC{Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
						CreateDeviceNodes: util.Pointer(false)}}},
					currentConfig: ConfigLXC{Features: featuresUnprivileged(true)},
					output:        map[string]any{"features": "fuse=1,keyctl=1,nesting=1"}},
				{name: `FUSE false Privileged`,
					config: ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						FUSE: util.Pointer(false)}}},
					currentConfig: ConfigLXC{Features: featuresPrivileged(true)},
					output:        map[string]any{"features": "mknod=1,mount=nfs;cifs,nesting=1"}},
				{name: `FUSE false Unprivileged`,
					config: ConfigLXC{Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
						FUSE: util.Pointer(false)}}},
					currentConfig: ConfigLXC{Features: featuresUnprivileged(true)},
					output:        map[string]any{"features": "mknod=1,keyctl=1,nesting=1"}},
				{name: `KeyCtl false Unprivileged`,
					config: ConfigLXC{Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
						KeyCtl: util.Pointer(false)}}},
					currentConfig: ConfigLXC{Features: featuresUnprivileged(true)},
					output:        map[string]any{"features": "mknod=1,fuse=1,nesting=1"}},
				{name: `NFS false Privileged`,
					config: ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						NFS: util.Pointer(false)}}},
					currentConfig: ConfigLXC{Features: featuresPrivileged(true)},
					output:        map[string]any{"features": "mknod=1,fuse=1,mount=cifs,nesting=1"}},
				{name: `SMB false Privileged`,
					config: ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						SMB: util.Pointer(false)}}},
					currentConfig: ConfigLXC{Features: featuresPrivileged(true)},
					output:        map[string]any{"features": "mknod=1,fuse=1,mount=nfs,nesting=1"}},
				{name: `Nesting false Privileged`,
					config: ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						Nesting: util.Pointer(false)}}},
					currentConfig: ConfigLXC{Features: featuresPrivileged(true)},
					output:        map[string]any{"features": "mknod=1,fuse=1,mount=nfs;cifs"}},
				{name: `Nesting false Unprivileged`,
					config: ConfigLXC{Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
						Nesting: util.Pointer(false)}}},
					currentConfig: ConfigLXC{Features: featuresUnprivileged(true)},
					output:        map[string]any{"features": "mknod=1,fuse=1,keyctl=1"}},
				{name: `delete Privileged`,
					config: ConfigLXC{Features: featuresPrivileged(false)},
					currentConfig: ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						CreateDeviceNodes: util.Pointer(true),
						FUSE:              util.Pointer(false),
						NFS:               util.Pointer(false),
						Nesting:           util.Pointer(true),
						SMB:               util.Pointer(true)}}},
					output: map[string]any{"delete": "features"}},
				{name: `delete Unprivileged`,
					config: ConfigLXC{Features: featuresUnprivileged(false)},
					currentConfig: ConfigLXC{Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
						CreateDeviceNodes: util.Pointer(true),
						FUSE:              util.Pointer(false),
						KeyCtl:            util.Pointer(true),
						Nesting:           util.Pointer(true)}}},
					output: map[string]any{"delete": "features"}},
				{name: `delete no effect nil Privileged`,
					config:        ConfigLXC{Features: featuresPrivileged(false)},
					currentConfig: ConfigLXC{Features: nil},
					omitDefaults:  lxcDefaultsAll},
				{name: `delete no effect nil Unprivileged`,
					config:        ConfigLXC{Features: featuresUnprivileged(false)},
					currentConfig: ConfigLXC{Features: nil},
					omitDefaults:  lxcDefaultsAll},
				{name: `same false Privileged`,
					config:        ConfigLXC{Features: featuresPrivileged(false)},
					currentConfig: ConfigLXC{Features: featuresPrivileged(false)},
					omitDefaults:  lxcDefaultsAll},
				{name: `same false Unprivileged`,
					config:        ConfigLXC{Features: featuresUnprivileged(false)},
					currentConfig: ConfigLXC{Features: featuresUnprivileged(false)},
					omitDefaults:  lxcDefaultsAll},
				{name: `same true Privileged`,
					config:        ConfigLXC{Features: featuresPrivileged(true)},
					currentConfig: ConfigLXC{Features: featuresPrivileged(true)},
					omitDefaults:  lxcDefaultsAll},
				{name: `same true Unprivileged`,
					config:        ConfigLXC{Features: featuresUnprivileged(true)},
					currentConfig: ConfigLXC{Features: featuresUnprivileged(true)},
					omitDefaults:  lxcDefaultsAll}}},
		{category: `ID`,
			create: []lxcTestCaseAPI{
				{name: `set`,
					config: ConfigLXC{ID: util.Pointer(GuestID(15))},
					output: map[string]any{"vmid": int(15)}}},
			update: []lxcTestCaseAPI{
				{name: `do nothing`,
					config:        ConfigLXC{ID: util.Pointer(GuestID(15))},
					currentConfig: ConfigLXC{ID: util.Pointer(GuestID(0))},
					omitDefaults:  lxcDefaultsAll}}},
		{category: `Memory`,
			createUpdate: []lxcTestCaseAPI{
				{name: `set`,
					config:        ConfigLXC{Memory: util.Pointer(LxcMemory(512))},
					currentConfig: ConfigLXC{Memory: util.Pointer(LxcMemory(256))},
					output:        map[string]any{"memory": LxcMemory(512)}}},
			update: []lxcTestCaseAPI{
				{name: `same`,
					config:        ConfigLXC{Memory: util.Pointer(LxcMemory(512))},
					currentConfig: ConfigLXC{Memory: util.Pointer(LxcMemory(512))},
					omitDefaults:  lxcDefaultsAll}}},
		{category: `Mount`,
			create: []lxcTestCaseAPI{
				{name: `BindMount minimal`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID0: LxcMount{BindMount: &LxcBindMount{
							GuestPath: util.Pointer(LxcMountPath("/mnt/test-dest")),
							HostPath:  util.Pointer(LxcHostPath("/mnt/test"))}}}},
					output: map[string]any{
						"mp0": string("/mnt/test,mp=/mnt/test-dest")}},
				{name: `BindMount.Options nil`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID1: LxcMount{BindMount: &LxcBindMount{
							Options: &LxcMountOptions{}}}}},
					output: map[string]any{
						"mp1": string("")}},
				{name: `BindMount.Options.Discard true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID2: LxcMount{BindMount: &LxcBindMount{
							Options: &LxcMountOptions{
								Discard: util.Pointer(true)}}}}},
					output: map[string]any{
						"mp2": string(",mountoptions=discard")}},
				{name: `BindMount.Options.LazyTime true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID3: LxcMount{BindMount: &LxcBindMount{
							Options: &LxcMountOptions{
								LazyTime: util.Pointer(true)}}}}},
					output: map[string]any{
						"mp3": string(",mountoptions=lazytime")}},
				{name: `BindMount.Options.NoATime true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID4: LxcMount{BindMount: &LxcBindMount{
							Options: &LxcMountOptions{
								NoATime: util.Pointer(true)}}}}},
					output: map[string]any{
						"mp4": string(",mountoptions=noatime")}},
				{name: `BindMount.Options.NoDevice true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID5: LxcMount{BindMount: &LxcBindMount{
							Options: &LxcMountOptions{
								NoDevice: util.Pointer(true)}}}}},
					output: map[string]any{
						"mp5": string(",mountoptions=nodev")}},
				{name: `BindMount.Options.NoExec true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID6: LxcMount{BindMount: &LxcBindMount{
							Options: &LxcMountOptions{
								NoExec: util.Pointer(true)}}}}},
					output: map[string]any{
						"mp6": string(",mountoptions=noexec")}},
				{name: `BindMount.Options.NoSuid true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID7: LxcMount{BindMount: &LxcBindMount{
							Options: &LxcMountOptions{NoSuid: util.Pointer(true)}}}}},
					output: map[string]any{
						"mp7": string(",mountoptions=nosuid")}},
				{name: `BindMount.ReadOnly false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID8: LxcMount{BindMount: &LxcBindMount{
							ReadOnly: util.Pointer(false)}}}},
					output: map[string]any{
						"mp8": string("")}},
				{name: `BindMount.ReadOnly true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID9: LxcMount{BindMount: &LxcBindMount{
							ReadOnly: util.Pointer(true)}}}},
					output: map[string]any{
						"mp9": string(",ro=1")}},
				{name: `BindMount.Replicate false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID10: LxcMount{BindMount: &LxcBindMount{
							Replicate: util.Pointer(false)}}}},
					output: map[string]any{
						"mp10": string(",replicate=0")}},
				{name: `BindMount.Replicate true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID11: LxcMount{BindMount: &LxcBindMount{
							Replicate: util.Pointer(true)}}}},
					output: map[string]any{
						"mp11": string("")}},
				{name: `DataMount Minimal 1G`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID0: LxcMount{DataMount: &LxcDataMount{
							Storage:         util.Pointer("local-zfs"),
							SizeInKibibytes: util.Pointer(LxcMountSize(1048576))}}}},
					output: map[string]any{
						"mp0": string("local-zfs:1")}},
				{name: `DataMount Minimal < 1G`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID1: LxcMount{DataMount: &LxcDataMount{
							Storage:         util.Pointer("local-zfs"),
							SizeInKibibytes: util.Pointer(LxcMountSize(1048000))}}}},
					output: map[string]any{
						"mp1": string("local-zfs:0.001")}},
				{name: `DataMount Minimal round down`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID2: LxcMount{DataMount: &LxcDataMount{
							Storage:         util.Pointer("local-zfs"),
							SizeInKibibytes: util.Pointer(LxcMountSize(2100000))}}}},
					output: map[string]any{
						"mp2": string("local-zfs:2")}},
				{name: `DataMount.ACL false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID5: LxcMount{DataMount: baseDataMount(LxcDataMount{
							ACL: util.Pointer(TriBoolFalse)})}}},
					output: map[string]any{
						"mp5": string("local-zfs:1,acl=0")}},
				{name: `DataMount.ACL none`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID3: LxcMount{DataMount: baseDataMount(LxcDataMount{
							ACL: util.Pointer(TriBoolNone)})}}},
					output: map[string]any{
						"mp3": string("local-zfs:1")}},
				{name: `DataMount.ACL true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID4: LxcMount{DataMount: baseDataMount(LxcDataMount{
							ACL: util.Pointer(TriBoolTrue)})}}},
					output: map[string]any{
						"mp4": string("local-zfs:1,acl=1")}},
				{name: `DataMount.Backup false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID7: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Backup: util.Pointer(false)})}}},
					output: map[string]any{
						"mp7": string("local-zfs:1")}},
				{name: `DataMount.Backup true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID6: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Backup: util.Pointer(true)})}}},
					output: map[string]any{
						"mp6": string("local-zfs:1,backup=1")}},
				{name: `DataMount.Options nil`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID8: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: &LxcMountOptions{}})}}},
					output: map[string]any{
						"mp8": string("local-zfs:1")}},
				{name: `DataMount.Options true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID9: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: &LxcMountOptions{
								Discard:  util.Pointer(true),
								LazyTime: util.Pointer(true),
								NoATime:  util.Pointer(true),
								NoDevice: util.Pointer(true),
								NoExec:   util.Pointer(true),
								NoSuid:   util.Pointer(true)}})}}},
					output: map[string]any{
						"mp9": string("local-zfs:1,mountoptions=discard;lazytime;noatime;nodev;noexec;nosuid")}},
				{name: `DataMount.Options false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID10: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: &LxcMountOptions{
								Discard:  util.Pointer(false),
								LazyTime: util.Pointer(false),
								NoATime:  util.Pointer(false),
								NoDevice: util.Pointer(false),
								NoExec:   util.Pointer(false),
								NoSuid:   util.Pointer(false)}})}}},
					output: map[string]any{
						"mp10": string("local-zfs:1")}},
				{name: `DataMount.Options.Discard true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID11: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: &LxcMountOptions{Discard: util.Pointer(true)}})}}},
					output: map[string]any{
						"mp11": string("local-zfs:1,mountoptions=discard")}},
				{name: `DataMount.Options.LazyTime true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID12: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: &LxcMountOptions{LazyTime: util.Pointer(true)}})}}},
					output: map[string]any{
						"mp12": string("local-zfs:1,mountoptions=lazytime")}},
				{name: `DataMount.Options.NoATime true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID13: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: &LxcMountOptions{NoATime: util.Pointer(true)}})}}},
					output: map[string]any{
						"mp13": string("local-zfs:1,mountoptions=noatime")}},
				{name: `DataMount.Options.NoDevice true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID14: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: &LxcMountOptions{NoDevice: util.Pointer(true)}})}}},
					output: map[string]any{
						"mp14": string("local-zfs:1,mountoptions=nodev")}},
				{name: `DataMount.Options.NoExec true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID15: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: &LxcMountOptions{NoExec: util.Pointer(true)}})}}},
					output: map[string]any{
						"mp15": string("local-zfs:1,mountoptions=noexec")}},
				{name: `DataMount.Options.NoSuid true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID16: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: &LxcMountOptions{NoSuid: util.Pointer(true)}})}}},
					output: map[string]any{
						"mp16": string("local-zfs:1,mountoptions=nosuid")}},
				{name: `DataMount.Path`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID17: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Path: util.Pointer(LxcMountPath("/mnt/test"))})}}},
					output: map[string]any{
						"mp17": string("local-zfs:1,mp=/mnt/test")}},
				{name: `DataMount.Quota false Privileged false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID19: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota: util.Pointer(false)})}},
						Privileged: util.Pointer(false)},
					output: map[string]any{
						"mp19": string("local-zfs:1")}},
				{name: `DataMount.Quota false Privileged true`,
					omitDefaults: lxcDefaultsAll,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID20: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota: util.Pointer(false)})}},
						Privileged: util.Pointer(true)},
					body: map[string]string{"unprivileged": "0"},
					output: map[string]any{
						"mp20": string("local-zfs:1")}},
				{name: `DataMount.Quota false Privileged unset`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID18: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota: util.Pointer(false)})}}},
					output: map[string]any{
						"mp18": string("local-zfs:1")}},
				{name: `DataMount.Quota true Privileged false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID22: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota: util.Pointer(true)})}},
						Privileged: util.Pointer(false)},
					output: map[string]any{
						"mp22": string("local-zfs:1")}},
				{name: `DataMount.Quota true Privileged true`,
					omitDefaults: lxcDefaultsAll,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID23: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota: util.Pointer(true)})}},
						Privileged: util.Pointer(true)},
					body: map[string]string{"unprivileged": "0"},
					output: map[string]any{
						"mp23": string("local-zfs:1,quota=1")}},
				{name: `DataMount.Quota true Privileged unset`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID24: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota: util.Pointer(true)})}}},
					output: map[string]any{
						"mp24": string("local-zfs:1")}},
				{name: `DataMount.ReadOnly false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID25: LxcMount{DataMount: baseDataMount(LxcDataMount{
							ReadOnly: util.Pointer(false)})}}},
					output: map[string]any{
						"mp25": string("local-zfs:1")}},
				{name: `DataMount.ReadOnly true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID26: LxcMount{DataMount: baseDataMount(LxcDataMount{
							ReadOnly: util.Pointer(true)})}}},
					output: map[string]any{
						"mp26": string("local-zfs:1,ro=1")}},
				{name: `DataMount.Replicate false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID27: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Replicate: util.Pointer(false)})}}},
					output: map[string]any{
						"mp27": string("local-zfs:1,replicate=0")}},
				{name: `DataMount.Replicate true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID28: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Replicate: util.Pointer(true)})}}},
					output: map[string]any{
						"mp28": string("local-zfs:1")}}},
			createUpdate: []lxcTestCaseAPI{
				{name: `Detach non-existing`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID29: LxcMount{Detach: true}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID30: LxcMount{}}},
					omitDefaults: lxcDefaultsUpdate}},
			update: []lxcTestCaseAPI{
				{name: `BindMount Detach existing`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID49: LxcMount{Detach: true}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID49: LxcMount{BindMount: &LxcBindMount{}}}},
					output: map[string]any{"delete": string("mp49")}},
				{name: `BindMount.GuestPath replace`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID50: LxcMount{BindMount: &LxcBindMount{
							GuestPath: util.Pointer(LxcMountPath("/mnt/test-dest"))}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID50: LxcMount{BindMount: bindMount()}}},
					output: map[string]any{"mp50": string("/mnt/data,mp=/mnt/test-dest,mountoptions=noatime;nodev;noexec;nosuid,ro=1,replicate=0")}},
				{name: `BindMount.HostPath replace`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID51: LxcMount{BindMount: &LxcBindMount{
							HostPath: util.Pointer(LxcHostPath("/mnt/new-data"))}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID51: LxcMount{BindMount: bindMount()}}},
					output: map[string]any{"mp51": string("/mnt/new-data,mp=/mnt/bind,mountoptions=noatime;nodev;noexec;nosuid,ro=1,replicate=0")}},
				{name: `BindMount.Options.Discard replace true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID52: LxcMount{BindMount: &LxcBindMount{
							Options: &LxcMountOptions{
								Discard: util.Pointer(true)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID52: LxcMount{BindMount: bindMount()}}},
					output: map[string]any{"mp52": string("/mnt/data,mp=/mnt/bind,mountoptions=discard;noatime;nodev;noexec;nosuid,ro=1,replicate=0")}},
				{name: `BindMount.Options.LazyTime replace true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID53: LxcMount{BindMount: &LxcBindMount{
							Options: &LxcMountOptions{
								LazyTime: util.Pointer(true)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID53: LxcMount{BindMount: bindMount()}}},
					output: map[string]any{"mp53": string("/mnt/data,mp=/mnt/bind,mountoptions=lazytime;noatime;nodev;noexec;nosuid,ro=1,replicate=0")}},
				{name: `BindMount.Options.NoATime replace false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID54: LxcMount{BindMount: &LxcBindMount{
							Options: &LxcMountOptions{
								NoATime: util.Pointer(false)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID54: LxcMount{BindMount: bindMount()}}},
					output: map[string]any{"mp54": string("/mnt/data,mp=/mnt/bind,mountoptions=nodev;noexec;nosuid,ro=1,replicate=0")}},
				{name: `BindMount.Options.NoDevice replace false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID55: LxcMount{BindMount: &LxcBindMount{
							Options: &LxcMountOptions{
								NoDevice: util.Pointer(false)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID55: LxcMount{BindMount: bindMount()}}},
					output: map[string]any{"mp55": string("/mnt/data,mp=/mnt/bind,mountoptions=noatime;noexec;nosuid,ro=1,replicate=0")}},
				{name: `BindMount.Options.NoExec replace false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID56: LxcMount{BindMount: &LxcBindMount{
							Options: &LxcMountOptions{
								NoExec: util.Pointer(false)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID56: LxcMount{BindMount: bindMount()}}},
					output: map[string]any{"mp56": string("/mnt/data,mp=/mnt/bind,mountoptions=noatime;nodev;nosuid,ro=1,replicate=0")}},
				{name: `BindMount.Options.NoSuid replace false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID57: LxcMount{BindMount: &LxcBindMount{
							Options: &LxcMountOptions{
								NoSuid: util.Pointer(false)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID57: LxcMount{BindMount: bindMount()}}},
					output: map[string]any{"mp57": string("/mnt/data,mp=/mnt/bind,mountoptions=noatime;nodev;noexec,ro=1,replicate=0")}},
				{name: `BindMount.Options all true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID96: LxcMount{BindMount: &LxcBindMount{
							Options: &LxcMountOptions{
								Discard:  util.Pointer(true),
								LazyTime: util.Pointer(true),
								NoATime:  util.Pointer(true),
								NoDevice: util.Pointer(true),
								NoExec:   util.Pointer(true),
								NoSuid:   util.Pointer(true)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID96: LxcMount{BindMount: &LxcBindMount{}}}},
					output: map[string]any{
						"mp96": string(",mountoptions=discard;lazytime;noatime;nodev;noexec;nosuid")}},
				{name: `BindMount.ReadOnly replace false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID58: LxcMount{BindMount: &LxcBindMount{
							ReadOnly: util.Pointer(false)}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID58: LxcMount{BindMount: bindMount()}}},
					output: map[string]any{"mp58": string("/mnt/data,mp=/mnt/bind,mountoptions=noatime;nodev;noexec;nosuid,replicate=0")}},
				{name: `BindMount.Replicate replace true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID59: LxcMount{BindMount: &LxcBindMount{
							Replicate: util.Pointer(true)}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID59: LxcMount{BindMount: bindMount()}}},
					output: map[string]any{"mp59": string("/mnt/data,mp=/mnt/bind,mountoptions=noatime;nodev;noexec;nosuid,ro=1")}},
				{name: `BindMount no change`,
					config:        ConfigLXC{Mounts: LxcMounts{LxcMountID110: LxcMount{BindMount: bindMount()}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID110: LxcMount{BindMount: bindMount()}}},
					omitDefaults:  lxcDefaultsAll},
				{name: `BindMount over DataMount`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID130: LxcMount{BindMount: &LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/mnt/data")),
							GuestPath: util.Pointer(LxcMountPath("/mnt/test-dest"))}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID130: LxcMount{DataMount: &LxcDataMount{
							SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
							Storage:         util.Pointer("local-zfs")}}}},
					output: map[string]any{"mp130": string("/mnt/data,mp=/mnt/test-dest")}},
				{name: `BindMount create`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID131: LxcMount{BindMount: &LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/mnt/data")),
							GuestPath: util.Pointer(LxcMountPath("/mnt/test-dest"))}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID135: LxcMount{}}},
					output: map[string]any{"mp131": string("/mnt/data,mp=/mnt/test-dest")}},
				{name: `DataMount Detach existing`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID50: LxcMount{Detach: true}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID50: LxcMount{DataMount: dataMount()}}},
					output: map[string]any{"delete": string("mp50")}},
				{name: `DataMount.ACL replace true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID60: LxcMount{DataMount: &LxcDataMount{
							ACL: util.Pointer(TriBoolTrue)}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID60: LxcMount{DataMount: dataMount()}}},
					output: map[string]any{
						"mp60": string("local-zfs:subvol-101-disk-0,size=1G,acl=1,backup=1,mountoptions=discard;lazytime;noatime;nosuid,mp=/mnt/data,ro=1,replicate=0")}},
				{name: `DataMount.Backup replace false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID61: LxcMount{DataMount: &LxcDataMount{
							Backup: util.Pointer(false)}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID61: LxcMount{DataMount: dataMount()}}},
					output: map[string]any{
						"mp61": string("local-zfs:subvol-101-disk-0,size=1G,acl=0,mountoptions=discard;lazytime;noatime;nosuid,mp=/mnt/data,ro=1,replicate=0")}},
				{name: `DataMount.Options.Discard replace false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID62: LxcMount{DataMount: &LxcDataMount{
							Options: &LxcMountOptions{
								Discard: util.Pointer(false)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID62: LxcMount{DataMount: dataMount()}}},
					output: map[string]any{
						"mp62": string("local-zfs:subvol-101-disk-0,size=1G,acl=0,backup=1,mountoptions=lazytime;noatime;nosuid,mp=/mnt/data,ro=1,replicate=0")}},
				{name: `DataMount.Options.LazyTime replace false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID63: LxcMount{DataMount: &LxcDataMount{
							Options: &LxcMountOptions{
								LazyTime: util.Pointer(false)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID63: LxcMount{DataMount: dataMount()}}},
					output: map[string]any{
						"mp63": string("local-zfs:subvol-101-disk-0,size=1G,acl=0,backup=1,mountoptions=discard;noatime;nosuid,mp=/mnt/data,ro=1,replicate=0")}},
				{name: `DataMount.Options.NoATime replace false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID64: LxcMount{DataMount: &LxcDataMount{
							Options: &LxcMountOptions{
								NoATime: util.Pointer(false)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID64: LxcMount{DataMount: dataMount()}}},
					output: map[string]any{
						"mp64": string("local-zfs:subvol-101-disk-0,size=1G,acl=0,backup=1,mountoptions=discard;lazytime;nosuid,mp=/mnt/data,ro=1,replicate=0")}},
				{name: `DataMount.Options.NoDevice replace true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID65: LxcMount{DataMount: &LxcDataMount{
							Options: &LxcMountOptions{
								NoDevice: util.Pointer(true)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID65: LxcMount{DataMount: dataMount()}}},
					output: map[string]any{
						"mp65": string("local-zfs:subvol-101-disk-0,size=1G,acl=0,backup=1,mountoptions=discard;lazytime;noatime;nodev;nosuid,mp=/mnt/data,ro=1,replicate=0")}},
				{name: `DataMount.Options.NoExec replace true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID66: LxcMount{DataMount: &LxcDataMount{
							Options: &LxcMountOptions{
								NoExec: util.Pointer(true)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID66: LxcMount{DataMount: dataMount()}}},
					output: map[string]any{
						"mp66": string("local-zfs:subvol-101-disk-0,size=1G,acl=0,backup=1,mountoptions=discard;lazytime;noatime;noexec;nosuid,mp=/mnt/data,ro=1,replicate=0")}},
				{name: `DataMount.Options.NoSuid replace true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID67: LxcMount{DataMount: &LxcDataMount{
							Options: &LxcMountOptions{
								NoSuid: util.Pointer(false)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID67: LxcMount{DataMount: dataMount()}}},
					output: map[string]any{
						"mp67": string("local-zfs:subvol-101-disk-0,size=1G,acl=0,backup=1,mountoptions=discard;lazytime;noatime,mp=/mnt/data,ro=1,replicate=0")}},
				{name: `DataMount.Options all false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID68: LxcMount{DataMount: &LxcDataMount{
							Options: &LxcMountOptions{
								Discard:  util.Pointer(false),
								LazyTime: util.Pointer(false),
								NoATime:  util.Pointer(false),
								NoDevice: util.Pointer(false),
								NoExec:   util.Pointer(false),
								NoSuid:   util.Pointer(false)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID68: LxcMount{DataMount: dataMount()}}},
					output: map[string]any{
						"mp68": string("local-zfs:subvol-101-disk-0,size=1G,acl=0,backup=1,mp=/mnt/data,ro=1,replicate=0")}},
				{name: `DataMount.Options all true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID90: LxcMount{DataMount: &LxcDataMount{
							Options: &LxcMountOptions{
								Discard:  util.Pointer(true),
								LazyTime: util.Pointer(true),
								NoATime:  util.Pointer(true),
								NoDevice: util.Pointer(true),
								NoExec:   util.Pointer(true),
								NoSuid:   util.Pointer(true)}}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID90: LxcMount{DataMount: &LxcDataMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
						rawDisk:         "local-zfs:subvol-101-disk-0"}}}},
					output: map[string]any{
						"mp90": string("local-zfs:subvol-101-disk-0,size=1G,mountoptions=discard;lazytime;noatime;nodev;noexec;nosuid")}},
				{name: `DataMount.Path replace`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID69: LxcMount{DataMount: &LxcDataMount{
							Path: util.Pointer(LxcMountPath("/opt/test"))}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID69: LxcMount{DataMount: dataMount()}}},
					output: map[string]any{
						"mp69": string("local-zfs:subvol-101-disk-0,size=1G,acl=0,backup=1,mountoptions=discard;lazytime;noatime;nosuid,mp=/opt/test,ro=1,replicate=0")}},
				{name: `DataMount.Quota replace false Privileged false`,
					omitDefaults: lxcDefaultsAll,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID80: LxcMount{DataMount: &LxcDataMount{
							Quota: util.Pointer(false)}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID80: LxcMount{DataMount: dataMount()}}}},
				{name: `DataMount.Quota replace false Privileged true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID81: LxcMount{DataMount: &LxcDataMount{
							Quota: util.Pointer(false)}}}},
					currentConfig: ConfigLXC{
						Privileged: util.Pointer(true),
						Mounts: LxcMounts{LxcMountID81: LxcMount{DataMount: &LxcDataMount{
							Storage:         util.Pointer("local-zfs"),
							SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
							ACL:             util.Pointer(TriBoolFalse),
							Backup:          util.Pointer(true),
							Quota:           util.Pointer(true),
							rawDisk:         "local-zfs:subvol-101-disk-0"}}}},
					output: map[string]any{
						"mp81": string("local-zfs:subvol-101-disk-0,size=1G,acl=0,backup=1")}},
				{name: `DataMount.Quota replace true Privileged false`,
					omitDefaults: lxcDefaultsAll,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID82: LxcMount{DataMount: &LxcDataMount{
							Quota: util.Pointer(true)}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID82: LxcMount{DataMount: dataMount()}}}},
				{name: `DataMount.Quota replace true Privileged true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID83: LxcMount{DataMount: &LxcDataMount{
							Quota: util.Pointer(true)}}}},
					currentConfig: ConfigLXC{
						Privileged: util.Pointer(true),
						Mounts: LxcMounts{LxcMountID83: LxcMount{DataMount: &LxcDataMount{
							Storage:         util.Pointer("local-zfs"),
							SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
							ACL:             util.Pointer(TriBoolFalse),
							Backup:          util.Pointer(true),
							Quota:           util.Pointer(false),
							rawDisk:         "local-zfs:subvol-101-disk-0"}}}},
					output: map[string]any{
						"mp83": string("local-zfs:subvol-101-disk-0,size=1G,acl=0,backup=1,quota=1")}},
				{name: `DataMount.ReadOnly replace false`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID70: LxcMount{DataMount: &LxcDataMount{
							ReadOnly: util.Pointer(false)}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID70: LxcMount{DataMount: dataMount()}}},
					output: map[string]any{
						"mp70": string("local-zfs:subvol-101-disk-0,size=1G,acl=0,backup=1,mountoptions=discard;lazytime;noatime;nosuid,mp=/mnt/data,replicate=0")}},
				{name: `DataMount.Replicate replace true`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID71: LxcMount{DataMount: &LxcDataMount{
							Replicate: util.Pointer(true)}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID71: LxcMount{DataMount: dataMount()}}},
					output: map[string]any{
						"mp71": string("local-zfs:subvol-101-disk-0,size=1G,acl=0,backup=1,mountoptions=discard;lazytime;noatime;nosuid,mp=/mnt/data,ro=1")}},
				{name: `DataMount.SizeInKibibytes increase, all other unchanged`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID72: LxcMount{DataMount: &LxcDataMount{
							SizeInKibibytes: util.Pointer(LxcMountSize(7340032))}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID72: LxcMount{DataMount: dataMount()}}},
					omitDefaults:  lxcDefaultsAll},
				{name: `DataMount.Storage change, all other unchanged`, // this is manaaged by other mechanisms
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID73: LxcMount{DataMount: &LxcDataMount{
							Storage: util.Pointer("test-storage")}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID73: LxcMount{DataMount: dataMount()}}},
					omitDefaults:  lxcDefaultsAll},
				{name: `DataMount over BindMount`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID131: LxcMount{DataMount: &LxcDataMount{
							SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
							Storage:         util.Pointer("local-zfs"),
						}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID131: LxcMount{BindMount: &LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/mnt/data")),
							GuestPath: util.Pointer(LxcMountPath("/mnt/test-dest"))}}}},
					output: map[string]any{"mp131": string("local-zfs:1")}},
				{name: `DataMount create`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID132: LxcMount{DataMount: &LxcDataMount{
							SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
							Storage:         util.Pointer("local-zfs"),
							ACL:             util.Pointer(TriBoolFalse),
							Path:            util.Pointer(LxcMountPath("/mnt/test-dest")),
							Backup:          util.Pointer(true)}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID135: LxcMount{}}},
					output: map[string]any{"mp132": string("local-zfs:1,acl=0,backup=1,mp=/mnt/test-dest")}},
				{name: `DataMount recreate`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID133: LxcMount{DataMount: &LxcDataMount{
							SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
							Storage:         util.Pointer("local-zfs")}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID133: LxcMount{DataMount: &LxcDataMount{
							SizeInKibibytes: util.Pointer(LxcMountSize(5242880)),
							Storage:         util.Pointer("local-lvm")}}}},
					output: map[string]any{"mp133": string("local-zfs:1")}},
				{name: `DataMount recreate inherit Storage`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID134: LxcMount{DataMount: &LxcDataMount{
							SizeInKibibytes: util.Pointer(LxcMountSize(1048576))}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID134: LxcMount{DataMount: &LxcDataMount{
							SizeInKibibytes: util.Pointer(LxcMountSize(5242880)),
							Storage:         util.Pointer("local-lvm")}}}},
					output: map[string]any{"mp134": string("local-lvm:1")}}}},
		{category: `Name`,
			createUpdate: []lxcTestCaseAPI{
				{name: `set`,
					config:        ConfigLXC{Name: util.Pointer(GuestName("test"))},
					currentConfig: ConfigLXC{Name: util.Pointer(GuestName("text"))},
					output:        map[string]any{"hostname": string("test")}}},
			update: []lxcTestCaseAPI{
				{name: `do nothing`,
					config:        ConfigLXC{Name: util.Pointer(GuestName("test"))},
					currentConfig: ConfigLXC{Name: util.Pointer(GuestName("test"))},
					omitDefaults:  lxcDefaultsAll}}},
		{category: `Node`,
			createUpdate: []lxcTestCaseAPI{
				{name: `do nothing`,
					config:        ConfigLXC{Node: util.Pointer(NodeName("test"))},
					currentConfig: ConfigLXC{Node: util.Pointer(NodeName("text"))},
					omitDefaults:  lxcDefaultsUpdate}}},
		{category: `OperatingSystem`,
			createUpdate: []lxcTestCaseAPI{
				{name: `do nothing`,
					config:        ConfigLXC{OperatingSystem: "test"},
					currentConfig: ConfigLXC{OperatingSystem: "text"},
					omitDefaults:  lxcDefaultsUpdate}}},
		{category: `Pool`,
			create: []lxcTestCaseAPI{
				{name: `set`,
					config: ConfigLXC{Pool: util.Pointer(PoolName("test"))},
					output: map[string]any{
						"pool": "test"},
					pool: "test"}},
			update: []lxcTestCaseAPI{
				{name: `do nothing`,
					config:        ConfigLXC{Pool: util.Pointer(PoolName("test"))},
					currentConfig: ConfigLXC{Pool: util.Pointer(PoolName("text"))},
					omitDefaults:  lxcDefaultsAll}}},
		{category: `Privileged`,
			create: []lxcTestCaseAPI{
				{name: `true`,
					config:       ConfigLXC{Privileged: util.Pointer(true)},
					body:         map[string]string{"unprivileged": "0"},
					omitDefaults: lxcDefaultsAll},
				{name: `false`,
					config: ConfigLXC{Privileged: util.Pointer(false)}}},
			update: []lxcTestCaseAPI{
				{name: `true no effect`,
					config:        ConfigLXC{Privileged: util.Pointer(true)},
					currentConfig: ConfigLXC{Privileged: util.Pointer(false)},
					omitDefaults:  lxcDefaultsAll},
				{name: `false no effect`,
					config:        ConfigLXC{Privileged: util.Pointer(false)},
					currentConfig: ConfigLXC{Privileged: util.Pointer(true)},
					omitDefaults:  lxcDefaultsAll}}},
		{category: `Protection`,
			create: []lxcTestCaseAPI{
				{name: `set false`,
					config:        ConfigLXC{Protection: util.Pointer(false)},
					currentConfig: ConfigLXC{}}},
			createUpdate: []lxcTestCaseAPI{
				{name: `set true`,
					config:        ConfigLXC{Protection: util.Pointer(true)},
					currentConfig: ConfigLXC{},
					output:        map[string]any{"protection": string("1")}}},
			update: []lxcTestCaseAPI{
				{name: `do nothing false`,
					config:        ConfigLXC{Protection: util.Pointer(false)},
					currentConfig: ConfigLXC{Protection: util.Pointer(false)},
					omitDefaults:  lxcDefaultsAll},
				{name: `do nothing true`,
					config:        ConfigLXC{Protection: util.Pointer(true)},
					currentConfig: ConfigLXC{Protection: util.Pointer(true)},
					omitDefaults:  lxcDefaultsAll},
				{name: `replace false`,
					config:        ConfigLXC{Protection: util.Pointer(false)},
					currentConfig: ConfigLXC{Protection: util.Pointer(true)},
					output:        map[string]any{"delete": string("protection")}},
				{name: `replace true`,
					config:        ConfigLXC{Protection: util.Pointer(true)},
					currentConfig: ConfigLXC{Protection: util.Pointer(false)},
					output:        map[string]any{"protection": string("1")}},
				{name: `set false`,
					config:        ConfigLXC{Protection: util.Pointer(false)},
					currentConfig: ConfigLXC{},
					output:        map[string]any{"delete": string("protection")}},
			}},
		{category: `StartAtNodeBoot`,
			create: []lxcTestCaseAPI{
				{name: `true`,
					config: ConfigLXC{StartAtNodeBoot: util.Pointer(true)},
					output: map[string]any{"onboot": int(1)}},
				{name: `false`,
					config: ConfigLXC{StartAtNodeBoot: util.Pointer(false)}}},
			update: []lxcTestCaseAPI{
				{name: `true to false`,
					config: ConfigLXC{StartAtNodeBoot: util.Pointer(false)},
					currentConfig: ConfigLXC{
						StartAtNodeBoot: util.Pointer(true)},
					output: map[string]any{"delete": string("onboot")}},
				{name: `false to true`,
					config: ConfigLXC{StartAtNodeBoot: util.Pointer(true)},
					currentConfig: ConfigLXC{
						StartAtNodeBoot: util.Pointer(false)},
					output: map[string]any{"onboot": int(1)}},
				{name: `true to true`,
					config: ConfigLXC{StartAtNodeBoot: util.Pointer(true)},
					currentConfig: ConfigLXC{
						StartAtNodeBoot: util.Pointer(true)},
					omitDefaults: lxcDefaultsAll},
				{name: `false to false`,
					config: ConfigLXC{StartAtNodeBoot: util.Pointer(false)},
					currentConfig: ConfigLXC{
						StartAtNodeBoot: util.Pointer(false)},
					omitDefaults: lxcDefaultsAll}}},
		{category: `StartupShutdown`,
			create: []lxcTestCaseAPI{
				{name: `Order no effect`,
					config: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						Order: util.Pointer(GuestStartupOrder(-1))}}},
				{name: `ShutdownTimeout no effect`,
					config: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						ShutdownTimeout: util.Pointer(TimeDuration(-1))}}},
				{name: `StartupDelay no effect`,
					config: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						StartupDelay: util.Pointer(TimeDuration(-1))}}}},
			createUpdate: []lxcTestCaseAPI{
				{name: `Order`,
					config: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						Order: util.Pointer(GuestStartupOrder(10))}},
					currentConfig: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						Order: util.Pointer(GuestStartupOrder(20))}},
					output: map[string]any{"startup": string("order=10")}},
				{name: `ShutdownTimeout`,
					config: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						ShutdownTimeout: util.Pointer(TimeDuration(74530))}},
					currentConfig: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						ShutdownTimeout: util.Pointer(TimeDuration(9362))}},
					output: map[string]any{"startup": string("down=74530")}},
				{name: `StartupDelay`,
					config: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						StartupDelay: util.Pointer(TimeDuration(200))}},
					currentConfig: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						StartupDelay: util.Pointer(TimeDuration(50))}},
					output: map[string]any{"startup": string("up=200")}},
				{name: `all`,
					config: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						Order:           util.Pointer(GuestStartupOrder(10)),
						StartupDelay:    util.Pointer(TimeDuration(200)),
						ShutdownTimeout: util.Pointer(TimeDuration(74530))}},
					currentConfig: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						Order:           util.Pointer(GuestStartupOrder(20)),
						StartupDelay:    util.Pointer(TimeDuration(50)),
						ShutdownTimeout: util.Pointer(TimeDuration(9362))}},
					output: map[string]any{
						"startup": string("order=10,up=200,down=74530")}}},
			update: []lxcTestCaseAPI{
				{name: `Order unset`,
					config: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						Order: util.Pointer(GuestStartupOrder(-1))}},
					currentConfig: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						Order:           util.Pointer(GuestStartupOrder(20)),
						ShutdownTimeout: util.Pointer(TimeDuration(9362)),
						StartupDelay:    util.Pointer(TimeDuration(50))}},
					output: map[string]any{"startup": string("up=50,down=9362")}},
				{name: `ShutdownTimeout unset`,
					config: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						ShutdownTimeout: util.Pointer(TimeDuration(-1))}},
					currentConfig: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						Order:           util.Pointer(GuestStartupOrder(20)),
						ShutdownTimeout: util.Pointer(TimeDuration(9362)),
						StartupDelay:    util.Pointer(TimeDuration(50))}},
					output: map[string]any{"startup": string("order=20,up=50")}},
				{name: `StartupDelay unset`,
					config: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						StartupDelay: util.Pointer(TimeDuration(-1))}},
					currentConfig: ConfigLXC{StartupShutdown: &StartupAndShutdown{
						Order:           util.Pointer(GuestStartupOrder(20)),
						ShutdownTimeout: util.Pointer(TimeDuration(9362)),
						StartupDelay:    util.Pointer(TimeDuration(50))}},
					output: map[string]any{"startup": string("order=20,down=9362")}}}},
		{category: `Swap`,
			createUpdate: []lxcTestCaseAPI{
				{name: `set`,
					config:        ConfigLXC{Swap: util.Pointer(LxcSwap(256))},
					currentConfig: ConfigLXC{Swap: util.Pointer(LxcSwap(128))},
					output:        map[string]any{"swap": int(256)}},
				{name: `set 0`,
					config:        ConfigLXC{Swap: util.Pointer(LxcSwap(0))},
					currentConfig: ConfigLXC{Swap: util.Pointer(LxcSwap(128))},
					output:        map[string]any{"swap": int(0)}}},
			update: []lxcTestCaseAPI{
				{name: `do nothing`,
					config:        ConfigLXC{Swap: util.Pointer(LxcSwap(256))},
					currentConfig: ConfigLXC{Swap: util.Pointer(LxcSwap(256))},
					omitDefaults:  lxcDefaultsAll}}},
		{category: `Tags`,
			create: []lxcTestCaseAPI{
				{name: `set`,
					config: ConfigLXC{Tags: &Tags{"test", "aaa"}},
					body:   map[string]string{"tags": "test,aaa"}},
				{name: `do nothing`,
					config: ConfigLXC{Tags: &Tags{}}}},
			update: []lxcTestCaseAPI{
				{name: `create`,
					config:       ConfigLXC{Tags: &Tags{"test", "aaa"}},
					body:         map[string]string{"tags": "test,aaa"},
					omitDefaults: lxcDefaultsAll},
				{name: `empty`,
					config:        ConfigLXC{Tags: &Tags{}},
					currentConfig: ConfigLXC{Tags: &Tags{"text", "test"}},
					body:          map[string]string{"tags": ""},
					omitDefaults:  lxcDefaultsAll},
				{name: `same, do nothing`,
					config:        ConfigLXC{Tags: &Tags{"test", "text"}},
					currentConfig: ConfigLXC{Tags: &Tags{"text", "test"}},
					omitDefaults:  lxcDefaultsAll},
				{name: `set`,
					config:        ConfigLXC{Tags: &Tags{"test", "aaa"}},
					currentConfig: ConfigLXC{Tags: &Tags{"text", "aaa"}},
					body:          map[string]string{"tags": "aaa,test"},
					omitDefaults:  lxcDefaultsAll},
				{name: `different order, do nothing`,
					config:        ConfigLXC{Tags: &Tags{"bbb", "aaa", "ccc"}},
					currentConfig: ConfigLXC{Tags: &Tags{"aaa", "ccc", "bbb"}},
					omitDefaults:  lxcDefaultsAll}}},
	}
	for _, test := range tests {
		for _, subTest := range append(test.create, test.createUpdate...) {
			name := test.category + "/Create/" + subTest.name
			t.Run(name, func(*testing.T) {
				tmpParams, tmpBody, pool := subTest.config.mapToApiCreate()
				clone := maps.Clone(subTest.body)
				if !(subTest.omitDefaults == lxcDefaultsAll || subTest.omitDefaults == lxcDefaultsCreate) {
					if clone == nil {
						clone = map[string]string{}
					}
					if _, isSet := clone["unprivileged"]; !isSet {
						clone["unprivileged"] = "1" // Default to unprivileged
					}
				}
				require.Equal(t, subTest.output, tmpParams, name)
				testParamsEqual(t, clone, tmpBody, name+" Body")
				require.Equal(t, subTest.pool, pool, name)
			})
		}
		for _, subTest := range append(test.update, test.createUpdate...) {
			name := test.category + "/Update/" + subTest.name
			t.Run(name, func(*testing.T) {
				tmpParams, tmpBody := subTest.config.mapToApiUpdate(subTest.currentConfig)
				clone := maps.Clone(subTest.output)
				if !(subTest.omitDefaults == lxcDefaultsAll || subTest.omitDefaults == lxcDefaultsUpdate) {
					if clone == nil {
						clone = map[string]any{}
					}
					if _, isSet := clone["digest"]; !isSet {
						clone["digest"] = "" // set empty digest
					}
				}
				require.Equal(t, clone, tmpParams, name)
				testParamsEqual(t, subTest.body, tmpBody, name+" Body")
			})
		}
	}
}

func Test_ConfigLXC_Validate(t *testing.T) {
	t.Parallel()
	baseConfig := func(config ConfigLXC) ConfigLXC {
		if config.BootMount == nil {
			config.BootMount = &LxcBootMount{
				SizeInKibibytes: util.Pointer(LxcMountSize(131072)),
				Storage:         util.Pointer("local-lvm")}
		}
		if config.CreateOptions == nil {
			config.CreateOptions = &LxcCreateOptions{
				OsTemplate: &LxcTemplate{
					Storage: "local",
					File:    "test-template"}}
		}
		return config
	}
	baseDataMount := func(config LxcDataMount) *LxcDataMount {
		if config.Path == nil {
			config.Path = util.Pointer(LxcMountPath("/mnt/test"))
		}
		if config.SizeInKibibytes == nil {
			config.SizeInKibibytes = util.Pointer(LxcMountSize(lxcMountSizeMinimum))
		}
		if config.Storage == nil {
			config.Storage = util.Pointer(string("test"))
		}
		return &config
	}
	publicKeys := func() []AuthorizedKey {
		data := test_data_guest.AuthorizedKey_Decoded_Input()
		keys := make([]AuthorizedKey, len(data))
		for i := range data {
			keys[i] = AuthorizedKey{Options: data[i].Options, PublicKey: data[i].PublicKey, Comment: data[i].Comment}
		}
		return keys
	}
	type test struct {
		name    string
		input   ConfigLXC
		current *ConfigLXC
		version Version
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
						current: &ConfigLXC{BootMount: &LxcBootMount{Storage: util.Pointer("text")}}},
					{name: `Quota`,
						input: baseConfig(ConfigLXC{
							BootMount: &LxcBootMount{
								SizeInKibibytes: util.Pointer(LxcMountSize(150000)),
								Storage:         util.Pointer("test"),
								Quota:           util.Pointer(true)},
							Privileged: util.Pointer(true)}),
						current: &ConfigLXC{
							BootMount:  &LxcBootMount{Storage: util.Pointer("text")},
							Privileged: util.Pointer(true)}}}},
			invalid: testType{
				create: []test{
					{name: `errors.New(ConfigLXC_Error_BootMountMissing)`,
						input: ConfigLXC{},
						err:   errors.New(ConfigLXC_Error_BootMountMissing)},
					{name: `errors.New(LxcBootMount_Error_NoStorageDuringCreation)`,
						input: ConfigLXC{BootMount: &LxcBootMount{}},
						err:   errors.New(LxcBootMount_Error_NoStorageDuringCreation)},
					{name: `errors.New(LxcBootMount_Error_NoSizeDuringCreation)`,
						input: ConfigLXC{BootMount: &LxcBootMount{
							Storage: util.Pointer("local-lvm")}},
						err: errors.New(LxcBootMount_Error_NoSizeDuringCreation)}},
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
							SizeInKibibytes: util.Pointer(lxcMountSizeMinimum - 1)}}),
						current: &ConfigLXC{BootMount: &LxcBootMount{
							SizeInKibibytes: util.Pointer(LxcMountSize(131071))}},
						err: errors.New(LxcMountSizeErrorMinimum)},
					{name: `error.New(LxcBootMount_Error_QuotaNotPrivileged) default`,
						input: baseConfig(ConfigLXC{
							BootMount: &LxcBootMount{
								SizeInKibibytes: util.Pointer(LxcMountSize(150000)),
								Storage:         util.Pointer("test"),
								Quota:           util.Pointer(true)}}),
						current: &ConfigLXC{BootMount: &LxcBootMount{Storage: util.Pointer("text")}},
						err:     errors.New(LxcBootMount_Error_QuotaNotPrivileged)},
					{name: `error.New(LxcBootMount_Error_QuotaNotPrivileged) false`,
						input: baseConfig(ConfigLXC{
							BootMount: &LxcBootMount{
								SizeInKibibytes: util.Pointer(LxcMountSize(150000)),
								Storage:         util.Pointer("test"),
								Quota:           util.Pointer(true)},
							Privileged: util.Pointer(false)}),
						current: &ConfigLXC{
							BootMount:  &LxcBootMount{Storage: util.Pointer("text")},
							Privileged: util.Pointer(false)},
						err: errors.New(LxcBootMount_Error_QuotaNotPrivileged)}}}},
		{category: `CPU`,
			valid: testType{
				createUpdate: []test{
					{name: `ALL`,
						input: baseConfig(ConfigLXC{CPU: &LxcCPU{
							Cores: util.Pointer(LxcCpuCores(2)),
							Limit: util.Pointer(LxcCpuLimit(3)),
							Units: util.Pointer(LxcCpuUnits(8))}}),
						current: &ConfigLXC{CPU: &LxcCPU{
							Cores: util.Pointer(LxcCpuCores(1)),
							Limit: util.Pointer(LxcCpuLimit(2)),
							Units: util.Pointer(LxcCpuUnits(3))}}}}},
			invalid: testType{
				createUpdate: []test{
					{name: `Cores maximum`,
						input: baseConfig(ConfigLXC{CPU: &LxcCPU{
							Cores: util.Pointer(LxcCpuCores(8193))}}),
						current: &ConfigLXC{CPU: &LxcCPU{
							Cores: util.Pointer(LxcCpuCores(1))}},
						err: errors.New(LxcCpuCores_Error_Invalid)},
					{name: `Limit maximum`,
						input: baseConfig(ConfigLXC{CPU: &LxcCPU{
							Limit: util.Pointer(LxcCpuLimit(8193))}}),
						current: &ConfigLXC{CPU: &LxcCPU{
							Limit: util.Pointer(LxcCpuLimit(2))}},
						err: errors.New(LxcCpuLimit_Error_Invalid)},
					{name: `Units maximum`,
						input: baseConfig(ConfigLXC{CPU: &LxcCPU{
							Units: util.Pointer(LxcCpuUnits(100001))}}),
						current: &ConfigLXC{CPU: &LxcCPU{
							Units: util.Pointer(LxcCpuUnits(3))}},
						err: errors.New(LxcCpuUnits_Error_Maximum)}}}},
		{category: `CreateOptions`,
			valid: testType{
				create: []test{
					{name: `all`,
						input: ConfigLXC{
							BootMount: &LxcBootMount{
								SizeInKibibytes: util.Pointer(LxcMountSize(131072)),
								Storage:         util.Pointer("local-lvm")},
							CreateOptions: &LxcCreateOptions{
								OsTemplate: &LxcTemplate{
									Storage: "local",
									File:    "test-template"},
								UserPassword:  util.Pointer("myPassword!"),
								PublicSSHkeys: publicKeys()}}},
					{name: `UserPassword`,
						input: ConfigLXC{
							BootMount: &LxcBootMount{
								SizeInKibibytes: util.Pointer(LxcMountSize(131072)),
								Storage:         util.Pointer("local-lvm")},
							CreateOptions: &LxcCreateOptions{
								OsTemplate: &LxcTemplate{
									Storage: "local",
									File:    "test-template"},
								UserPassword: util.Pointer("")}}},
					{name: `PublicSSHkeys`,
						input: ConfigLXC{
							BootMount: &LxcBootMount{
								SizeInKibibytes: util.Pointer(LxcMountSize(131072)),
								Storage:         util.Pointer("local-lvm")},
							CreateOptions: &LxcCreateOptions{
								OsTemplate: &LxcTemplate{
									Storage: "local",
									File:    "test-template"},
								PublicSSHkeys: []AuthorizedKey{}}}}}},
			invalid: testType{
				create: []test{
					{name: `errors.New(ConfigLXC_Error_CreateOptionsMissing)`,
						input: ConfigLXC{BootMount: &LxcBootMount{
							SizeInKibibytes: util.Pointer(LxcMountSize(131072)),
							Storage:         util.Pointer("local-lvm")}},
						err: errors.New(ConfigLXC_Error_CreateOptionsMissing)},
					{name: `errors.New(LxcCreateOptions_Error_TemplateMissing)`,
						input: ConfigLXC{
							BootMount: &LxcBootMount{
								SizeInKibibytes: util.Pointer(LxcMountSize(131072)),
								Storage:         util.Pointer("local-lvm")},
							CreateOptions: &LxcCreateOptions{}},
						err: errors.New(LxcCreateOptions_Error_TemplateMissing)},
					{name: `errors.New(LxcTemplate_Error_StorageMissing)`,
						input: ConfigLXC{
							BootMount: &LxcBootMount{
								SizeInKibibytes: util.Pointer(LxcMountSize(131072)),
								Storage:         util.Pointer("local-lvm")},
							CreateOptions: &LxcCreateOptions{
								OsTemplate: &LxcTemplate{}}},
						err: errors.New(LxcTemplate_Error_StorageMissing)},
					{name: `errors.New(LxcTemplate_Error_StorageMissing)`,
						input: ConfigLXC{
							BootMount: &LxcBootMount{
								SizeInKibibytes: util.Pointer(LxcMountSize(131072)),
								Storage:         util.Pointer("local-lvm")},
							CreateOptions: &LxcCreateOptions{
								OsTemplate: &LxcTemplate{Storage: "local"}}},
						err: errors.New(LxcTemplate_Error_FileMissing)}}}},
		{category: `Features`,
			valid: testType{
				create: []test{
					{name: `privileged`,
						input: baseConfig(ConfigLXC{Privileged: util.Pointer(true),
							Features: &LxcFeatures{
								Privileged: &PrivilegedFeatures{}}})},
					{name: `unprivileged`,
						input: baseConfig(ConfigLXC{Privileged: util.Pointer(false),
							Features: &LxcFeatures{
								Unprivileged: &UnprivilegedFeatures{}}})},
				},
				update: []test{
					{name: `privileged`,
						input: baseConfig(ConfigLXC{
							Features: &LxcFeatures{
								Privileged: &PrivilegedFeatures{}}}),
						current: &ConfigLXC{Privileged: util.Pointer(true)}},
					{name: `unprivileged`,
						input: baseConfig(ConfigLXC{
							Features: &LxcFeatures{
								Unprivileged: &UnprivilegedFeatures{}}}),
						current: &ConfigLXC{Privileged: util.Pointer(false)},
					},
				},
			},
			invalid: testType{
				create: []test{
					{name: `privilege default errors.New(LxcFeatures_Error_PrivilegedInUnprivileged)`,
						input: baseConfig(ConfigLXC{Features: &LxcFeatures{
							Privileged: &PrivilegedFeatures{}}}),
						err: errors.New(LxcFeatures_Error_PrivilegedInUnprivileged)},
					{name: `errors.New(LxcFeatures_Error_PrivilegedInUnprivileged)`,
						input: baseConfig(ConfigLXC{Privileged: util.Pointer(false),
							Features: &LxcFeatures{
								Privileged: &PrivilegedFeatures{}}}),
						err: errors.New(LxcFeatures_Error_PrivilegedInUnprivileged)},
					{name: `errors.New(LxcFeatures_Error_UnprivilegedInPrivileged)`,
						input: baseConfig(ConfigLXC{Privileged: util.Pointer(true),
							Features: &LxcFeatures{
								Unprivileged: &UnprivilegedFeatures{}}}),
						err: errors.New(LxcFeatures_Error_UnprivilegedInPrivileged)}},
				createUpdate: []test{
					{name: `errors.New(LxcFeatures_Error_MutuallyExclusive)`,
						input: baseConfig(ConfigLXC{Features: &LxcFeatures{
							Privileged:   &PrivilegedFeatures{},
							Unprivileged: &UnprivilegedFeatures{}}}),
						current: &ConfigLXC{Privileged: util.Pointer(false)},
						err:     errors.New(LxcFeatures_Error_MutuallyExclusive)}},
				update: []test{
					{name: `errors.New(LxcFeatures_Error_PrivilegedInUnprivileged)`,
						input: ConfigLXC{Features: &LxcFeatures{
							Privileged: &PrivilegedFeatures{}}},
						current: &ConfigLXC{Privileged: util.Pointer(false)},
						err:     errors.New(LxcFeatures_Error_PrivilegedInUnprivileged)},
					{name: `errors.New(LxcFeatures_Error_UnprivilegedInPrivileged)`,
						input: ConfigLXC{Features: &LxcFeatures{
							Unprivileged: &UnprivilegedFeatures{}}},
						current: &ConfigLXC{Privileged: util.Pointer(true)},
						err:     errors.New(LxcFeatures_Error_UnprivilegedInPrivileged)}},
			},
		},
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
						input:   baseConfig(ConfigLXC{ID: util.Pointer(GuestID(GuestIdMinimum - 1))}),
						current: &ConfigLXC{ID: util.Pointer(GuestID(0))},
						err:     errors.New(GuestID_Error_Minimum)},
					{name: `maximum`,
						input:   baseConfig(ConfigLXC{ID: util.Pointer(GuestID(GuestIdMaximum + 1))}),
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
		{category: `Mount`,
			valid: testType{
				createUpdate: []test{
					{name: `detach == true`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID200: LxcMount{
								BindMount: &LxcBindMount{},
								DataMount: &LxcDataMount{},
								Detach:    true}}}),
						current: &ConfigLXC{Mounts: LxcMounts{LxcMountID200: LxcMount{}}}}}},
			invalid: testType{
				createUpdate: []test{
					{name: `errors.New(LxcMountErrorMutuallyExclusive)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID100: LxcMount{
								BindMount: &LxcBindMount{},
								DataMount: &LxcDataMount{}}}}),
						current: &ConfigLXC{
							Mounts: LxcMounts{
								LxcMountID100: LxcMount{}}},
						err: errors.New(LxcMountErrorMutuallyExclusive)}}}},
		{category: `Mount.BindMount`,
			valid: testType{
				createUpdate: []test{
					{name: `minimal`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID230: LxcMount{BindMount: &LxcBindMount{
								HostPath:  util.Pointer(LxcHostPath("/mnt/test")),
								GuestPath: util.Pointer(LxcMountPath("/mnt/opt"))}}}}),
						current: &ConfigLXC{Mounts: LxcMounts{
							LxcMountID230: LxcMount{BindMount: &LxcBindMount{
								HostPath:  util.Pointer(LxcHostPath("/mnt/aaa")),
								GuestPath: util.Pointer(LxcMountPath("/mnt/bla"))}}}}}}},
			invalid: testType{
				create: []test{
					{name: `errors.New(LxcBindMountErrorHostPathRequired)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID200: LxcMount{BindMount: &LxcBindMount{}}},
						}),
						err: errors.New(LxcBindMountErrorHostPathRequired)},
					{name: `errors.New(LxcBindMountErrorGuestPathRequired)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID200: LxcMount{BindMount: &LxcBindMount{
								HostPath: util.Pointer(LxcHostPath("/mnt/test"))}}}}),
						err: errors.New(LxcBindMountErrorGuestPathRequired)}},
				createUpdate: []test{
					{name: `errors.New(LxcHostPathErrorInvalid)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID201: LxcMount{BindMount: &LxcBindMount{
								HostPath:  util.Pointer(LxcHostPath("")),
								GuestPath: util.Pointer(LxcMountPath("/mnt/test"))}}}}),
						current: &ConfigLXC{Mounts: LxcMounts{
							LxcMountID201: LxcMount{BindMount: &LxcBindMount{}}}},
						err: errors.New(LxcHostPathErrorInvalid)},
					{name: `errors.New(LxcHostPathErrorRelative)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID201: LxcMount{BindMount: &LxcBindMount{
								HostPath:  util.Pointer(LxcHostPath("./mnt/test")),
								GuestPath: util.Pointer(LxcMountPath("/mnt/test"))}}}}),
						current: &ConfigLXC{Mounts: LxcMounts{
							LxcMountID201: LxcMount{BindMount: &LxcBindMount{}}}},
						err: errors.New(LxcHostPathErrorRelative)},
					{name: `errors.New(LxcHostPathErrorInvalidCharacter)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID201: LxcMount{BindMount: &LxcBindMount{
								HostPath:  util.Pointer(LxcHostPath("/mnt/,test")),
								GuestPath: util.Pointer(LxcMountPath("/mnt/test"))}}}}),
						current: &ConfigLXC{Mounts: LxcMounts{
							LxcMountID201: LxcMount{BindMount: &LxcBindMount{}}}},
						err: errors.New(LxcHostPathErrorInvalidCharacter)},
					{name: `errors.New(LxcMountPathErrorInvalid)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID202: LxcMount{BindMount: &LxcBindMount{
								HostPath:  util.Pointer(LxcHostPath("/mnt/test")),
								GuestPath: util.Pointer(LxcMountPath(""))}}}}),
						current: &ConfigLXC{Mounts: LxcMounts{
							LxcMountID202: LxcMount{BindMount: &LxcBindMount{}}}},
						err: errors.New(LxcMountPathErrorInvalid)},
					{name: `errors.New(LxcMountPathErrorInvalidCharacter)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID202: LxcMount{BindMount: &LxcBindMount{
								HostPath:  util.Pointer(LxcHostPath("/mnt/test")),
								GuestPath: util.Pointer(LxcMountPath("mnt/test/aaa"))}}}}),
						current: &ConfigLXC{Mounts: LxcMounts{
							LxcMountID202: LxcMount{BindMount: &LxcBindMount{}}}},
						err: errors.New(LxcMountPathErrorRelative)},
					{name: `errors.New(LxcMountPathErrorInvalidCharacter)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID202: LxcMount{BindMount: &LxcBindMount{
								HostPath:  util.Pointer(LxcHostPath("/mnt/test")),
								GuestPath: util.Pointer(LxcMountPath("/mnt/test,/aaa"))}}}}),
						current: &ConfigLXC{Mounts: LxcMounts{
							LxcMountID202: LxcMount{BindMount: &LxcBindMount{}}}},
						err: errors.New(LxcMountPathErrorInvalidCharacter)}}}},
		{category: `Mount.DataMount`,
			valid: testType{
				createUpdate: []test{
					{name: `Quota == true`,
						input: baseConfig(ConfigLXC{
							Privileged: util.Pointer(true),
							Mounts: LxcMounts{
								LxcMountID130: LxcMount{
									DataMount: baseDataMount(LxcDataMount{
										Quota: util.Pointer(true)})}}}),
						current: &ConfigLXC{
							Privileged: util.Pointer(true),
							Mounts: LxcMounts{LxcMountID130: LxcMount{
								DataMount: &LxcDataMount{}}}}}}},
			invalid: testType{
				create: []test{
					{name: `errors.New(LxcDataMountErrorPathRequired)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID215: LxcMount{DataMount: &LxcDataMount{}}}}),
						err: errors.New(LxcDataMountErrorPathRequired)},
					{name: `errors.New(LxcDataMountErrorSizeRequired)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID216: LxcMount{DataMount: &LxcDataMount{
								Path: util.Pointer(LxcMountPath(""))}}}}),
						err: errors.New(LxcDataMountErrorSizeRequired)},
					{name: `errors.New(LxcDataMountErrorStorageRequired)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID5: LxcMount{DataMount: &LxcDataMount{
								Path:            util.Pointer(LxcMountPath("")),
								SizeInKibibytes: util.Pointer(LxcMountSize(0))}}}}),
						err: errors.New(LxcDataMountErrorStorageRequired)}},
				createUpdate: []test{
					{name: `errors.New(TriBool_Error_Invalid)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID101: LxcMount{DataMount: baseDataMount(LxcDataMount{
								ACL: util.Pointer(TriBool(5))})}}}),
						current: &ConfigLXC{Mounts: LxcMounts{
							LxcMountID101: LxcMount{}}},
						err: errors.New(TriBool_Error_Invalid)},
					{name: `Quote == false, errors.New(LxcDataMountErrorQuotaUnprivileged)`,
						input: baseConfig(ConfigLXC{
							Mounts: LxcMounts{
								LxcMountID102: LxcMount{DataMount: baseDataMount(LxcDataMount{
									Quota: util.Pointer(false)})}},
							Privileged: util.Pointer(false)}),
						current: &ConfigLXC{
							Mounts: LxcMounts{
								LxcMountID102: LxcMount{DataMount: &LxcDataMount{}}},
							Privileged: util.Pointer(false)},
						err: errors.New(LxcDataMountErrorQuotaUnprivileged)},
					{name: `Quote == true, errors.New(LxcDataMountErrorQuotaUnprivileged)`,
						input: baseConfig(ConfigLXC{
							Mounts: LxcMounts{
								LxcMountID102: LxcMount{DataMount: baseDataMount(LxcDataMount{
									Quota: util.Pointer(true)})}},
							Privileged: util.Pointer(false)}),
						current: &ConfigLXC{
							Mounts: LxcMounts{
								LxcMountID102: LxcMount{DataMount: &LxcDataMount{}}},
							Privileged: util.Pointer(false)},
						err: errors.New(LxcDataMountErrorQuotaUnprivileged)},
					{name: `errors.New(LxcMountPathErrorInvalid)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID103: LxcMount{DataMount: baseDataMount(LxcDataMount{
								Path: util.Pointer(LxcMountPath(""))})}}}),
						current: &ConfigLXC{Mounts: LxcMounts{
							LxcMountID103: LxcMount{DataMount: &LxcDataMount{}}}},
						err: errors.New(LxcMountPathErrorInvalid)},
					{name: `errors.New(lxcMountSizeMinimum)`,
						input: baseConfig(ConfigLXC{Mounts: LxcMounts{
							LxcMountID104: LxcMount{DataMount: baseDataMount(LxcDataMount{
								SizeInKibibytes: util.Pointer(LxcMountSize(10))})}}}),
						current: &ConfigLXC{Mounts: LxcMounts{
							LxcMountID104: LxcMount{DataMount: &LxcDataMount{}}}},
						err: errors.New(LxcMountSizeErrorMinimum)}}}},
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
						err:     errors.New(GuestNameErrorEmpty)}}}},
		{category: `Networks`,
			valid: testType{
				create: []test{
					{name: `all`,
						input: baseConfig(ConfigLXC{Networks: LxcNetworks{
							LxcNetworkID0: LxcNetwork{
								Name:          util.Pointer(LxcNetworkName("my_net")),
								Bridge:        util.Pointer("vmbr0"),
								IPv4:          util.Pointer(LxcIPv4{DHCP: true}),
								IPv6:          util.Pointer(LxcIPv6{SLAAC: true}),
								Mtu:           util.Pointer(MTU(1500)),
								NativeVlan:    util.Pointer(Vlan(23)),
								RateLimitKBps: util.Pointer(GuestNetworkRate(45)),
								TaggedVlans:   util.Pointer(Vlans{Vlan(12), Vlan(23), Vlan(45)})},
							LxcNetworkID5: LxcNetwork{
								Name:          util.Pointer(LxcNetworkName("my_net2")),
								Bridge:        util.Pointer("vmbr0"),
								IPv4:          util.Pointer(LxcIPv4{Manual: true}),
								IPv6:          util.Pointer(LxcIPv6{Manual: true}),
								RateLimitKBps: util.Pointer(GuestNetworkRate(45)),
								TaggedVlans:   util.Pointer(Vlans{Vlan(12), Vlan(23), Vlan(45)})},
							LxcNetworkID3: LxcNetwork{
								Name:   util.Pointer(LxcNetworkName("eth3")),
								Bridge: util.Pointer("vmbr0"),
								IPv4: util.Pointer(LxcIPv4{
									Address: util.Pointer(IPv4CIDR("192.168.0.10/24")),
									Gateway: util.Pointer(IPv4Address("192.168.0.1"))}),
								IPv6: util.Pointer(LxcIPv6{
									Address: util.Pointer(IPv6CIDR("2001:db8::1234/64")),
									Gateway: util.Pointer(IPv6Address("2001:db8::1"))})}}})},
					{name: `minimum`,
						input: baseConfig(ConfigLXC{Networks: LxcNetworks{
							LxcNetworkID10: LxcNetwork{
								Name:   util.Pointer(LxcNetworkName("my_net")),
								Bridge: util.Pointer("vmbr0")}}})}},
				createUpdate: []test{
					{name: `delete non existing`,
						input: baseConfig(ConfigLXC{Networks: LxcNetworks{
							LxcNetworkID1: LxcNetwork{Delete: true}}}),
						current: &ConfigLXC{Networks: LxcNetworks{}}}},
				update: []test{
					{name: `all`,
						input: baseConfig(ConfigLXC{Networks: LxcNetworks{
							LxcNetworkID0: LxcNetwork{
								Name:          util.Pointer(LxcNetworkName("eth0")),
								Bridge:        util.Pointer("vmbr0"),
								IPv4:          util.Pointer(LxcIPv4{DHCP: true}),
								IPv6:          util.Pointer(LxcIPv6{SLAAC: true}),
								Mtu:           util.Pointer(MTU(1500)),
								NativeVlan:    util.Pointer(Vlan(23)),
								RateLimitKBps: util.Pointer(GuestNetworkRate(45)),
								TaggedVlans:   util.Pointer(Vlans{Vlan(12), Vlan(23), Vlan(45)})},
							LxcNetworkID5: LxcNetwork{
								Name:          util.Pointer(LxcNetworkName("eth2")),
								Bridge:        util.Pointer("vmbr0"),
								IPv4:          util.Pointer(LxcIPv4{Manual: true}),
								IPv6:          util.Pointer(LxcIPv6{Manual: true}),
								RateLimitKBps: util.Pointer(GuestNetworkRate(45)),
								TaggedVlans:   util.Pointer(Vlans{Vlan(12), Vlan(23), Vlan(45)})},
							LxcNetworkID3: LxcNetwork{
								Name:   util.Pointer(LxcNetworkName("my_net")),
								Bridge: util.Pointer("vmbr0"),
								IPv4: util.Pointer(LxcIPv4{
									Address: util.Pointer(IPv4CIDR("192.168.0.10/24")),
									Gateway: util.Pointer(IPv4Address("192.168.0.1"))}),
								IPv6: util.Pointer(LxcIPv6{
									Address: util.Pointer(IPv6CIDR("2001:db8::1234/64")),
									Gateway: util.Pointer(IPv6Address("2001:db8::1"))})},
							LxcNetworkID10: LxcNetwork{
								Name:   util.Pointer(LxcNetworkName("lo0")),
								Bridge: util.Pointer("vmbr0")}}}),
						current: &ConfigLXC{Networks: LxcNetworks{
							LxcNetworkID3: LxcNetwork{},
							LxcNetworkID11: LxcNetwork{
								Name:          util.Pointer(LxcNetworkName("eth6")),
								Bridge:        util.Pointer("vmbr0"),
								IPv4:          util.Pointer(LxcIPv4{DHCP: true}),
								IPv6:          util.Pointer(LxcIPv6{SLAAC: true}),
								Mtu:           util.Pointer(MTU(1500)),
								NativeVlan:    util.Pointer(Vlan(23)),
								RateLimitKBps: util.Pointer(GuestNetworkRate(45))}}}},
					{name: `minimum`,
						input: baseConfig(ConfigLXC{Networks: LxcNetworks{
							LxcNetworkID12: LxcNetwork{}}}),
						current: &ConfigLXC{Networks: LxcNetworks{
							LxcNetworkID12: LxcNetwork{
								Name:   util.Pointer(LxcNetworkName("text")),
								Bridge: util.Pointer("vmbr0")}}}},
					{name: `duplicate net overwrite`,
						input: baseConfig(ConfigLXC{Networks: LxcNetworks{
							LxcNetworkID13: LxcNetwork{
								Name: util.Pointer(LxcNetworkName("text"))},
							LxcNetworkID12: LxcNetwork{
								Name: util.Pointer(LxcNetworkName("test"))},
							LxcNetworkID11: LxcNetwork{
								Name: util.Pointer(LxcNetworkName("net"))}}}),
						current: &ConfigLXC{Networks: LxcNetworks{
							LxcNetworkID13: LxcNetwork{
								Name: util.Pointer(LxcNetworkName("test"))},
							LxcNetworkID12: LxcNetwork{
								Name: util.Pointer(LxcNetworkName("text"))},
							LxcNetworkID11: LxcNetwork{
								Name: util.Pointer(LxcNetworkName("net"))}}}}}},
			invalid: testType{
				create: []test{
					{name: `errors.New(LxcNetwork_Error_BridgeRequired)`,
						input: baseConfig(ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: LxcNetwork{}}}),
						err:   errors.New(LxcNetwork_Error_BridgeRequired)}},
				createUpdate: []test{
					{name: `errors.New(LxcNetworks_Error_Amount)`,
						input: baseConfig(ConfigLXC{Networks: LxcNetworks{
							0: LxcNetwork{}, 1: LxcNetwork{}, 2: LxcNetwork{},
							3: LxcNetwork{}, 4: LxcNetwork{}, 5: LxcNetwork{},
							6: LxcNetwork{}, 7: LxcNetwork{}, 8: LxcNetwork{},
							9: LxcNetwork{}, 10: LxcNetwork{}, 11: LxcNetwork{},
							12: LxcNetwork{}, 13: LxcNetwork{}, 14: LxcNetwork{},
							15: LxcNetwork{}, 16: LxcNetwork{}}}),
						current: &ConfigLXC{Networks: LxcNetworks{}},
						err:     errors.New(LxcNetworks_Error_Amount)},
					{name: `errors.New(LxcNetworkID_Error_Invalid)`,
						input:   baseConfig(ConfigLXC{Networks: LxcNetworks{45: LxcNetwork{}}}),
						current: &ConfigLXC{Networks: LxcNetworks{LxcNetworkID11: LxcNetwork{}}},
						err:     errors.New(LxcNetworkID_Error_Invalid)},
					{name: `errors.New(LxcNetworkName_Error_LengthMinimum)`,
						input: baseConfig(ConfigLXC{Networks: LxcNetworks{LxcNetworkID14: LxcNetwork{
							Bridge: util.Pointer("vmbr0"),
							Name:   util.Pointer(LxcNetworkName(""))}}}),
						current: &ConfigLXC{Networks: LxcNetworks{LxcNetworkID14: LxcNetwork{
							Name: util.Pointer(LxcNetworkName("eth0"))}}},
						err: errors.New(LxcNetworkName_Error_LengthMinimum)}}}},
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
						current: &ConfigLXC{Pool: util.Pointer(PoolName("text"))}},
					{name: `empty`,
						input:   baseConfig(ConfigLXC{Pool: util.Pointer(PoolName(""))}),
						current: &ConfigLXC{Pool: util.Pointer(PoolName("text"))}}}},
			invalid: testType{
				createUpdate: []test{
					{name: `characters`,
						input:   baseConfig(ConfigLXC{Pool: util.Pointer(PoolName("^&$%"))}),
						current: &ConfigLXC{Pool: util.Pointer(PoolName("text"))},
						err:     errors.New(PoolName_Error_Characters)}}}},
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
				require.Equal(t, subTest.err, subTest.input.Validate(nil, subTest.version), name)
			})
		}
		for _, subTest := range append(test.valid.update, test.valid.createUpdate...) {
			name := test.category + "/Valid/Update"
			if len(test.valid.update)+len(test.valid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.NotNil(t, subTest.current)
				require.Equal(t, subTest.err, subTest.input.Validate(subTest.current, subTest.version), name)
			})
		}
		for _, subTest := range append(test.invalid.create, test.invalid.createUpdate...) {
			name := test.category + "/Invalid/Create"
			if len(test.invalid.create)+len(test.invalid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.Equal(t, subTest.err, subTest.input.Validate(nil, subTest.version), name)
			})
		}
		for _, subTest := range append(test.invalid.update, test.invalid.createUpdate...) {
			name := test.category + "/Invalid/Update"
			if len(test.invalid.update)+len(test.invalid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.NotNil(t, subTest.current)
				require.Equal(t, subTest.err, subTest.input.Validate(subTest.current, subTest.version), name)
			})
		}
	}
}

func Test_ConfigLXC_Update(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		vmr    *VmRef
		client *Client
		err    error
	}{
		{name: `client nil`,
			err: errors.New(Client_Error_Nil)},
		{name: `client not initialized`,
			client: util.Pointer(Client{}),
			err:    errors.New(Client_Error_NotInitialized)},
		{name: `vmr nil`,
			client: fakeClient(),
			err:    errors.New(VmRef_Error_Nil)},
	}
	for _, test := range tests {
		var err error
		require.NotPanics(t, func() { err = ConfigLXC{}.Update(context.Background(), true, test.vmr, test.client) })
		require.Equal(t, test.err, err)
	}
}

func Test_ConfigLXC_UpdateNoCheck(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		vmr    *VmRef
		client *Client
		err    error
	}{
		{name: `client nil`,
			err: errors.New(Client_Error_Nil)},
		{name: `client not initialized`,
			client: util.Pointer(Client{}),
			err:    errors.New(Client_Error_NotInitialized)},
		{name: `vmr nil`,
			client: fakeClient(),
			err:    errors.New(VmRef_Error_Nil)},
	}
	for _, test := range tests {
		var err error
		require.NotPanics(t, func() { err = ConfigLXC{}.UpdateNoCheck(context.Background(), true, test.vmr, test.client) })
		require.Equal(t, test.err, err)
	}
}

func testData_RawConfigLXC_Architecture_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{
			{name: `amd64`,
				input:  rawConfigLXC{a: map[string]any{"arch": "amd64"}},
				output: lxcGetBaseConfig(ConfigLXC{Architecture: "amd64"})},
			{name: `""`,
				input:  rawConfigLXC{a: map[string]any{"arch": ""}},
				output: lxcGetBaseConfig(ConfigLXC{Architecture: ""})}}
	}
}

func testData_RawConfigLXC_Description_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{
			{name: `test`,
				input:  rawConfigLXC{a: map[string]any{"description": "test"}},
				output: lxcGetBaseConfig(ConfigLXC{Description: util.Pointer("test")})},
			{name: `""`,
				input:  rawConfigLXC{a: map[string]any{"description": ""}},
				output: lxcGetBaseConfig(ConfigLXC{Description: util.Pointer("")})}}
	}
}

func testData_RawConfigLXC_Digest_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{{
			input: rawConfigLXC{a: map[string]any{"digest": "af064923bbf2301596aac4c273ba32178ebc4a96"}},
			output: lxcGetBaseConfig(ConfigLXC{
				Digest: [sha1.Size]byte{
					0xaf, 0x06, 0x49, 0x23, 0xbb, 0xf2, 0x30, 0x15, 0x96, 0xaa,
					0xc4, 0xc2, 0x73, 0xba, 0x32, 0x17, 0x8e, 0xbc, 0x4a, 0x96},
				rawDigest: "af064923bbf2301596aac4c273ba32178ebc4a96"})}}
	}
}

func testData_RawConfigLXC_DNS_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{
			{name: `all`,
				input: rawConfigLXC{a: map[string]any{
					"nameserver":   "1.1.1.1 8.8.8.8 9.9.9.9",
					"searchdomain": "example.com"}},
				output: lxcGetBaseConfig(ConfigLXC{DNS: &GuestDNS{
					NameServers: util.Pointer([]netip.Addr{
						parseIP("1.1.1.1"),
						parseIP("8.8.8.8"),
						parseIP("9.9.9.9")}),
					SearchDomain: util.Pointer("example.com")}})},
			{name: `NameServers`,
				input: rawConfigLXC{a: map[string]any{"nameserver": "8.8.8.8"}},
				output: lxcGetBaseConfig(ConfigLXC{DNS: &GuestDNS{
					NameServers:  util.Pointer([]netip.Addr{parseIP("8.8.8.8")}),
					SearchDomain: util.Pointer("")}})},
			{name: `SearchDomain`,
				input: rawConfigLXC{a: map[string]any{"searchdomain": "example.com"}},
				output: lxcGetBaseConfig(ConfigLXC{DNS: &GuestDNS{
					NameServers:  util.Pointer([]netip.Addr(nil)),
					SearchDomain: util.Pointer("example.com")}})}}
	}
}

func testData_RawConfigLXC_ID_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{{
			input:  rawConfigLXC{guestID: 15},
			output: lxcGetBaseConfig(ConfigLXC{ID: util.Pointer(GuestID(15))})}}
	}
}

func testData_RawConfigLXC_Memory_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{{
			input:  rawConfigLXC{a: map[string]any{"memory": float64(512)}},
			output: lxcGetBaseConfig(ConfigLXC{Memory: util.Pointer(LxcMemory(512))})}}
	}
}

func testData_RawConfigLXC_Name_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{{
			input:  rawConfigLXC{a: map[string]any{"hostname": "test"}},
			output: lxcGetBaseConfig(ConfigLXC{Name: util.Pointer(GuestName("test"))})}}
	}
}

func testData_RawConfigLXC_Node_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{{
			input:  rawConfigLXC{node: "test"},
			output: lxcGetBaseConfig(ConfigLXC{Node: util.Pointer(NodeName("test"))})}}
	}
}

func testData_RawConfigLXC_OperatingSystem_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{{
			input:  rawConfigLXC{a: map[string]any{"ostype": "test"}},
			output: lxcGetBaseConfig(ConfigLXC{OperatingSystem: "test"})}}
	}
}

func testData_RawConfigLXC_Pool_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{{
			pool:   new(PoolName("test")),
			output: lxcGetBaseConfig(ConfigLXC{Pool: util.Pointer(PoolName("test"))})}}
	}
}

func testData_RawConfigLXC_Privileged_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{
			{name: `true`,
				input:  rawConfigLXC{a: map[string]any{"unprivileged": float64(0)}},
				output: lxcGetBaseConfig(ConfigLXC{Privileged: util.Pointer(true)})},
			{name: `false`,
				input:  rawConfigLXC{a: map[string]any{"unprivileged": float64(1)}},
				output: lxcGetBaseConfig(ConfigLXC{Privileged: util.Pointer(false)})},
			{name: `default true`,
				input:  rawConfigLXC{a: map[string]any{}},
				output: lxcGetBaseConfig(ConfigLXC{Privileged: util.Pointer(true)})}}
	}
}

func testData_RawConfigLXC_Protection_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{
			{name: `false`,
				input:  rawConfigLXC{a: map[string]any{}},
				output: lxcGetBaseConfig(ConfigLXC{Protection: util.Pointer(false)})},
			{name: `true`,
				input:  rawConfigLXC{a: map[string]any{"protection": float64(1)}},
				output: lxcGetBaseConfig(ConfigLXC{Protection: util.Pointer(true)})}}
	}
}

func testData_RawConfigLXC_StartAtNodeBoot_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{
			{name: `true`,
				input:  rawConfigLXC{a: map[string]any{"onboot": float64(1)}},
				output: lxcGetBaseConfig(ConfigLXC{StartAtNodeBoot: util.Pointer(true)})},
			{name: `false`,
				input:  rawConfigLXC{a: map[string]any{"onboot": float64(0)}},
				output: lxcGetBaseConfig(ConfigLXC{StartAtNodeBoot: util.Pointer(false)})},
			{name: `default false`,
				input:  rawConfigLXC{a: map[string]any{}},
				output: lxcGetBaseConfig(ConfigLXC{StartAtNodeBoot: util.Pointer(false)})}}
	}
}

func testData_RawConfigLXC_Swap_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{
			{name: `set`,
				input:  rawConfigLXC{a: map[string]any{"swap": float64(256)}},
				output: lxcGetBaseConfig(ConfigLXC{Swap: util.Pointer(LxcSwap(256))})},
			{name: `set 0`,
				input:  rawConfigLXC{a: map[string]any{"swap": float64(0)}},
				output: lxcGetBaseConfig(ConfigLXC{Swap: util.Pointer(LxcSwap(0))})}}
	}
}

func testData_RawConfigLXC_State_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{{
			state:  2,
			output: lxcGetBaseConfig(ConfigLXC{State: util.Pointer(PowerStateRunning)})}}
	}
}

func Test_RawConfigLXC_GetArchitecture(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Architecture_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, output.Architecture, raw.GetArchitecture())
		})
}

func Test_RawConfigLXC_Architecture_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Architecture_Get().Test(t)
}

func Test_RawConfigLXC_GetDescription(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Description_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, output.Description, raw.GetDescription())
		})
}

func Test_RawConfigLXC_Description_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Description_Get().Test(t)
}

func Test_RawConfigLXC_GetDigest(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Digest_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, output.Digest, raw.GetDigest())
		})
}

func Test_RawConfigLXC_Digest_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Digest_Get().Test(t)
}

func Test_RawConfigLXC_GetDNS(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_DNS_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, output.DNS, raw.GetDNS())
		})
}

func Test_RawConfigLXC_DNS_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_DNS_Get().Test(t)
}

func Test_RawConfigLXC_GetID(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_ID_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, *output.ID, raw.GetID())
		})
}

func Test_RawConfigLXC_ID_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_ID_Get().Test(t)
}

func Test_RawConfigLXC_GetMemory(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Memory_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, *output.Memory, raw.GetMemory())
		})
}

func Test_RawConfigLXC_Memory_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Memory_Get().Test(t)
}

func Test_RawConfigLXC_GetName(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Name_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, *output.Name, raw.GetName())
		})
}

func Test_RawConfigLXC_Name_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Name_Get().Test(t)
}

func Test_RawConfigLXC_GetNode(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Node_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, *output.Node, raw.GetNode())
		})
}

func Test_RawConfigLXC_Node_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Node_Get().Test(t)
}

func Test_RawConfigLXC_GetOperatingSystem(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_OperatingSystem_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, output.OperatingSystem, raw.GetOperatingSystem())
		})
}

func Test_RawConfigLXC_OperatingSystem_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_OperatingSystem_Get().Test(t)
}

func Test_RawConfigLXC_Pool_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Pool_Get().Test(t)
}

func Test_RawConfigLXC_GetPrivileged(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Privileged_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, *output.Privileged, raw.GetPrivileged())
		})
}

func Test_RawConfigLXC_Privileged_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Privileged_Get().Test(t)
}

func Test_RawConfigLXC_GetProtection(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Protection_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, *output.Protection, raw.GetProtection())
		})
}

func Test_RawConfigLXC_Protection_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Protection_Get().Test(t)
}

func Test_RawConfigLXC_GetStartAtNodeBoot(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_StartAtNodeBoot_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, *output.StartAtNodeBoot, raw.GetStartAtNodeBoot())
		})
}

func Test_RawConfigLXC_StartAtNodeBoot_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_StartAtNodeBoot_Get().Test(t)
}

func Test_RawConfigLXC_GetSwap(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Swap_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, *output.Swap, raw.GetSwap())
		})
}

func Test_RawConfigLXC_Swap_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Swap_Get().Test(t)
}

func Test_RawConfigLXC_State_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_State_Get().Test(t)
}

func Test_LxcMemory_String(t *testing.T) {
	t.Parallel()
	require.Equal(t, "583421", LxcMemory(583421).String())
}

func Test_LxcSwap_String(t *testing.T) {
	t.Parallel()
	require.Equal(t, "8423", LxcSwap(8423).String())
}
