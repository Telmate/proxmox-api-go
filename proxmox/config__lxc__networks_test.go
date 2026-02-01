package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_lxc"
	"github.com/stretchr/testify/require"
)

func Test_LxcNetwork_Validate(t *testing.T) {
	t.Parallel()
	baseConfig := func(config LxcNetwork) LxcNetwork {
		if config.Bridge == nil {
			config.Bridge = util.Pointer("vmbr0")
		}
		if config.Name == nil {
			config.Name = util.Pointer(LxcNetworkName("eth0"))
		}
		return config
	}
	type test struct {
		name    string
		config  LxcNetwork
		current *LxcNetwork // current will be used for update and ignored for create
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
					config: LxcNetwork{Bridge: util.Pointer("vmbr0")},
					err:    errors.New(LxcNetwork_Error_NameRequired)}},
			createUpdate: []test{
				{name: `IPv4 errors.New(LxcIPv4_Error_MutuallyExclusiveAddress)`,
					config: baseConfig(LxcNetwork{IPv4: &LxcIPv4{
						DHCP:    true,
						Address: util.Pointer(IPv4CIDR("192.168.0.10/24"))}}),
					current: util.Pointer(LxcNetwork{}),
					err:     errors.New(LxcIPv4_Error_MutuallyExclusiveAddress)},
				{name: `IPv6 errors.New(LxcIPv6_Error_MutuallyExclusive)`,
					config: baseConfig(LxcNetwork{IPv6: &LxcIPv6{
						DHCP:  true,
						SLAAC: true}}),
					current: util.Pointer(LxcNetwork{}),
					err:     errors.New(LxcIPv6_Error_MutuallyExclusive)},
				{name: `Mtu errors.New(MTU_Error_Invalid)`,
					config: baseConfig(LxcNetwork{
						Mtu: util.Pointer(MTU(100))}),
					current: util.Pointer(LxcNetwork{}),
					err:     errors.New(MTU_Error_Invalid)},
				{name: `Name errors.New(LxcNetworkName_Error_Invalid)`,
					config: baseConfig(LxcNetwork{
						Name: util.Pointer(LxcNetworkName(test_data_lxc.LxcNetworkName_Character_Illegal()[0]))}),
					current: util.Pointer(LxcNetwork{}),
					err:     errors.New(LxcNetworkName_Error_Invalid)},
				{name: `NativeVlan errors.New(Vlan_Error_Invalid)`,
					config: baseConfig(LxcNetwork{
						NativeVlan: util.Pointer(Vlan(4096))}),
					current: util.Pointer(LxcNetwork{}),
					err:     errors.New(Vlan_Error_Invalid)},
				{name: `RateLimitKBps errors.New(GuestNetworkRate_Error_Invalid)`,
					config: baseConfig(LxcNetwork{
						RateLimitKBps: util.Pointer(GuestNetworkRate(1024000000))}),
					current: util.Pointer(LxcNetwork{}),
					err:     errors.New(GuestNetworkRate_Error_Invalid)},
				{name: `TaggedVlans errors.New(Vlan_Error_Invalid)`,
					config: baseConfig(LxcNetwork{
						TaggedVlans: util.Pointer(Vlans{Vlan(4096)})}),
					current: util.Pointer(LxcNetwork{}),
					err:     errors.New(Vlan_Error_Invalid)}}},
		valid: testType{
			create: []test{
				{name: `Valid minimum`,
					config: baseConfig(LxcNetwork{})}},
			update: []test{
				{name: `Valid minimum`,
					config:  LxcNetwork{},
					current: util.Pointer(LxcNetwork{})}}},
	}
	var name string
	for _, subTest := range append(tests.valid.create, tests.valid.createUpdate...) {
		name = "Valid/Create/" + subTest.name
		t.Run(name, func(*testing.T) {
			require.Equal(t, subTest.err, subTest.config.Validate(nil), name)
		})
	}
	for _, subTest := range append(tests.valid.update, tests.valid.createUpdate...) {
		name = "Valid/Update/" + subTest.name
		t.Run(name, func(*testing.T) {
			require.NotNil(t, subTest.current)
			require.Equal(t, subTest.err, subTest.config.Validate(subTest.current), name)
		})
	}
	for _, subTest := range append(tests.invalid.create, tests.invalid.createUpdate...) {
		name = "Invalid/Create/" + subTest.name
		t.Run(name, func(*testing.T) {
			require.Equal(t, subTest.err, subTest.config.Validate(nil), name)
		})
	}
	for _, subTest := range append(tests.invalid.update, tests.invalid.createUpdate...) {
		name = "Invalid/Update/" + subTest.name
		t.Run(name, func(*testing.T) {
			require.NotNil(t, subTest.current)
			require.Equal(t, subTest.err, subTest.config.Validate(subTest.current), name)
		})
	}
}

