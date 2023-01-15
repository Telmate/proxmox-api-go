package cli_user_test

import (
	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
	"testing"
)

func Test_User_List(t *testing.T) {
	Test := cliTest.Test{
		Expected: `"userid":"root@pam"`,
		ReqErr:   false,
		Contains: true,
		Args:     []string{"-i", "list", "users"},
	}
	Test.StandardTest(t)
}
