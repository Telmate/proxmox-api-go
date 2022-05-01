package cli_user_test

import (
	"testing"
	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
)

func Test_List_Users(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	tests := []struct {
		name   string
		args  []string
		expected string
	}{{
		name: "List_User_root@pam",
		args: []string{"-i","list","users"},
		expected: `"userid":"root@pam"`,
	}}
	
	for _, test := range tests {
		cliTest.ListTest(t,test.args,test.expected)
	}
}