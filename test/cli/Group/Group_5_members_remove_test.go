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
// remove some users from group
// check some users in group
// remove more users from group (empty)
// check no users in group
// Delete group and users

func Test_Group_5_Cleanup(t *testing.T) {
	group_sub_tests.Cleanup(t, "group5")
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "group5-user50", Realm: "pve"})
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "group5-user51", Realm: "pve"})
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "group5-user52", Realm: "pve"})
}

func Test_Group_5_Create_Users(t *testing.T) {
	user_sub_tests.Set(t, proxmox.ConfigUser{User: proxmox.UserID{Name: "group5-user50", Realm: "pve"}})
	user_sub_tests.Set(t, proxmox.ConfigUser{User: proxmox.UserID{Name: "group5-user51", Realm: "pve"}})
	user_sub_tests.Set(t, proxmox.ConfigUser{User: proxmox.UserID{Name: "group5-user52", Realm: "pve"}})
}

func Test_Group_5_Create_Group(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group5)"},
		Args:     []string{"-i", "set", "group", "group5", "--members=group5-user50@pve,group5-user51@pve,group5-user52@pve"},
	}
	Test.StandardTest(t)
}

func Test_Group_5_Get_Members_Full(t *testing.T) {
	group_sub_tests.Get(t, proxmox.ConfigGroup{
		Name: "group5",
		Members: &[]proxmox.UserID{
			{Name: "group5-user50", Realm: "pve"},
			{Name: "group5-user51", Realm: "pve"},
			{Name: "group5-user52", Realm: "pve"},
		},
	})
}

func Test_Group_5_Remove_Members_Partial(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group5)"},
		Args:     []string{"-i", "member", "group", "remove", "group5", "group5-user50@pve"},
	}
	Test.StandardTest(t)
}

func Test_Group_5_Get_Members_Partial(t *testing.T) {
	group_sub_tests.Get(t, proxmox.ConfigGroup{
		Name: "group5",
		Members: &[]proxmox.UserID{
			{Name: "group5-user51", Realm: "pve"},
			{Name: "group5-user52", Realm: "pve"},
		},
	})
}

func Test_Group_5_Remove_Members_Empty(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group5)"},
		Args:     []string{"-i", "member", "group", "remove", "group5", "group5-user51@pve,group5-user52@pve"},
	}
	Test.StandardTest(t)
}

func Test_Group_5_Get_Members_Empty(t *testing.T) {
	group_sub_tests.Get(t, proxmox.ConfigGroup{
		Name:    "group5",
		Members: &[]proxmox.UserID{},
	})
}

func Test_Group_5_Delete(t *testing.T) {
	group_sub_tests.Delete(t, "group5")
	user_sub_tests.Delete(t, proxmox.UserID{Name: "group5-user50", Realm: "pve"})
	user_sub_tests.Delete(t, proxmox.UserID{Name: "group5-user51", Realm: "pve"})
	user_sub_tests.Delete(t, proxmox.UserID{Name: "group5-user52", Realm: "pve"})
}
