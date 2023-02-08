package cli_user_test

// Create group and add user to it.
// list users without group information.
// list users with group information.
// delete group

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
)

func Test_User_2_Cleanup(t *testing.T) {
	// remove group
	Test := &cliTest.Test{
		ReqErr:      true,
		ErrContains: "user2-group",
		Args:        []string{"-i", "delete", "group", "user2-group"},
	}
	Test.StandardTest(t)
}

func Test_User_2_Set_Group(t *testing.T) {
	Test := cliTest.Test{
		Contains: []string{"(user2-group)"},
		Args:     []string{"-i", "set", "group", "user2-group", "--members=root@pam"},
	}
	Test.StandardTest(t)
}

func Test_User_2_List_Without_Group(t *testing.T) {
	Test := cliTest.Test{
		Contains:    []string{`"name":"root","realm":"pam"`},
		NotContains: []string{"user2-group"},
		Args:        []string{"-i", "list", "users"},
	}
	Test.StandardTest(t)
}

func Test_User_2_List_With_Group(t *testing.T) {
	Test := cliTest.Test{
		Contains: []string{`"name":"root","realm":"pam"`, "user2-group"},
		Args:     []string{"-i", "list", "users", "--groups"},
	}
	Test.StandardTest(t)
}

func Test_User_2_Group_Delete(t *testing.T) {
	Test := &cliTest.Test{
		Args: []string{"-i", "delete", "group", "user2-group"},
	}
	Test.StandardTest(t)
}