func Test_LxcNetworks_Validate(t *testing.T) {
	t.Parallel()
	type testInput struct {
		config  LxcNetworks
		current LxcNetworks
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
					LxcNetworkID15: {Name: util.Pointer(LxcNetworkName("eth0"))},
					LxcNetworkID12: {Name: util.Pointer(LxcNetworkName("eth0"))}}},
			output: errors.New(LxcNetworks_Error_DuplicateName)},
		{name: `Invalid duplicate name, update`,
			input: testInput{
				config: LxcNetworks{
					LxcNetworkID5:  {},
					LxcNetworkID12: {Name: util.Pointer(LxcNetworkName("eth0"))}},
				current: LxcNetworks{
					LxcNetworkID12: {},
					LxcNetworkID15: {Name: util.Pointer(LxcNetworkName("eth0"))}}},
			output: errors.New(LxcNetworks_Error_DuplicateName)},
		{name: `Invalid id errors.New(LxcNetworkID_Error_Invalid)`,
			input: testInput{
				config: LxcNetworks{
					16: {Name: util.Pointer(LxcNetworkName("eth0"))}}},
			output: errors.New(LxcNetworkID_Error_Invalid)},
		{name: `Invalid errors.New(LxcNetwork_Error_BridgeRequired)`,
			input: testInput{
				config: LxcNetworks{
					LxcNetworkID0: {}}},
			output: errors.New(LxcNetwork_Error_BridgeRequired)},
		{name: `Valid duplicate name, update`,
			input: testInput{
				config: LxcNetworks{
					LxcNetworkID12: {Name: util.Pointer(LxcNetworkName("replaced1"))},
					LxcNetworkID15: {Delete: true},
					LxcNetworkID3:  {Name: util.Pointer(LxcNetworkName("switch2"))},
					LxcNetworkID8:  {Name: util.Pointer(LxcNetworkName("switch1"))}},
				current: LxcNetworks{
					LxcNetworkID12: {Name: util.Pointer(LxcNetworkName("eth1"))},
					LxcNetworkID15: {Name: util.Pointer(LxcNetworkName("replaced1"))},
					LxcNetworkID3:  {Name: util.Pointer(LxcNetworkName("switch1"))},
					LxcNetworkID8:  {Name: util.Pointer(LxcNetworkName("switch2"))}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.config.Validate(test.input.current))
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
				Gateway: util.Pointer(IPv4Address("invalid"))},
			output: errors.New(IPv4Address_Error_Invalid)},
		{name: `Invalid errors.New(IPv4CIDR_Error_Invalid)`,
			input: LxcIPv4{
				Address: util.Pointer(IPv4CIDR("invalid"))},
			output: errors.New(IPv4CIDR_Error_Invalid)},
		{name: `Invalid errors.New(LxcIPv4_Error_MutuallyExclusive)`,
			input: LxcIPv4{
				DHCP:   true,
				Manual: true},
			output: errors.New(LxcIPv4_Error_MutuallyExclusive)},
		{name: `Invalid errors.New(LxcIPv4_Error_MutuallyExclusiveAddress) dhcp`,
			input: LxcIPv4{
				DHCP:    true,
				Address: util.Pointer(IPv4CIDR("192.168.0.10/24"))},
			output: errors.New(LxcIPv4_Error_MutuallyExclusiveAddress)},
		{name: `Invalid errors.New(LxcIPv4_Error_MutuallyExclusiveAddress) manual`,
			input: LxcIPv4{
				Manual:  true,
				Address: util.Pointer(IPv4CIDR("192.168.0.10/24"))},
			output: errors.New(LxcIPv4_Error_MutuallyExclusiveAddress)},
		{name: `Invalid errors.New(LxcIPv4_Error_MutuallyExclusiveGateway) dhcp`,
			input: LxcIPv4{
				DHCP:    true,
				Gateway: util.Pointer(IPv4Address("192.168.0.1"))},
			output: errors.New(LxcIPv4_Error_MutuallyExclusiveGateway)},
		{name: `Invalid errors.New(LxcIPv4_Error_MutuallyExclusiveGateway) manual`,
			input: LxcIPv4{
				Manual:  true,
				Gateway: util.Pointer(IPv4Address("192.168.0.1"))},
			output: errors.New(LxcIPv4_Error_MutuallyExclusiveGateway)},
		{name: `Valid Address`,
			input: LxcIPv4{Address: util.Pointer(IPv4CIDR("192.168.0.10/24"))}},
		{name: `Valid Address and Gateway`,
			input: LxcIPv4{
				Address: util.Pointer(IPv4CIDR("192.168.0.10/24")),
				Gateway: util.Pointer(IPv4Address("192.168.0.1"))}},
		{name: `Valid DHCP`,
			input: LxcIPv4{DHCP: true}},
		{name: `Valid Gateway`,
			input: LxcIPv4{
				Gateway: util.Pointer(IPv4Address("192.168.0.1"))}},
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
				Gateway: util.Pointer(IPv6Address("invalid"))},
			output: errors.New(IPv6Address_Error_Invalid)},
		{name: `Invalid errors.New(IPv6CIDR_Error_Invalid)`,
			input: LxcIPv6{
				Address: util.Pointer(IPv6CIDR("invalid"))},
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
				Address: util.Pointer(IPv6CIDR("2001:db8::2/64"))},
			output: errors.New(LxcIPv6_Error_MutuallyExclusiveAddress)},
		{name: `Invalid errors.New(LxcIPv6_Error_MutuallyExclusiveAddress) manual`,
			input: LxcIPv6{
				Manual:  true,
				Address: util.Pointer(IPv6CIDR("2001:db8::2/64"))},
			output: errors.New(LxcIPv6_Error_MutuallyExclusiveAddress)},
		{name: `Invalid errors.New(LxcIPv6_Error_MutuallyExclusiveAddress) slaac`,
			input: LxcIPv6{
				SLAAC:   true,
				Address: util.Pointer(IPv6CIDR("2001:db8::2/64"))},
			output: errors.New(LxcIPv6_Error_MutuallyExclusiveAddress)},
		{name: `Invalid errors.New(LxcIPv6_Error_MutuallyExclusiveGateway) dhcp`,
			input: LxcIPv6{
				DHCP:    true,
				Gateway: util.Pointer(IPv6Address("2001:db8::3"))},
			output: errors.New(LxcIPv6_Error_MutuallyExclusiveGateway)},
		{name: `Invalid errors.New(LxcIPv6_Error_MutuallyExclusiveGateway) manual`,
			input: LxcIPv6{
				Manual:  true,
				Gateway: util.Pointer(IPv6Address("2001:db8::3"))},
			output: errors.New(LxcIPv6_Error_MutuallyExclusiveGateway)},
		{name: `Invalid errors.New(LxcIPv6_Error_MutuallyExclusiveGateway) slaac`,
			input: LxcIPv6{
				SLAAC:   true,
				Gateway: util.Pointer(IPv6Address("2001:db8::3"))},
			output: errors.New(LxcIPv6_Error_MutuallyExclusiveGateway)},
		{name: `Valid Address`,
			input: LxcIPv6{Address: util.Pointer(IPv6CIDR("2001:db8::2/64"))}},
		{name: `Valid Address and Gateway`,
			input: LxcIPv6{
				Address: util.Pointer(IPv6CIDR("2001:db8::2/64")),
				Gateway: util.Pointer(IPv6Address("2001:db8::3"))}},
		{name: `Valid DHCP`,
			input: LxcIPv6{DHCP: true}},
		{name: `Valid Gateway`,
			input: LxcIPv6{
				Gateway: util.Pointer(IPv6Address("2001:db8::3"))}},
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
