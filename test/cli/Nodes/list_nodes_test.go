package cli_node_test

import (
	"testing"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
)

func Test_List_Nodes(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	tests := []struct {
		name   string
		args  []string
		expected string
	}{{
		name: "List_User_root@pam",
		args: []string{"-i","list","nodes"},
		expected: `"id":"node/pve"`,
	}}
	
	for _, test := range tests {
		cliTest.ListTest(t,test.args,test.expected)
	}
}