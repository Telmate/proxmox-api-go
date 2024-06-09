package proxmox

import (
	"crypto"
	"errors"
	"testing"

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
		name   string
		input  CloudInit
		output error
	}{
		{name: `Valid CloudInit CloudInitCustom FilePath`,
			input: CloudInit{Custom: &CloudInitCustom{
				Meta: &CloudInitSnippet{
					FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Max_Legal())}}}},
		{name: `Valid CloudInit CloudInitCustom FilePath empty`,
			input: CloudInit{Custom: &CloudInitCustom{Network: &CloudInitSnippet{}}}},
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
			input:  CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{32: ""}},
			output: errors.New(QemuNetworkInterfaceID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
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
