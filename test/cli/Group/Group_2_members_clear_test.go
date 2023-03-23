package group_test

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
	"github.com/perimeter-81/proxmox-api-go/test/cli/Group/group_sub_tests"
	"github.com/perimeter-81/proxmox-api-go/test/cli/Users/user_sub_tests"
)

// create users
// create group and add users to group
// check users in group
// clear group membership
// check no users in group
// Delete group and users

func Test_Group_2_Cleanup(t *testing.T) {
	group_sub_tests.Cleanup(t, "group2")
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "group2-user20", Realm: "pve"})
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "group2-user21", Realm: "pve"})
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "group2-user22", Realm: "pve"})
}

func Test_Group_2_Create_Users(t *testing.T) {
	user_sub_tests.Set(t, proxmox.ConfigUser{User: proxmox.UserID{Name: "group2-user20", Realm: "pve"}})
	user_sub_tests.Set(t, proxmox.ConfigUser{User: proxmox.UserID{Name: "group2-user21", Realm: "pve"}})
	user_sub_tests.Set(t, proxmox.ConfigUser{User: proxmox.UserID{Name: "group2-user22", Realm: "pve"}})
}

func Test_Group_2_Create_Group(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group2)"},
		Args:     []string{"-i", "set", "group", "group2", "--members=group2-user20@pve,group2-user21@pve,group2-user22@pve"},
	}
	Test.StandardTest(t)
}

func Test_Group_2_Get_Members_Full(t *testing.T) {
	group_sub_tests.Get(t, proxmox.ConfigGroup{
		Name: "group2",
		Members: &[]proxmox.UserID{
			{Name: "group2-user20", Realm: "pve"},
			{Name: "group2-user21", Realm: "pve"},
			{Name: "group2-user22", Realm: "pve"},
		},
	})
}

func Test_Group_2_Clear_Members(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group2)"},
		Args:     []string{"-i", "member", "group", "clear", "group2"},
	}
	Test.StandardTest(t)
}

func Test_Group_2_Get_Members_Empty(t *testing.T) {
	Test := &cliTest.Test{
		NotContains: []string{"group2-user20", "group2-user21", "group2-user22"},
		Args:        []string{"-i", "get", "group", "group2"},
	}
	group_sub_tests.Get_Test(t, proxmox.ConfigGroup{
		Name:    "group2",
		Members: &[]proxmox.UserID{},
	}, Test.StandardTest(t))
}

func Test_Group_2_Delete(t *testing.T) {
	group_sub_tests.Delete(t, "group2")
	user_sub_tests.Delete(t, proxmox.UserID{Name: "group2-user20", Realm: "pve"})
	user_sub_tests.Delete(t, proxmox.UserID{Name: "group2-user21", Realm: "pve"})
	user_sub_tests.Delete(t, proxmox.UserID{Name: "group2-user22", Realm: "pve"})
}
