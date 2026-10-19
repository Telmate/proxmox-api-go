package proxmox

import (
	"errors"
	"net"
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_lxc"
	"github.com/stretchr/testify/require"
)

func testData_RawConfigLXC_Networks_Get() lxcTestGetFunc {
	baseConfig := func(config ConfigLXC) *ConfigLXC {
		return lxcGetBaseConfig(config)
	}
	baseNetwork := func(config LxcNetwork) LxcNetwork {
		if config.Bridge == nil {
			config.Bridge = new("")
		}
		if config.Connected == nil {
			config.Connected = new(true)
		}
		if config.Firewall == nil {
			config.Firewall = new(false)
		}
		if config.Name == nil {
			config.Name = new(LxcNetworkName(""))
		}
		if config.MAC == nil {
			var mac net.HardwareAddr
			config.MAC = new(mac)
		}
		return config
	}
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{
			{name: `all`,
				input: rawConfigLXC{a: map[string]any{"net0": "name=eth0,bridge=vmbr0,ip=192.168.0.23/24,gw=12.168.0.1,rate=0.810,trunks=101,hwaddr=00:A1:22:b3:44:55,tag=100,link_down=1,firewall=1,ip6=2001:db8::1/64,gw6=2001:db8::2,mtu=896"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID0: baseNetwork(LxcNetwork{
						Bridge:    new("vmbr0"),
						Connected: new(false),
						Firewall:  new(true),
						IPv4: &LxcIPv4{
							Address: new(IPv4CIDR("192.168.0.23/24")),
							Gateway: new(IPv4Address("12.168.0.1"))},
						IPv6: &LxcIPv6{
							Address: new(IPv6CIDR("2001:db8::1/64")),
							Gateway: new(IPv6Address("2001:db8::2"))},
						MAC:           new(parseMAC("00:a1:22:B3:44:55")),
						Mtu:           new(MTU(896)),
						Name:          new(LxcNetworkName("eth0")),
						NativeVlan:    new(Vlan(100)),
						RateLimitKBps: new(GuestNetworkRate(810)),
						TaggedVlans:   new(Vlans{Vlan(101)}),
						mac:           "00:A1:22:b3:44:55"})}})},
			{name: `Bridge`,
				input: rawConfigLXC{a: map[string]any{"net0": "bridge=vmbr0"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID0: baseNetwork(LxcNetwork{Bridge: new("vmbr0")})}})},
			{name: `Connected`,
				input: rawConfigLXC{a: map[string]any{"net1": "link_down=1"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID1: baseNetwork(LxcNetwork{Connected: new(false)})}})},
			{name: `Firewall`,
				input: rawConfigLXC{a: map[string]any{"net2": "firewall=1"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID2: baseNetwork(LxcNetwork{Firewall: new(true)})}})},
			{name: `HostManaged false`,
				input: rawConfigLXC{
					a:       map[string]any{"net12": ""},
					version: Version{Major: 9, Minor: 1}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID12: baseNetwork(LxcNetwork{HostManaged: new(false)})}})},
			{name: `HostManaged true`,
				input: rawConfigLXC{
					a:       map[string]any{"net12": "host-managed=1"},
					version: Version{Major: 9, Minor: 1}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID12: baseNetwork(LxcNetwork{HostManaged: new(true)})}})},
			{name: `IPv4 Address`,
				input: rawConfigLXC{a: map[string]any{"net3": "ip=192.168.0.10/24"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID3: baseNetwork(LxcNetwork{IPv4: &LxcIPv4{
						Address: new(IPv4CIDR("192.168.0.10/24"))}})}})},
			{name: `IPv4 DHCP`,
				input: rawConfigLXC{a: map[string]any{"net4": "ip=dhcp"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID4: baseNetwork(LxcNetwork{IPv4: &LxcIPv4{
						DHCP: true}})}})},
			{name: `IPv4 Gateway`,
				input: rawConfigLXC{a: map[string]any{"net5": "gw=1.1.1.1"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID5: baseNetwork(LxcNetwork{IPv4: &LxcIPv4{
						Gateway: new(IPv4Address("1.1.1.1"))}})}})},
			{name: `IPv4 Manual`,
				input: rawConfigLXC{a: map[string]any{"net6": "ip=manual"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID6: baseNetwork(LxcNetwork{IPv4: &LxcIPv4{
						Manual: true}})}})},
			{name: `IPv6 Address`,
				input: rawConfigLXC{a: map[string]any{"net7": "ip6=2001:db8::1/64"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID7: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
						Address: new(IPv6CIDR("2001:db8::1/64"))}})}})},
			{name: `IPv6 DHCP`,
				input: rawConfigLXC{a: map[string]any{"net8": "ip6=dhcp"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID8: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
						DHCP: true}})}})},
			{name: `IPv6 Gateway`,
				input: rawConfigLXC{a: map[string]any{"net9": "gw6=2001:db8::2"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID9: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
						Gateway: new(IPv6Address("2001:db8::2"))}})}})},
			{name: `IPv6 Manual`,
				input: rawConfigLXC{a: map[string]any{"net10": "ip6=manual"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID10: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
						Manual: true}})}})},
			{name: `IPv6 SLAAC`,
				input: rawConfigLXC{a: map[string]any{"net11": "ip6=auto"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID11: baseNetwork(LxcNetwork{IPv6: &LxcIPv6{
						SLAAC: true}})}})},
			{name: `MAC`,
				input: rawConfigLXC{a: map[string]any{"net12": "hwaddr=00:A1:22:b3:44:55"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID12: baseNetwork(LxcNetwork{
						MAC: new(parseMAC("00:a1:22:B3:44:55")),
						mac: "00:A1:22:b3:44:55"})}})},
			{name: `Mtu`,
				input: rawConfigLXC{a: map[string]any{"net13": "mtu=1321"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID13: baseNetwork(LxcNetwork{Mtu: new(MTU(1321))})}})},
			{name: `Name`,
				input: rawConfigLXC{a: map[string]any{"net13": "name=eth0"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID13: baseNetwork(LxcNetwork{Name: new(LxcNetworkName("eth0"))})}})},
			{name: `NativeVlan`,
				input: rawConfigLXC{a: map[string]any{"net14": "tag=100"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID14: baseNetwork(LxcNetwork{NativeVlan: new(Vlan(100))})}})},
			{name: `RateLimitKBps`,
				input: rawConfigLXC{a: map[string]any{"net15": "rate=95.649"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID15: baseNetwork(LxcNetwork{RateLimitKBps: new(GuestNetworkRate(95649))})}})},
			{name: `TaggedVlans`,
				input: rawConfigLXC{a: map[string]any{"net0": "trunks=200;100;300"}},
				output: baseConfig(ConfigLXC{Networks: LxcNetworks{
					LxcNetworkID0: baseNetwork(LxcNetwork{TaggedVlans: &Vlans{Vlan(100), Vlan(200), Vlan(300)}})}})},
		}
	}
}

