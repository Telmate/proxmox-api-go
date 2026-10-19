package proxmox

import (
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func testData_RawConfigLXC_Mounts_Get() lxcTestGetFunc {
	baseBindMount := func(config LxcBindMount) *LxcBindMount {
		if config.GuestPath == nil {
			config.GuestPath = new(LxcMountPath(""))
		}
		if config.HostPath == nil {
			config.HostPath = new(LxcHostPath(""))
		}
		if config.ReadOnly == nil {
			config.ReadOnly = new(false)
		}
		if config.Replicate == nil {
			config.Replicate = new(true)
		}
		return &config
	}
	baseDataMount := func(config LxcDataMount) *LxcDataMount {
		if config.ACL == nil {
			config.ACL = new(TriBoolNone)
		}
		if config.Backup == nil {
			config.Backup = new(false)
		}
		if config.Path == nil {
			config.Path = new(LxcMountPath(""))
		}
		if config.ReadOnly == nil {
			config.ReadOnly = new(false)
		}
		if config.Replicate == nil {
			config.Replicate = new(true)
		}
		if config.SizeInKibibytes == nil {
			config.SizeInKibibytes = new(LxcMountSize(0))
		}
		return &config
	}
	baseMountOptions := func(config LxcMountOptions) *LxcMountOptions {
		if config.Discard == nil {
			config.Discard = new(false)
		}
		if config.LazyTime == nil {
			config.LazyTime = new(false)
		}
		if config.NoATime == nil {
			config.NoATime = new(false)
		}
		if config.NoDevice == nil {
			config.NoDevice = new(false)
		}
		if config.NoExec == nil {
			config.NoExec = new(false)
		}
		if config.NoSuid == nil {
			config.NoSuid = new(false)
		}
		return &config
	}
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{
			{name: `BindMount minimal`,
				input: rawConfigLXC{a: map[string]any{"mp0": "/host/path,mp=/guest/path"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID0: LxcMount{BindMount: baseBindMount(LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path"))})}}})},
			{name: `BindMount.ReadOnly false`,
				input: rawConfigLXC{a: map[string]any{"mp1": "/host/path,mp=/guest/path"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID1: LxcMount{BindMount: baseBindMount(LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path")),
						ReadOnly:  new(false)})}}})},
			{name: `BindMount.ReadOnly true`,
				input: rawConfigLXC{a: map[string]any{"mp2": "/host/path,mp=/guest/path,ro=1"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID2: LxcMount{BindMount: baseBindMount(LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path")),
						ReadOnly:  new(true)})}}})},
			{name: `BindMount.Replicate false`,
				input: rawConfigLXC{a: map[string]any{"mp3": "/host/path,mp=/guest/path,replicate=0"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID3: LxcMount{BindMount: baseBindMount(LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path")),
						Replicate: new(false)})}}})},
			{name: `BindMount.Replicate true`,
				input: rawConfigLXC{a: map[string]any{"mp4": "/host/path,mp=/guest/path"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID4: LxcMount{BindMount: baseBindMount(LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path")),
						Replicate: new(true)})}}})},
			{name: `BindMount.Options.Discard true`,
				input: rawConfigLXC{a: map[string]any{"mp5": "/host/path,mp=/guest/path,mountoptions=discard"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID5: LxcMount{BindMount: baseBindMount(LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path")),
						Options: baseMountOptions(LxcMountOptions{
							Discard: new(true)})})}}})},
			{name: `BindMount.Options.LazyTime true`,
				input: rawConfigLXC{a: map[string]any{"mp6": "/host/path,mp=/guest/path,mountoptions=lazytime"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID6: LxcMount{BindMount: baseBindMount(LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path")),
						Options: baseMountOptions(LxcMountOptions{
							LazyTime: new(true)})})}}})},
			{name: `BindMount.Options.NoATime true`,
				input: rawConfigLXC{a: map[string]any{"mp7": "/host/path,mp=/guest/path,mountoptions=noatime"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID7: LxcMount{BindMount: baseBindMount(LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path")),
						Options: baseMountOptions(LxcMountOptions{
							NoATime: new(true)})})}}})},
			{name: `BindMount.Options.NoDevice true`,
				input: rawConfigLXC{a: map[string]any{"mp8": "/host/path,mp=/guest/path,mountoptions=nodev"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID8: LxcMount{BindMount: baseBindMount(LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path")),
						Options: baseMountOptions(LxcMountOptions{
							NoDevice: new(true)})})}}})},
			{name: `BindMount.Options.NoExec true`,
				input: rawConfigLXC{a: map[string]any{"mp9": "/host/path,mp=/guest/path,mountoptions=noexec"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID9: LxcMount{BindMount: baseBindMount(LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path")),
						Options: baseMountOptions(LxcMountOptions{
							NoExec: new(true)})})}}})},
			{name: `BindMount.Options.NoSuid true`,
				input: rawConfigLXC{a: map[string]any{"mp10": "/host/path,mp=/guest/path,mountoptions=nosuid"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID10: LxcMount{BindMount: baseBindMount(LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path")),
						Options: baseMountOptions(LxcMountOptions{
							NoSuid: new(true)})})}}})},
			{name: `BindMount.Options all true`,
				input: rawConfigLXC{a: map[string]any{"mp11": "/host/path,mp=/guest/path,mountoptions=discard;lazytime;noatime;nodev;noexec;nosuid"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID11: LxcMount{BindMount: baseBindMount(LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path")),
						Options: &LxcMountOptions{
							Discard:  new(true),
							LazyTime: new(true),
							NoATime:  new(true),
							NoDevice: new(true),
							NoExec:   new(true),
							NoSuid:   new(true)}})}}})},
			{name: `DataMount minimal ext4 (privileged`,
				input: rawConfigLXC{a: map[string]any{"mp100": "local-ext4:100/vm-100-disk-0.raw"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID100: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Quota:   new(false),
						Storage: new("local-ext4"),
						rawDisk: "local-ext4:100/vm-100-disk-0.raw"})}}})},
			{name: `DataMount minimal zfs (privileged`,
				input: rawConfigLXC{a: map[string]any{"mp101": "local-zfs:subvol-100-disk-1"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID101: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Quota:   new(false),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.ACL false (privileged`,
				input: rawConfigLXC{a: map[string]any{"mp102": "local-zfs:subvol-100-disk-1,acl=0"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID102: LxcMount{DataMount: baseDataMount(LxcDataMount{
						ACL:     new(TriBoolFalse),
						Quota:   new(false),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.ACL true (privileged)`,
				input: rawConfigLXC{a: map[string]any{"mp103": "local-zfs:subvol-100-disk-1,acl=1"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID103: LxcMount{DataMount: baseDataMount(LxcDataMount{
						ACL:     new(TriBoolTrue),
						Quota:   new(false),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Backup false (privileged)`,
				input: rawConfigLXC{a: map[string]any{"mp104": "local-zfs:subvol-100-disk-1,backup=0"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID104: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Backup:  new(false),
						Quota:   new(false),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Backup true (privileged)`,
				input: rawConfigLXC{a: map[string]any{"mp105": "local-zfs:subvol-100-disk-1,backup=1"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID105: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Backup:  new(true),
						Quota:   new(false),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Options.Discard true`,
				input: rawConfigLXC{a: map[string]any{"mp201": "local-zfs:subvol-100-disk-1,mountoptions=discard"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID201: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Options: baseMountOptions(LxcMountOptions{
							Discard: new(true)}),
						Quota:   new(false),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Options.LazyTime true`,
				input: rawConfigLXC{a: map[string]any{"mp203": "local-zfs:subvol-100-disk-1,mountoptions=lazytime"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID203: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Options: baseMountOptions(LxcMountOptions{
							LazyTime: new(true)}),
						Quota:   new(false),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Options.NoATime true`,
				input: rawConfigLXC{a: map[string]any{"mp204": "local-zfs:subvol-100-disk-1,mountoptions=noatime"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID204: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Options: baseMountOptions(LxcMountOptions{
							NoATime: new(true)}),
						Quota:   new(false),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Options.NoDevice true`,
				input: rawConfigLXC{a: map[string]any{"mp205": "local-zfs:subvol-100-disk-1,mountoptions=nodev"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID205: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Options: baseMountOptions(LxcMountOptions{
							NoDevice: new(true)}),
						Quota:   new(false),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Options.NoExec true`,
				input: rawConfigLXC{a: map[string]any{"mp206": "local-zfs:subvol-100-disk-1,mountoptions=noexec"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID206: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Options: baseMountOptions(LxcMountOptions{
							NoExec: new(true)}),
						Quota:   new(false),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Options.NoSuid true`,
				input: rawConfigLXC{a: map[string]any{"mp207": "local-zfs:subvol-100-disk-1,mountoptions=nosuid"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID207: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Options: baseMountOptions(LxcMountOptions{
							NoSuid: new(true)}),
						Quota:   new(false),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Options all true`,
				input: rawConfigLXC{a: map[string]any{"mp208": "local-zfs:subvol-100-disk-1,mountoptions=noexec;nosuid;lazytime;discard;noatime;nodev"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID208: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Options: &LxcMountOptions{
							Discard:  new(true),
							LazyTime: new(true),
							NoATime:  new(true),
							NoDevice: new(true),
							NoExec:   new(true),
							NoSuid:   new(true)},
						Quota:   new(false),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Path (privileged)`,
				input: rawConfigLXC{a: map[string]any{"mp106": "local-zfs:subvol-100-disk-1,mp=/mnt/test"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID106: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Quota:   new(false),
						Path:    new(LxcMountPath("/mnt/test")),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Quota false Privilege false`,
				input: rawConfigLXC{a: map[string]any{
					"mp107":        "local-zfs:subvol-100-disk-1,quota=0",
					"unprivileged": float64(1)}},
				output: lxcGetBaseConfig(ConfigLXC{
					Privileged: new(false),
					Mounts: LxcMounts{
						LxcMountID107: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Storage: new("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Quota false Privilege true`,
				input: rawConfigLXC{a: map[string]any{
					"mp108":        "local-zfs:subvol-100-disk-1,quota=0",
					"unprivileged": float64(0)}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID108: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Quota:   new(false),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Quota true Privilege false`,
				input: rawConfigLXC{a: map[string]any{
					"mp109":        "local-zfs:subvol-100-disk-1,quota=1",
					"unprivileged": float64(1)}},
				output: lxcGetBaseConfig(ConfigLXC{
					Privileged: new(false),
					Mounts: LxcMounts{
						LxcMountID109: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Storage: new("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Quota true Privilege true`,
				input: rawConfigLXC{a: map[string]any{
					"mp110":        "local-zfs:subvol-100-disk-1,quota=1",
					"unprivileged": float64(0)}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID110: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Quota:   new(true),
						Storage: new("local-zfs"),
						rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.ReadOnly false (privileged)`,
				input: rawConfigLXC{a: map[string]any{"mp111": "local-zfs:subvol-100-disk-1,ro=0"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID111: LxcMount{DataMount: baseDataMount(LxcDataMount{
						ReadOnly: new(false),
						Quota:    new(false),
						Storage:  new("local-zfs"),
						rawDisk:  "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.ReadOnly true (privileged)`,
				input: rawConfigLXC{a: map[string]any{"mp112": "local-zfs:subvol-100-disk-1,ro=1"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID112: LxcMount{DataMount: baseDataMount(LxcDataMount{
						ReadOnly: new(true),
						Quota:    new(false),
						Storage:  new("local-zfs"),
						rawDisk:  "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Replicate false (privileged)`,
				input: rawConfigLXC{a: map[string]any{"mp113": "local-zfs:subvol-100-disk-1,replicate=0"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID113: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Replicate: new(false),
						Quota:     new(false),
						Storage:   new("local-zfs"),
						rawDisk:   "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.Replicate true (privileged)`,
				input: rawConfigLXC{a: map[string]any{"mp114": "local-zfs:subvol-100-disk-1,replicate=1"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID114: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Replicate: new(true),
						Quota:     new(false),
						Storage:   new("local-zfs"),
						rawDisk:   "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.SizeInKibibytes 1T (privileged)`,
				input: rawConfigLXC{a: map[string]any{"mp115": "local-zfs:subvol-100-disk-1,size=1T"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID115: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Quota:           new(false),
						SizeInKibibytes: new(LxcMountSize(1073741824)),
						Storage:         new("local-zfs"),
						rawDisk:         "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.SizeInKibibytes 1G (privileged)`,
				input: rawConfigLXC{a: map[string]any{"mp116": "local-zfs:subvol-100-disk-1,size=1G"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID116: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Quota:           new(false),
						SizeInKibibytes: new(LxcMountSize(1048576)),
						Storage:         new("local-zfs"),
						rawDisk:         "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.SizeInKibibytes 12M (privileged)`,
				input: rawConfigLXC{a: map[string]any{"mp117": "local-zfs:subvol-100-disk-1,size=12M"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID117: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Quota:           new(false),
						SizeInKibibytes: new(LxcMountSize(12288)),
						Storage:         new("local-zfs"),
						rawDisk:         "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount.SizeInKibibytes 18K  (privileged)`,
				input: rawConfigLXC{a: map[string]any{"mp118": "local-zfs:subvol-100-disk-1,size=18K"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID118: LxcMount{DataMount: baseDataMount(LxcDataMount{
						Quota:           new(false),
						SizeInKibibytes: new(LxcMountSize(18)),
						Storage:         new("local-zfs"),
						rawDisk:         "local-zfs:subvol-100-disk-1"})}}})},
			{name: `DataMount all (privileged)`,
				input: rawConfigLXC{a: map[string]any{"mp150": "local-zfs:subvol-100-disk-1,size=18K,acl=0,backup=1,quota=1,mountoptions=lazytime;noexec;discard,mp=/opt/test,replicate=1,ro=1"}},
				output: lxcGetBaseConfig(ConfigLXC{Mounts: LxcMounts{
					LxcMountID150: LxcMount{DataMount: baseDataMount(LxcDataMount{
						ACL:    new(TriBoolFalse),
						Backup: new(true),
						Options: &LxcMountOptions{
							Discard:  new(true),
							LazyTime: new(true),
							NoATime:  new(false),
							NoDevice: new(false),
							NoExec:   new(true),
							NoSuid:   new(false)},
						Path:            new(LxcMountPath("/opt/test")),
						Quota:           new(true),
						ReadOnly:        new(true),
						Replicate:       new(true),
						SizeInKibibytes: new(LxcMountSize(18)),
						Storage:         new("local-zfs"),
						rawDisk:         "local-zfs:subvol-100-disk-1"})}}})}}
	}
}

func testData_RawConfigLXC_BootMount_Get() lxcTestGetFunc {
	baseBootMount := func(config LxcBootMount) *LxcBootMount {
		if config.ACL == nil {
			config.ACL = new(TriBoolNone)
		}
		if config.Quota == nil {
			config.Quota = new(false)
		}
		if config.Replicate == nil {
			config.Replicate = new(true)
		}
		if config.Storage == nil {
			config.Storage = new("")
		}
		if config.SizeInKibibytes == nil {
			config.SizeInKibibytes = new(LxcMountSize(0))
		}
		return &config
	}
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{
			{name: `ACL true`,
				input: rawConfigLXC{a: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,acl=1"}},
				output: lxcGetBaseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
					ACL:     new(TriBoolTrue),
					Storage: new("local-zfs"),
					rawDisk: "local-zfs:subvol-101-disk-0"})})},
			{name: `ACL false`,
				input: rawConfigLXC{a: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,acl=0"}},
				output: lxcGetBaseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
					ACL:     new(TriBoolFalse),
					Storage: new("local-zfs"),
					rawDisk: "local-zfs:subvol-101-disk-0"})})},
			{name: `Options Discard true`,
				input: rawConfigLXC{a: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=discard"}},
				output: lxcGetBaseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
					Options: &LxcBootMountOptions{
						Discard:  new(true),
						LazyTime: new(false),
						NoATime:  new(false),
						NoSuid:   new(false)},
					Storage: new("local-zfs"),
					rawDisk: "local-zfs:subvol-101-disk-0"})})},
			{name: `Options LazyTime true`,
				input: rawConfigLXC{a: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=lazytime"}},
				output: lxcGetBaseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
					Options: &LxcBootMountOptions{
						Discard:  new(false),
						LazyTime: new(true),
						NoATime:  new(false),
						NoSuid:   new(false)},
					Storage: new("local-zfs"),
					rawDisk: "local-zfs:subvol-101-disk-0"})})},
			{name: `Options NoATime true`,
				input: rawConfigLXC{a: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=noatime"}},
				output: lxcGetBaseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
					Options: &LxcBootMountOptions{
						Discard:  new(false),
						LazyTime: new(false),
						NoATime:  new(true),
						NoSuid:   new(false)},
					Storage: new("local-zfs"),
					rawDisk: "local-zfs:subvol-101-disk-0"})})},
			{name: `Options NoSuid true`,
				input: rawConfigLXC{a: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=nosuid"}},
				output: lxcGetBaseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
					ACL: new(TriBoolNone),
					Options: &LxcBootMountOptions{
						Discard:  new(false),
						LazyTime: new(false),
						NoATime:  new(false),
						NoSuid:   new(true)},
					Storage: new("local-zfs"),
					rawDisk: "local-zfs:subvol-101-disk-0"})})},
			{name: `Quota false`,
				input: rawConfigLXC{a: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0"}},
				output: lxcGetBaseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
					Quota:   new(false),
					Storage: new("local-zfs"),
					rawDisk: "local-zfs:subvol-101-disk-0"})})},
			{name: `Quota true`,
				input: rawConfigLXC{a: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,quota=1"}},
				output: lxcGetBaseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
					Quota:   new(true),
					Storage: new("local-zfs"),
					rawDisk: "local-zfs:subvol-101-disk-0"})})},
			{name: `Replicate false`,
				input: rawConfigLXC{a: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,replicate=0"}},
				output: lxcGetBaseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
					Replicate: new(false),
					Storage:   new("local-zfs"),
					rawDisk:   "local-zfs:subvol-101-disk-0"})})},
			{name: `Replicate true`,
				input: rawConfigLXC{a: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,replicate=1"}},
				output: lxcGetBaseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
					Replicate: new(true),
					Storage:   new("local-zfs"),
					rawDisk:   "local-zfs:subvol-101-disk-0"})})},
			{name: `SizeInKibibytes`,
				input: rawConfigLXC{a: map[string]any{"rootfs": "local-ext4:subvol-101-disk-0,size=999M"}},
				output: lxcGetBaseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
					Storage:         new("local-ext4"),
					SizeInKibibytes: new(LxcMountSize(1022976)),
					rawDisk:         "local-ext4:subvol-101-disk-0"})})},
			{name: `all`,
				input: rawConfigLXC{a: map[string]any{"rootfs": "local-ext4:subvol-101-disk-0,acl=1,mountoptions=discard;lazytime;noatime;nosuid,size=1G,quota=1,replicate=1"}},
				output: lxcGetBaseConfig(ConfigLXC{BootMount: &LxcBootMount{
					ACL:   new(TriBoolTrue),
					Quota: new(true),
					Options: &LxcBootMountOptions{
						Discard:  new(true),
						LazyTime: new(true),
						NoATime:  new(true),
						NoSuid:   new(true)},
					Replicate:       new(true),
					Storage:         new("local-ext4"),
					SizeInKibibytes: new(LxcMountSize(1048576)),
					rawDisk:         "local-ext4:subvol-101-disk-0"}})}}
	}
}

