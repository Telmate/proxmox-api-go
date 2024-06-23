package proxmox

import (
	"crypto"
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_qemu"
	"github.com/stretchr/testify/require"
)

func Test_sshKeyUrlDecode(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output []crypto.PublicKey
	}{
		{name: "Decode",
			input:  test_data_qemu.PublicKey_Encoded_Input(),
			output: test_data_qemu.PublicKey_Decoded_Output()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, sshKeyUrlDecode(test.input))
		})
	}
}

// Test the encoding logic to encode the ssh keys
func Test_sshKeyUrlEncode(t *testing.T) {
	tests := []struct {
		name   string
		input  []crypto.PublicKey
		output string
	}{
		{name: "Encode",
			input:  test_data_qemu.PublicKey_Decoded_Input(),
			output: test_data_qemu.PublicKey_Encoded_Output()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, sshKeyUrlEncode(test.input))
		})
	}
}

func Test_CloudInit_Validate(t *testing.T) {
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

func Test_IPv4Address_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  IPv4Address
		output error
	}{
		{name: `Valid`,
			input: "192.168.45.1"},
		{name: "Valid empty"},
		{name: `Invalid is CIDR`,
			input:  "192.168.45.1/24",
			output: errors.New(IPv4Address_Error_Invalid)},
		{name: `Invalid is IPv6`,
			input:  "3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc",
			output: errors.New(IPv4Address_Error_Invalid)},
		{name: `Invalid is gibberish`,
			input:  "ABCDEFG123",
			output: errors.New(IPv4Address_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_IPv4CIDR_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  IPv4CIDR
		output error
	}{
		{name: `Valid`,
			input: "192.168.45.0/24"},
		{name: `Valid empty`},
		{name: `Invalid only IP no CIDR`,
			input:  "192.168.45.0",
			output: errors.New(IPv4CIDR_Error_Invalid)},
		{name: `Invalid is IPv6`,
			input:  "2001:0db8:85a3::/64",
			output: errors.New(IPv4CIDR_Error_Invalid)},
		{name: `Invalid gibberish`,
			input:  "ABCDEFG123",
			output: errors.New(IPv4CIDR_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_IPv6Address_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  IPv6Address
		output error
	}{
		{name: `Valid`,
			input: "3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc"},
		{name: `Valid empty`},
		{name: `Invalid is CIDR`,
			input:  "2001:0db8:85a3::/64",
			output: errors.New(IPv6Address_Error_Invalid)},
		{name: `Invalid is IPv4`,
			input:  "192.168.45.0",
			output: errors.New(IPv6Address_Error_Invalid)},
		{name: `Invalid is gibberish`,
			input:  "ABCDEFG123",
			output: errors.New(IPv6Address_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_IPv6CIDR_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  IPv6CIDR
		output error
	}{
		{name: `Valid`,
			input: "2001:0db8:85a3::/64"},
		{name: `Valid empty`},
		{name: `Invalid only IP no CIDR`,
			input:  "2001:0db8:85a3::",
			output: errors.New(IPv6CIDR_Error_Invalid)},
		{name: `Invalid is IPv4`,
			input:  "192.168.45.0/24",
			output: errors.New(IPv6CIDR_Error_Invalid)},
		{name: `Invalid gibberish`,
			input:  "ABCDEFG123",
			output: errors.New(IPv6CIDR_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
