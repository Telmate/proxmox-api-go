package proxmox

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_qemu"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func Test_AuthorizedKey_Parse(t *testing.T) {
	const key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEY5T2JQgiL5Z5Yuy4yXuUYglVJlpsokHFXR1hvnCVYW"
	parsePublicKey := func(rawKey string) ssh.PublicKey {
		key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(rawKey))
		failError(err)
		return key
	}
	type output struct {
		err error
		key *AuthorizedKey
	}
	tests := []struct {
		name   string
		input  string
		output output
		holder *AuthorizedKey
	}{
		{name: `errors.New(AuthorizedKey_ErrorNilPointer)`,
			input:  key,
			output: output{err: errors.New(AuthorizedKey_Error_NilPointer)}},
		{name: `Normal`,
			input: key + " test comment",
			output: output{key: &AuthorizedKey{
				PublicKey: parsePublicKey(key),
				Comment:   "test comment"}},
			holder: &AuthorizedKey{}},
		{name: `Key only`,
			input:  key,
			output: output{key: &AuthorizedKey{PublicKey: parsePublicKey(key)}},
			holder: &AuthorizedKey{}},
		{name: `Comment only`,
			input:  "test comment",
			output: output{key: &AuthorizedKey{}, err: errors.New("")},
			holder: &AuthorizedKey{}},
		{name: `Key with options`,
			input:  `no-pty,command="/path/to/script.sh" ` + key,
			holder: &AuthorizedKey{},
			output: output{key: &AuthorizedKey{
				Options:   []string{`no-pty`, `command="/path/to/script.sh"`},
				PublicKey: parsePublicKey(key)}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.holder.Parse([]byte(test.input))
			if test.output.err != nil {
				require.ErrorContains(t, err, test.output.err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.output.key, test.holder)
		})
	}
}

func Test_AuthorizedKey_MarshalJSON(t *testing.T) {
	const key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEY5T2JQgiL5Z5Yuy4yXuUYglVJlpsokHFXR1hvnCVYW"
	parsePublicKey := func(rawKey string) ssh.PublicKey {
		key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(rawKey))
		failError(err)
		return key
	}
	type testData struct {
		Keys []AuthorizedKey `json:"keys"`
	}
	tests := []struct {
		name   string
		input  testData
		output string
		err    error
	}{
		{name: `empty key`,
			input:  testData{Keys: []AuthorizedKey{{}}},
			output: `{"keys":[""]}`},
		{name: `normal`,
			input: testData{Keys: []AuthorizedKey{
				{PublicKey: parsePublicKey(key), Comment: "test comment"}}},
			output: `{"keys":["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEY5T2JQgiL5Z5Yuy4yXuUYglVJlpsokHFXR1hvnCVYW test comment"]}`},
		{name: `key only`,
			input: testData{Keys: []AuthorizedKey{
				{PublicKey: parsePublicKey(key)}}},
			output: `{"keys":["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEY5T2JQgiL5Z5Yuy4yXuUYglVJlpsokHFXR1hvnCVYW"]}`},
		{name: `key with options`,
			input: testData{Keys: []AuthorizedKey{
				{Options: []string{`no-pty`, `command="/path/to/script.sh"`},
					PublicKey: parsePublicKey(key)}}},
			output: `{"keys":["no-pty,command=\"/path/to/script.sh\" ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEY5T2JQgiL5Z5Yuy4yXuUYglVJlpsokHFXR1hvnCVYW"]}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := json.Marshal(test.input)
			require.Equal(t, test.output, string(output))
			if test.err == nil {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, test.err.Error())
			}
		})
	}
}

func Test_AuthorizedKey_String(t *testing.T) {
	const key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEY5T2JQgiL5Z5Yuy4yXuUYglVJlpsokHFXR1hvnCVYW"
	parsePublicKey := func(rawKey string) ssh.PublicKey {
		key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(rawKey))
		failError(err)
		return key
	}
	tests := []struct {
		name   string
		input  AuthorizedKey
		output string
	}{
		{name: `Empty key`, input: AuthorizedKey{}, output: ""},
		{name: `Normal`,
			input: AuthorizedKey{
				PublicKey: parsePublicKey(key),
				Comment:   "test comment"},
			output: key + " test comment"},
		{name: `Key only`,
			input: AuthorizedKey{
				PublicKey: parsePublicKey(key)},
			output: key},
		{name: `Comment only`,
			input: AuthorizedKey{
				Comment: "test comment"},
			output: ""},
		{name: `Comment whitespace`,
			input: AuthorizedKey{
				PublicKey: parsePublicKey(key),
				Comment:   " "},
			output: key},
		{name: `Key with options`,
			input: AuthorizedKey{
				Options:   []string{`no-pty`, `command="/path/to/script.sh"`},
				PublicKey: parsePublicKey(key)},
			output: `no-pty,command="/path/to/script.sh" ` + key},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.String())
		})
	}
}

func Test_AuthorizedKey_UnmarshalJSON(t *testing.T) {
	const key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEY5T2JQgiL5Z5Yuy4yXuUYglVJlpsokHFXR1hvnCVYW"
	parsePublicKey := func(rawKey string) ssh.PublicKey {
		key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(rawKey))
		failError(err)
		return key
	}
	type testData struct {
		Keys []AuthorizedKey `json:"keys"`
	}
	tests := []struct {
		name   string
		input  string
		output testData
		err    error
	}{
		{name: `valid normal`,
			input: `{"keys":["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEY5T2JQgiL5Z5Yuy4yXuUYglVJlpsokHFXR1hvnCVYW test comment"]}`,
			output: testData{Keys: []AuthorizedKey{
				{PublicKey: parsePublicKey(key), Comment: "test comment"}}}},
		{name: `valid key only`,
			input: `{"keys":["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEY5T2JQgiL5Z5Yuy4yXuUYglVJlpsokHFXR1hvnCVYW"]}`,
			output: testData{Keys: []AuthorizedKey{
				{PublicKey: parsePublicKey(key)}}}},
		{name: `valid with options`,
			input: `{"keys":["no-pty,command=\"/path/to/script.sh\" ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEY5T2JQgiL5Z5Yuy4yXuUYglVJlpsokHFXR1hvnCVYW"]}`,
			output: testData{Keys: []AuthorizedKey{
				{Options: []string{`no-pty`, `command="/path/to/script.sh"`},
					PublicKey: parsePublicKey(key)}}}},
		{name: `invalid`,
			input:  `{"keys":[" test comment"]}`,
			output: testData{Keys: []AuthorizedKey{{}}},
			err:    errors.New("")},
		{name: `invalid broken json`,
			input: `{"keys":["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEY5T2JQgiL5Z5Yuy4yXuUYglVJlpsokHFXR1hvnCVYW test comment}]`,
			err:   errors.New("")},
		{name: `invalid empty key`,
			input:  `{"keys":[1]}`,
			output: testData{Keys: []AuthorizedKey{{}}},
			err:    errors.New(AuthorizedKey_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output testData
			err := json.Unmarshal([]byte(test.input), &output)
			// output, err := json.Marshal(test.input)
			require.Equal(t, test.output, output)
			if test.err == nil {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, test.err.Error())
			}
		})
	}
}