func Test_LxcBootMount_markMountChanges_Unsafe(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   LxcBootMount
		current *LxcBootMount
		output  lxcUpdateChanges
	}{
		{name: `resize`,
			input: LxcBootMount{
				SizeInKibibytes: new(LxcMountSize(1051648))},
			current: &LxcBootMount{
				SizeInKibibytes: new(LxcMountSize(1048576))},
			output: lxcUpdateChanges{
				resize: []lxcMountResize{{
					sizeInKibibytes: LxcMountSize(1051648),
					id:              "rootfs"}}}},
		{name: `move`,
			input: LxcBootMount{
				Storage: new("local-zfs")},
			current: &LxcBootMount{
				Storage: new("local-ext")},
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
	t.Parallel()
	tests := []struct {
		name    string
		input   LxcBindMount
		current *LxcBindMount
		output  error
	}{
		{name: `valid update HostPath`,
			input: LxcBindMount{
				HostPath: new(LxcHostPath("/test/new"))},
			current: &LxcBindMount{
				HostPath:  new(LxcHostPath("/test/old")),
				GuestPath: new(LxcMountPath("/test/guest"))}},
		{name: `valid update LxcMountPath`,
			input: LxcBindMount{
				GuestPath: new(LxcMountPath("/test/new"))},
			current: &LxcBindMount{
				HostPath:  new(LxcHostPath("/test/old")),
				GuestPath: new(LxcMountPath("/test/guest"))}},
		{name: `invalid create`,
			input:  LxcBindMount{},
			output: errors.New(LxcBindMountErrorHostPathRequired)},
		{name: `invalid update`,
			input: LxcBindMount{
				HostPath: new(LxcHostPath("./test"))},
			current: &LxcBindMount{
				HostPath:  new(LxcHostPath("/test/old")),
				GuestPath: new(LxcMountPath("/test/guest"))},
			output: errors.New(LxcHostPathErrorRelative)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate(test.current))
		})
	}
}

