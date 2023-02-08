package cli_user_test

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_cli"
)

func Test_User_1_Cleanup(t *testing.T) {
	Test := cliTest.Test{
		ReqErr:      true,
		ErrContains: "test-user1@pve",
		Args:        []string{"-i", "delete", "user", "test-user1@pve"},
	}
	Test.StandardTest(t)
}

func Test_User_1_Set_Empty_Without_Password(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"enable": false,
	"expire": 0
}`,
		Contains: []string{"(test-user1@pve)"},
		Args:     []string{"-i", "set", "user", "test-user1@pve"},
	}
	Test.StandardTest(t)
}

func Test_User_1_Login_Password_Not_Set(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.LoginTest{
		UserID:   "test-user1@pve",
		Password: "Enter123!",
		ReqErr:   true,
	}
	Test.Login(t)
}

func Test_User_1_Get_Empty(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: test_data_cli.User_Empty_testData(1),
		Args:       []string{"-i", "get", "user", "test-user1@pve"},
	}
	Test.StandardTest(t)
}

func Test_User_1_Set_Full_With_Password(t *testing.T) {
	Test := cliTest.Test{
		InputJson: test_data_cli.User_Full_testData(1),
		Contains:  []string{"(test-user1@pve)"},
		Args:      []string{"-i", "set", "user", "test-user1@pve", "Enter123!"},
	}
	Test.StandardTest(t)
}

func Test_User_1_Login_Password_Set(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.LoginTest{
		UserID:   "test-user1@pve",
		Password: "Enter123!",
		ReqErr:   false,
	}
	Test.Login(t)
}

func Test_User_1_Get_Full(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: test_data_cli.User_Full_testData(1),
		Args:       []string{"-i", "get", "user", "test-user1@pve"},
	}
	Test.StandardTest(t)
}

func Test_User_1_Delete(t *testing.T) {
	Test := cliTest.Test{
		Expected: "",
		ReqErr:   false,
		Args:     []string{"-i", "delete", "user", "test-user1@pve"},
	}
	Test.StandardTest(t)
}
