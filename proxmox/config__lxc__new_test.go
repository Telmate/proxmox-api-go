package proxmox

import (
	"context"
	"crypto/sha1"
	"errors"
	"maps"
	"net"
	"net/netip"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_guest"
	"github.com/stretchr/testify/require"
)

func Test_CpuArchitecture_String(t *testing.T) {
	require.Equal(t, "amd64", CpuArchitecture("amd64").String())
}

func Test_ConfigLXC_mapToAPI(t *testing.T) {
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
	parseMAC := func(rawMAC string) net.HardwareAddr {
		mac, err := net.ParseMAC(rawMAC)
		failPanic(err)
		return mac
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
	network := func() LxcNetwork {
		return LxcNetwork{
			Bridge:    util.Pointer("vmbr0"),
			Connected: util.Pointer(false),
			Firewall:  util.Pointer(true),
			IPv4: &LxcIPv4{
				Address: util.Pointer(IPv4CIDR("192.168.10.12/24")),
				Gateway: util.Pointer(IPv4Address("192.168.10.1"))},
			IPv6: &LxcIPv6{
				Address: util.Pointer(IPv6CIDR("2001:db8::1234/64")),
				Gateway: util.Pointer(IPv6Address("2001:db8::1"))},
			MAC:           util.Pointer(parseMAC("52:A4:00:12:b4:56")),
			Mtu:           util.Pointer(MTU(1500)),
			Name:          util.Pointer(LxcNetworkName("my_net")),
			NativeVlan:    util.Pointer(Vlan(23)),
			RateLimitKBps: util.Pointer(GuestNetworkRate(45)),
			TaggedVlans:   util.Pointer(Vlans{12, 23, 45}),
			mac:           "52:A4:00:12:b4:56",
		}
	}
	publicKeys := func() []AuthorizedKey {
		data := test_data_guest.AuthorizedKey_Decoded_Input()
		keys := make([]AuthorizedKey, len(data))
		for i := range data {
			keys[i] = AuthorizedKey{Options: data[i].Options, PublicKey: data[i].PublicKey, Comment: data[i].Comment}
		}
		return keys
	}
	type defaults uint8
	const (
		none defaults = iota
		create
		update
		all
	)
	type test struct {
		name          string
		config        ConfigLXC
		currentConfig ConfigLXC
		output        map[string]any
		omitDefaults  defaults // opt out of default values in returned map
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
					omitDefaults: all,
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
					omitDefaults: all,
					output:       map[string]any{"rootfs": ",quota=1"}}},
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
			update: []test{
				{name: `all storage change, no api `,
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
					omitDefaults: all,
					output:       map[string]any{}},
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
					omitDefaults: all,
					output:       map[string]any{}},
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
					omitDefaults: all,
					output:       map[string]any{}},
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
					omitDefaults: all,
					output:       map[string]any{}},
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
					omitDefaults: all,
					output:       map[string]any{}}}},
		{category: `CPU`,
			createUpdate: []test{
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
					omitDefaults:  update,
					output:        map[string]any{}},
				{name: `Limit delete no effect`,
					config:        ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(0))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{}},
					omitDefaults:  update,
					output:        map[string]any{}},
				{name: `Units delete no effect`,
					config:        ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(0))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{}},
					omitDefaults:  update,
					output:        map[string]any{}}},
			update: []test{
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
					omitDefaults:  all,
					output:        map[string]any{}},
				{name: `Limit same`,
					config:        ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(2))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(2))}},
					omitDefaults:  all,
					output:        map[string]any{}},
				{name: `Units same`,
					config:        ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(3))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(3))}},
					omitDefaults:  all,
					output:        map[string]any{}},
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
					omitDefaults:  all,
					output:        map[string]any{}},
				{name: `Limit delete no current`,
					config:        ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(0))}},
					currentConfig: ConfigLXC{},
					omitDefaults:  all,
					output:        map[string]any{}},
				{name: `Units delete no current`,
					config:        ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(0))}},
					currentConfig: ConfigLXC{},
					omitDefaults:  all,
					output:        map[string]any{}}}},
		{category: `CreateOptions`,
			create: []test{
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
						"ssh-public-keys": string(test_data_guest.AuthorizedKey_Encoded_Output())}},
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
						"ssh-public-keys": string(test_data_guest.AuthorizedKey_Encoded_Output())}},
				{name: `PublicSSHkeys empty`,
					config: ConfigLXC{CreateOptions: &LxcCreateOptions{
						PublicSSHkeys: []AuthorizedKey{}}},
					output: map[string]any{}}},
			update: []test{
				{name: `all do nothing`,
					config: ConfigLXC{CreateOptions: &LxcCreateOptions{
						OsTemplate: &LxcTemplate{
							Storage: "local",
							File:    "test-template"},
						UserPassword:  util.Pointer("myPassword!"),
						PublicSSHkeys: publicKeys()}},
					omitDefaults: all,
					output:       map[string]any{}}}},
		{category: `Description`,
			createUpdate: []test{
				{name: `set`,
					config:        ConfigLXC{Description: util.Pointer("test")},
					currentConfig: ConfigLXC{Description: util.Pointer("text")},
					output:        map[string]any{"description": "test"}},
				{name: `delete no effect`,
					config:        ConfigLXC{Description: util.Pointer("")},
					currentConfig: ConfigLXC{Description: util.Pointer("")},
					omitDefaults:  update,
					output:        map[string]any{}}},
			update: []test{
				{name: `delete`,
					config:        ConfigLXC{Description: util.Pointer("")},
					currentConfig: ConfigLXC{Description: util.Pointer("test")},
					output:        map[string]any{"delete": "description"}},
				{name: `same`,
					config:        ConfigLXC{Description: util.Pointer("test")},
					currentConfig: ConfigLXC{Description: util.Pointer("test")},
					omitDefaults:  all,
					output:        map[string]any{}}}},
		{category: `Digest`,
			update: []test{
				{name: `not set when only setting`,
					config:        ConfigLXC{},
					currentConfig: ConfigLXC{rawDigest: "af064923bbf2301596aac4c273ba32178ebc4a96"},
					omitDefaults:  update,
					output:        map[string]any{}},
				{name: `set`,
					config:        ConfigLXC{Description: util.Pointer("test")},
					currentConfig: ConfigLXC{rawDigest: "af064923bbf2301596aac4c273ba32178ebc4a96"},
					omitDefaults:  update,
					output: map[string]any{
						"digest":      "af064923bbf2301596aac4c273ba32178ebc4a96",
						"description": "test"}}}},
		{category: `DNS`,
			createUpdate: []test{
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
			create: []test{
				{name: `do nothing`,
					config: ConfigLXC{DNS: &GuestDNS{
						NameServers:  util.Pointer([]netip.Addr{}),
						SearchDomain: util.Pointer("")}},
					output: map[string]any{}}},
			update: []test{
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
					omitDefaults: all,
					output:       map[string]any{}}}},
		{category: `Features`,
			create: []test{
				{name: `all false Privileged`,
					config: ConfigLXC{Features: featuresPrivileged(false)},
					output: map[string]any{}},
				{name: `all false Unprivileged`,
					config: ConfigLXC{Features: featuresUnprivileged(false)},
					output: map[string]any{}}},
			createUpdate: []test{
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
					omitDefaults:  update,
					output:        map[string]any{}},
				{name: `delete no effect false Unprivileged`,
					config:        ConfigLXC{Features: featuresUnprivileged(false)},
					currentConfig: ConfigLXC{Features: featuresUnprivileged(false)},
					omitDefaults:  update,
					output:        map[string]any{}},
				{name: `only top-level set, no effect`,
					config:        ConfigLXC{Features: &LxcFeatures{}},
					currentConfig: ConfigLXC{Features: &LxcFeatures{}},
					omitDefaults:  update,
					output:        map[string]any{}}},
			update: []test{
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
					omitDefaults:  all,
					output:        map[string]any{}},
				{name: `delete no effect nil Unprivileged`,
					config:        ConfigLXC{Features: featuresUnprivileged(false)},
					currentConfig: ConfigLXC{Features: nil},
					omitDefaults:  all,
					output:        map[string]any{}},
				{name: `same false Privileged`,
					config:        ConfigLXC{Features: featuresPrivileged(false)},
					currentConfig: ConfigLXC{Features: featuresPrivileged(false)},
					omitDefaults:  all,
					output:        map[string]any{}},
				{name: `same false Unprivileged`,
					config:        ConfigLXC{Features: featuresUnprivileged(false)},
					currentConfig: ConfigLXC{Features: featuresUnprivileged(false)},
					omitDefaults:  all,
					output:        map[string]any{}},
				{name: `same true Privileged`,
					config:        ConfigLXC{Features: featuresPrivileged(true)},
					currentConfig: ConfigLXC{Features: featuresPrivileged(true)},
					omitDefaults:  all,
					output:        map[string]any{}},
				{name: `same true Unprivileged`,
					config:        ConfigLXC{Features: featuresUnprivileged(true)},
					currentConfig: ConfigLXC{Features: featuresUnprivileged(true)},
					omitDefaults:  all,
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
					omitDefaults:  all,
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
					omitDefaults:  all,
					output:        map[string]any{}}}},
		{category: `Mount`,
			create: []test{
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
					omitDefaults: all,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID19: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota: util.Pointer(false)})}},
						Privileged: util.Pointer(false)},
					output: map[string]any{
						"mp19":         string("local-zfs:1"),
						"unprivileged": int(1)}},
				{name: `DataMount.Quota false Privileged true`,
					omitDefaults: all,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID20: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota: util.Pointer(false)})}},
						Privileged: util.Pointer(true)},
					output: map[string]any{
						"mp20": string("local-zfs:1")}},
				{name: `DataMount.Quota false Privileged unset`,
					omitDefaults: all,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID18: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota: util.Pointer(false)})}}},
					output: map[string]any{
						"mp18":         string("local-zfs:1"),
						"unprivileged": int(1)}},
				{name: `DataMount.Quota true Privileged false`,
					omitDefaults: all,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID22: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota: util.Pointer(true)})}},
						Privileged: util.Pointer(false)},
					output: map[string]any{
						"mp22":         string("local-zfs:1"),
						"unprivileged": int(1)}},
				{name: `DataMount.Quota true Privileged true`,
					omitDefaults: all,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID23: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota: util.Pointer(true)})}},
						Privileged: util.Pointer(true)},
					output: map[string]any{
						"mp23": string("local-zfs:1,quota=1")}},
				{name: `DataMount.Quota true Privileged unset`,
					omitDefaults: all,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID24: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota: util.Pointer(true)})}}},
					output: map[string]any{
						"mp24":         string("local-zfs:1"),
						"unprivileged": int(1)}},
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
			createUpdate: []test{
				{name: `Detach non-existing`,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID29: LxcMount{Detach: true}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{
						LxcMountID30: LxcMount{}}},
					omitDefaults: update,
					output:       map[string]any{}}},
			update: []test{
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
					omitDefaults:  all,
					output:        map[string]any{}},
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
					omitDefaults: all,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID80: LxcMount{DataMount: &LxcDataMount{
							Quota: util.Pointer(false)}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID80: LxcMount{DataMount: dataMount()}}},
					output:        map[string]any{}},
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
					omitDefaults: all,
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID82: LxcMount{DataMount: &LxcDataMount{
							Quota: util.Pointer(true)}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID82: LxcMount{DataMount: dataMount()}}},
					output:        map[string]any{}},
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
					omitDefaults:  all,
					output:        map[string]any{}},
				{name: `DataMount.Storage change, all other unchanged`, // this is manaaged by other mechanisms
					config: ConfigLXC{Mounts: LxcMounts{
						LxcMountID73: LxcMount{DataMount: &LxcDataMount{
							Storage: util.Pointer("test-storage")}}}},
					currentConfig: ConfigLXC{Mounts: LxcMounts{LxcMountID73: LxcMount{DataMount: dataMount()}}},
					omitDefaults:  all,
					output:        map[string]any{}},
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
			createUpdate: []test{
				{name: `set`,
					config:        ConfigLXC{Name: util.Pointer(GuestName("test"))},
					currentConfig: ConfigLXC{Name: util.Pointer(GuestName("text"))},
					output:        map[string]any{"hostname": string("test")}}},
			update: []test{
				{name: `do nothing`,
					config:        ConfigLXC{Name: util.Pointer(GuestName("test"))},
					currentConfig: ConfigLXC{Name: util.Pointer(GuestName("test"))},
					omitDefaults:  all,
					output:        map[string]any{}}}},
		{category: `Networks`,
			create: []test{
				{name: `Delete`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID3: LxcNetwork{Delete: true}}},
					output: map[string]any{}}},
			createUpdate: []test{
				{name: `create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID1: LxcNetwork{Bridge: util.Pointer("vmbr0")}}},
					currentConfig: ConfigLXC{},
					output: map[string]any{
						"net1": ",bridge=vmbr0"}},
				{name: `delete no effect`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID1: LxcNetwork{Delete: true}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID5: LxcNetwork{Bridge: util.Pointer("vmbr0")}}},
					omitDefaults: update,
					output:       map[string]any{}},
				{name: `Bridge`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID0: LxcNetwork{Bridge: util.Pointer("vmbr0")}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID0: LxcNetwork{Bridge: util.Pointer("vmbr1")}}},
					output: map[string]any{
						"net0": ",bridge=vmbr0"}},
				{name: `Connected true`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID1: LxcNetwork{Connected: util.Pointer(true)}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID1: LxcNetwork{Connected: util.Pointer(false)}}},
					output: map[string]any{
						"net1": ""}},
				{name: `Connected false`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID2: LxcNetwork{Connected: util.Pointer(false)}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID2: LxcNetwork{Connected: util.Pointer(true)}}},
					output: map[string]any{
						"net2": ",link_down=1"}},
				{name: `Firewall true`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID4: LxcNetwork{Firewall: util.Pointer(true)}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID4: LxcNetwork{Firewall: util.Pointer(false)}}},
					output: map[string]any{"net4": ",firewall=1"}},
				{name: `Firewall false`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID5: LxcNetwork{Firewall: util.Pointer(false)}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID5: LxcNetwork{Firewall: util.Pointer(true)}}},
					output: map[string]any{"net5": ""}},
				{name: `IPv4.Address create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID6: LxcNetwork{IPv4: &LxcIPv4{
							Address: util.Pointer(IPv4CIDR("10.0.0.10/24"))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID6: LxcNetwork{}}},
					output: map[string]any{"net6": ",ip=10.0.0.10/24"}},
				{name: `IPv4.Address empty`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID6: LxcNetwork{IPv4: &LxcIPv4{
							Address: util.Pointer(IPv4CIDR(""))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID6: LxcNetwork{IPv4: &LxcIPv4{
							Address: util.Pointer(IPv4CIDR("10.0.0.10/24"))}}}},
					output: map[string]any{"net6": ""}},
				{name: `IPv4.DHCP create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID7: LxcNetwork{IPv4: &LxcIPv4{
							DHCP: true}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID7: LxcNetwork{}}},
					output: map[string]any{"net7": ",ip=dhcp"}},
				{name: `IPv4.DHCP false`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID7: LxcNetwork{IPv4: &LxcIPv4{
							DHCP: false}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID7: LxcNetwork{IPv4: &LxcIPv4{
							DHCP: true}}}},
					output: map[string]any{"net7": ""}},
				{name: `IPv4.Gateway create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID8: LxcNetwork{IPv4: &LxcIPv4{
							Gateway: util.Pointer(IPv4Address("10.0.0.1"))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID8: LxcNetwork{}}},
					output: map[string]any{"net8": ",gw=10.0.0.1"}},
				{name: `IPv4.Gateway empty`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID8: LxcNetwork{IPv4: &LxcIPv4{
							Gateway: util.Pointer(IPv4Address(""))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID8: LxcNetwork{IPv4: &LxcIPv4{
							Gateway: util.Pointer(IPv4Address("10.0.0.1"))}}}},
					output: map[string]any{"net8": ""}},
				{name: `IPv4.Manual create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID9: LxcNetwork{IPv4: &LxcIPv4{
							Manual: true}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID9: LxcNetwork{}}},
					output: map[string]any{"net9": ",ip=manual"}},
				{name: `IPv4.Manual false`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID9: LxcNetwork{IPv4: &LxcIPv4{
							Manual: false}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID9: LxcNetwork{IPv4: &LxcIPv4{
							Manual: true}}}},
					output: map[string]any{"net9": ""}},
				{name: `IPv6.Address create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID10: LxcNetwork{IPv6: &LxcIPv6{
							Address: util.Pointer(IPv6CIDR("2001:db8::1/64"))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID10: LxcNetwork{}}},
					output: map[string]any{"net10": ",ip6=2001:db8::1/64"}},
				{name: `IPv6.Address empty`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID10: LxcNetwork{IPv6: &LxcIPv6{
							Address: util.Pointer(IPv6CIDR(""))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID10: LxcNetwork{IPv6: &LxcIPv6{
							Address: util.Pointer(IPv6CIDR("2001:db8::1/64"))}}}},
					output: map[string]any{"net10": ""}},
				{name: `IPv6.DHCP create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID11: LxcNetwork{IPv6: &LxcIPv6{
							DHCP: true}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID11: LxcNetwork{}}},
					output: map[string]any{"net11": ",ip6=dhcp"}},
				{name: `IPv6.DHCP false`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID11: LxcNetwork{IPv6: &LxcIPv6{
							DHCP: false}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID11: LxcNetwork{IPv6: &LxcIPv6{
							DHCP: true}}}},
					output: map[string]any{"net11": ""}},
				{name: `IPv6.Gateway create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID12: LxcNetwork{IPv6: &LxcIPv6{
							Gateway: util.Pointer(IPv6Address("2001:db8::2"))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID12: LxcNetwork{}}},
					output: map[string]any{"net12": ",gw6=2001:db8::2"}},
				{name: `IPv6.Gateway empty`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID12: LxcNetwork{IPv6: &LxcIPv6{
							Gateway: util.Pointer(IPv6Address(""))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID12: LxcNetwork{IPv6: &LxcIPv6{
							Gateway: util.Pointer(IPv6Address("2001:db8::2"))}}}},
					output: map[string]any{"net12": ""}},
				{name: `IPv6.SLAAC create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID13: LxcNetwork{IPv6: &LxcIPv6{
							SLAAC: true}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID13: LxcNetwork{}}},
					output: map[string]any{"net13": ",ip6=auto"}},
				{name: `IPv6.SLAAC false`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID13: LxcNetwork{IPv6: &LxcIPv6{
							SLAAC: false}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID13: LxcNetwork{IPv6: &LxcIPv6{
							SLAAC: true}}}},
					output: map[string]any{"net13": ""}},
				{name: `IPv6.Manual create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID14: LxcNetwork{IPv6: util.Pointer(LxcIPv6{
							Manual: true})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID14: LxcNetwork{}}},
					output: map[string]any{"net14": ",ip6=manual"}},
				{name: `MAC set`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID15: LxcNetwork{MAC: util.Pointer(parseMAC("00:11:22:33:44:55"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID15: LxcNetwork{MAC: util.Pointer(parseMAC("00:11:22:33:44:66"))}}},
					output: map[string]any{"net15": ",hwaddr=00:11:22:33:44:55"}},
				{name: `MAC unset`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID5: LxcNetwork{MAC: util.Pointer(net.HardwareAddr{})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID5: LxcNetwork{MAC: util.Pointer(parseMAC("00:11:22:33:44:66"))}}},
					output: map[string]any{"net5": ""}},
				{name: `Mtu set`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID0: LxcNetwork{Mtu: util.Pointer(MTU(1500))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID0: LxcNetwork{Mtu: util.Pointer(MTU(1400))}}},
					output: map[string]any{"net0": ",mtu=1500"}},
				{name: `Mtu unset`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID1: LxcNetwork{Mtu: util.Pointer(MTU(0))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID1: LxcNetwork{Mtu: util.Pointer(MTU(1400))}}},
					output: map[string]any{"net1": ""}},
				{name: `Name`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID2: LxcNetwork{Name: util.Pointer(LxcNetworkName("test0"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID2: LxcNetwork{Name: util.Pointer(LxcNetworkName("text0"))}}},
					output: map[string]any{
						"net2": "name=test0"}},
				{name: `NativeVlan set`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID3: LxcNetwork{NativeVlan: util.Pointer(Vlan(100))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID3: LxcNetwork{NativeVlan: util.Pointer(Vlan(200))}}},
					output: map[string]any{"net3": ",tag=100"}},
				{name: `NativeVlan unset`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID4: LxcNetwork{NativeVlan: util.Pointer(Vlan(0))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID4: LxcNetwork{NativeVlan: util.Pointer(Vlan(200))}}},
					output: map[string]any{"net4": ""}},
				{name: `RateLimitKBps set`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID5: LxcNetwork{RateLimitKBps: util.Pointer(GuestNetworkRate(1023))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID5: LxcNetwork{RateLimitKBps: util.Pointer(GuestNetworkRate(1024))}}},
					output: map[string]any{"net5": ",rate=1.023"}},
				{name: `RateLimitKBps unset`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID6: LxcNetwork{RateLimitKBps: util.Pointer(GuestNetworkRate(0))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID6: LxcNetwork{RateLimitKBps: util.Pointer(GuestNetworkRate(1024))}}},
					output: map[string]any{"net6": ""}},
				{name: `TaggedVlans set`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID7: LxcNetwork{TaggedVlans: util.Pointer(Vlans{Vlan(100), Vlan(200)})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID7: LxcNetwork{TaggedVlans: util.Pointer(Vlans{Vlan(100), Vlan(300)})}}},
					output: map[string]any{"net7": ",trunks=100;200"}},
				{name: `TaggedVlans unset`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID8: LxcNetwork{TaggedVlans: util.Pointer(Vlans{})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID8: LxcNetwork{TaggedVlans: util.Pointer(Vlans{Vlan(100), Vlan(200)})}}},
					output: map[string]any{"net8": ""}}},
			update: []test{
				{name: `create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID0: network()}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID1: network()}},
					output: map[string]any{"net0": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `delete`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: LxcNetwork{Delete: true}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: network()}},
					output:        map[string]any{"delete": "net0"}},
				{name: `no change`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: network()}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: network()}},
					omitDefaults:  all,
					output:        map[string]any{}},
				{name: `Bridge replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: LxcNetwork{Bridge: util.Pointer("vmbr3")}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: network()}},
					output:        map[string]any{"net0": "name=my_net,bridge=vmbr3,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `Connected replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID1: LxcNetwork{Connected: util.Pointer(true)}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID1: network()}},
					output:        map[string]any{"net1": "name=my_net,bridge=vmbr0,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `Firewall replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID2: LxcNetwork{Firewall: util.Pointer(false)}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID2: network()}},
					output:        map[string]any{"net2": "name=my_net,bridge=vmbr0,link_down=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv4.Address inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0")),
						IPv4: &LxcIPv4{}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: LxcNetwork{IPv4: &LxcIPv4{Address: util.Pointer(IPv4CIDR("192.168.1.34/24"))}}}},
					output:        map[string]any{"net4": "name=test0,ip=192.168.1.34/24"}},
				{name: `IPv4.Address replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID3: LxcNetwork{IPv4: &LxcIPv4{Address: util.Pointer(IPv4CIDR("10.0.0.2/24"))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID3: network()}},
					output:        map[string]any{"net3": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=10.0.0.2/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv4.DHCP inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: LxcNetwork{IPv4: &LxcIPv4{DHCP: true}}}},
					output:        map[string]any{"net4": "name=test0,ip=dhcp"}},
				{name: `IPv4.DHCP replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: LxcNetwork{IPv4: &LxcIPv4{DHCP: true}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: network()}},
					output:        map[string]any{"net4": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=dhcp,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv4.Manual inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID8: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID8: LxcNetwork{IPv4: &LxcIPv4{Manual: true}}}},
					output:        map[string]any{"net8": "name=test0,ip=manual"}},
				{name: `IPv4.Manual replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID6: LxcNetwork{IPv4: &LxcIPv4{Manual: true}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID6: network()}},
					output:        map[string]any{"net6": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=manual,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv4.Gateway inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID5: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0")),
						IPv4: &LxcIPv4{}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID5: LxcNetwork{IPv4: &LxcIPv4{Gateway: util.Pointer(IPv4Address("1.1.1.1"))}}}},
					output:        map[string]any{"net5": "name=test0,gw=1.1.1.1"}},
				{name: `IPv4.Gateway replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID5: LxcNetwork{IPv4: &LxcIPv4{Gateway: util.Pointer(IPv4Address("1.1.1.1"))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID5: network()}},
					output:        map[string]any{"net5": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=1.1.1.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv6.Address inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0")),
						IPv6: util.Pointer(LxcIPv6{})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: LxcNetwork{IPv6: &LxcIPv6{Address: util.Pointer(IPv6CIDR("2001:db8::2/64"))}}}},
					output:        map[string]any{"net12": "name=test0,ip6=2001:db8::2/64"}},
				{name: `IPv6.Address replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID7: LxcNetwork{IPv6: &LxcIPv6{Address: util.Pointer(IPv6CIDR("2001:db8::2/64"))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID7: network()}},
					output:        map[string]any{"net7": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::2/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv6.DHCP inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: LxcNetwork{IPv6: &LxcIPv6{DHCP: true}}}},
					output:        map[string]any{"net12": "name=test0,ip6=dhcp"}},
				{name: `IPv6.DHCP replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID8: LxcNetwork{IPv6: &LxcIPv6{DHCP: true}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID8: network()}},
					output:        map[string]any{"net8": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=dhcp,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv6.SLAAC inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: LxcNetwork{IPv6: &LxcIPv6{SLAAC: true}}}},
					output:        map[string]any{"net13": "name=test0,ip6=auto"}},
				{name: `IPv6.SLAAC replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID10: LxcNetwork{IPv6: &LxcIPv6{SLAAC: true}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID10: network()}},
					output:        map[string]any{"net10": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=auto,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv6.Manual inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: LxcNetwork{IPv6: &LxcIPv6{Manual: true}}}},
					output:        map[string]any{"net13": "name=test0,ip6=manual"}},
				{name: `IPv6.Manual replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID11: LxcNetwork{IPv6: &LxcIPv6{Manual: true}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID11: network()}},
					output:        map[string]any{"net11": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=manual,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv6.Gateway inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID9: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0")),
						IPv6: &LxcIPv6{}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID9: LxcNetwork{IPv6: &LxcIPv6{Gateway: util.Pointer(IPv6Address("2001:db8::3"))}}}},
					output:        map[string]any{"net9": "name=test0,gw6=2001:db8::3"}},
				{name: `IPv6.Gateway replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID9: LxcNetwork{IPv6: &LxcIPv6{Gateway: util.Pointer(IPv6Address("2001:db8::3"))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID9: network()}},
					output:        map[string]any{"net9": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::3,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `MAC replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: LxcNetwork{MAC: util.Pointer(parseMAC("00:11:a2:B3:44:66"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: network()}},
					output:        map[string]any{"net12": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=00:11:A2:B3:44:66,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `Name replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: LxcNetwork{Name: util.Pointer(LxcNetworkName("test0"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: network()}},
					output:        map[string]any{"net13": "name=test0,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `NativeVlan replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID14: LxcNetwork{NativeVlan: util.Pointer(Vlan(200))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID14: network()}},
					output:        map[string]any{"net14": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=200,rate=0.045,trunks=12;23;45"}},
				{name: `RateLimitKBps replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID15: LxcNetwork{RateLimitKBps: util.Pointer(GuestNetworkRate(2040))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID15: network()}},
					output:        map[string]any{"net15": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=2.04,trunks=12;23;45"}},
				{name: `TaggedVlans replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: LxcNetwork{TaggedVlans: util.Pointer(Vlans{Vlan(200), Vlan(100), Vlan(300)})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: network()}},
					output:        map[string]any{"net0": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=100;200;300"}}}},
		{category: `Node`,
			createUpdate: []test{
				{name: `do nothing`,
					config:        ConfigLXC{Node: util.Pointer(NodeName("test"))},
					currentConfig: ConfigLXC{Node: util.Pointer(NodeName("text"))},
					omitDefaults:  update,
					output:        map[string]any{}}}},
		{category: `OperatingSystem`,
			createUpdate: []test{
				{name: `do nothing`,
					config:        ConfigLXC{OperatingSystem: "test"},
					currentConfig: ConfigLXC{OperatingSystem: "text"},
					omitDefaults:  update,
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
					omitDefaults:  all,
					output:        map[string]any{}}}},
		{category: `Privileged`,
			create: []test{
				{name: `true`,
					config:       ConfigLXC{Privileged: util.Pointer(true)},
					omitDefaults: all,
					output:       map[string]any{}},
				{name: `false`,
					config: ConfigLXC{Privileged: util.Pointer(false)},
					output: map[string]any{"unprivileged": int(1)}}},
			update: []test{
				{name: `true no effect`,
					config:        ConfigLXC{Privileged: util.Pointer(true)},
					currentConfig: ConfigLXC{Privileged: util.Pointer(false)},
					omitDefaults:  all,
					output:        map[string]any{}},
				{name: `false no effect`,
					config:        ConfigLXC{Privileged: util.Pointer(false)},
					currentConfig: ConfigLXC{Privileged: util.Pointer(true)},
					omitDefaults:  all,
					output:        map[string]any{}}}},
		{category: `Protection`,
			create: []test{
				{name: `set false`,
					config:        ConfigLXC{Protection: util.Pointer(false)},
					currentConfig: ConfigLXC{},
					output:        map[string]any{}}},
			createUpdate: []test{
				{name: `set true`,
					config:        ConfigLXC{Protection: util.Pointer(true)},
					currentConfig: ConfigLXC{},
					output:        map[string]any{"protection": string("1")}}},
			update: []test{
				{name: `do nothing false`,
					config:        ConfigLXC{Protection: util.Pointer(false)},
					currentConfig: ConfigLXC{Protection: util.Pointer(false)},
					omitDefaults:  all,
					output:        map[string]any{}},
				{name: `do nothing true`,
					config:        ConfigLXC{Protection: util.Pointer(true)},
					currentConfig: ConfigLXC{Protection: util.Pointer(true)},
					omitDefaults:  all,
					output:        map[string]any{}},
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
		{category: `Swap`,
			createUpdate: []test{
				{name: `set`,
					config:        ConfigLXC{Swap: util.Pointer(LxcSwap(256))},
					currentConfig: ConfigLXC{Swap: util.Pointer(LxcSwap(128))},
					output:        map[string]any{"swap": int(256)}},
				{name: `set 0`,
					config:        ConfigLXC{Swap: util.Pointer(LxcSwap(0))},
					currentConfig: ConfigLXC{Swap: util.Pointer(LxcSwap(128))},
					output:        map[string]any{"swap": int(0)}}},
			update: []test{
				{name: `do nothing`,
					config:        ConfigLXC{Swap: util.Pointer(LxcSwap(256))},
					currentConfig: ConfigLXC{Swap: util.Pointer(LxcSwap(256))},
					omitDefaults:  all,
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
					omitDefaults:  all,
					output:        map[string]any{}}}},
	}
	for _, test := range tests {
		for _, subTest := range append(test.create, test.createUpdate...) {
			name := test.category + "/Create/" + subTest.name
			t.Run(name, func(*testing.T) {
				tmpParams, pool := subTest.config.mapToApiCreate()
				clone := maps.Clone(subTest.output)
				if !(subTest.omitDefaults == all || subTest.omitDefaults == create) {
					if _, isSet := clone["unprivileged"]; !isSet {
						clone["unprivileged"] = int(1) // Default to unprivileged
					}
				}
				require.Equal(t, clone, tmpParams, name)
				require.Equal(t, subTest.pool, pool, name)
			})
		}
		for _, subTest := range append(test.update, test.createUpdate...) {
			name := test.category + "/Update/" + subTest.name
			t.Run(name, func(*testing.T) {
				tmpParams := subTest.config.mapToApiUpdate(subTest.currentConfig)
				clone := maps.Clone(subTest.output)
				if !(subTest.omitDefaults == all || subTest.omitDefaults == update) {
					if _, isSet := clone["digest"]; !isSet {
						clone["digest"] = "" // set empty digest
					}
				}
				require.Equal(t, clone, tmpParams, name)
			})
		}
	}
}

func Test_ConfigLXC_Validate(t *testing.T) {
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

func Test_ConfigLXC_Update(t *testing.T) {
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

func Test_RawConfigLXC_Get(t *testing.T) {
	parseIP := func(rawIP string) netip.Addr {
		ip, err := netip.ParseAddr(rawIP)
		failPanic(err)
		return ip
	}
	parseMAC := func(rawMAC string) net.HardwareAddr {
		mac, err := net.ParseMAC(rawMAC)
		failPanic(err)
		return mac
	}
	baseConfig := func(config ConfigLXC) *ConfigLXC {
		if config.ID == nil {
			config.ID = util.Pointer(GuestID(0))
		}
		if config.Memory == nil {
			config.Memory = util.Pointer(LxcMemory(0))
		}
		if config.Name == nil {
			config.Name = util.Pointer(GuestName(""))
		}
		if config.Networks == nil {
			config.Networks = make(LxcNetworks)
		}
		if config.Node == nil {
			config.Node = util.Pointer(NodeName(""))
		}
		if config.Privileged == nil {
			config.Privileged = util.Pointer(true)
		}
		if config.Protection == nil {
			config.Protection = util.Pointer(false)
		}
		if config.Swap == nil {
			config.Swap = util.Pointer(LxcSwap(0))
		}
		return &config
	}
	baseBootMount := func(config LxcBootMount) *LxcBootMount {
		if config.ACL == nil {
			config.ACL = util.Pointer(TriBoolNone)
		}
		if config.Quota == nil {
			config.Quota = util.Pointer(false)
		}
		if config.Replicate == nil {
			config.Replicate = util.Pointer(true)
		}
		if config.Storage == nil {
			config.Storage = util.Pointer("")
		}
		if config.SizeInKibibytes == nil {
			config.SizeInKibibytes = util.Pointer(LxcMountSize(0))
		}
		return &config
	}
	baseBindMount := func(config LxcBindMount) *LxcBindMount {
		if config.GuestPath == nil {
			config.GuestPath = util.Pointer(LxcMountPath(""))
		}
		if config.HostPath == nil {
			config.HostPath = util.Pointer(LxcHostPath(""))
		}
		if config.ReadOnly == nil {
			config.ReadOnly = util.Pointer(false)
		}
		if config.Replicate == nil {
			config.Replicate = util.Pointer(true)
		}
		return &config
	}
	baseDataMount := func(config LxcDataMount) *LxcDataMount {
		if config.ACL == nil {
			config.ACL = util.Pointer(TriBoolNone)
		}
		if config.Backup == nil {
			config.Backup = util.Pointer(false)
		}
		if config.Path == nil {
			config.Path = util.Pointer(LxcMountPath(""))
		}
		if config.ReadOnly == nil {
			config.ReadOnly = util.Pointer(false)
		}
		if config.Replicate == nil {
			config.Replicate = util.Pointer(true)
		}
		if config.SizeInKibibytes == nil {
			config.SizeInKibibytes = util.Pointer(LxcMountSize(0))
		}
		return &config
	}
	baseMountOptions := func(config LxcMountOptions) *LxcMountOptions {
		if config.Discard == nil {
			config.Discard = util.Pointer(false)
		}
		if config.LazyTime == nil {
			config.LazyTime = util.Pointer(false)
		}
		if config.NoATime == nil {
			config.NoATime = util.Pointer(false)
		}
		if config.NoDevice == nil {
			config.NoDevice = util.Pointer(false)
		}
		if config.NoExec == nil {
			config.NoExec = util.Pointer(false)
		}
		if config.NoSuid == nil {
			config.NoSuid = util.Pointer(false)
		}
		return &config
	}
	baseNetwork := func(config LxcNetwork) LxcNetwork {
		if config.Bridge == nil {
			config.Bridge = util.Pointer("")
		}
		if config.Connected == nil {
			config.Connected = util.Pointer(true)
		}
		if config.Firewall == nil {
			config.Firewall = util.Pointer(false)
		}
		if config.Name == nil {
			config.Name = util.Pointer(LxcNetworkName(""))
		}
		if config.MAC == nil {
			var mac net.HardwareAddr
			config.MAC = util.Pointer(mac)
		}
		return config
	}
	type test struct {
		name   string
		input  map[string]any
		vmr    VmRef
		state  PowerState
		output *ConfigLXC
		err    error
	}
	tests := []struct {
		category string
		tests    []test
	}{
		{category: `Error`,
			tests: []test{{err: errors.New("this should propagate")}}},
		{category: `Architecture`,
			tests: []test{
				{name: `amd64`,
					input:  map[string]any{"arch": "amd64"},
					output: baseConfig(ConfigLXC{Architecture: "amd64"})},
				{name: `""`,
					input:  map[string]any{"arch": ""},
					output: baseConfig(ConfigLXC{Architecture: ""})}}},
		{category: `BootMount`,
			tests: []test{
				{name: `ACL true`,
					input: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,acl=1"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						ACL:     util.Pointer(TriBoolTrue),
						Storage: util.Pointer("local-zfs"),
						rawDisk: "local-zfs:subvol-101-disk-0"})})},
				{name: `ACL false`,
					input: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,acl=0"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						ACL:     util.Pointer(TriBoolFalse),
						Storage: util.Pointer("local-zfs"),
						rawDisk: "local-zfs:subvol-101-disk-0"})})},
				{name: `Options Discard true`,
					input: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=discard"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(true),
							LazyTime: util.Pointer(false),
							NoATime:  util.Pointer(false),
							NoSuid:   util.Pointer(false)},
						Storage: util.Pointer("local-zfs"),
						rawDisk: "local-zfs:subvol-101-disk-0"})})},
				{name: `Options LazyTime true`,
					input: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=lazytime"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(false),
							LazyTime: util.Pointer(true),
							NoATime:  util.Pointer(false),
							NoSuid:   util.Pointer(false)},
						Storage: util.Pointer("local-zfs"),
						rawDisk: "local-zfs:subvol-101-disk-0"})})},
				{name: `Options NoATime true`,
					input: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=noatime"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(false),
							LazyTime: util.Pointer(false),
							NoATime:  util.Pointer(true),
							NoSuid:   util.Pointer(false)},
						Storage: util.Pointer("local-zfs"),
						rawDisk: "local-zfs:subvol-101-disk-0"})})},
				{name: `Options NoSuid true`,
					input: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=nosuid"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						ACL: util.Pointer(TriBoolNone),
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(false),
							LazyTime: util.Pointer(false),
							NoATime:  util.Pointer(false),
							NoSuid:   util.Pointer(true)},
						Storage: util.Pointer("local-zfs"),
						rawDisk: "local-zfs:subvol-101-disk-0"})})},
				{name: `Quota false`,
					input: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Quota:   util.Pointer(false),
						Storage: util.Pointer("local-zfs"),
						rawDisk: "local-zfs:subvol-101-disk-0"})})},
				{name: `Quota true`,
					input: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,quota=1"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Quota:   util.Pointer(true),
						Storage: util.Pointer("local-zfs"),
						rawDisk: "local-zfs:subvol-101-disk-0"})})},
				{name: `Replicate false`,
					input: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,replicate=0"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Replicate: util.Pointer(false),
						Storage:   util.Pointer("local-zfs"),
						rawDisk:   "local-zfs:subvol-101-disk-0"})})},
				{name: `Replicate true`,
					input: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,replicate=1"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Replicate: util.Pointer(true),
						Storage:   util.Pointer("local-zfs"),
						rawDisk:   "local-zfs:subvol-101-disk-0"})})},
				{name: `SizeInKibibytes`,
					input: map[string]any{"rootfs": "local-ext4:subvol-101-disk-0,size=999M"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Storage:         util.Pointer("local-ext4"),
						SizeInKibibytes: util.Pointer(LxcMountSize(1022976)),
						rawDisk:         "local-ext4:subvol-101-disk-0"})})},
				{name: `all`,
					input: map[string]any{"rootfs": "local-ext4:subvol-101-disk-0,acl=1,mountoptions=discard;lazytime;noatime;nosuid,size=1G,quota=1,replicate=1"},
					output: baseConfig(ConfigLXC{BootMount: &LxcBootMount{
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
						rawDisk:         "local-ext4:subvol-101-disk-0"}})}}},
		{category: `CPU`,
			tests: []test{
				{name: `Cores`,
					input:  map[string]any{"cores": float64(1)},
					output: baseConfig(ConfigLXC{CPU: &LxcCPU{Cores: util.Pointer(LxcCpuCores(1))}})},
				{name: `Limit`,
					input:  map[string]any{"cpulimit": string("2")},
					output: baseConfig(ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(2))}})},
				{name: `Units`,
					input:  map[string]any{"cpuunits": float64(3)},
					output: baseConfig(ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(3))}})}}},
		{category: `Description`,
			tests: []test{
				{name: `test`,
					input:  map[string]any{"description": "test"},
					output: baseConfig(ConfigLXC{Description: util.Pointer("test")})},
				{name: `""`,
					input:  map[string]any{"description": ""},
					output: baseConfig(ConfigLXC{Description: util.Pointer("")})}}},
		{category: `Digest`,
			tests: []test{
				{name: `set`,
					input: map[string]any{"digest": "af064923bbf2301596aac4c273ba32178ebc4a96"},
					output: baseConfig(ConfigLXC{
						Digest: [sha1.Size]byte{
							0xaf, 0x06, 0x49, 0x23, 0xbb, 0xf2, 0x30, 0x15, 0x96, 0xaa,
							0xc4, 0xc2, 0x73, 0xba, 0x32, 0x17, 0x8e, 0xbc, 0x4a, 0x96},
						rawDigest: "af064923bbf2301596aac4c273ba32178ebc4a96"})}}},
		{category: `DNS`,
			tests: []test{
				{name: `all`,
					input: map[string]any{
						"nameserver":   "1.1.1.1 8.8.8.8 9.9.9.9",
						"searchdomain": "example.com"},
					output: baseConfig(ConfigLXC{DNS: &GuestDNS{
						NameServers: util.Pointer([]netip.Addr{
							parseIP("1.1.1.1"),
							parseIP("8.8.8.8"),
							parseIP("9.9.9.9")}),
						SearchDomain: util.Pointer("example.com")}})},
				{name: `NameServers`,
					input: map[string]any{"nameserver": "8.8.8.8"},
					output: baseConfig(ConfigLXC{DNS: &GuestDNS{
						NameServers:  util.Pointer([]netip.Addr{parseIP("8.8.8.8")}),
						SearchDomain: util.Pointer("")}})},
				{name: `SearchDomain`,
					input: map[string]any{"searchdomain": "example.com"},
					output: baseConfig(ConfigLXC{DNS: &GuestDNS{
						NameServers:  util.Pointer([]netip.Addr(nil)),
						SearchDomain: util.Pointer("example.com")}})}}},
		{category: `Features`,
			tests: []test{
				{name: `CreateDeviceNodes Privileged`,
					input: map[string]any{"features": string("mknod=1")},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						CreateDeviceNodes: util.Pointer(true),
						FUSE:              util.Pointer(false),
						NFS:               util.Pointer(false),
						Nesting:           util.Pointer(false),
						SMB:               util.Pointer(false)}}})},
				{name: `CreateDeviceNodes Unprivileged`,
					input: map[string]any{
						"features":     string("mknod=1"),
						"unprivileged": float64(1)},
					output: baseConfig(ConfigLXC{
						Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
							CreateDeviceNodes: util.Pointer(true),
							FUSE:              util.Pointer(false),
							KeyCtl:            util.Pointer(false),
							Nesting:           util.Pointer(false)}},
						Privileged: util.Pointer(false)})},
				{name: `FUSE Privileged`,
					input: map[string]any{"features": string("fuse=1")},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						CreateDeviceNodes: util.Pointer(false),
						FUSE:              util.Pointer(true),
						NFS:               util.Pointer(false),
						Nesting:           util.Pointer(false),
						SMB:               util.Pointer(false)}}})},
				{name: `FUSE Unprivileged`,
					input: map[string]any{
						"features":     string("fuse=1"),
						"unprivileged": float64(1)},
					output: baseConfig(ConfigLXC{
						Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
							CreateDeviceNodes: util.Pointer(false),
							FUSE:              util.Pointer(true),
							KeyCtl:            util.Pointer(false),
							Nesting:           util.Pointer(false)}},
						Privileged: util.Pointer(false)})},
				{name: `KeyCtl Unprivileged`,
					input: map[string]any{
						"features":     string("keyctl=1"),
						"unprivileged": float64(1)},
					output: baseConfig(ConfigLXC{
						Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
							CreateDeviceNodes: util.Pointer(false),
							FUSE:              util.Pointer(false),
							KeyCtl:            util.Pointer(true),
							Nesting:           util.Pointer(false)}},
						Privileged: util.Pointer(false)})},
				{name: `NFS Privileged`,
					input: map[string]any{"features": string("mount=nfs")},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						CreateDeviceNodes: util.Pointer(false),
						FUSE:              util.Pointer(false),
						NFS:               util.Pointer(true),
						Nesting:           util.Pointer(false),
						SMB:               util.Pointer(false)}}})},
				{name: `NFS and SMB Privileged`,
					input: map[string]any{
						"features":     string("mount=nfs;cifs"),
						"unprivileged": float64(0)},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						CreateDeviceNodes: util.Pointer(false),
						FUSE:              util.Pointer(false),
						NFS:               util.Pointer(true),
						Nesting:           util.Pointer(false),
						SMB:               util.Pointer(true)}}})},
				{name: `Nesting Privileged`,
					input: map[string]any{"features": string("nesting=1")},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						CreateDeviceNodes: util.Pointer(false),
						FUSE:              util.Pointer(false),
						NFS:               util.Pointer(false),
						Nesting:           util.Pointer(true),
						SMB:               util.Pointer(false)}}})},
				{name: `Nesting Unprivileged`,
					input: map[string]any{
						"features":     string("nesting=1"),
						"unprivileged": float64(1)},
					output: baseConfig(ConfigLXC{
						Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
							CreateDeviceNodes: util.Pointer(false),
							FUSE:              util.Pointer(false),
							KeyCtl:            util.Pointer(false),
							Nesting:           util.Pointer(true)}},
						Privileged: util.Pointer(false)})},
				{name: `SMB Privileged`,
					input: map[string]any{"features": string("mount=cifs")},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						CreateDeviceNodes: util.Pointer(false),
						FUSE:              util.Pointer(false),
						NFS:               util.Pointer(false),
						Nesting:           util.Pointer(false),
						SMB:               util.Pointer(true)}}})},
				{name: `SMB and NFS Privileged`,
					input: map[string]any{"features": string("mount=cifs;nfs")},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
						CreateDeviceNodes: util.Pointer(false),
						FUSE:              util.Pointer(false),
						NFS:               util.Pointer(true),
						Nesting:           util.Pointer(false),
						SMB:               util.Pointer(true)}}})}}},
		{category: `ID`,
			tests: []test{
				{name: `set`,
					vmr:    VmRef{vmId: 15},
					output: baseConfig(ConfigLXC{ID: util.Pointer(GuestID(15))})}}},
		{category: `Memory`,
			tests: []test{
				{name: `set`,
					input:  map[string]any{"memory": float64(512)},
					output: baseConfig(ConfigLXC{Memory: util.Pointer(LxcMemory(512))})}}},
		{category: `Name`,
			tests: []test{
				{name: `set`,
					input:  map[string]any{"hostname": "test"},
					output: baseConfig(ConfigLXC{Name: util.Pointer(GuestName("test"))})}}},
		{category: `Mounts`,
			tests: []test{
				{name: `BindMount minimal`,
					input: map[string]any{"mp0": "/host/path,mp=/guest/path"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID0: LxcMount{BindMount: baseBindMount(LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/host/path")),
							GuestPath: util.Pointer(LxcMountPath("/guest/path"))})}}})},
				{name: `BindMount.ReadOnly false`,
					input: map[string]any{"mp1": "/host/path,mp=/guest/path"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID1: LxcMount{BindMount: baseBindMount(LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/host/path")),
							GuestPath: util.Pointer(LxcMountPath("/guest/path")),
							ReadOnly:  util.Pointer(false)})}}})},
				{name: `BindMount.ReadOnly true`,
					input: map[string]any{"mp2": "/host/path,mp=/guest/path,ro=1"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID2: LxcMount{BindMount: baseBindMount(LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/host/path")),
							GuestPath: util.Pointer(LxcMountPath("/guest/path")),
							ReadOnly:  util.Pointer(true)})}}})},
				{name: `BindMount.Replicate false`,
					input: map[string]any{"mp3": "/host/path,mp=/guest/path,replicate=0"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID3: LxcMount{BindMount: baseBindMount(LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/host/path")),
							GuestPath: util.Pointer(LxcMountPath("/guest/path")),
							Replicate: util.Pointer(false)})}}})},
				{name: `BindMount.Replicate true`,
					input: map[string]any{"mp4": "/host/path,mp=/guest/path"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID4: LxcMount{BindMount: baseBindMount(LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/host/path")),
							GuestPath: util.Pointer(LxcMountPath("/guest/path")),
							Replicate: util.Pointer(true)})}}})},
				{name: `BindMount.Options.Discard true`,
					input: map[string]any{"mp5": "/host/path,mp=/guest/path,mountoptions=discard"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID5: LxcMount{BindMount: baseBindMount(LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/host/path")),
							GuestPath: util.Pointer(LxcMountPath("/guest/path")),
							Options: baseMountOptions(LxcMountOptions{
								Discard: util.Pointer(true)})})}}})},
				{name: `BindMount.Options.LazyTime true`,
					input: map[string]any{"mp6": "/host/path,mp=/guest/path,mountoptions=lazytime"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID6: LxcMount{BindMount: baseBindMount(LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/host/path")),
							GuestPath: util.Pointer(LxcMountPath("/guest/path")),
							Options: baseMountOptions(LxcMountOptions{
								LazyTime: util.Pointer(true)})})}}})},
				{name: `BindMount.Options.NoATime true`,
					input: map[string]any{"mp7": "/host/path,mp=/guest/path,mountoptions=noatime"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID7: LxcMount{BindMount: baseBindMount(LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/host/path")),
							GuestPath: util.Pointer(LxcMountPath("/guest/path")),
							Options: baseMountOptions(LxcMountOptions{
								NoATime: util.Pointer(true)})})}}})},
				{name: `BindMount.Options.NoDevice true`,
					input: map[string]any{"mp8": "/host/path,mp=/guest/path,mountoptions=nodev"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID8: LxcMount{BindMount: baseBindMount(LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/host/path")),
							GuestPath: util.Pointer(LxcMountPath("/guest/path")),
							Options: baseMountOptions(LxcMountOptions{
								NoDevice: util.Pointer(true)})})}}})},
				{name: `BindMount.Options.NoExec true`,
					input: map[string]any{"mp9": "/host/path,mp=/guest/path,mountoptions=noexec"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID9: LxcMount{BindMount: baseBindMount(LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/host/path")),
							GuestPath: util.Pointer(LxcMountPath("/guest/path")),
							Options: baseMountOptions(LxcMountOptions{
								NoExec: util.Pointer(true)})})}}})},
				{name: `BindMount.Options.NoSuid true`,
					input: map[string]any{"mp10": "/host/path,mp=/guest/path,mountoptions=nosuid"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID10: LxcMount{BindMount: baseBindMount(LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/host/path")),
							GuestPath: util.Pointer(LxcMountPath("/guest/path")),
							Options: baseMountOptions(LxcMountOptions{
								NoSuid: util.Pointer(true)})})}}})},
				{name: `BindMount.Options all true`,
					input: map[string]any{"mp11": "/host/path,mp=/guest/path,mountoptions=discard;lazytime;noatime;nodev;noexec;nosuid"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID11: LxcMount{BindMount: baseBindMount(LxcBindMount{
							HostPath:  util.Pointer(LxcHostPath("/host/path")),
							GuestPath: util.Pointer(LxcMountPath("/guest/path")),
							Options: &LxcMountOptions{
								Discard:  util.Pointer(true),
								LazyTime: util.Pointer(true),
								NoATime:  util.Pointer(true),
								NoDevice: util.Pointer(true),
								NoExec:   util.Pointer(true),
								NoSuid:   util.Pointer(true)}})}}})},
				{name: `DataMount minimal ext4 (privileged`,
					input: map[string]any{"mp100": "local-ext4:100/vm-100-disk-0.raw"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID100: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota:   util.Pointer(false),
							Storage: util.Pointer("local-ext4"),
							rawDisk: "local-ext4:100/vm-100-disk-0.raw"})}}})},
				{name: `DataMount minimal zfs (privileged`,
					input: map[string]any{"mp101": "local-zfs:subvol-100-disk-1"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID101: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota:   util.Pointer(false),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.ACL false (privileged`,
					input: map[string]any{"mp102": "local-zfs:subvol-100-disk-1,acl=0"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID102: LxcMount{DataMount: baseDataMount(LxcDataMount{
							ACL:     util.Pointer(TriBoolFalse),
							Quota:   util.Pointer(false),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.ACL true (privileged)`,
					input: map[string]any{"mp103": "local-zfs:subvol-100-disk-1,acl=1"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID103: LxcMount{DataMount: baseDataMount(LxcDataMount{
							ACL:     util.Pointer(TriBoolTrue),
							Quota:   util.Pointer(false),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Backup false (privileged)`,
					input: map[string]any{"mp104": "local-zfs:subvol-100-disk-1,backup=0"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID104: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Backup:  util.Pointer(false),
							Quota:   util.Pointer(false),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Backup true (privileged)`,
					input: map[string]any{"mp105": "local-zfs:subvol-100-disk-1,backup=1"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID105: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Backup:  util.Pointer(true),
							Quota:   util.Pointer(false),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Options.Discard true`,
					input: map[string]any{"mp201": "local-zfs:subvol-100-disk-1,mountoptions=discard"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID201: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: baseMountOptions(LxcMountOptions{
								Discard: util.Pointer(true)}),
							Quota:   util.Pointer(false),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Options.LazyTime true`,
					input: map[string]any{"mp203": "local-zfs:subvol-100-disk-1,mountoptions=lazytime"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID203: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: baseMountOptions(LxcMountOptions{
								LazyTime: util.Pointer(true)}),
							Quota:   util.Pointer(false),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Options.NoATime true`,
					input: map[string]any{"mp204": "local-zfs:subvol-100-disk-1,mountoptions=noatime"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID204: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: baseMountOptions(LxcMountOptions{
								NoATime: util.Pointer(true)}),
							Quota:   util.Pointer(false),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Options.NoDevice true`,
					input: map[string]any{"mp205": "local-zfs:subvol-100-disk-1,mountoptions=nodev"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID205: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: baseMountOptions(LxcMountOptions{
								NoDevice: util.Pointer(true)}),
							Quota:   util.Pointer(false),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Options.NoExec true`,
					input: map[string]any{"mp206": "local-zfs:subvol-100-disk-1,mountoptions=noexec"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID206: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: baseMountOptions(LxcMountOptions{
								NoExec: util.Pointer(true)}),
							Quota:   util.Pointer(false),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Options.NoSuid true`,
					input: map[string]any{"mp207": "local-zfs:subvol-100-disk-1,mountoptions=nosuid"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID207: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: baseMountOptions(LxcMountOptions{
								NoSuid: util.Pointer(true)}),
							Quota:   util.Pointer(false),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Options all true`,
					input: map[string]any{"mp208": "local-zfs:subvol-100-disk-1,mountoptions=noexec;nosuid;lazytime;discard;noatime;nodev"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID208: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Options: &LxcMountOptions{
								Discard:  util.Pointer(true),
								LazyTime: util.Pointer(true),
								NoATime:  util.Pointer(true),
								NoDevice: util.Pointer(true),
								NoExec:   util.Pointer(true),
								NoSuid:   util.Pointer(true)},
							Quota:   util.Pointer(false),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Path (privileged)`,
					input: map[string]any{"mp106": "local-zfs:subvol-100-disk-1,mp=/mnt/test"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID106: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota:   util.Pointer(false),
							Path:    util.Pointer(LxcMountPath("/mnt/test")),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Quota false Privilege false`,
					input: map[string]any{
						"mp107":        "local-zfs:subvol-100-disk-1,quota=0",
						"unprivileged": float64(1)},
					output: baseConfig(ConfigLXC{
						Privileged: util.Pointer(false),
						Mounts: LxcMounts{
							LxcMountID107: LxcMount{DataMount: baseDataMount(LxcDataMount{
								Storage: util.Pointer("local-zfs"),
								rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Quota false Privilege true`,
					input: map[string]any{
						"mp108":        "local-zfs:subvol-100-disk-1,quota=0",
						"unprivileged": float64(0)},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID108: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota:   util.Pointer(false),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Quota true Privilege false`,
					input: map[string]any{
						"mp109":        "local-zfs:subvol-100-disk-1,quota=1",
						"unprivileged": float64(1)},
					output: baseConfig(ConfigLXC{
						Privileged: util.Pointer(false),
						Mounts: LxcMounts{
							LxcMountID109: LxcMount{DataMount: baseDataMount(LxcDataMount{
								Storage: util.Pointer("local-zfs"),
								rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Quota true Privilege true`,
					input: map[string]any{
						"mp110":        "local-zfs:subvol-100-disk-1,quota=1",
						"unprivileged": float64(0)},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID110: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota:   util.Pointer(true),
							Storage: util.Pointer("local-zfs"),
							rawDisk: "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.ReadOnly false (privileged)`,
					input: map[string]any{"mp111": "local-zfs:subvol-100-disk-1,ro=0"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID111: LxcMount{DataMount: baseDataMount(LxcDataMount{
							ReadOnly: util.Pointer(false),
							Quota:    util.Pointer(false),
							Storage:  util.Pointer("local-zfs"),
							rawDisk:  "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.ReadOnly true (privileged)`,
					input: map[string]any{"mp112": "local-zfs:subvol-100-disk-1,ro=1"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID112: LxcMount{DataMount: baseDataMount(LxcDataMount{
							ReadOnly: util.Pointer(true),
							Quota:    util.Pointer(false),
							Storage:  util.Pointer("local-zfs"),
							rawDisk:  "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Replicate false (privileged)`,
					input: map[string]any{"mp113": "local-zfs:subvol-100-disk-1,replicate=0"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID113: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Replicate: util.Pointer(false),
							Quota:     util.Pointer(false),
							Storage:   util.Pointer("local-zfs"),
							rawDisk:   "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.Replicate true (privileged)`,
					input: map[string]any{"mp114": "local-zfs:subvol-100-disk-1,replicate=1"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID114: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Replicate: util.Pointer(true),
							Quota:     util.Pointer(false),
							Storage:   util.Pointer("local-zfs"),
							rawDisk:   "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.SizeInKibibytes 1T (privileged)`,
					input: map[string]any{"mp115": "local-zfs:subvol-100-disk-1,size=1T"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID115: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota:           util.Pointer(false),
							SizeInKibibytes: util.Pointer(LxcMountSize(1073741824)),
							Storage:         util.Pointer("local-zfs"),
							rawDisk:         "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.SizeInKibibytes 1G (privileged)`,
					input: map[string]any{"mp116": "local-zfs:subvol-100-disk-1,size=1G"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID116: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota:           util.Pointer(false),
							SizeInKibibytes: util.Pointer(LxcMountSize(1048576)),
							Storage:         util.Pointer("local-zfs"),
							rawDisk:         "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.SizeInKibibytes 12M (privileged)`,
					input: map[string]any{"mp117": "local-zfs:subvol-100-disk-1,size=12M"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID117: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota:           util.Pointer(false),
							SizeInKibibytes: util.Pointer(LxcMountSize(12288)),
							Storage:         util.Pointer("local-zfs"),
							rawDisk:         "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount.SizeInKibibytes 18K  (privileged)`,
					input: map[string]any{"mp118": "local-zfs:subvol-100-disk-1,size=18K"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID118: LxcMount{DataMount: baseDataMount(LxcDataMount{
							Quota:           util.Pointer(false),
							SizeInKibibytes: util.Pointer(LxcMountSize(18)),
							Storage:         util.Pointer("local-zfs"),
							rawDisk:         "local-zfs:subvol-100-disk-1"})}}})},
				{name: `DataMount all (privileged)`,
					input: map[string]any{"mp150": "local-zfs:subvol-100-disk-1,size=18K,acl=0,backup=1,quota=1,mountoptions=lazytime;noexec;discard,mp=/opt/test,replicate=1,ro=1"},
					output: baseConfig(ConfigLXC{Mounts: LxcMounts{
						LxcMountID150: LxcMount{DataMount: baseDataMount(LxcDataMount{
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
							rawDisk:         "local-zfs:subvol-100-disk-1"})}}})}}},
		{category: `Networks`,
			tests: []test{
				{name: `all`,
					input: map[string]any{"net0": "name=eth0,bridge=vmbr0,ip=192.168.0.23/24,gw=12.168.0.1,rate=0.810,trunks=101,hwaddr=00:A1:22:b3:44:55,tag=100,link_down=1,firewall=1,ip6=2001:db8::1/64,gw6=2001:db8::2,mtu=896"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID0: baseNetwork(LxcNetwork{
							Bridge:    util.Pointer("vmbr0"),
							Connected: util.Pointer(false),
							Firewall:  util.Pointer(true),
							IPv4: &LxcIPv4{
								Address: util.Pointer(IPv4CIDR("192.168.0.23/24")),
								Gateway: util.Pointer(IPv4Address("12.168.0.1"))},
							IPv6: &LxcIPv6{
								Address: util.Pointer(IPv6CIDR("2001:db8::1/64")),
								Gateway: util.Pointer(IPv6Address("2001:db8::2"))},
							MAC:           util.Pointer(parseMAC("00:a1:22:B3:44:55")),
							Mtu:           util.Pointer(MTU(896)),
							Name:          util.Pointer(LxcNetworkName("eth0")),
							NativeVlan:    util.Pointer(Vlan(100)),
							RateLimitKBps: util.Pointer(GuestNetworkRate(810)),
							TaggedVlans:   util.Pointer(Vlans{Vlan(101)}),
							mac:           "00:A1:22:b3:44:55"})}})},
				{name: `Bridge`,
					input: map[string]any{"net0": "bridge=vmbr0"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID0: baseNetwork(LxcNetwork{Bridge: util.Pointer("vmbr0")})}})},
				{name: `Connected`,
					input: map[string]any{"net1": "link_down=1"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID1: baseNetwork(LxcNetwork{Connected: util.Pointer(false)})}})},
				{name: `Firewall`,
					input: map[string]any{"net2": "firewall=1"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID2: baseNetwork(LxcNetwork{Firewall: util.Pointer(true)})}})},
				{name: `IPv4 Address`,
					input: map[string]any{"net3": "ip=192.168.0.10/24"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID3: baseNetwork(LxcNetwork{IPv4: &LxcIPv4{
							Address: util.Pointer(IPv4CIDR("192.168.0.10/24"))}})}})},
				{name: `IPv4 DHCP`,
					input: map[string]any{"net4": "ip=dhcp"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID4: baseNetwork(LxcNetwork{IPv4: &LxcIPv4{
							DHCP: true}})}})},
				{name: `IPv4 Gateway`,
					input: map[string]any{"net5": "gw=1.1.1.1"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID5: baseNetwork(LxcNetwork{IPv4: &LxcIPv4{
							Gateway: util.Pointer(IPv4Address("1.1.1.1"))}})}})},
				{name: `IPv4 Manual`,
					input: map[string]any{"net6": "ip=manual"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID6: baseNetwork(LxcNetwork{IPv4: &LxcIPv4{
							Manual: true}})}})},
				{name: `IPv6 Address`,
					input: map[string]any{"net7": "ip6=2001:db8::1/64"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID7: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
							Address: util.Pointer(IPv6CIDR("2001:db8::1/64"))}})}})},
				{name: `IPv6 DHCP`,
					input: map[string]any{"net8": "ip6=dhcp"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID8: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
							DHCP: true}})}})},
				{name: `IPv6 Gateway`,
					input: map[string]any{"net9": "gw6=2001:db8::2"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID9: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
							Gateway: util.Pointer(IPv6Address("2001:db8::2"))}})}})},
				{name: `IPv6 Manual`,
					input: map[string]any{"net10": "ip6=manual"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID10: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
							Manual: true}})}})},
				{name: `IPv6 SLAAC`,
					input: map[string]any{"net11": "ip6=auto"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID11: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
							SLAAC: true}})}})},
				{name: `MAC`,
					input: map[string]any{"net12": "hwaddr=00:A1:22:b3:44:55"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID12: baseNetwork(LxcNetwork{
							MAC: util.Pointer(parseMAC("00:a1:22:B3:44:55")),
							mac: "00:A1:22:b3:44:55"})}})},
				{name: `Mtu`,
					input: map[string]any{"net13": "mtu=1321"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID13: baseNetwork(LxcNetwork{Mtu: util.Pointer(MTU(1321))})}})},
				{name: `Name`,
					input: map[string]any{"net13": "name=eth0"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID13: baseNetwork(LxcNetwork{Name: util.Pointer(LxcNetworkName("eth0"))})}})},
				{name: `NativeVlan`,
					input: map[string]any{"net14": "tag=100"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID14: baseNetwork(LxcNetwork{NativeVlan: util.Pointer(Vlan(100))})}})},
				{name: `RateLimitKBps`,
					input: map[string]any{"net15": "rate=95.649"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID15: baseNetwork(LxcNetwork{RateLimitKBps: util.Pointer(GuestNetworkRate(95649))})}})},
				{name: `TaggedVlans`,
					input: map[string]any{"net0": "trunks=200;100;300"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID0: baseNetwork(LxcNetwork{TaggedVlans: &Vlans{Vlan(100), Vlan(200), Vlan(300)}})}})}}},
		{category: `Node`,
			tests: []test{
				{name: `set`,
					vmr:    VmRef{node: "test"},
					output: baseConfig(ConfigLXC{Node: util.Pointer(NodeName("test"))})}}},
		{category: `OperatingSystem`,
			tests: []test{
				{name: `set`,
					input:  map[string]any{"ostype": "test"},
					output: baseConfig(ConfigLXC{OperatingSystem: "test"})}}},
		{category: `Pool`,
			tests: []test{
				{name: `set`,
					vmr:    VmRef{pool: "test"},
					output: baseConfig(ConfigLXC{Pool: util.Pointer(PoolName("test"))})}}},
		{category: `Privileged`,
			tests: []test{
				{name: `true`,
					input:  map[string]any{"unprivileged": float64(0)},
					output: baseConfig(ConfigLXC{Privileged: util.Pointer(true)})},
				{name: `false`,
					input:  map[string]any{"unprivileged": float64(1)},
					output: baseConfig(ConfigLXC{Privileged: util.Pointer(false)})},
				{name: `default true`,
					input:  map[string]any{},
					output: baseConfig(ConfigLXC{Privileged: util.Pointer(true)})}}},
		{category: `Protection`,
			tests: []test{
				{name: `false`,
					input:  map[string]any{},
					output: baseConfig(ConfigLXC{Protection: util.Pointer(false)})},
				{name: `true`,
					input:  map[string]any{"protection": float64(1)},
					output: baseConfig(ConfigLXC{Protection: util.Pointer(true)})}}},
		{category: `Swap`,
			tests: []test{
				{name: `set`,
					input:  map[string]any{"swap": float64(256)},
					output: baseConfig(ConfigLXC{Swap: util.Pointer(LxcSwap(256))})},
				{name: `set 0`,
					input:  map[string]any{"swap": float64(0)},
					output: baseConfig(ConfigLXC{Swap: util.Pointer(LxcSwap(0))})}}},
		{category: `State`,
			tests: []test{
				{name: `set`,
					state:  PowerStateRunning,
					output: baseConfig(ConfigLXC{State: util.Pointer(PowerStateRunning)})}}},
		{category: `Tags`,
			tests: []test{
				{name: `set`,
					input:  map[string]any{"tags": "test"},
					output: baseConfig(ConfigLXC{Tags: &Tags{"test"}})}}},
	}
	for _, test := range tests {
		for _, subTest := range test.tests {
			name := test.category
			if len(test.tests) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				raw, err := guestGetLxcRawConfig_Unsafe(context.Background(), &subTest.vmr, &mockClientAPI{
					getGuestConfigFunc: func(ctx context.Context, vmr *VmRef) (map[string]any, error) {
						return subTest.input, subTest.err
					}})
				require.Equal(t, subTest.err, err, name)
				if subTest.output != nil {
					require.Equal(t, subTest.output, raw.Get(subTest.vmr, subTest.state), name)
				}
			})
		}
	}
}

func Test_RawConfigLXC_GetDigest(t *testing.T) {
	set := func(raw map[string]any) *rawConfigLXC { return &rawConfigLXC{a: raw} }
	require.Equal(t,
		[sha1.Size]byte{
			0xaf, 0x06, 0x49, 0x23, 0xbb, 0xf2, 0x30, 0x15, 0x96, 0xaa,
			0xc4, 0xc2, 0x73, 0xba, 0x32, 0x17, 0x8e, 0xbc, 0x4a, 0x96},
		set(map[string]any{"digest": "af064923bbf2301596aac4c273ba32178ebc4a96"}).GetDigest(), "")
}

func Test_RawConfigLXC_GetPrivileged(t *testing.T) {
	set := func(raw map[string]any) *rawConfigLXC { return &rawConfigLXC{a: raw} }
	require.Equal(t, true, set(map[string]any{}).GetPrivileged())
}

func Test_LxcMemory_String(t *testing.T) {
	require.Equal(t, "583421", LxcMemory(583421).String())
}

func Test_LxcSwap_String(t *testing.T) {
	require.Equal(t, "8423", LxcSwap(8423).String())
}
