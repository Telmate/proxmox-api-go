package proxmox

import (
	"errors"
	"net/netip"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_guest"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_qemu"
	"github.com/stretchr/testify/require"
)

func testData_RawConfigQemu_CloudInit_Get() qemuTestGetFunc {
	publicKeys := func() *[]AuthorizedKey {
		rawOutput := test_data_guest.AuthorizedKey_Decoded_Output()
		output := make([]AuthorizedKey, len(rawOutput))
		for i := range rawOutput {
			output[i] = AuthorizedKey{Options: rawOutput[i].Options, PublicKey: rawOutput[i].PublicKey, Comment: rawOutput[i].Comment}
		}
		return &output
	}
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `ALL`,
				input: rawConfigQemu{a: map[string]any{
					"cicustom":     string("meta=local-zfs:ci-meta.yml,network=local-lvm:ci-network.yml,user=folder:ci-user.yml,vendor=local:snippets/ci-custom.yml"),
					"searchdomain": string("example.com"),
					"nameserver":   string("1.1.1.1 8.8.8.8 9.9.9.9"),
					"ipconfig0":    string("ip=dhcp,ip6=dhcp"),
					"ipconfig19":   string(""),
					"ipconfig31":   string("ip=10.20.4.7/22"),
					"sshkeys":      test_data_guest.AuthorizedKey_Encoded_Input(),
					"ciupgrade":    float64(1),
					"cipassword":   string("Enter123!"),
					"ciuser":       string("root")}},
				output: testQemuBaseConfig_get(ConfigQemu{CloudInit: &CloudInit{
					Custom: &CloudInitCustom{
						Meta: &CloudInitSnippet{
							FilePath: "ci-meta.yml",
							Storage:  "local-zfs"},
						Network: &CloudInitSnippet{
							FilePath: "ci-network.yml",
							Storage:  "local-lvm"},
						User: &CloudInitSnippet{
							FilePath: "ci-user.yml",
							Storage:  "folder"},
						Vendor: &CloudInitSnippet{
							FilePath: "snippets/ci-custom.yml",
							Storage:  "local"}},
					DNS: &GuestDNS{
						SearchDomain: new("example.com"),
						NameServers:  &[]netip.Addr{parseIP("1.1.1.1"), parseIP("8.8.8.8"), parseIP("9.9.9.9")}},
					NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID0: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{DHCP: true},
							IPv6: &CloudInitIPv6Config{DHCP: true}},
						QemuNetworkInterfaceID31: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{Address: new(IPv4CIDR("10.20.4.7/22"))}}},
					PublicSSHkeys:   publicKeys(),
					UpgradePackages: new(true),
					UserPassword:    new("Enter123!"),
					Username:        new("root")}})},
			{name: `Custom Meta`,
				input: rawConfigQemu{a: map[string]any{
					"cicustom": string("meta=local-zfs:ci-meta.yml")}},
				output: testQemuBaseConfig_get(ConfigQemu{CloudInit: &CloudInit{
					Custom:            &CloudInitCustom{Meta: &CloudInitSnippet{FilePath: "ci-meta.yml", Storage: "local-zfs"}},
					NetworkInterfaces: CloudInitNetworkInterfaces{}}})},
			{name: `Custom Network`,
				input: rawConfigQemu{a: map[string]any{
					"cicustom": string("network=local-lvm:ci-network.yml")}},
				output: testQemuBaseConfig_get(ConfigQemu{CloudInit: &CloudInit{
					Custom:            &CloudInitCustom{Network: &CloudInitSnippet{FilePath: "ci-network.yml", Storage: "local-lvm"}},
					NetworkInterfaces: CloudInitNetworkInterfaces{}}})},
			{name: `Custom User`,
				input: rawConfigQemu{a: map[string]any{
					"cicustom": string("user=folder:ci-user.yml")}},
				output: testQemuBaseConfig_get(ConfigQemu{CloudInit: &CloudInit{
					Custom:            &CloudInitCustom{User: &CloudInitSnippet{FilePath: "ci-user.yml", Storage: "folder"}},
					NetworkInterfaces: CloudInitNetworkInterfaces{}}})},
			{name: `Custom Vendor`,
				input: rawConfigQemu{a: map[string]any{
					"cicustom": string("vendor=local:snippets/ci-custom.yml")}},
				output: testQemuBaseConfig_get(ConfigQemu{CloudInit: &CloudInit{
					Custom:            &CloudInitCustom{Vendor: &CloudInitSnippet{FilePath: "snippets/ci-custom.yml", Storage: "local"}},
					NetworkInterfaces: CloudInitNetworkInterfaces{}}})},
			{name: `DNS SearchDomain`,
				input: rawConfigQemu{a: map[string]any{
					"searchdomain": string("example.com")}},
				output: testQemuBaseConfig_get(ConfigQemu{CloudInit: &CloudInit{
					DNS: &GuestDNS{
						SearchDomain: util.Pointer("example.com"),
						NameServers:  util.Pointer(uninitializedArray[netip.Addr]())},
					NetworkInterfaces: CloudInitNetworkInterfaces{}}})},
			{name: `DNS SearchDomain empty`,
				input: rawConfigQemu{a: map[string]any{
					"searchdomain": string(" ")}},
				output: testQemuBaseConfig_get(ConfigQemu{})},
			{name: `DNS NameServers`,
				input: rawConfigQemu{a: map[string]any{
					"nameserver": string("1.1.1.1 8.8.8.8 9.9.9.9")}},
				output: testQemuBaseConfig_get(ConfigQemu{CloudInit: &CloudInit{
					DNS: &GuestDNS{
						SearchDomain: util.Pointer(""),
						NameServers:  &[]netip.Addr{parseIP("1.1.1.1"), parseIP("8.8.8.8"), parseIP("9.9.9.9")}},
					NetworkInterfaces: CloudInitNetworkInterfaces{}}})},
			{name: `NetworkInterfaces`,
				input: rawConfigQemu{a: map[string]any{
					"ipconfig0":  string("ip=dhcp,ip6=dhcp"),
					"ipconfig1":  string("ip6=auto"),
					"ipconfig2":  string("ip=192.168.1.10/24,gw=192.168.56.1,ip6=2001:0db8:abcd::/48,gw6=2001:0db8:abcd::1"),
					"ipconfig19": string(""),
					"ipconfig20": string(" "), // this single space is on purpose to test if it is ignored
					"ipconfig31": string("ip=10.20.4.7/22")}},
				output: testQemuBaseConfig_get(ConfigQemu{CloudInit: &CloudInit{
					NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID0: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{DHCP: true},
							IPv6: &CloudInitIPv6Config{DHCP: true}},
						QemuNetworkInterfaceID1: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{SLAAC: true}},
						QemuNetworkInterfaceID2: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{
								Address: util.Pointer(IPv4CIDR("192.168.1.10/24")),
								Gateway: util.Pointer(IPv4Address("192.168.56.1"))},
							IPv6: &CloudInitIPv6Config{
								Address: util.Pointer(IPv6CIDR("2001:0db8:abcd::/48")),
								Gateway: util.Pointer(IPv6Address("2001:0db8:abcd::1"))}},
						QemuNetworkInterfaceID31: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("10.20.4.7/22"))}}}}})},
			{name: `PublicSSHkeys`,
				input: rawConfigQemu{a: map[string]any{
					"sshkeys": test_data_guest.AuthorizedKey_Encoded_Input()}},
				output: testQemuBaseConfig_get(ConfigQemu{CloudInit: &CloudInit{
					NetworkInterfaces: CloudInitNetworkInterfaces{},
					PublicSSHkeys:     publicKeys()}})},
			{name: `UpgradePackages`,
				input: rawConfigQemu{a: map[string]any{
					"ciupgrade": float64(0)}},
				output: testQemuBaseConfig_get(ConfigQemu{CloudInit: &CloudInit{
					NetworkInterfaces: CloudInitNetworkInterfaces{},
					UpgradePackages:   util.Pointer(false)}})},
			{name: `UserPassword`,
				input: rawConfigQemu{a: map[string]any{
					"cipassword": string("Enter123!")}},
				output: testQemuBaseConfig_get(ConfigQemu{CloudInit: &CloudInit{
					NetworkInterfaces: CloudInitNetworkInterfaces{},
					UserPassword:      util.Pointer("Enter123!")}})},
			{name: `Username`,
				input: rawConfigQemu{a: map[string]any{
					"ciuser": string("root")}},
				output: testQemuBaseConfig_get(ConfigQemu{CloudInit: &CloudInit{
					NetworkInterfaces: CloudInitNetworkInterfaces{},
					Username:          util.Pointer("root")}})},
			{name: `Username empty`,
				input: rawConfigQemu{a: map[string]any{
					"ciuser": string(" ")}},
				output: testQemuBaseConfig_get(ConfigQemu{})}}
	}
}