func Test_sshKeyUrlDecode(t *testing.T) {
	rawOutput := test_data_qemu.PublicKey_Decoded_Output()
	input := test_data_qemu.PublicKey_Encoded_Input()
	output := make([]AuthorizedKey, len(rawOutput))
	for i := range rawOutput {
		output[i] = AuthorizedKey{Options: rawOutput[i].Options, PublicKey: rawOutput[i].PublicKey, Comment: rawOutput[i].Comment}
	}
	require.Equal(t, output, sshKeyUrlDecode(input))
}

// Test the encoding logic to encode the ssh keys
func Test_sshKeyUrlEncode(t *testing.T) {
	rawInput := test_data_qemu.PublicKey_Decoded_Input()
	input := make([]AuthorizedKey, len(rawInput))
	for i := range rawInput {
		input[i] = AuthorizedKey{Options: rawInput[i].Options, PublicKey: rawInput[i].PublicKey, Comment: rawInput[i].Comment}
	}

	tests := []struct {
		name   string
		input  []AuthorizedKey
		output string
	}{
		{name: `multiple keys`,
			input:  input,
			output: test_data_qemu.PublicKey_Encoded_Output()},
		{name: `empty key`,
			input:  []AuthorizedKey{{}},
			output: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, sshKeyUrlEncode(test.input))
		})
	}
}

func Benchmark_sshKeyUrlDecode(b *testing.B) {
	input := test_data_qemu.PublicKey_Encoded_Input()

	b.ResetTimer() // Reset timer to exclude setup time
	for i := 0; i < b.N; i++ {
		_ = sshKeyUrlDecode(input)
	}
}

func Benchmark_sshKeyUrlEncode(b *testing.B) {
	rawInput := test_data_qemu.PublicKey_Decoded_Input()
	input := make([]AuthorizedKey, len(rawInput))
	for i := range rawInput {
		input[i] = AuthorizedKey{PublicKey: rawInput[i].PublicKey, Comment: rawInput[i].Comment}
	}

	b.ResetTimer() // Reset timer to exclude setup time
	for i := 0; i < b.N; i++ {
		_ = sshKeyUrlEncode(input)
	}
}