func Test_LxcDataMount_Validate(t *testing.T) {
	t.Parallel()
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
				SizeInKibibytes: new(LxcMountSize(131070))},
			current: &LxcDataMount{
				Storage:         new("test"),
				SizeInKibibytes: new(LxcMountSize(131072))},
			output: errors.New(LxcMountSizeErrorMinimum)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate(test.current, test.privileged))
		})
	}
}

func Test_LxcHostPath_String(t *testing.T) {
	t.Parallel()
	require.Equal(t, string("/mnt/test"), LxcHostPath("/mnt/test").String())
}

func Test_LxcMounts_Validate(t *testing.T) {
	t.Parallel()
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
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path"))}}},
			current: LxcMounts{
				LxcMountID200: LxcMount{
					BindMount: &LxcBindMount{}}}},
		{name: `valid create`,
			input: LxcMounts{
				LxcMountID10: LxcMount{
					DataMount: &LxcDataMount{
						Path:            new(LxcMountPath("/opt/test")),
						SizeInKibibytes: new(LxcMountSize(68343802)),
						Storage:         new("test")}}}},
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
	t.Parallel()
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
					ACL:             new(TriBoolFalse),
					Backup:          new(true),
					SizeInKibibytes: new(LxcMountSize(53763756)),
					Storage:         new("test"),
					Path:            new(LxcMountPath("/mnt/opt"))}}},
		{name: `valid update`,
			privileged: true,
			input: LxcMount{
				DataMount: &LxcDataMount{
					ACL:       new(TriBoolFalse),
					Replicate: new(false),
					Storage:   new("test"),
					Quota:     new(false),
					ReadOnly:  new(false)}},
			current: &LxcMount{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate(test.current, test.privileged))
		})
	}
}

