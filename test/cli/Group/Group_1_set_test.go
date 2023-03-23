package group_test

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
	"github.com/perimeter-81/proxmox-api-go/test/cli/Group/group_sub_tests"
	"github.com/perimeter-81/proxmox-api-go/test/cli/Users/user_sub_tests"
	"github.com/perimeter-81/proxmox-api-go/test/data/test_data_cli"
)

// Create group with minimal option populated
// Check empty
// Update group with --members not defined (should not update memberships)
// Check no changes
// Update group with all options populated
// Check if populated
// Delete items
// Check items deleted

func Test_Group_1_Cleanup(t *testing.T) {
	group_sub_tests.Cleanup(t, "group1")
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "group1-user10", Realm: "pve"})
}

func Test_Group_1_Create_User(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"group1-user10@pve"},
		Args:     []string{"-i", "set", "user", "group1-user10@pve"},
	}
	Test.StandardTest(t)
}

func Test_Group_1_Set_Empty_Create(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group1)"},
		Args:     []string{"-i", "set", "group", "group1"},
	}
	Test.StandardTest(t)
}

func Test_Group_1_Get_Empty_0(t *testing.T) {
	group_sub_tests.Get(t, test_data_cli.Group_Get_Empty_testData(1))
}

func Test_Group_1_Set_MembersNotDefined_Update(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group1)"},
		Args:     []string{"-i", "set", "group", "group1"},
	}
	Test.StandardTest(t)
}

func Test_Group_1_Get_Empty_1(t *testing.T) {
	group_sub_tests.Get(t, test_data_cli.Group_Get_Empty_testData(1))
}

func Test_Group_1_Set_Full_Update(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group1)"},
		Args:     []string{"-i", "set", "group", "group1", "comment", "--members=root@pam,group1-user10@pve,group1-user11@pve"},
	}
	Test.StandardTest(t)
}

func Test_Group_1_Get_Full(t *testing.T) {
	Test := &cliTest.Test{
		NotContains: []string{"group1-user11"},
		Args:        []string{"-i", "get", "group", "group1"},
	}
	out := Test.StandardTest(t)
	group_sub_tests.Get_Test(t, test_data_cli.Group_Get_Full_testData(1), out)
}

func Test_Group_1_Delete_Group(t *testing.T) {
	group_sub_tests.Delete(t, "group1")
}

func Test_Group_1_Delete_User(t *testing.T) {
	user_sub_tests.Delete(t, proxmox.UserID{Name: "group1-user10", Realm: "pve"})
}

func Test_Group_1_List_Group_NotExistent(t *testing.T) {
	Test := &cliTest.Test{
		NotContains: []string{"group1"},
		Args:        []string{"-i", "list", "groups"},
	}
	Test.StandardTest(t)
}

func Test_Group_1_List_User_NotExistent(t *testing.T) {
	Test := &cliTest.Test{
		NotContains: []string{"group1-user10@pve"},
		Args:        []string{"-i", "list", "users"},
	}
	Test.StandardTest(t)
}
