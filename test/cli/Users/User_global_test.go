package cli_user_test

import (
	"testing"
	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
)

func Test_User_List(t *testing.T){
	Test := cliTest.Test{
		Expected: `"userid":"root@pam"`,
		ReqErr: false,
		Contains: true,
		Args: []string{"-i","list","users"},
	}
	Test.StandardTest(t)
}