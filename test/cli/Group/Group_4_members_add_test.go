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
// add users to group
// check users in group
// Delete group and users

func Test_Group_4_Cleanup(t *testing.T) {
	group_sub_tests.Cleanup(t, "group4")
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "group4-user40", Realm: "pve"})
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "group4-user41", Realm: "pve"})
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "group4-user42", Realm: "pve"})
}

func Test_Group_4_Create_Users(t *testing.T) {
	user_sub_tests.Set(t, proxmox.ConfigUser{User: proxmox.UserID{Name: "group4-user40", Realm: "pve"}})
	user_sub_tests.Set(t, proxmox.ConfigUser{User: proxmox.UserID{Name: "group4-user41", Realm: "pve"}})
	user_sub_tests.Set(t, proxmox.ConfigUser{User: proxmox.UserID{Name: "group4-user42", Realm: "pve"}})
}

func Test_Group_4_Create_Group(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group4)"},
		Args:     []string{"-i", "set", "group", "group4", "--members=group4-user40@pve"},
	}
	Test.StandardTest(t)
}

func Test_Group_4_Get_Members_Partial(t *testing.T) {
	group_sub_tests.Get(t, proxmox.ConfigGroup{
		Name: "group4",
		Members: &[]proxmox.UserID{
			{Name: "group4-user40", Realm: "pve"},
		},
	})
}

func Test_Group_4_Add_Members(t *testing.T) {
	Test := &cliTest.Test{
		Contains: []string{"(group4)"},
		Args:     []string{"-i", "member", "group", "add", "group4", "group4-user41@pve,group4-user42@pve"},
	}
	Test.StandardTest(t)
}

func Test_Group_4_Get_Members_Full(t *testing.T) {
	group_sub_tests.Get(t, proxmox.ConfigGroup{
		Name: "group4",
		Members: &[]proxmox.UserID{
			{Name: "group4-user40", Realm: "pve"},
			{Name: "group4-user41", Realm: "pve"},
			{Name: "group4-user42", Realm: "pve"},
		},
	})
}

func Test_Group_4_Delete(t *testing.T) {
	group_sub_tests.Delete(t, "group4")
	user_sub_tests.Delete(t, proxmox.UserID{Name: "group4-user40", Realm: "pve"})
	user_sub_tests.Delete(t, proxmox.UserID{Name: "group4-user41", Realm: "pve"})
	user_sub_tests.Delete(t, proxmox.UserID{Name: "group4-user42", Realm: "pve"})
}
