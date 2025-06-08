package proxmox

import (
	"errors"
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
	feature := func(value bool) *LxcFeatures {
		return &LxcFeatures{
			CreateDeviceNodes: util.Pointer(value),
			FUSE:              util.Pointer(value),
			KeyCtl:            util.Pointer(value),
			NFS:               util.Pointer(value),
			Nesting:           util.Pointer(value),
			SMB:               util.Pointer(value)}
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
				{name: `Quota false`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						Quota: util.Pointer(false)}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						Quota: util.Pointer(false)}},
					output: map[string]any{"rootfs": ""}},
				{name: `Quota true`,
					config: ConfigLXC{BootMount: &LxcBootMount{
						Quota: util.Pointer(true)}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						Quota: util.Pointer(false)}},
					output: map[string]any{"rootfs": ",quota=1"}},
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
				{name: `all storage change`,
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
						rawDisk:   "subvol-101-disk-0"}},
					output: map[string]any{"rootfs": "local-zfs:subvol-101-disk-0,size=1G,acl=1,mountoptions=discard;lazytime;noatime;nosuid,quota=1,replicate=0"}},
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
						SizeInKibibytes: util.Pointer(LxcMountSize(2621440)),
						Storage:         util.Pointer("local-ext4")}},
					currentConfig: ConfigLXC{BootMount: &LxcBootMount{
						SizeInKibibytes: util.Pointer(LxcMountSize(2097152)),
						Storage:         util.Pointer("local-zfs"),
						rawDisk:         "subvol-101-disk-0"}},
					output: map[string]any{"rootfs": "local-ext4:subvol-101-disk-0,size=2560M"}},
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
					output: map[string]any{}}}},
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
					output:        map[string]any{}},
				{name: `Limit delete no effect`,
					config:        ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(0))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{}},
					output:        map[string]any{}},
				{name: `Units delete no effect`,
					config:        ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(0))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{}},
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
					output:        map[string]any{}},
				{name: `Limit same`,
					config:        ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(2))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(2))}},
					output:        map[string]any{}},
				{name: `Units same`,
					config:        ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(3))}},
					currentConfig: ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(3))}},
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
					output:        map[string]any{}},
				{name: `Limit delete no current`,
					config:        ConfigLXC{CPU: &LxcCPU{Limit: util.Pointer(LxcCpuLimit(0))}},
					currentConfig: ConfigLXC{},
					output:        map[string]any{}},
				{name: `Units delete no current`,
					config:        ConfigLXC{CPU: &LxcCPU{Units: util.Pointer(LxcCpuUnits(0))}},
					currentConfig: ConfigLXC{},
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
					output: map[string]any{}}}},
		{category: `Features`,
			createUpdate: []test{
				{name: `CreateDeviceNodes`,
					config:        ConfigLXC{Features: &LxcFeatures{CreateDeviceNodes: util.Pointer(true)}},
					currentConfig: ConfigLXC{Features: feature(false)},
					output:        map[string]any{"features": "mknod=1"}},
				{name: `FUSE`,
					config:        ConfigLXC{Features: &LxcFeatures{FUSE: util.Pointer(true)}},
					currentConfig: ConfigLXC{Features: feature(false)},
					output:        map[string]any{"features": "fuse=1"}},
				{name: `KeyCtl`,
					config:        ConfigLXC{Features: &LxcFeatures{KeyCtl: util.Pointer(true)}},
					currentConfig: ConfigLXC{Features: feature(false)},
					output:        map[string]any{"features": "keyctl=1"}},
				{name: `NFS`,
					config:        ConfigLXC{Features: &LxcFeatures{NFS: util.Pointer(true)}},
					currentConfig: ConfigLXC{Features: feature(false)},
					output:        map[string]any{"features": "mount=nfs"}},
				{name: `smb`,
					config:        ConfigLXC{Features: &LxcFeatures{SMB: util.Pointer(true)}},
					currentConfig: ConfigLXC{Features: feature(false)},
					output:        map[string]any{"features": "mount=cifs"}},
				{name: `Nesting`,
					config:        ConfigLXC{Features: &LxcFeatures{Nesting: util.Pointer(true)}},
					currentConfig: ConfigLXC{Features: feature(false)},
					output:        map[string]any{"features": "nesting=1"}},
				{name: `NFS and SMB`,
					config:        ConfigLXC{Features: &LxcFeatures{NFS: util.Pointer(true), SMB: util.Pointer(true)}},
					currentConfig: ConfigLXC{Features: feature(false)},
					output:        map[string]any{"features": "mount=nfs;cifs"}},
				{name: `delete no effect false`,
					config:        ConfigLXC{Features: feature(false)},
					currentConfig: ConfigLXC{Features: feature(false)},
					output:        map[string]any{}}},
			update: []test{
				{name: `CreateDeviceNodes false`,
					config:        ConfigLXC{Features: &LxcFeatures{CreateDeviceNodes: util.Pointer(false)}},
					currentConfig: ConfigLXC{Features: feature(true)},
					output:        map[string]any{"features": "fuse=1,keyctl=1,mount=nfs;cifs,nesting=1"}},
				{name: `FUSE false`,
					config:        ConfigLXC{Features: &LxcFeatures{FUSE: util.Pointer(false)}},
					currentConfig: ConfigLXC{Features: feature(true)},
					output:        map[string]any{"features": "mknod=1,keyctl=1,mount=nfs;cifs,nesting=1"}},
				{name: `KeyCtl false`,
					config:        ConfigLXC{Features: &LxcFeatures{KeyCtl: util.Pointer(false)}},
					currentConfig: ConfigLXC{Features: feature(true)},
					output:        map[string]any{"features": "mknod=1,fuse=1,mount=nfs;cifs,nesting=1"}},
				{name: `NFS false`,
					config:        ConfigLXC{Features: &LxcFeatures{NFS: util.Pointer(false)}},
					currentConfig: ConfigLXC{Features: feature(true)},
					output:        map[string]any{"features": "mknod=1,fuse=1,keyctl=1,mount=cifs,nesting=1"}},
				{name: `SMB false`,
					config:        ConfigLXC{Features: &LxcFeatures{SMB: util.Pointer(false)}},
					currentConfig: ConfigLXC{Features: feature(true)},
					output:        map[string]any{"features": "mknod=1,fuse=1,keyctl=1,mount=nfs,nesting=1"}},
				{name: `Nesting false`,
					config:        ConfigLXC{Features: &LxcFeatures{Nesting: util.Pointer(false)}},
					currentConfig: ConfigLXC{Features: feature(true)},
					output:        map[string]any{"features": "mknod=1,fuse=1,keyctl=1,mount=nfs;cifs"}},
				{name: `delete`,
					config: ConfigLXC{Features: feature(false)},
					currentConfig: ConfigLXC{Features: &LxcFeatures{
						CreateDeviceNodes: util.Pointer(true),
						FUSE:              util.Pointer(false),
						KeyCtl:            util.Pointer(true),
						NFS:               util.Pointer(false),
						Nesting:           util.Pointer(true),
						SMB:               util.Pointer(true)}},
					output: map[string]any{"delete": "features"}},
				{name: `delete no effect nil`,
					config:        ConfigLXC{Features: feature(false)},
					currentConfig: ConfigLXC{Features: nil},
					output:        map[string]any{}},
				{name: `same true`,
					config:        ConfigLXC{Features: feature(true)},
					currentConfig: ConfigLXC{Features: feature(true)},
					output:        map[string]any{}},
				{name: `same true`,
					config:        ConfigLXC{Features: feature(false)},
					currentConfig: ConfigLXC{Features: feature(false)},
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
					output: map[string]any{}},
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
				{name: `IPv4 Address create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID6: LxcNetwork{IPv4: util.Pointer(LxcIPv4{
							Address: util.Pointer(IPv4CIDR("10.0.0.10/24"))})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID6: LxcNetwork{}}},
					output: map[string]any{"net6": ",ip=10.0.0.10/24"}},
				{name: `IPv4 DHCP create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID7: LxcNetwork{IPv4: util.Pointer(LxcIPv4{
							DHCP: true})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID7: LxcNetwork{}}},
					output: map[string]any{"net7": ",ip=dhcp"}},
				{name: `IPv4 Gateway create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID8: LxcNetwork{IPv4: util.Pointer(LxcIPv4{
							Gateway: util.Pointer(IPv4Address("10.0.0.1"))})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID8: LxcNetwork{}}},
					output: map[string]any{"net8": ",gw=10.0.0.1"}},
				{name: `IPv4 Manual create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID9: LxcNetwork{IPv4: util.Pointer(LxcIPv4{
							Manual: true})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID9: LxcNetwork{}}},
					output: map[string]any{"net9": ",ip=manual"}},
				{name: `IPv6 Address create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID10: LxcNetwork{IPv6: util.Pointer(LxcIPv6{
							Address: util.Pointer(IPv6CIDR("2001:db8::1/64"))})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID10: LxcNetwork{}}},
					output: map[string]any{"net10": ",ip6=2001:db8::1/64"}},
				{name: `IPv6 DHCP create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID11: LxcNetwork{IPv6: util.Pointer(LxcIPv6{
							DHCP: true})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID11: LxcNetwork{}}},
					output: map[string]any{"net11": ",ip6=dhcp"}},
				{name: `IPv6 Gateway create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID12: LxcNetwork{IPv6: util.Pointer(LxcIPv6{
							Gateway: util.Pointer(IPv6Address("2001:db8::2"))})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID12: LxcNetwork{}}},
					output: map[string]any{"net12": ",gw6=2001:db8::2"}},
				{name: `IPv6 SLAAC create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID13: LxcNetwork{IPv6: util.Pointer(LxcIPv6{
							SLAAC: true})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID13: LxcNetwork{}}},
					output: map[string]any{"net13": ",ip6=auto"}},
				{name: `IPv6 Manual create`,
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
				{name: `IPv4 inherit Address`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0")),
						IPv4: util.Pointer(LxcIPv4{})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: LxcNetwork{IPv4: util.Pointer(LxcIPv4{Address: util.Pointer(IPv4CIDR("192.168.1.34/24"))})}}},
					output:        map[string]any{"net4": "name=test0,ip=192.168.1.34/24"}},
				{name: `IPv4 inherit DHCP`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0")),
						IPv4: util.Pointer(LxcIPv4{})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: LxcNetwork{IPv4: util.Pointer(LxcIPv4{DHCP: true})}}},
					output:        map[string]any{"net4": "name=test0,ip=dhcp"}},
				{name: `IPv4 inherit Manual`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID8: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0")),
						IPv4: util.Pointer(LxcIPv4{})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID8: LxcNetwork{IPv4: util.Pointer(LxcIPv4{Manual: true})}}},
					output:        map[string]any{"net8": "name=test0,ip=manual"}},
				{name: `IPv4 replace Address`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID3: LxcNetwork{IPv4: util.Pointer(LxcIPv4{Address: util.Pointer(IPv4CIDR("10.0.0.2/24"))})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID3: network()}},
					output:        map[string]any{"net3": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=10.0.0.2/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv4 replace DHCP`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: LxcNetwork{IPv4: util.Pointer(LxcIPv4{DHCP: true})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: network()}},
					output:        map[string]any{"net4": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=dhcp,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv4 replace Gateway`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID5: LxcNetwork{IPv4: util.Pointer(LxcIPv4{Gateway: util.Pointer(IPv4Address("1.1.1.1"))})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID5: network()}},
					output:        map[string]any{"net5": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=1.1.1.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv4 replace Manual`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID6: LxcNetwork{IPv4: util.Pointer(LxcIPv4{Manual: true})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID6: network()}},
					output:        map[string]any{"net6": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=manual,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv6 inherit Address`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0")),
						IPv6: util.Pointer(LxcIPv6{})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: LxcNetwork{IPv6: util.Pointer(LxcIPv6{Address: util.Pointer(IPv6CIDR("2001:db8::2/64"))})}}},
					output:        map[string]any{"net12": "name=test0,ip6=2001:db8::2/64"}},
				{name: `IPv6 inherit DHCP`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0")),
						IPv6: util.Pointer(LxcIPv6{})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: LxcNetwork{IPv6: util.Pointer(LxcIPv6{DHCP: true})}}},
					output:        map[string]any{"net12": "name=test0,ip6=dhcp"}},
				{name: `IPv6 inherit SLAAC`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0")),
						IPv6: util.Pointer(LxcIPv6{})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: LxcNetwork{IPv6: util.Pointer(LxcIPv6{SLAAC: true})}}},
					output:        map[string]any{"net13": "name=test0,ip6=auto"}},
				{name: `IPv6 inherit Manual`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: LxcNetwork{
						Name: util.Pointer(LxcNetworkName("test0")),
						IPv6: util.Pointer(LxcIPv6{})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: LxcNetwork{IPv6: util.Pointer(LxcIPv6{Manual: true})}}},
					output:        map[string]any{"net13": "name=test0,ip6=manual"}},
				{name: `IPv6 replace Address`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID7: LxcNetwork{IPv6: util.Pointer(LxcIPv6{Address: util.Pointer(IPv6CIDR("2001:db8::2/64"))})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID7: network()}},
					output:        map[string]any{"net7": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::2/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv6 replace DHCP`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID8: LxcNetwork{IPv6: util.Pointer(LxcIPv6{DHCP: true})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID8: network()}},
					output:        map[string]any{"net8": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=dhcp,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv6 replace Gateway`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID9: LxcNetwork{IPv6: util.Pointer(LxcIPv6{Gateway: util.Pointer(IPv6Address("2001:db8::3"))})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID9: network()}},
					output:        map[string]any{"net9": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::3,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv6 replace SLAAC`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID10: LxcNetwork{IPv6: util.Pointer(LxcIPv6{SLAAC: true})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID10: network()}},
					output:        map[string]any{"net10": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=auto,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
				{name: `IPv6 replace Manual`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID11: LxcNetwork{IPv6: util.Pointer(LxcIPv6{Manual: true})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID11: network()}},
					output:        map[string]any{"net11": "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=manual,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"}},
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
						current: &ConfigLXC{BootMount: &LxcBootMount{Storage: util.Pointer("text")}}}}},
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
							SizeInKibibytes: util.Pointer(lxcMountSize_Minimum - 1)}}),
						current: &ConfigLXC{BootMount: &LxcBootMount{
							SizeInKibibytes: util.Pointer(LxcMountSize(131071))}},
						err: errors.New(LxcMountSize_Error_Minimum)}}}},
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
								Name: util.Pointer(LxcNetworkName("net"))}}},
					},
				},
			},
			invalid: testType{
				create: []test{
					{name: `errors.New(LxcNetwork_Error_BridgeRequired)`,
						input: baseConfig(ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: LxcNetwork{}}}),
						err:   errors.New(LxcNetwork_Error_BridgeRequired)},
				},
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
				},
			},
		},
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
		if config.Networks == nil {
			config.Networks = make(LxcNetworks)
		}
		if config.Node == nil {
			config.Node = util.Pointer(NodeName(""))
		}
		if config.Privileged == nil {
			config.Privileged = util.Pointer(false)
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
		return config
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
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						ACL:     util.Pointer(TriBoolTrue),
						Storage: util.Pointer("local-zfs"),
						rawDisk: "subvol-101-disk-0"})})},
				{name: `ACL false`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,acl=0"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						ACL:     util.Pointer(TriBoolFalse),
						Storage: util.Pointer("local-zfs"),
						rawDisk: "subvol-101-disk-0"})})},
				{name: `Options Discard true`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=discard"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(true),
							LazyTime: util.Pointer(false),
							NoATime:  util.Pointer(false),
							NoSuid:   util.Pointer(false)},
						Storage: util.Pointer("local-zfs"),
						rawDisk: "subvol-101-disk-0"})})},
				{name: `Options LazyTime true`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=lazytime"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(false),
							LazyTime: util.Pointer(true),
							NoATime:  util.Pointer(false),
							NoSuid:   util.Pointer(false)},
						Storage: util.Pointer("local-zfs"),
						rawDisk: "subvol-101-disk-0"})})},
				{name: `Options NoATime true`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=noatime"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(false),
							LazyTime: util.Pointer(false),
							NoATime:  util.Pointer(true),
							NoSuid:   util.Pointer(false)},
						Storage: util.Pointer("local-zfs"),
						rawDisk: "subvol-101-disk-0"})})},
				{name: `Options NoSuid true`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,mountoptions=nosuid"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						ACL: util.Pointer(TriBoolNone),
						Options: &LxcBootMountOptions{
							Discard:  util.Pointer(false),
							LazyTime: util.Pointer(false),
							NoATime:  util.Pointer(false),
							NoSuid:   util.Pointer(true)},
						Storage: util.Pointer("local-zfs"),
						rawDisk: "subvol-101-disk-0"})})},
				{name: `Quota false`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Quota:   util.Pointer(false),
						Storage: util.Pointer("local-zfs"),
						rawDisk: "subvol-101-disk-0"})})},
				{name: `Quota true`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,quota=1"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Quota:   util.Pointer(true),
						Storage: util.Pointer("local-zfs"),
						rawDisk: "subvol-101-disk-0"})})},
				{name: `Replicate false`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,replicate=0"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Replicate: util.Pointer(false),
						Storage:   util.Pointer("local-zfs"),
						rawDisk:   "subvol-101-disk-0"})})},
				{name: `Replicate true`,
					input: RawConfigLXC{"rootfs": "local-zfs:subvol-101-disk-0,replicate=1"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Replicate: util.Pointer(true),
						Storage:   util.Pointer("local-zfs"),
						rawDisk:   "subvol-101-disk-0"})})},
				{name: `SizeInKibibytes`,
					input: RawConfigLXC{"rootfs": "local-ext4:subvol-101-disk-0,size=999M"},
					output: baseConfig(ConfigLXC{BootMount: baseBootMount(LxcBootMount{
						Storage:         util.Pointer("local-ext4"),
						SizeInKibibytes: util.Pointer(LxcMountSize(1022976)),
						rawDisk:         "subvol-101-disk-0"})})},
				{name: `all`,
					input: RawConfigLXC{"rootfs": "local-ext4:subvol-101-disk-0,acl=1,mountoptions=discard;lazytime;noatime;nosuid,size=1G,quota=1,replicate=1"},
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
						rawDisk:         "subvol-101-disk-0"}})}}},
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
					input:  RawConfigLXC{"description": "test"},
					output: baseConfig(ConfigLXC{Description: util.Pointer("test")})},
				{name: `""`,
					input:  RawConfigLXC{"description": ""},
					output: baseConfig(ConfigLXC{Description: util.Pointer("")})}}},
		{category: `DNS`,
			tests: []test{
				{name: `all`,
					input: RawConfigLXC{
						"nameserver":   "1.1.1.1 8.8.8.8 9.9.9.9",
						"searchdomain": "example.com"},
					output: baseConfig(ConfigLXC{DNS: &GuestDNS{
						NameServers: util.Pointer([]netip.Addr{
							parseIP("1.1.1.1"),
							parseIP("8.8.8.8"),
							parseIP("9.9.9.9")}),
						SearchDomain: util.Pointer("example.com")}})},
				{name: `NameServers`,
					input: RawConfigLXC{"nameserver": "8.8.8.8"},
					output: baseConfig(ConfigLXC{DNS: &GuestDNS{
						NameServers:  util.Pointer([]netip.Addr{parseIP("8.8.8.8")}),
						SearchDomain: util.Pointer("")}})},
				{name: `SearchDomain`,
					input: RawConfigLXC{"searchdomain": "example.com"},
					output: baseConfig(ConfigLXC{DNS: &GuestDNS{
						NameServers:  util.Pointer([]netip.Addr(nil)),
						SearchDomain: util.Pointer("example.com")}})}}},
		{category: `Features`,
			tests: []test{
				{name: `CreateDeviceNodes`,
					input: map[string]any{"features": "mknod=1"},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{
						CreateDeviceNodes: util.Pointer(true),
						FUSE:              util.Pointer(false),
						KeyCtl:            util.Pointer(false),
						NFS:               util.Pointer(false),
						Nesting:           util.Pointer(false),
						SMB:               util.Pointer(false)}})},
				{name: `FUSE`,
					input: map[string]any{"features": "fuse=1"},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{
						CreateDeviceNodes: util.Pointer(false),
						FUSE:              util.Pointer(true),
						KeyCtl:            util.Pointer(false),
						NFS:               util.Pointer(false),
						Nesting:           util.Pointer(false),
						SMB:               util.Pointer(false)}})},
				{name: `KeyCtl`,
					input: map[string]any{"features": "keyctl=1"},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{
						CreateDeviceNodes: util.Pointer(false),
						FUSE:              util.Pointer(false),
						KeyCtl:            util.Pointer(true),
						NFS:               util.Pointer(false),
						Nesting:           util.Pointer(false),
						SMB:               util.Pointer(false)}})},
				{name: `NFS`,
					input: map[string]any{"features": "mount=nfs"},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{
						CreateDeviceNodes: util.Pointer(false),
						FUSE:              util.Pointer(false),
						KeyCtl:            util.Pointer(false),
						NFS:               util.Pointer(true),
						Nesting:           util.Pointer(false),
						SMB:               util.Pointer(false)}})},
				{name: `NFS and SMB`,
					input: map[string]any{"features": "mount=nfs;cifs"},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{
						CreateDeviceNodes: util.Pointer(false),
						FUSE:              util.Pointer(false),
						KeyCtl:            util.Pointer(false),
						NFS:               util.Pointer(true),
						Nesting:           util.Pointer(false),
						SMB:               util.Pointer(true)}})},
				{name: `Nesting`,
					input: map[string]any{"features": "nesting=1"},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{
						CreateDeviceNodes: util.Pointer(false),
						FUSE:              util.Pointer(false),
						KeyCtl:            util.Pointer(false),
						NFS:               util.Pointer(false),
						Nesting:           util.Pointer(true),
						SMB:               util.Pointer(false)}})},
				{name: `SMB`,
					input: map[string]any{"features": "mount=cifs"},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{
						CreateDeviceNodes: util.Pointer(false),
						FUSE:              util.Pointer(false),
						KeyCtl:            util.Pointer(false),
						NFS:               util.Pointer(false),
						Nesting:           util.Pointer(false),
						SMB:               util.Pointer(true)}})},
				{name: `SMB and NFS`,
					input: map[string]any{"features": "mount=cifs;nfs"},
					output: baseConfig(ConfigLXC{Features: &LxcFeatures{
						CreateDeviceNodes: util.Pointer(false),
						FUSE:              util.Pointer(false),
						KeyCtl:            util.Pointer(false),
						NFS:               util.Pointer(true),
						Nesting:           util.Pointer(false),
						SMB:               util.Pointer(true)}})}}},
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
		{category: `Name`,
			tests: []test{
				{name: `set`,
					input:  RawConfigLXC{"name": "test"},
					output: baseConfig(ConfigLXC{Name: util.Pointer(GuestName("test"))})}}},
		{category: `Networks`,
			tests: []test{
				{name: `all`,
					input: RawConfigLXC{"net0": "name=eth0,bridge=vmbr0,ip=192.168.0.23/24,gw=12.168.0.1,rate=0.810,trunks=101,hwaddr=00:A1:22:b3:44:55,tag=100,link_down=1,firewall=1,ip6=2001:db8::1/64,gw6=2001:db8::2,mtu=896"},
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
					input: RawConfigLXC{"net0": "bridge=vmbr0"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID0: baseNetwork(LxcNetwork{Bridge: util.Pointer("vmbr0")})}})},
				{name: `Connected`,
					input: RawConfigLXC{"net1": "link_down=1"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID1: baseNetwork(LxcNetwork{Connected: util.Pointer(false)})}})},
				{name: `Firewall`,
					input: RawConfigLXC{"net2": "firewall=1"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID2: baseNetwork(LxcNetwork{Firewall: util.Pointer(true)})}})},
				{name: `IPv4 Address`,
					input: RawConfigLXC{"net3": "ip=192.168.0.10/24"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID3: baseNetwork(LxcNetwork{IPv4: &LxcIPv4{
							Address: util.Pointer(IPv4CIDR("192.168.0.10/24"))}})}})},
				{name: `IPv4 DHCP`,
					input: RawConfigLXC{"net4": "ip=dhcp"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID4: baseNetwork(LxcNetwork{IPv4: &LxcIPv4{
							DHCP: true}})}})},
				{name: `IPv4 Gateway`,
					input: RawConfigLXC{"net5": "gw=1.1.1.1"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID5: baseNetwork(LxcNetwork{IPv4: &LxcIPv4{
							Gateway: util.Pointer(IPv4Address("1.1.1.1"))}})}})},
				{name: `IPv4 Manual`,
					input: RawConfigLXC{"net6": "ip=manual"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID6: baseNetwork(LxcNetwork{IPv4: &LxcIPv4{
							Manual: true}})}})},
				{name: `IPv6 Address`,
					input: RawConfigLXC{"net7": "ip6=2001:db8::1/64"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID7: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
							Address: util.Pointer(IPv6CIDR("2001:db8::1/64"))}})}})},
				{name: `IPv6 DHCP`,
					input: RawConfigLXC{"net8": "ip6=dhcp"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID8: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
							DHCP: true}})}})},
				{name: `IPv6 Gateway`,
					input: RawConfigLXC{"net9": "gw6=2001:db8::2"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID9: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
							Gateway: util.Pointer(IPv6Address("2001:db8::2"))}})}})},
				{name: `IPv6 Manual`,
					input: RawConfigLXC{"net10": "ip6=manual"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID10: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
							Manual: true}})}})},
				{name: `IPv6 SLAAC`,
					input: RawConfigLXC{"net11": "ip6=auto"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID11: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
							SLAAC: true}})}})},
				{name: `MAC`,
					input: RawConfigLXC{"net12": "hwaddr=00:A1:22:b3:44:55"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID12: baseNetwork(LxcNetwork{
							MAC: util.Pointer(parseMAC("00:a1:22:B3:44:55")),
							mac: "00:A1:22:b3:44:55"})}})},
				{name: `Mtu`,
					input: RawConfigLXC{"net13": "mtu=1321"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID13: baseNetwork(LxcNetwork{Mtu: util.Pointer(MTU(1321))})}})},
				{name: `Name`,
					input: RawConfigLXC{"net13": "name=eth0"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID13: baseNetwork(LxcNetwork{Name: util.Pointer(LxcNetworkName("eth0"))})}})},
				{name: `NativeVlan`,
					input: RawConfigLXC{"net14": "tag=100"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID14: baseNetwork(LxcNetwork{NativeVlan: util.Pointer(Vlan(100))})}})},
				{name: `RateLimitKBps`,
					input: RawConfigLXC{"net15": "rate=95.649"},
					output: baseConfig(ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID15: baseNetwork(LxcNetwork{RateLimitKBps: util.Pointer(GuestNetworkRate(95649))})}})},
				{name: `TaggedVlans`,
					input: RawConfigLXC{"net0": "trunks=200;100;300"},
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
		{category: `Swap`,
			tests: []test{
				{name: `set`,
					input:  RawConfigLXC{"swap": float64(256)},
					output: baseConfig(ConfigLXC{Swap: util.Pointer(LxcSwap(256))})},
				{name: `set 0`,
					input:  RawConfigLXC{"swap": float64(0)},
					output: baseConfig(ConfigLXC{Swap: util.Pointer(LxcSwap(0))})}}},
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

func Test_LxcMemory_String(t *testing.T) {
	require.Equal(t, "583421", LxcMemory(583421).String())
}

func Test_LxcSwap_String(t *testing.T) {
	require.Equal(t, "8423", LxcSwap(8423).String())
}
