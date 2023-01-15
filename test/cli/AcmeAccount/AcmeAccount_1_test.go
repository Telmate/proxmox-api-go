package cli_acmeaccount_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
)

func Test_AcmeAccount_1_Cleanup(t *testing.T) {
	Test := cliTest.Test{
		ReqErr:      true,
		ErrContains: "test-1",
		Args:        []string{"-i", "delete", "acmeaccount", "test-1"},
	}
	Test.StandardTest(t)
}

func Test_AcmeAccount_1_Set(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"contact": [
		"a@nonexistantdomain.com"
	],
	"directory": "https://acme-staging-v02.api.letsencrypt.org/directory",
	"tos": true
}`,
		Expected: "(test-1)",
		Contains: true,
		Args:     []string{"-i", "create", "acmeaccount", "test-1"},
	}
	Test.StandardTest(t)
}

func Test_AcmeAccount_1_Get(t *testing.T) {
	Test := cliTest.Test{
		OutputJson: `
{
	"name": "test-1",
	"contact": [
		"a@nonexistantdomain.com"
	],
	"directory": "https://acme-staging-v02.api.letsencrypt.org/directory",
	"tos": true
}`,
		Args: []string{"-i", "get", "acmeaccount", "test-1"},
	}
	Test.StandardTest(t)
}

func Test_AcmeAccount_1_Delete(t *testing.T) {
	Test := cliTest.Test{
		Expected: "",
		ReqErr:   false,
		Args:     []string{"-i", "delete", "acmeaccount", "test-1"},
	}
	Test.StandardTest(t)
}
