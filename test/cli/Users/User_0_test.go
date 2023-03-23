package cli_user_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
	"github.com/perimeter-81/proxmox-api-go/test/cli/Group/group_sub_tests"
	"github.com/perimeter-81/proxmox-api-go/test/cli/Users/user_sub_tests"
	"github.com/perimeter-81/proxmox-api-go/test/data/test_data_cli"
)

func Test_User_0_Cleanup(t *testing.T) {
	user_sub_tests.Cleanup(t, proxmox.UserID{Name: "test-user0", Realm: "pve"})
	group_sub_tests.Cleanup(t, proxmox.GroupName("user0-group0"))
}

func Test_User_0_Set_Group(t *testing.T) {
	Test := cliTest.Test{
		Contains: []string{"(user0-group0)"},
		Args:     []string{"-i", "set", "group", "user0-group0"},
	}
	Test.StandardTest(t)
}

func Test_User_0_Set_Full_With_Password_Set(t *testing.T) {
	Test := cliTest.Test{
		InputJson: test_data_cli.User_Full_testData(0),
		Contains:  []string{"(test-user0@pve)"},
		Args:      []string{"-i", "set", "user", "test-user0@pve", "Enter123!"},
	}
	Test.StandardTest(t)
}

func Test_User_0_Login_Password_Set(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.LoginTest{
		UserID:   "test-user0@pve",
		Password: "Enter123!",
		ReqErr:   false,
	}
	Test.Login(t)
}

func Test_User_0_Change_Password(t *testing.T) {
	Test := cliTest.Test{
		Contains: []string{"(test-user0@pve)"},
		Args:     []string{"-i", "set", "user", "test-user0@pve", "aBc123!"},
	}
	Test.StandardTest(t)
}

func Test_User_0_Login_Password_Changed(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.LoginTest{
		UserID:   "test-user0@pve",
		Password: "aBc123!",
		ReqErr:   false,
	}
	Test.Login(t)
}

func Test_User_0_Get_Full(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: test_data_cli.User_Full_testData(0),
		Args:       []string{"-i", "get", "user", "test-user0@pve"},
	}
	Test.StandardTest(t)
}

func Test_User_0_Set_Empty(t *testing.T) {
	user_sub_tests.Set(t, test_data_cli.User_Empty_testData(0))
}

func Test_User_0_Get_Empty(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: test_data_cli.User_Empty_testData(0),
		Args:       []string{"-i", "get", "user", "test-user0@pve"},
	}
	Test.StandardTest(t)
}

func Test_User_0_Delete(t *testing.T) {
	user_sub_tests.Delete(t, proxmox.UserID{Name: "test-user0", Realm: "pve"})
	group_sub_tests.Delete(t, proxmox.GroupName("user0-group0"))
}
