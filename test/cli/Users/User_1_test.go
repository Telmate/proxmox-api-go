package cli_user_test

import (
	"testing"
	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
)

func Test_User_1_Cleanup(t *testing.T){
	Test := cliTest.Test{
		ReqErr: true,
		ErrContains: "test-user1@pve",
		Args: []string{"-i","delete","user","test-user1@pve"},
	}
	Test.StandardTest(t)
}

func Test_User_1_Set_Empty_Without_Password(t *testing.T){
	Test := cliTest.Test{
		InputJson: `
{
	"enable": false,
	"expire": 0
}`,
		Expected: "(test-user1@pve)",
		Contains: true,
		Args: []string{"-i","set","user","test-user1@pve"},
	}
	Test.StandardTest(t)
}

// Test Login (error)

func Test_User_1_Get_Empty(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: `
{
	"userid": "test-user1@pve",
	"enable": false,
	"expire": 0
}`,
		Args: []string{"-i","get","user","test-user1@pve"},
	}
	Test.StandardTest(t)
}

func Test_User_1_Set_Full_With_Password(t *testing.T){
	Test := cliTest.Test{
		InputJson: `
{
	"comment": "this is a comment",
	"email": "b.wayne@proxmox.com",
	"enable": true,
	"expire": 99999999,
	"firstname": "Bruce",
	"lastname": "Wayne",
	"groups": [
	],
	"keys": "2fa key"
}`,
		Expected: "(test-user1@pve)",
		Contains: true,
		Args: []string{"-i","set","user","test-user1@pve","Enter123!"},
	}
	Test.StandardTest(t)
}

// Test Login (no error)

func Test_User_1_Get_Full(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: `
{
	"comment": "this is a comment",
	"userid": "test-user1@pve",
	"email": "b.wayne@proxmox.com",
	"enable": true,
	"expire": 99999999,
	"firstname": "Bruce",
	"lastname": "Wayne",
	"keys": "2fa key"
}`,
		Args: []string{"-i","get","user","test-user1@pve"},
	}
	Test.StandardTest(t)
}

func Test_User_1_Delete(t *testing.T){
	Test := cliTest.Test{
		Expected: "",
		ReqErr: false,
		Args: []string{"-i","delete","user","test-user1@pve"},
	}
	Test.StandardTest(t)
}