package cli_user_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
)

func Test_User_0_Cleanup(t *testing.T) {
	Test := cliTest.Test{
		ReqErr:      true,
		ErrContains: "test-user0@pve",
		Args:        []string{"-i", "delete", "user", "test-user0@pve"},
	}
	Test.StandardTest(t)
}

// Set groups

func Test_User_0_Set_Full_With_Password_Set(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"comment": "this is a comment",
	"email": "b.wayne@proxmox.com",
	"enable": true,
	"expire": 253370811600,
	"firstname": "Bruce",
	"lastname": "Wayne",
	"groups": [
	],
	"keys": "2fa key"
}`,
		Expected: "(test-user0@pve)",
		Contains: true,
		Args:     []string{"-i", "set", "user", "test-user0@pve", "Enter123!"},
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
		Expected: "(test-user0@pve)",
		Contains: true,
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
		OutputJson: `
{
	"user": {
		"name": "test-user0",
		"realm": "pve"
	},
	"comment": "this is a comment",
	"email": "b.wayne@proxmox.com",
	"enable": true,
	"expire": 253370811600,
	"firstname": "Bruce",
	"keys": "2fa key",
	"lastname": "Wayne"
}`,
		Args: []string{"-i", "get", "user", "test-user0@pve"},
	}
	Test.StandardTest(t)
}

func Test_User_0_Set_Empty(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"comment": "",
	"email": "",
	"enable": false,
	"expire": 0,
	"firstname": "",
	"lastname": "",
	"groups": [
	],
	"keys": ""
}`,
		Expected: "(test-user0@pve)",
		Contains: true,
		Args:     []string{"-i", "set", "user", "test-user0@pve"},
	}
	Test.StandardTest(t)
}

func Test_User_0_Get_Empty(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: `
{
	"user": {
		"name": "test-user0",
		"realm": "pve"
	},
	"enable": false,
	"expire": 0
}`,
		Args: []string{"-i", "get", "user", "test-user0@pve"},
	}
	Test.StandardTest(t)
}

func Test_User_0_Delete(t *testing.T) {
	Test := cliTest.Test{
		Expected: "",
		ReqErr:   false,
		Args:     []string{"-i", "delete", "user", "test-user0@pve"},
	}
	Test.StandardTest(t)
}