func Test_CloudInit_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   CloudInit
		version Version
		output  error
	}{
		{name: `Valid CloudInit CloudInitCustom FilePath`,
			input: CloudInit{Custom: &CloudInitCustom{
				Meta: &CloudInitSnippet{
					FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Max_Legal())}}}},
		{name: `Valid CloudInit CloudInitCustom FilePath empty`,
			input: CloudInit{Custom: &CloudInitCustom{Network: &CloudInitSnippet{}}}},
		{name: `Valid CloudInit CloudInitNetworkInterfaces IPv4 Address`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID0: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("192.168.45.1/24"))})}}}},
		{name: `Valid CloudInit CloudInitNetworkInterfaces IPv4 Address empty`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID1: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR(""))})}}}},
		{name: `Valid CloudInit CloudInitNetworkInterfaces IPv4 DHCP Address empty`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID2: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{
					Address: util.Pointer(IPv4CIDR("")),
					DHCP:    true})}}}},
		{name: `Valid CloudInit CloudInitNetworkInterfaces IPv4 DHCP Gateway empty`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID3: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{
					Gateway: util.Pointer(IPv4Address("")),
					DHCP:    true})}}}},
		{name: `Valid CloudInit CloudInitNetworkInterfaces IPv4 Gateway`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID4: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address("192.168.45.1"))})}}}},
		{name: `Valid CloudInit CloudInitNetworkInterfaces IPv4 Gateway empty`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID4: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address(""))})}}}},
		{name: `Valid CloudInit CloudInitNetworkInterfaces IPv6 Address`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID9: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/64"))})}}}},
		{name: `Valid CloudInit CloudInitNetworkInterfaces IPv6 Address empty`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID10: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR(""))})}}}},
		{name: `Valid CloudInit CloudInitNetworkInterfaces IPv6 DHCP Address empty`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID11: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Address: util.Pointer(IPv6CIDR("")),
					DHCP:    true})}}}},
		{name: `Valid CloudInit CloudInitNetworkInterfaces IPv6 DHCP Gateway empty`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID12: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Gateway: util.Pointer(IPv6Address("")),
					DHCP:    true})}}}},
		{name: `Valid CloudInit CloudInitNetworkInterfaces IPv6 Gateway`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID13: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc"))})}}}},
		{name: `Valid CloudInit CloudInitNetworkInterfaces IPv6 Gateway empty`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID14: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address(""))})}}}},
		{name: `Valid CloudInit CloudInitNetworkInterfaces IPv6 SLAAC Address empty`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID15: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Address: util.Pointer(IPv6CIDR("")),
					SLAAC:   true})}}}},
		{name: `Valid CloudInit CloudInitNetworkInterfaces IPv6 SLAAC Gateway empty`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID16: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Gateway: util.Pointer(IPv6Address("")),
					SLAAC:   true})}}}},
		{name: `Valid CloudInit UpgradePackages v7`,
			version: Version{Major: 7, Minor: 255, Patch: 255},
			input:   CloudInit{UpgradePackages: util.Pointer(false)}},
		{name: `Valid CloudInit UpgradePackages v8`,
			version: Version{Major: 8},
			input:   CloudInit{UpgradePackages: util.Pointer(true)}},
		{name: `Invalid errors.New(CloudInit_Error_UpgradePackagesPre8)`,
			version: Version{Major: 7, Minor: 255, Patch: 255},
			input:   CloudInit{UpgradePackages: util.Pointer(true)},
			output:  errors.New(CloudInit_Error_UpgradePackagesPre8)},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_InvalidCharacters)`,
			input: CloudInit{Custom: &CloudInitCustom{User: &CloudInitSnippet{
				FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Character_Illegal()[0])}}},
			output: errors.New(CloudInitSnippetPath_Error_InvalidCharacters)},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_InvalidPath)`,
			input: CloudInit{Custom: &CloudInitCustom{Vendor: &CloudInitSnippet{
				FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_InvalidPath())}}},
			output: errors.New(CloudInitSnippetPath_Error_InvalidPath)},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_MaxLength)`,
			input: CloudInit{Custom: &CloudInitCustom{Meta: &CloudInitSnippet{
				FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Max_Illegal())}}},
			output: errors.New(CloudInitSnippetPath_Error_MaxLength)},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_Relative)`,
			input: CloudInit{Custom: &CloudInitCustom{Network: &CloudInitSnippet{
				FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Relative())}}},
			output: errors.New(CloudInitSnippetPath_Error_Relative)},
		{name: `Invalid errors.New(QemuNetworkInterfaceID_Error_Invalid)`,
			input:  CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{32: CloudInitNetworkConfig{}}},
			output: errors.New(QemuNetworkInterfaceID_Error_Invalid)},
		{name: `Invalid CloudInit CloudInitNetworkInterfaces IPv4 Address Mutually exclusive with DHCP`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID5: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{
					Address: util.Pointer(IPv4CIDR("192.168.45.1/24")),
					DHCP:    true})}}},
			output: errors.New(CloudInitIPv4Config_Error_DhcpAddressMutuallyExclusive)},
		{name: `Invalid CloudInit CloudInitNetworkInterfaces IPv4 Gateway Mutually exclusive with DHCP`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID6: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{
					Gateway: util.Pointer(IPv4Address("192.168.45.1")),
					DHCP:    true})}}},
			output: errors.New(CloudInitIPv4Config_Error_DhcpGatewayMutuallyExclusive)},
		{name: `Invalid CloudInit CloudInitNetworkInterfaces IPv4 Address errors.New(IPv4CIDR_Error_Invalid)`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID7: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("192.168.45.1"))})}}},
			output: errors.New(IPv4CIDR_Error_Invalid)},
		{name: `Invalid CloudInit CloudInitNetworkInterfaces IPv4 Gateway errors.New(IPv4Address_Error_Invalid)`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID8: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address("192.168.45.1/24"))})}}},
			output: errors.New(IPv4Address_Error_Invalid)},
		{name: `Invalid CloudInit CloudInitNetworkInterfaces IPv6 Address Mutually exclusive with DHCP`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID17: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/64")),
					DHCP:    true})}}},
			output: errors.New(CloudInitIPv6Config_Error_DhcpAddressMutuallyExclusive)},
		{name: `Invalid CloudInit CloudInitNetworkInterfaces IPv6 Address Mutually exclusive with SLAAC`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID18: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/64")),
					SLAAC:   true})}}},
			output: errors.New(CloudInitIPv6Config_Error_SlaacAddressMutuallyExclusive)},
		{name: `Invalid CloudInit CloudInitNetworkInterfaces IPv6 DHCP Mutually exclusive with SLAAC`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID19: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					DHCP:  true,
					SLAAC: true})}}},
			output: errors.New(CloudInitIPv6Config_Error_DhcpSlaacMutuallyExclusive)},
		{name: `Invalid CloudInit CloudInitNetworkInterfaces IPv6 Gateway Mutually exclusive with DHCP`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID20: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Gateway: util.Pointer(IPv6Address("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc")),
					DHCP:    true})}}},
			output: errors.New(CloudInitIPv6Config_Error_DhcpGatewayMutuallyExclusive)},
		{name: `Invalid CloudInit CloudInitNetworkInterfaces IPv6 Gateway Mutually exclusive with SLAAC`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID21: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Gateway: util.Pointer(IPv6Address("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc")),
					SLAAC:   true})}}},
			output: errors.New(CloudInitIPv6Config_Error_SlaacGatewayMutuallyExclusive)},
		{name: `Invalid CloudInit CloudInitNetworkInterfaces IPv6 Address errors.New(IPv6CIDR_Error_Invalid)`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID22: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc"))})}}},
			output: errors.New(IPv6CIDR_Error_Invalid)},
		{name: `Invalid CloudInit CloudInitNetworkInterfaces IPv6 Gateway errors.New(IPv6Address_Error_Invalid)`,
			input: CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID23: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("2001:0db8:85a3::/64"))})}}},
			output: errors.New(IPv6Address_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate(test.version))
		})
	}
}

