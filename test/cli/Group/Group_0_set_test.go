package group_test

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
	"github.com/perimeter-81/proxmox-api-go/test/cli/Group/group_sub_tests"
	"github.com/perimeter-81/proxmox-api-go/test/cli/Users/user_sub_tests"
	"github.com/perimeter-81/proxmox-api-go/test/data/test_data_cli"
)

// Create group with all option populated
// Check if populated
// Update group with --members not defined (should not update memberships)
// Check no changes
// Update group with all options empty
// Check empty
// Delete items
// Check items deleted

func Test_Group_0_Cleanup(t *testing.T) {
	group_sub_tests.Cleanup(t, "group0")
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "group0-user00", Realm: "pve"})
}

func Test_Group_0_Create_User(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"group0-user00@pve"},
		Args:     []string{"-i", "set", "user", "group0-user00@pve"},
	}
	Test.StandardTest(t)
}

func Test_Group_0_Set_Full_Create(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group0)"},
		Args:     []string{"-i", "set", "group", "group0", "comment", "--members=root@pam,group0-user00@pve,group0-user01@pve"},
	}
	Test.StandardTest(t)
}

func Test_Group_0_Get_Full_0(t *testing.T) {
	Test := &cliTest.Test{
		NotContains: []string{"group0-user01"},
		Args:        []string{"-i", "get", "group", "group0"},
	}
	out := Test.StandardTest(t)
	group_sub_tests.Get_Test(t, test_data_cli.Group_Get_Full_testData(0), out)
}

func Test_Group_0_Set_MembersNotDefined_Update(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group0)"},
		Args:     []string{"-i", "set", "group", "group0", "comment"},
	}
	Test.StandardTest(t)
}

func Test_Group_0_Get_Full_1(t *testing.T) {
	group_sub_tests.Get(t, test_data_cli.Group_Get_Full_testData(0))
}

func Test_Group_0_Set_Empty_Create(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group0)"},
		Args:     []string{"-i", "set", "group", "group0", "--members="},
	}
	Test.StandardTest(t)
}

func Test_Group_0_Get_Empty(t *testing.T) {
	group_sub_tests.Get(t, test_data_cli.Group_Get_Empty_testData(0))
}

func Test_Group_0_Delete_Group(t *testing.T) {
	group_sub_tests.Delete(t, "group0")
}

func Test_Group_0_Delete_User(t *testing.T) {
	user_sub_tests.Delete(t, proxmox.UserID{Name: "group0-user00", Realm: "pve"})
}

func Test_Group_0_List_Group_NotExistent(t *testing.T) {
	Test := &cliTest.Test{
		NotContains: []string{"group0"},
		Args:        []string{"-i", "list", "groups"},
	}
	Test.StandardTest(t)
}

func Test_Group_0_List_User_NotExistent(t *testing.T) {
	Test := &cliTest.Test{
		NotContains: []string{"group0-user00@pve"},
		Args:        []string{"-i", "list", "users"},
	}
	Test.StandardTest(t)
}