func Test_ConfigLXC_LxcNetworks_MapToApi(t *testing.T) {
	t.Parallel()
	network := func() LxcNetwork {
		return LxcNetwork{
			Bridge:    new("vmbr0"),
			Connected: new(false),
			Firewall:  new(true),
			IPv4: &LxcIPv4{
				Address: new(IPv4CIDR("192.168.10.12/24")),
				Gateway: new(IPv4Address("192.168.10.1"))},
			IPv6: &LxcIPv6{
				Address: new(IPv6CIDR("2001:db8::1234/64")),
				Gateway: new(IPv6Address("2001:db8::1"))},
			MAC:           new(parseMAC("52:A4:00:12:b4:56")),
			Mtu:           new(MTU(1500)),
			Name:          new(LxcNetworkName("my_net")),
			NativeVlan:    new(Vlan(23)),
			RateLimitKBps: new(GuestNetworkRate(45)),
			TaggedVlans:   new(Vlans{12, 23, 45}),
			mac:           "52:A4:00:12:b4:56",
		}
	}
	networkCurrent := func(add LxcNetwork) LxcNetwork {
		current := LxcNetwork{
			Bridge:        new("vmbr0"),
			Connected:     new(false),
			Firewall:      new(true),
			HostManaged:   add.HostManaged,
			MAC:           new(parseMAC("52:A4:00:12:b4:56")),
			Mtu:           new(MTU(1500)),
			Name:          new(LxcNetworkName("my_net")),
			mac:           "52:A4:00:12:b4:56",
			IPv4:          add.IPv4,
			IPv6:          add.IPv6,
			NativeVlan:    add.NativeVlan,
			TaggedVlans:   add.TaggedVlans,
			RateLimitKBps: add.RateLimitKBps,
		}
		if add.Bridge != nil {
			current.Bridge = add.Bridge
		}
		if add.Connected != nil {
			current.Connected = add.Connected
		}
		if add.Firewall != nil {
			current.Firewall = add.Firewall
		}
		if add.MAC != nil {
			current.MAC = add.MAC
		}
		if add.Mtu != nil {
			current.Mtu = add.Mtu
		}
		if add.Name != nil {
			current.Name = add.Name
		}
		return current
	}
	tests := lxcTestsApiFunc(func() lxcTestsAPI {
		return lxcTestsAPI{
			create: []lxcTestCaseAPI{
				{name: `Delete`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID3: LxcNetwork{Delete: true}}}},
				{name: `Bridge`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID0: LxcNetwork{Bridge: new("vmbr0")}}},
					body: map[string]string{"net0": "%2Cbridge%3Dvmbr0"}}, // ",bridge=vmbr0"
				{name: `Connected true`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID1: LxcNetwork{Connected: new(true)}}},
					body: map[string]string{"net1": ""}},
				{name: `Connected false`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID2: LxcNetwork{Connected: new(false)}}},
					body: map[string]string{"net2": "%2Clink_down%3D1"}}, // ",link_down=1"
				{name: `Firewall true`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID4: LxcNetwork{Firewall: new(true)}}},
					body: map[string]string{"net4": "%2Cfirewall%3D1"}}, // ",firewall=1"
				{name: `Firewall false`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID5: LxcNetwork{Firewall: new(false)}}},
					body: map[string]string{"net5": ""}},
				{name: `HostManaged true`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID13: LxcNetwork{HostManaged: new(true)}}},
					body: map[string]string{"net13": "%2Chost-managed%3D1"}}, // ",host-managed=1"
				{name: `HostManaged false`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID14: LxcNetwork{HostManaged: new(false)}}},
					body: map[string]string{"net14": ""}},
				{name: `IPv4.Address create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID6: LxcNetwork{IPv4: &LxcIPv4{
							Address: new(IPv4CIDR("10.0.0.10/24"))}}}},
					body: map[string]string{"net6": "%2Cip%3D10.0.0.10%2F24"}}, // ",ip=10.0.0.10/24"
				{name: `IPv4.Address empty`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID6: LxcNetwork{IPv4: &LxcIPv4{
							Address: new(IPv4CIDR(""))}}}},
					body: map[string]string{"net6": ""}},
				{name: `IPv4.DHCP create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID7: LxcNetwork{IPv4: &LxcIPv4{
							DHCP: true}}}},
					body: map[string]string{"net7": "%2Cip%3Ddhcp"}}, // ",ip=dhcp"
				{name: `IPv4.DHCP false`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID7: LxcNetwork{IPv4: &LxcIPv4{
							DHCP: false}}}},
					body: map[string]string{"net7": ""}},
				{name: `IPv4.Gateway create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID8: LxcNetwork{IPv4: &LxcIPv4{
							Gateway: new(IPv4Address("10.0.0.1"))}}}},
					body: map[string]string{"net8": "%2Cgw%3D10.0.0.1"}}, // ",gw=10.0.0.1"
				{name: `IPv4.Gateway empty`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID8: LxcNetwork{IPv4: &LxcIPv4{
							Gateway: new(IPv4Address(""))}}}},
					body: map[string]string{"net8": ""}},
				{name: `IPv4.Manual create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID9: LxcNetwork{IPv4: &LxcIPv4{
							Manual: true}}}},
					body: map[string]string{"net9": "%2Cip%3Dmanual"}}, // ",ip=manual"
				{name: `IPv4.Manual false`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID9: LxcNetwork{IPv4: &LxcIPv4{
							Manual: false}}}},
					body: map[string]string{"net9": ""}},
				{name: `IPv6.Address create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID10: LxcNetwork{IPv6: &LxcIPv6{
							Address: new(IPv6CIDR("2001:db8::1/64"))}}}},
					body: map[string]string{"net10": "%2Cip6%3D2001%3Adb8%3A%3A1%2F64"}}, // ",ip6=2001:db8::1/64"
				{name: `IPv6.Address empty`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID10: LxcNetwork{IPv6: &LxcIPv6{
							Address: new(IPv6CIDR(""))}}}},
					body: map[string]string{"net10": ""}},
				{name: `IPv6.DHCP create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID11: LxcNetwork{IPv6: &LxcIPv6{
							DHCP: true}}}},
					body: map[string]string{"net11": "%2Cip6%3Ddhcp"}}, // ",ip6=dhcp"
				{name: `IPv6.DHCP false`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID11: LxcNetwork{IPv6: &LxcIPv6{
							DHCP: false}}}},
					body: map[string]string{"net11": ""}},
				{name: `IPv6.Gateway create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID12: LxcNetwork{IPv6: &LxcIPv6{
							Gateway: new(IPv6Address("2001:db8::2"))}}}},
					body: map[string]string{"net12": "%2Cgw6%3D2001%3Adb8%3A%3A2"}}, // ",gw6=2001:db8::2"
				{name: `IPv6.Gateway empty`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID12: LxcNetwork{IPv6: &LxcIPv6{
							Gateway: new(IPv6Address(""))}}}},
					body: map[string]string{"net12": ""}},
				{name: `IPv6.SLAAC create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID13: LxcNetwork{IPv6: &LxcIPv6{
							SLAAC: true}}}},
					body: map[string]string{"net13": "%2Cip6%3Dauto"}}, // ",ip6=auto"
				{name: `IPv6.SLAAC false`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID13: LxcNetwork{IPv6: &LxcIPv6{
							SLAAC: false}}}},
					body: map[string]string{"net13": ""}},
				{name: `IPv6.Manual create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID14: LxcNetwork{IPv6: new(LxcIPv6{
							Manual: true})}}},
					body: map[string]string{"net14": "%2Cip6%3Dmanual"}}, // ",ip6=manual"
				{name: `MAC set`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID15: LxcNetwork{MAC: new(parseMAC("00:11:22:33:44:55"))}}},
					body: map[string]string{"net15": "%2Chwaddr%3D00%3A11%3A22%3A33%3A44%3A55"}}, // ",hwaddr=00:11:22:33:44:55"
				{name: `MAC unset`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID5: LxcNetwork{MAC: new(net.HardwareAddr{})}}},
					body: map[string]string{"net5": ""}},
				{name: `Mtu set`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID0: LxcNetwork{Mtu: new(MTU(1500))}}},
					body: map[string]string{"net0": "%2Cmtu%3D1500"}}, // ",mtu=1500"
				{name: `Mtu unset`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID1: LxcNetwork{Mtu: new(MTU(0))}}},
					body: map[string]string{"net1": ""}},
				{name: `Name`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID2: LxcNetwork{Name: new(LxcNetworkName("test0"))}}},
					body: map[string]string{"net2": "name%3Dtest0"}}, // "name=test0"
				{name: `NativeVlan set`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID3: LxcNetwork{NativeVlan: new(Vlan(100))}}},
					body: map[string]string{"net3": "%2Ctag%3D100"}}, // ",tag=100"
				{name: `NativeVlan unset`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID4: LxcNetwork{NativeVlan: new(Vlan(0))}}},
					body: map[string]string{"net4": ""}},
				{name: `RateLimitKBps set`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID5: LxcNetwork{RateLimitKBps: new(GuestNetworkRate(1023))}}},
					body: map[string]string{"net5": "%2Crate%3D1.023"}}, // ",rate=1.023"
				{name: `RateLimitKBps unset`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID6: LxcNetwork{RateLimitKBps: new(GuestNetworkRate(0))}}},
					body: map[string]string{"net6": ""}},
				{name: `TaggedVlans set`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID7: LxcNetwork{TaggedVlans: new(Vlans{Vlan(100), Vlan(200)})}}},
					body: map[string]string{"net7": "%2Ctrunks%3D100%3B200"}}, // ",trunks=100;200"
				{name: `TaggedVlans unset`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID8: LxcNetwork{TaggedVlans: new(Vlans{})}}},
					body: map[string]string{"net8": ""}}},
			createUpdate: []lxcTestCaseAPI{
				{name: `create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID1: LxcNetwork{Bridge: new("vmbr0")}}},
					currentConfig: ConfigLXC{},
					body:          map[string]string{"net1": "%2Cbridge%3Dvmbr0"}}, // ",bridge=vmbr0"
				{name: `delete no effect`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID1: LxcNetwork{Delete: true}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID5: LxcNetwork{Bridge: new("vmbr0")}}},
					omitDefaults: lxcDefaultsUpdate}},
			update: []lxcTestCaseAPI{
				{name: `create`,
					config: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID0: network()}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{
						LxcNetworkID1: network()}},
					body: map[string]string{"net0": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3AB4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `delete`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: LxcNetwork{Delete: true}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: network()}},
					body:          map[string]string{"delete": "net0"}},
				{name: `no change`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: network()}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: network()}},
					omitDefaults:  lxcDefaultsAll},
				{name: `Bridge replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: LxcNetwork{Bridge: new("vmbr3")}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: network()}},
					body:          map[string]string{"net0": "name%3Dmy_net%2Cbridge%3Dvmbr3%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr3,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `Connected replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID1: LxcNetwork{Connected: new(true)}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID1: network()}},
					body:          map[string]string{"net1": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Cfirewall%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `Firewall replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID2: LxcNetwork{Firewall: new(false)}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID2: network()}},
					body:          map[string]string{"net2": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,host-managed=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `Host Managed replace`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID2: LxcNetwork{HostManaged: new(true)}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID2: func() LxcNetwork {
						a := network()
						a.HostManaged = new(false)
						return a
					}()}},
					body: map[string]string{"net2": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Chost-managed%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,host-managed=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `Host Managed inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID2: LxcNetwork{Name: new(LxcNetworkName("change"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID2: func() LxcNetwork {
						a := network()
						a.HostManaged = new(true)
						return a
					}()}},
					body: map[string]string{"net2": "name%3Dchange%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Chost-managed%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=change,bridge=vmbr0,link_down=1,firewall=1,host-managed=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `IPv4.Address add`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID2: LxcNetwork{
						IPv4: &LxcIPv4{Address: new(IPv4CIDR("192.168.1.34/24"))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID2: networkCurrent(LxcNetwork{})}},
					body:          map[string]string{"net2": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.1.34%2F24%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.1.34/24,hwaddr=52:A4:00:12:b4:56,mtu=1500"
				{name: `IPv4.Address inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: LxcNetwork{
						Name: new(LxcNetworkName("test0")),
						IPv4: &LxcIPv4{}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: networkCurrent(LxcNetwork{
						IPv4: &LxcIPv4{Address: new(IPv4CIDR("192.168.1.34/24"))}})}},
					body: map[string]string{"net4": "name%3Dtest0%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.1.34%2F24%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500"}}, // "name=test0,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.1.34/24,hwaddr=52:A4:00:12:b4:56,mtu=1500"
				{name: `IPv4.Address replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID3: LxcNetwork{IPv4: &LxcIPv4{Address: new(IPv4CIDR("10.0.0.2/24"))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID3: network()}},
					body:          map[string]string{"net3": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D10.0.0.2%2F24%2Cgw%3D192.168.10.1%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=10.0.0.2/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `IPv4.DHCP inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: LxcNetwork{
						Name: new(LxcNetworkName("test0"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: networkCurrent(LxcNetwork{IPv4: &LxcIPv4{DHCP: true}})}},
					body:          map[string]string{"net4": "name%3Dtest0%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3Ddhcp%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500"}}, // "name=test0,bridge=vmbr0,link_down=1,firewall=1,ip=dhcp,hwaddr=52:A4:00:12:b4:56,mtu=1500"
				{name: `IPv4.DHCP replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: LxcNetwork{IPv4: &LxcIPv4{DHCP: true}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID4: network()}},
					body:          map[string]string{"net4": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3Ddhcp%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=dhcp,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `IPv4.Manual inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID8: LxcNetwork{
						Name: new(LxcNetworkName("test0"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID8: networkCurrent(LxcNetwork{IPv4: &LxcIPv4{Manual: true}})}},
					body:          map[string]string{"net8": "name%3Dtest0%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3Dmanual%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500"}}, // "name=test0,bridge=vmbr0,link_down=1,firewall=1,ip=manual,hwaddr=52:A4:00:12:b4:56,mtu=1500"
				{name: `IPv4.Manual replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID6: LxcNetwork{IPv4: &LxcIPv4{Manual: true}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID6: network()}},
					body:          map[string]string{"net6": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3Dmanual%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=manual,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `IPv4.Gateway inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID5: LxcNetwork{
						Name: new(LxcNetworkName("test0")),
						IPv4: &LxcIPv4{}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID5: networkCurrent(LxcNetwork{IPv4: &LxcIPv4{Gateway: new(IPv4Address("1.1.1.1"))}})}},
					body:          map[string]string{"net5": "name%3Dtest0%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cgw%3D1.1.1.1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500"}}, // "name=test0,bridge=vmbr0,link_down=1,firewall=1,gw=1.1.1.1,hwaddr=52:A4:00:12:b4:56,mtu=1500"
				{name: `IPv4.Gateway replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID5: LxcNetwork{IPv4: &LxcIPv4{Gateway: new(IPv4Address("1.1.1.1"))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID5: network()}},
					body:          map[string]string{"net5": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D1.1.1.1%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=1.1.1.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `IPv6.Address add`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: LxcNetwork{
						IPv6: &LxcIPv6{Address: new(IPv6CIDR("2001:db8::2/64"))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: networkCurrent(LxcNetwork{})}},
					body:          map[string]string{"net12": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip6%3D2001%3Adb8%3A%3A2%2F64%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip6=2001:db8::2/64,hwaddr=52:A4:00:12:b4:56,mtu=1500"
				{name: `IPv6.Address inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: LxcNetwork{
						Name: new(LxcNetworkName("test0")),
						IPv6: &LxcIPv6{}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: networkCurrent(LxcNetwork{
						IPv6: &LxcIPv6{Address: new(IPv6CIDR("2001:db8::2/64"))}})}},
					body: map[string]string{"net12": "name%3Dtest0%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip6%3D2001%3Adb8%3A%3A2%2F64%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500"}}, // "name=test0,bridge=vmbr0,link_down=1,firewall=1,ip6=2001:db8::2/64,hwaddr=52:A4:00:12:b4:56,mtu=1500"
				{name: `IPv6.Address replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID7: LxcNetwork{IPv6: &LxcIPv6{Address: new(IPv6CIDR("2001:db8::2/64"))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID7: network()}},
					body:          map[string]string{"net7": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3D2001%3Adb8%3A%3A2%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::2/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `IPv6.DHCP inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: LxcNetwork{
						Name: new(LxcNetworkName("test0"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: networkCurrent(LxcNetwork{IPv6: &LxcIPv6{DHCP: true}})}},
					body:          map[string]string{"net12": "name%3Dtest0%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip6%3Ddhcp%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500"}}, // "name=test0,ip6=dhcp" "name=test0,bridge=vmbr0,link_down=1,firewall=1,ip6=dhcp,hwaddr=52:A4:00:12:b4:56,mtu=1500"
				{name: `IPv6.DHCP replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID8: LxcNetwork{IPv6: &LxcIPv6{DHCP: true}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID8: network()}},
					body:          map[string]string{"net8": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3Ddhcp%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=dhcp,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `IPv6.SLAAC inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: LxcNetwork{
						Name: new(LxcNetworkName("test0"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: networkCurrent(LxcNetwork{IPv6: &LxcIPv6{SLAAC: true}})}},
					body:          map[string]string{"net13": "name%3Dtest0%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip6%3Dauto%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500"}}, // "name=test0,bridge=vmbr0,link_down=1,firewall=1,ip6=auto,hwaddr=52:A4:00:12:b4:56,mtu=1500"
				{name: `IPv6.SLAAC replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID10: LxcNetwork{IPv6: &LxcIPv6{SLAAC: true}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID10: network()}},
					body:          map[string]string{"net10": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3Dauto%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=auto,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `IPv6.Manual inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: LxcNetwork{
						Name: new(LxcNetworkName("test0"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: networkCurrent(LxcNetwork{IPv6: &LxcIPv6{Manual: true}})}},
					body:          map[string]string{"net13": "name%3Dtest0%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip6%3Dmanual%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500"}}, // "name=test0,bridge=vmbr0,link_down=1,firewall=1,ip6=manual,hwaddr=52:A4:00:12:b4:56,mtu=1500"
				{name: `IPv6.Manual replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID11: LxcNetwork{IPv6: &LxcIPv6{Manual: true}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID11: network()}},
					body:          map[string]string{"net11": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3Dmanual%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=manual,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `IPv6.Gateway inherit`,
					config: ConfigLXC{Networks: LxcNetworks{LxcNetworkID9: LxcNetwork{
						Name: new(LxcNetworkName("test0")),
						IPv6: &LxcIPv6{}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID9: networkCurrent(LxcNetwork{IPv6: &LxcIPv6{Gateway: new(IPv6Address("2001:db8::3"))}})}},
					body:          map[string]string{"net9": "name%3Dtest0%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cgw6%3D2001%3Adb8%3A%3A3%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500"}}, // "name=test0,bridge=vmbr0,link_down=1,firewall=1,gw6=2001:db8::3,hwaddr=52:A4:00:12:b4:56,mtu=1500"
				{name: `IPv6.Gateway replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID9: LxcNetwork{IPv6: &LxcIPv6{Gateway: new(IPv6Address("2001:db8::3"))}}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID9: network()}},
					body:          map[string]string{"net9": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A3%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::3,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `MAC replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: LxcNetwork{MAC: new(parseMAC("00:11:a2:B3:44:66"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID12: network()}},
					body:          map[string]string{"net12": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D00%3A11%3AA2%3AB3%3A44%3A66%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=00:11:A2:B3:44:66,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `Name replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: LxcNetwork{Name: new(LxcNetworkName("test0"))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID13: network()}},
					body:          map[string]string{"net13": "name%3Dtest0%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=test0,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=12;23;45"
				{name: `NativeVlan replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID14: LxcNetwork{NativeVlan: new(Vlan(200))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID14: network()}},
					body:          map[string]string{"net14": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D200%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=200,rate=0.045,trunks=12;23;45"
				{name: `RateLimitKBps replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID15: LxcNetwork{RateLimitKBps: new(GuestNetworkRate(2040))}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID15: network()}},
					body:          map[string]string{"net15": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D2.04%2Ctrunks%3D12%3B23%3B45"}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=2.04,trunks=12;23;45"
				{name: `TaggedVlans replace`,
					config:        ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: LxcNetwork{TaggedVlans: new(Vlans{Vlan(200), Vlan(100), Vlan(300)})}}},
					currentConfig: ConfigLXC{Networks: LxcNetworks{LxcNetworkID0: network()}},
					body:          map[string]string{"net0": "name%3Dmy_net%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cfirewall%3D1%2Cip%3D192.168.10.12%2F24%2Cgw%3D192.168.10.1%2Cip6%3D2001%3Adb8%3A%3A1234%2F64%2Cgw6%3D2001%3Adb8%3A%3A1%2Chwaddr%3D52%3AA4%3A00%3A12%3Ab4%3A56%2Cmtu%3D1500%2Ctag%3D23%2Crate%3D0.045%2Ctrunks%3D100%3B200%3B300"}}}, // "name=my_net,bridge=vmbr0,link_down=1,firewall=1,ip=192.168.10.12/24,gw=192.168.10.1,ip6=2001:db8::1234/64,gw6=2001:db8::1,hwaddr=52:A4:00:12:b4:56,mtu=1500,tag=23,rate=0.045,trunks=100;200;300"
		}
	})
	tests.Test(t)
}

func Test_RawConfigLXC_LxcNetworks_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Networks_Get().Test(t)
}

func Test_LxcNetwork_Validate(t *testing.T) {
	t.Parallel()
	baseConfig := func(config LxcNetwork) LxcNetwork {
		if config.Bridge == nil {
			config.Bridge = new("vmbr0")
		}
		if config.Name == nil {
			config.Name = new(LxcNetworkName("eth0"))
		}
		return config
	}
	type test struct {
		name    string
		config  LxcNetwork
		current *LxcNetwork // current will be used for update and ignored for create
		version Version
		err     error
	}
	type testType struct {
		create       []test
		createUpdate []test
		update       []test
	}
	tests := struct {
		valid   testType
		invalid testType
	}{
		invalid: testType{
			create: []test{
				{name: `Bridge errors.New(LxcNetwork_Error_BridgeRequired)`,
					config: LxcNetwork{},
					err:    errors.New(LxcNetwork_Error_BridgeRequired)},
				{name: `Name errors.New(LxcNetwork_Error_NameRequired)`,
					config: LxcNetwork{Bridge: new("vmbr0")},
					err:    errors.New(LxcNetwork_Error_NameRequired)}},
			createUpdate: []test{
				{name: `errors.New(LxcNetwork_Error_HostManaged)`,
					config:  baseConfig(LxcNetwork{HostManaged: new(true)}),
					current: &LxcNetwork{},
					version: Version{Major: 9},
					err:     errors.New(LxcNetwork_Error_HostManaged)},
				{name: `IPv4 errors.New(LxcIPv4_Error_MutuallyExclusiveAddress)`,
					config: baseConfig(LxcNetwork{IPv4: &LxcIPv4{
						DHCP:    true,
						Address: new(IPv4CIDR("192.168.0.10/24"))}}),
					current: &LxcNetwork{},
					err:     errors.New(LxcIPv4_Error_MutuallyExclusiveAddress)},
				{name: `IPv6 errors.New(LxcIPv6_Error_MutuallyExclusive)`,
					config: baseConfig(LxcNetwork{IPv6: &LxcIPv6{
						DHCP:  true,
						SLAAC: true}}),
					current: &LxcNetwork{},
					err:     errors.New(LxcIPv6_Error_MutuallyExclusive)},
				{name: `Mtu errors.New(MTU_Error_Invalid)`,
					config: baseConfig(LxcNetwork{
						Mtu: new(MTU(100))}),
					current: &LxcNetwork{},
					err:     errors.New(MTU_Error_Invalid)},
				{name: `Name errors.New(LxcNetworkName_Error_Invalid)`,
					config: baseConfig(LxcNetwork{
						Name: new(LxcNetworkName(test_data_lxc.LxcNetworkName_Character_Illegal()[0]))}),
					current: &LxcNetwork{},
					err:     errors.New(LxcNetworkName_Error_Invalid)},
				{name: `NativeVlan errors.New(Vlan_Error_Invalid)`,
					config: baseConfig(LxcNetwork{
						NativeVlan: new(Vlan(4096))}),
					current: &LxcNetwork{},
					err:     errors.New(Vlan_Error_Invalid)},
				{name: `RateLimitKBps errors.New(GuestNetworkRate_Error_Invalid)`,
					config: baseConfig(LxcNetwork{
						RateLimitKBps: new(GuestNetworkRate(1024000000))}),
					current: &LxcNetwork{},
					err:     errors.New(GuestNetworkRate_Error_Invalid)},
				{name: `TaggedVlans errors.New(Vlan_Error_Invalid)`,
					config: baseConfig(LxcNetwork{
						TaggedVlans: new(Vlans{Vlan(4096)})}),
					current: &LxcNetwork{},
					err:     errors.New(Vlan_Error_Invalid)}}},
		valid: testType{
			create: []test{
				{name: `minimum`,
					config: baseConfig(LxcNetwork{})}},
			createUpdate: []test{
				{name: `Host Managed v9.1`,
					config:  baseConfig(LxcNetwork{HostManaged: new(true)}),
					current: &LxcNetwork{},
					version: Version{Major: 9, Minor: 1}}},
			update: []test{
				{name: `minimum`,
					config:  LxcNetwork{},
					current: &LxcNetwork{}}}},
	}
	var name string
	for _, subTest := range append(tests.valid.create, tests.valid.createUpdate...) {
		name = "Valid/Create/" + subTest.name
		t.Run(name, func(*testing.T) {
			require.Equal(t, subTest.err, subTest.config.Validate(nil, subTest.version), name)
		})
	}
	for _, subTest := range append(tests.valid.update, tests.valid.createUpdate...) {
		name = "Valid/Update/" + subTest.name
		t.Run(name, func(*testing.T) {
			require.NotNil(t, subTest.current)
			require.Equal(t, subTest.err, subTest.config.Validate(subTest.current, subTest.version), name)
		})
	}
	for _, subTest := range append(tests.invalid.create, tests.invalid.createUpdate...) {
		name = "Invalid/Create/" + subTest.name
		t.Run(name, func(*testing.T) {
			require.Equal(t, subTest.err, subTest.config.Validate(nil, subTest.version), name)
		})
	}
	for _, subTest := range append(tests.invalid.update, tests.invalid.createUpdate...) {
		name = "Invalid/Update/" + subTest.name
		t.Run(name, func(*testing.T) {
			require.NotNil(t, subTest.current)
			require.Equal(t, subTest.err, subTest.config.Validate(subTest.current, subTest.version), name)
		})
	}
}

func Test_LxcNetworks_Validate(t *testing.T) {
	t.Parallel()
	type testInput struct {
		config  LxcNetworks
		current LxcNetworks
		version Version
	}
	tests := []struct {
		name   string
		input  testInput
		output error
	}{
		{name: `Invalid errors.New(LxcNetworks_Error_Amount)`,
			input:  testInput{config: LxcNetworks{0: {}, 1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}, 7: {}, 8: {}, 9: {}, 10: {}, 11: {}, 12: {}, 13: {}, 14: {}, 15: {}, 16: {}}},
			output: errors.New(LxcNetworks_Error_Amount)},
		{name: `Invalid duplicate name, create`,
			input: testInput{
				config: LxcNetworks{
					LxcNetworkID7:  {},
					LxcNetworkID15: {Name: new(LxcNetworkName("eth0"))},
					LxcNetworkID12: {Name: new(LxcNetworkName("eth0"))}}},
			output: errors.New(LxcNetworks_Error_DuplicateName)},
		{name: `Invalid duplicate name, update`,
			input: testInput{
				config: LxcNetworks{
					LxcNetworkID5:  {},
					LxcNetworkID12: {Name: new(LxcNetworkName("eth0"))}},
				current: LxcNetworks{
					LxcNetworkID12: {},
					LxcNetworkID15: {Name: new(LxcNetworkName("eth0"))}}},
			output: errors.New(LxcNetworks_Error_DuplicateName)},
		{name: `Invalid id errors.New(LxcNetworkID_Error_Invalid)`,
			input: testInput{
				config: LxcNetworks{
					16: {Name: new(LxcNetworkName("eth0"))}}},
			output: errors.New(LxcNetworkID_Error_Invalid)},
		{name: `Invalid errors.New(LxcNetwork_Error_BridgeRequired)`,
			input: testInput{
				config: LxcNetworks{
					LxcNetworkID0: {}}},
			output: errors.New(LxcNetwork_Error_BridgeRequired)},
		{name: `Valid duplicate name, update`,
			input: testInput{
				config: LxcNetworks{
					LxcNetworkID12: {Name: new(LxcNetworkName("replaced1"))},
					LxcNetworkID15: {Delete: true},
					LxcNetworkID3:  {Name: new(LxcNetworkName("switch2"))},
					LxcNetworkID8:  {Name: new(LxcNetworkName("switch1"))}},
				current: LxcNetworks{
					LxcNetworkID12: {Name: new(LxcNetworkName("eth1"))},
					LxcNetworkID15: {Name: new(LxcNetworkName("replaced1"))},
					LxcNetworkID3:  {Name: new(LxcNetworkName("switch1"))},
					LxcNetworkID8:  {Name: new(LxcNetworkName("switch2"))}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.config.Validate(test.input.current, test.input.version))
		})
	}
}

func Test_LxcNetworkID_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  LxcNetworkID
		output error
	}{
		{name: `Valid minimum`,
			input: 0},
		{name: `Valid maximum`,
			input: 15},
		{name: `Invalid`,
			input:  16,
			output: errors.New(LxcNetworkID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_LxcNetworkName_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  []string
		output error
	}{
		{name: `Valid`,
			input: test_data_lxc.LxcNetworkName_Legal()},
		{name: `Invalid errors.New(LxcNetworkName_Error_LengthMinimum)`,
			input:  []string{test_data_lxc.LxcNetworkName_Min_Illegal()},
			output: errors.New(LxcNetworkName_Error_LengthMinimum)},
		{name: `Invalid errors.New(LxcNetworkName_Error_LengthMaximum)`,
			input:  []string{test_data_lxc.LxcNetworkName_Max_Illegal()},
			output: errors.New(LxcNetworkName_Error_LengthMaximum)},
		{name: `Invalid errors.New(LxcNetworkName_Error_Invalid)`,
			input:  test_data_lxc.LxcNetworkName_Character_Illegal(),
			output: errors.New(LxcNetworkName_Error_Invalid)},
		{name: `Invalid errors.New(LxcNetworkName_Error_Invalid)`,
			input:  test_data_lxc.LxcNetworkName_Special_Illegal(),
			output: errors.New(LxcNetworkName_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, input := range test.input {
				require.Equal(t, test.output, LxcNetworkName(input).Validate())
			}
		})
	}
}

func Test_LxcIPv4_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  LxcIPv4
		output error
	}{
		{name: `Invalid errors.New(IPv4Address_Error_Invalid)`,
			input: LxcIPv4{
				Gateway: new(IPv4Address("invalid"))},
			output: errors.New(IPv4Address_Error_Invalid)},
		{name: `Invalid errors.New(IPv4CIDR_Error_Invalid)`,
			input: LxcIPv4{
				Address: new(IPv4CIDR("invalid"))},
			output: errors.New(IPv4CIDR_Error_Invalid)},
		{name: `Invalid errors.New(LxcIPv4_Error_MutuallyExclusive)`,
			input: LxcIPv4{
				DHCP:   true,
				Manual: true},
			output: errors.New(LxcIPv4_Error_MutuallyExclusive)},
		{name: `Invalid errors.New(LxcIPv4_Error_MutuallyExclusiveAddress) dhcp`,
			input: LxcIPv4{
				DHCP:    true,
				Address: new(IPv4CIDR("192.168.0.10/24"))},
			output: errors.New(LxcIPv4_Error_MutuallyExclusiveAddress)},
		{name: `Invalid errors.New(LxcIPv4_Error_MutuallyExclusiveAddress) manual`,
			input: LxcIPv4{
				Manual:  true,
				Address: new(IPv4CIDR("192.168.0.10/24"))},
			output: errors.New(LxcIPv4_Error_MutuallyExclusiveAddress)},
		{name: `Invalid errors.New(LxcIPv4_Error_MutuallyExclusiveGateway) dhcp`,
			input: LxcIPv4{
				DHCP:    true,
				Gateway: new(IPv4Address("192.168.0.1"))},
			output: errors.New(LxcIPv4_Error_MutuallyExclusiveGateway)},
		{name: `Invalid errors.New(LxcIPv4_Error_MutuallyExclusiveGateway) manual`,
			input: LxcIPv4{
				Manual:  true,
				Gateway: new(IPv4Address("192.168.0.1"))},
			output: errors.New(LxcIPv4_Error_MutuallyExclusiveGateway)},
		{name: `Valid Address`,
			input: LxcIPv4{Address: new(IPv4CIDR("192.168.0.10/24"))}},
		{name: `Valid Address and Gateway`,
			input: LxcIPv4{
				Address: new(IPv4CIDR("192.168.0.10/24")),
				Gateway: new(IPv4Address("192.168.0.1"))}},
		{name: `Valid DHCP`,
			input: LxcIPv4{DHCP: true}},
		{name: `Valid Gateway`,
			input: LxcIPv4{
				Gateway: new(IPv4Address("192.168.0.1"))}},
		{name: `Valid Manual`,
			input: LxcIPv4{Manual: true}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_LxcIPv6_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  LxcIPv6
		output error
	}{
		{name: `Invalid errors.New(IPv6Address_Error_Invalid)`,
			input: LxcIPv6{
				Gateway: new(IPv6Address("invalid"))},
			output: errors.New(IPv6Address_Error_Invalid)},
		{name: `Invalid errors.New(IPv6CIDR_Error_Invalid)`,
			input: LxcIPv6{
				Address: new(IPv6CIDR("invalid"))},
			output: errors.New(IPv6CIDR_Error_Invalid)},
		{name: `Invalid errors.New(LxcIPv6_Error_MutuallyExclusive) dhcp and manual`,
			input: LxcIPv6{
				DHCP:   true,
				Manual: true},
			output: errors.New(LxcIPv6_Error_MutuallyExclusive)},
		{name: `Invalid errors.New(LxcIPv6_Error_MutuallyExclusive) dhcp and slaac`,
			input: LxcIPv6{
				DHCP:  true,
				SLAAC: true},
			output: errors.New(LxcIPv6_Error_MutuallyExclusive)},
		{name: `Invalid errors.New(LxcIPv6_Error_MutuallyExclusive) manual and slaac`,
			input: LxcIPv6{
				Manual: true,
				SLAAC:  true},
			output: errors.New(LxcIPv6_Error_MutuallyExclusive)},
		{name: `Invalid errors.New(LxcIPv6_Error_MutuallyExclusiveAddress) dhcp`,
			input: LxcIPv6{
				DHCP:    true,
				Address: new(IPv6CIDR("2001:db8::2/64"))},
			output: errors.New(LxcIPv6_Error_MutuallyExclusiveAddress)},
		{name: `Invalid errors.New(LxcIPv6_Error_MutuallyExclusiveAddress) manual`,
			input: LxcIPv6{
				Manual:  true,
				Address: new(IPv6CIDR("2001:db8::2/64"))},
			output: errors.New(LxcIPv6_Error_MutuallyExclusiveAddress)},
		{name: `Invalid errors.New(LxcIPv6_Error_MutuallyExclusiveAddress) slaac`,
			input: LxcIPv6{
				SLAAC:   true,
				Address: new(IPv6CIDR("2001:db8::2/64"))},
			output: errors.New(LxcIPv6_Error_MutuallyExclusiveAddress)},
		{name: `Invalid errors.New(LxcIPv6_Error_MutuallyExclusiveGateway) dhcp`,
			input: LxcIPv6{
				DHCP:    true,
				Gateway: new(IPv6Address("2001:db8::3"))},
			output: errors.New(LxcIPv6_Error_MutuallyExclusiveGateway)},
		{name: `Invalid errors.New(LxcIPv6_Error_MutuallyExclusiveGateway) manual`,
			input: LxcIPv6{
				Manual:  true,
				Gateway: new(IPv6Address("2001:db8::3"))},
			output: errors.New(LxcIPv6_Error_MutuallyExclusiveGateway)},
		{name: `Invalid errors.New(LxcIPv6_Error_MutuallyExclusiveGateway) slaac`,
			input: LxcIPv6{
				SLAAC:   true,
				Gateway: new(IPv6Address("2001:db8::3"))},
			output: errors.New(LxcIPv6_Error_MutuallyExclusiveGateway)},
		{name: `Valid Address`,
			input: LxcIPv6{Address: new(IPv6CIDR("2001:db8::2/64"))}},
		{name: `Valid Address and Gateway`,
			input: LxcIPv6{
				Address: new(IPv6CIDR("2001:db8::2/64")),
				Gateway: new(IPv6Address("2001:db8::3"))}},
		{name: `Valid DHCP`,
			input: LxcIPv6{DHCP: true}},
		{name: `Valid Gateway`,
			input: LxcIPv6{
				Gateway: new(IPv6Address("2001:db8::3"))}},
		{name: `Valid Manual`,
			input: LxcIPv6{Manual: true}},
		{name: `Valid SLAAC`,
			input: LxcIPv6{SLAAC: true}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