func Test_CloudInitCustom_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  CloudInitCustom
		output error
	}{
		{name: `Valid CloudInitCustom FilePath`,
			input: CloudInitCustom{Meta: &CloudInitSnippet{
				FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Max_Legal())}}},
		{name: `Valid CloudInitCustom FilePath empty`,
			input: CloudInitCustom{Network: &CloudInitSnippet{}}},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_InvalidCharacters`,
			input: CloudInitCustom{User: &CloudInitSnippet{
				FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Character_Illegal()[0])}},
			output: errors.New(CloudInitSnippetPath_Error_InvalidCharacters)},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_InvalidPath)`,
			input: CloudInitCustom{Vendor: &CloudInitSnippet{
				FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_InvalidPath())}},
			output: errors.New(CloudInitSnippetPath_Error_InvalidPath)},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_MaxLength)`,
			input: CloudInitCustom{Meta: &CloudInitSnippet{
				FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Max_Illegal())}},
			output: errors.New(CloudInitSnippetPath_Error_MaxLength)},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_Relative)`,
			input: CloudInitCustom{Network: &CloudInitSnippet{
				FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Relative())}},
			output: errors.New(CloudInitSnippetPath_Error_Relative)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_CloudInitSnippet_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  CloudInitSnippet
		output error
	}{
		{name: `Valid FilePath`,
			input: CloudInitSnippet{FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Max_Legal())}},
		{name: `Valid FilePath empty`},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_InvalidCharacters)`,
			input:  CloudInitSnippet{FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Character_Illegal()[0])},
			output: errors.New(CloudInitSnippetPath_Error_InvalidCharacters)},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_InvalidPath)`,
			input:  CloudInitSnippet{FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_InvalidPath())},
			output: errors.New(CloudInitSnippetPath_Error_InvalidPath)},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_MaxLength)`,
			input:  CloudInitSnippet{FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Max_Illegal())},
			output: errors.New(CloudInitSnippetPath_Error_MaxLength)},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_Relative)`,
			input:  CloudInitSnippet{FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Relative())},
			output: errors.New(CloudInitSnippetPath_Error_Relative)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_CloudInitSnippetPath_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  []string
		output error
	}{
		{name: `Valid`,
			input: test_data_qemu.CloudInitSnippetPath_Legal()},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_Empty)`,
			input:  []string{test_data_qemu.CloudInitSnippetPath_Min_Illegal()},
			output: errors.New(CloudInitSnippetPath_Error_Empty)},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_InvalidCharacters)`,
			input:  test_data_qemu.CloudInitSnippetPath_Character_Illegal(),
			output: errors.New(CloudInitSnippetPath_Error_InvalidCharacters)},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_InvalidPath)`,
			input:  []string{test_data_qemu.CloudInitSnippetPath_InvalidPath()},
			output: errors.New(CloudInitSnippetPath_Error_InvalidPath)},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_MaxLength)`,
			input:  []string{test_data_qemu.CloudInitSnippetPath_Max_Illegal()},
			output: errors.New(CloudInitSnippetPath_Error_MaxLength)},
		{name: `Invalid errors.New(CloudInitSnippetPath_Error_Relative)`,
			input:  []string{test_data_qemu.CloudInitSnippetPath_Relative()},
			output: errors.New(CloudInitSnippetPath_Error_Relative)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, input := range test.input {
				require.Equal(t, test.output, CloudInitSnippetPath(input).Validate())
			}
		})
	}
}

