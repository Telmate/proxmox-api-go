package cli_user_test

// Create group and add user to it.
// list users without group information.
// list users with group information.
// delete group

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
	"github.com/perimeter-81/proxmox-api-go/test/cli/Group/group_sub_tests"
)

func Test_User_2_Cleanup(t *testing.T) {
	group_sub_tests.Cleanup(t, "user2-group")
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
	group_sub_tests.Delete(t, "user2-group")
}