func Test_LxcMountPath_Validate(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	baseMount := func() LxcMount {
		return LxcMount{
			DataMount: &LxcDataMount{
				SizeInKibibytes: new(LxcMountSize(1048576)),
				Storage:         new("local-ext4")}}
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
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path"))}}},
			current: LxcMounts{}},
		{name: `create BindMount current not empty`,
			input: LxcMounts{
				LxcMountID100: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path"))}}},
			current: LxcMounts{
				LxcMountID100: LxcMount{}}},
		{name: `create DataMount current empty`,
			input: LxcMounts{
				LxcMountID101: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: new(LxcMountSize(1048576)),
						Storage:         new("local-ext4")}}},
			current: LxcMounts{}},
		{name: `create DataMount current not empty`,
			input: LxcMounts{
				LxcMountID101: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: new(LxcMountSize(1048576)),
						Storage:         new("local-ext4")}}},
			current: LxcMounts{LxcMountID101: LxcMount{}}},
		{name: `detach BindMount`,
			input: LxcMounts{
				LxcMountID100: LxcMount{
					Detach: true}},
			current: LxcMounts{
				LxcMountID100: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path"))}}},
			output: lxcUpdateChanges{
				offState: true}},
		{name: `detach DataMount`,
			input: LxcMounts{
				LxcMountID101: LxcMount{
					Detach: true}},
			current: LxcMounts{
				LxcMountID101: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: new(LxcMountSize(1048576)),
						Storage:         new("local-ext4")}}},
			output: lxcUpdateChanges{
				offState: true}},
		{name: `recreate BindMount due to host path`,
			input: LxcMounts{
				LxcMountID100: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  new(LxcHostPath("/new/host/path")),
						GuestPath: new(LxcMountPath("/guest/path"))}}},
			current: LxcMounts{
				LxcMountID100: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path"))}}},
			output: lxcUpdateChanges{
				offState: true}},
		{name: `recreate BindMount due to guest path`,
			input: LxcMounts{
				LxcMountID100: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/new/guest/path"))}}},
			current: LxcMounts{
				LxcMountID100: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path"))}}},
			output: lxcUpdateChanges{
				offState: true}},
		{name: `recreate DataMount due to resize`,
			input: LxcMounts{
				LxcMountID200: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: new(LxcMountSize(1048575)),
						Storage:         new("local-zfs")}}},
			current: LxcMounts{
				LxcMountID200: baseMount()},
			output: lxcUpdateChanges{
				offState: true}},
		{name: `replace BindMount with DataMount`,
			input: LxcMounts{
				LxcMountID105: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: new(LxcMountSize(1051648)),
						Storage:         new("local-zfs")}}},
			current: LxcMounts{
				LxcMountID105: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path"))}}},
			output: lxcUpdateChanges{
				offState: true}},
		{name: `replace DataMount with BindMount`,
			input: LxcMounts{
				LxcMountID105: LxcMount{
					BindMount: &LxcBindMount{
						HostPath:  new(LxcHostPath("/host/path")),
						GuestPath: new(LxcMountPath("/guest/path"))}}},
			current: LxcMounts{
				LxcMountID105: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: new(LxcMountSize(1051648)),
						Storage:         new("local-zfs")}}},
			output: lxcUpdateChanges{
				offState: true}},
		{name: `resize`,
			input: LxcMounts{
				LxcMountID105: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: new(LxcMountSize(1051648))}}},
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
						Storage: new("local-zfs")}}},
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
						SizeInKibibytes: new(LxcMountSize(1051648)),
						Storage:         new("local-zfs")}}},
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
						SizeInKibibytes: new(LxcMountSize(1051648)),
						Storage:         new("local-zfs")}},
				LxcMountID80: LxcMount{
					DataMount: &LxcDataMount{
						SizeInKibibytes: new(LxcMountSize(1051648)),
						Storage:         new("local-zfs")}}},
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
	t.Parallel()
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
	t.Parallel()
	tests := []struct {
		name   string
		input  lxcMountMove
		delete bool
		output map[string]string
	}{
		{name: `delete false`,
			input: lxcMountMove{
				id:      "100",
				storage: "local-ext4"},
			output: map[string]string{
				"storage": "local-ext4",
				"volume":  "100"}},
		{name: `delete true`,
			input: lxcMountMove{
				id:      "100",
				storage: "local-ext4"},
			delete: true,
			output: map[string]string{
				"storage": "local-ext4",
				"volume":  "100",
				"delete":  "1"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testParamsEqual(t, test.output, test.input.mapToAPI(test.delete))
		})
	}
}

func Test_RawConfigLXC_GetMounts(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Mounts_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, output.Mounts, raw.GetMounts())
		})
}

func Test_RawConfigLXC_Mounts_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Mounts_Get().Test(t)
}

func Test_RawConfigLXC_GetBootMount(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_BootMount_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, output.BootMount, raw.GetBootMount())
		})
}

func Test_RawConfigLXC_BootMount_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_BootMount_Get().Test(t)
}
