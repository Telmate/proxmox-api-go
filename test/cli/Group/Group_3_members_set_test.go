package group_test

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
	"github.com/perimeter-81/proxmox-api-go/test/cli/Group/group_sub_tests"
	"github.com/perimeter-81/proxmox-api-go/test/cli/Users/user_sub_tests"
)

// create users
// create group and add some users to group
// check users in group and not in group
// set group membership (add and remove)
// check users in group and not in group
// set group membership (add and remove)
// check users in group and not in group
// set group membership (clear)
// check no users in group
// set group membership (Full)
// check users in group
// Delete group and users

func Test_Group_3_Cleanup(t *testing.T) {
	group_sub_tests.Cleanup(t, "group3")
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "group3-user30", Realm: "pve"})
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "group3-user31", Realm: "pve"})
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "group3-user32", Realm: "pve"})
}

func Test_Group_3_Create_Users(t *testing.T) {
	user_sub_tests.Set(t, proxmox.ConfigUser{User: proxmox.UserID{Name: "group3-user30", Realm: "pve"}})
	user_sub_tests.Set(t, proxmox.ConfigUser{User: proxmox.UserID{Name: "group3-user31", Realm: "pve"}})
	user_sub_tests.Set(t, proxmox.ConfigUser{User: proxmox.UserID{Name: "group3-user32", Realm: "pve"}})
}

func Test_Group_3_Create_Group(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group3)"},
		Args:     []string{"-i", "set", "group", "group3", "--members=group3-user31@pve,group3-user32@pve"},
	}
	Test.StandardTest(t)
}

func Test_Group_3_Get_Members_Partial_0(t *testing.T) {
	group_sub_tests.Get(t, proxmox.ConfigGroup{
		Name: "group3",
		Members: &[]proxmox.UserID{
			{Name: "group3-user31", Realm: "pve"},
			{Name: "group3-user32", Realm: "pve"},
		},
	})
}

func Test_Group_3_Set_Members_Partial_0(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group3)"},
		Args:     []string{"-i", "member", "group", "set", "group3", "group3-user30@pve"},
	}
	Test.StandardTest(t)
}

func Test_Group_3_Get_Members_Partial_1(t *testing.T) {
	group_sub_tests.Get(t, proxmox.ConfigGroup{
		Name: "group3",
		Members: &[]proxmox.UserID{
			{Name: "group3-user30", Realm: "pve"},
		},
	})
}

func Test_Group_3_Set_Members_Partial_1(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group3)"},
		Args:     []string{"-i", "member", "group", "set", "group3", "group3-user31@pve,group3-user32@pve"},
	}
	Test.StandardTest(t)
}

func Test_Group_3_Get_Members_Partial_2(t *testing.T) {
	group_sub_tests.Get(t, proxmox.ConfigGroup{
		Name: "group3",
		Members: &[]proxmox.UserID{
			{Name: "group3-user31", Realm: "pve"},
			{Name: "group3-user32", Realm: "pve"},
		},
	})
}

func Test_Group_3_Set_Members_Empty(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group3)"},
		Args:     []string{"-i", "member", "group", "set", "group3"},
	}
	Test.StandardTest(t)
}

func Test_Group_3_Get_Members_Empty(t *testing.T) {
	group_sub_tests.Get(t, proxmox.ConfigGroup{
		Name:    "group3",
		Members: &[]proxmox.UserID{},
	})
}

func Test_Group_3_Set_Members_Full(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group3)"},
		Args:     []string{"-i", "member", "group", "set", "group3", "group3-user30@pve,group3-user31@pve,group3-user32@pve"},
	}
	Test.StandardTest(t)
}

func Test_Group_3_Get_Members_Full(t *testing.T) {
	group_sub_tests.Get(t, proxmox.ConfigGroup{
		Name: "group3",
		Members: &[]proxmox.UserID{
			{Name: "group3-user30", Realm: "pve"},
			{Name: "group3-user31", Realm: "pve"},
			{Name: "group3-user32", Realm: "pve"},
		},
	})
}

func Test_Group_3_Delete(t *testing.T) {
	group_sub_tests.Delete(t, "group3")
	user_sub_tests.Delete(t, proxmox.UserID{Name: "group3-user30", Realm: "pve"})
	user_sub_tests.Delete(t, proxmox.UserID{Name: "group3-user31", Realm: "pve"})
	user_sub_tests.Delete(t, proxmox.UserID{Name: "group3-user32", Realm: "pve"})
}