func Test_CloudInitNetworkInterfaces_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  CloudInitNetworkInterfaces
		output error
	}{
		{name: `Valid IPv4 Address`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID0: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("192.168.45.1/24"))})}}},
		{name: `Valid IPv4 Address empty`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID1: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR(""))})}}},
		{name: `Valid IPv4 DHCP Address empty`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID2: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{
					Address: util.Pointer(IPv4CIDR("")),
					DHCP:    true})}}},
		{name: `Valid IPv4 DHCP Gateway empty`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID3: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{
					Gateway: util.Pointer(IPv4Address("")),
					DHCP:    true})}}},
		{name: `Valid IPv4 Gateway`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID4: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address("192.168.45.1"))})}}},
		{name: `Valid IPv4 Gateway empty`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID5: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address(""))})}}},
		{name: `Valid IPv6 Address`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID9: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/64"))})}}},
		{name: `Valid IPv6 Address empty`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID10: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR(""))})}}},
		{name: `Valid IPv6 DHCP Address empty`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID11: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Address: util.Pointer(IPv6CIDR("")),
					DHCP:    true})}}},
		{name: `Valid IPv6 DHCP Gateway empty`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID12: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Gateway: util.Pointer(IPv6Address("")),
					DHCP:    true})}}},
		{name: `Valid IPv6 Gateway`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID13: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc"))})}}},
		{name: `Valid IPv6 Gateway empty`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID14: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address(""))})}}},
		{name: `Valid IPv6 SLAAC Address empty`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID15: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Address: util.Pointer(IPv6CIDR("")),
					SLAAC:   true})}}},
		{name: `Valid IPv6 SLAAC Gateway empty`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID16: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Gateway: util.Pointer(IPv6Address("")),
					SLAAC:   true})}}},
		{name: `Invalid IPv4 Address Mutually exclusive with DHCP`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID5: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{
					Address: util.Pointer(IPv4CIDR("192.168.45.1/24")),
					DHCP:    true})}},
			output: errors.New(CloudInitIPv4Config_Error_DhcpAddressMutuallyExclusive)},
		{name: `Invalid IPv4 Gateway Mutually exclusive with DHCP`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID6: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{
					Gateway: util.Pointer(IPv4Address("192.168.45.1")),
					DHCP:    true})}},
			output: errors.New(CloudInitIPv4Config_Error_DhcpGatewayMutuallyExclusive)},
		{name: `Invalid IPv4 Address errors.New(IPv4CIDR_Error_Invalid)`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID7: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("192.168.45.1"))})}},
			output: errors.New(IPv4CIDR_Error_Invalid)},
		{name: `Invalid IPv4 Gateway errors.New(IPv4Address_Error_Invalid)`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID8: CloudInitNetworkConfig{
				IPv4: util.Pointer(CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address("192.168.45.1/24"))})}},
			output: errors.New(IPv4Address_Error_Invalid)},
		{name: `Invalid IPv6 Address Mutually exclusive with DHCP`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID17: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/64")),
					DHCP:    true})}},
			output: errors.New(CloudInitIPv6Config_Error_DhcpAddressMutuallyExclusive)},
		{name: `Invalid IPv6 Address Mutually exclusive with SLAAC`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID18: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/64")),
					SLAAC:   true})}},
			output: errors.New(CloudInitIPv6Config_Error_SlaacAddressMutuallyExclusive)},
		{name: `Invalid IPv6 DHCP Mutually exclusive with SLAAC`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID19: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					DHCP:  true,
					SLAAC: true})}},
			output: errors.New(CloudInitIPv6Config_Error_DhcpSlaacMutuallyExclusive)},
		{name: `Invalid IPv6 Gateway Mutually exclusive with DHCP`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID20: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Gateway: util.Pointer(IPv6Address("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc")),
					DHCP:    true})}},
			output: errors.New(CloudInitIPv6Config_Error_DhcpGatewayMutuallyExclusive)},
		{name: `Invalid IPv6 Gateway Mutually exclusive with SLAAC`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID21: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{
					Gateway: util.Pointer(IPv6Address("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc")),
					SLAAC:   true})}},
			output: errors.New(CloudInitIPv6Config_Error_SlaacGatewayMutuallyExclusive)},
		{name: `Invalid IPv6 Address errors.New(IPv6CIDR_Error_Invalid)`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID22: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc"))})}},
			output: errors.New(IPv6CIDR_Error_Invalid)},
		{name: `Invalid IPv6 Gateway errors.New(IPv6Address_Error_Invalid)`,
			input: CloudInitNetworkInterfaces{QemuNetworkInterfaceID23: CloudInitNetworkConfig{
				IPv6: util.Pointer(CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("2001:0db8:85a3::/64"))})}},
			output: errors.New(IPv6Address_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_CloudInitIPv4Config_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  CloudInitIPv4Config
		output error
	}{
		{name: `Valid Address`,
			input: CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("192.168.45.1/24"))}},
		{name: `Valid Address empty`,
			input: CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR(""))}},
		{name: `Valid DHCP Address empty`,
			input: CloudInitIPv4Config{
				Address: util.Pointer(IPv4CIDR("")),
				DHCP:    true}},
		{name: `Valid DHCP Gateway empty`,
			input: CloudInitIPv4Config{
				Gateway: util.Pointer(IPv4Address("")),
				DHCP:    true}},
		{name: `Valid Gateway`,
			input: CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address("192.168.45.1"))}},
		{name: `Valid Gateway empty`,
			input: CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address(""))}},
		{name: `Invalid Address Mutually exclusive with DHCP`,
			input: CloudInitIPv4Config{
				Address: util.Pointer(IPv4CIDR("192.168.45.1/24")),
				DHCP:    true},
			output: errors.New(CloudInitIPv4Config_Error_DhcpAddressMutuallyExclusive)},
		{name: `Invalid Gateway Mutually exclusive with DHCP`,
			input: CloudInitIPv4Config{
				Gateway: util.Pointer(IPv4Address("192.168.45.1")),
				DHCP:    true},
			output: errors.New(CloudInitIPv4Config_Error_DhcpGatewayMutuallyExclusive)},
		{name: `Invalid Address errors.New(IPv4CIDR_Error_Invalid)`,
			input:  CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("192.168.45.1"))},
			output: errors.New(IPv4CIDR_Error_Invalid)},
		{name: `Invalid Gateway errors.New(IPv4Address_Error_Invalid)`,
			input:  CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address("192.168.45.1/24"))},
			output: errors.New(IPv4Address_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_CloudInitIPv6Config_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  CloudInitIPv6Config
		output error
	}{
		{name: `Valid Address`,
			input: CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/64"))}},
		{name: `Valid Address empty`,
			input: CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR(""))}},
		{name: `Valid DHCP Address empty`,
			input: CloudInitIPv6Config{
				Address: util.Pointer(IPv6CIDR("")),
				DHCP:    true}},
		{name: `Valid DHCP Gateway empty`,
			input: CloudInitIPv6Config{
				Gateway: util.Pointer(IPv6Address("")),
				DHCP:    true}},
		{name: `Valid Gateway`,
			input: CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc"))}},
		{name: `Valid Gateway empty`,
			input: CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address(""))}},
		{name: `Valid SLAAC Address empty`,
			input: CloudInitIPv6Config{
				Address: util.Pointer(IPv6CIDR("")),
				SLAAC:   true}},
		{name: `Valid SLAAC Gateway empty`,
			input: CloudInitIPv6Config{
				Gateway: util.Pointer(IPv6Address("")),
				SLAAC:   true}},
		{name: `Invalid Address Mutually exclusive with DHCP`,
			input: CloudInitIPv6Config{
				Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/64")),
				DHCP:    true},
			output: errors.New(CloudInitIPv6Config_Error_DhcpAddressMutuallyExclusive)},
		{name: `Invalid Address Mutually exclusive with SLAAC`,
			input: CloudInitIPv6Config{
				Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/64")),
				SLAAC:   true},
			output: errors.New(CloudInitIPv6Config_Error_SlaacAddressMutuallyExclusive)},
		{name: `Invalid DHCP Mutually exclusive with SLAAC`,
			input: CloudInitIPv6Config{
				DHCP:  true,
				SLAAC: true},
			output: errors.New(CloudInitIPv6Config_Error_DhcpSlaacMutuallyExclusive)},
		{name: `Invalid Gateway Mutually exclusive with DHCP`,
			input: CloudInitIPv6Config{
				Gateway: util.Pointer(IPv6Address("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc")),
				DHCP:    true},
			output: errors.New(CloudInitIPv6Config_Error_DhcpGatewayMutuallyExclusive)},
		{name: `Invalid Gateway Mutually exclusive with SLAAC`,
			input: CloudInitIPv6Config{
				Gateway: util.Pointer(IPv6Address("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc")),
				SLAAC:   true},
			output: errors.New(CloudInitIPv6Config_Error_SlaacGatewayMutuallyExclusive)},
		{name: `Invalid Address errors.New(IPv6CIDR_Error_Invalid)`,
			input:  CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc"))},
			output: errors.New(IPv6CIDR_Error_Invalid)},
		{name: `Invalid Gateway errors.New(IPv6Address_Error_Invalid)`,
			input:  CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("2001:0db8:85a3::/64"))},
			output: errors.New(IPv6Address_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_RawConfigQemu_GetCloudInit(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_CloudInit_Get().Inject(t,
		func(t *testing.T, raw RawConfigQemu, output *ConfigQemu, _ error) {
			require.Equal(t, output.CloudInit, raw.GetCloudInit())
		})
}

func Test_RawConfigQemu_CloudInit_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_CloudInit_Get().Test(t)
}
