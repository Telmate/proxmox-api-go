package cli_acmeaccount_test

import (
	"testing"

	_ "github.com/Bluearchive/proxmox-api-go/cli/command/commands"
	cliTest "github.com/Bluearchive/proxmox-api-go/test/cli"
)

func Test_AcmeAccount_0_Cleanup(t *testing.T) {
	Test := cliTest.Test{
		ReqErr:      true,
		ErrContains: "test-0",
		Args:        []string{"-i", "delete", "acmeaccount", "test-0"},
	}
	Test.StandardTest(t)
}

func Test_AcmeAccount_0_Set(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"contact": [
		"a@nonexistantdomain.com",
		"b@nonexistantdomain.com",
		"c@nonexistantdomain.com",
		"d@nonexistantdomain.com"
	],
	"directory": "https://acme-staging-v02.api.letsencrypt.org/directory",
	"tos": true
}`,
		Contains: []string{"(test-0)"},
		Args:     []string{"-i", "create", "acmeaccount", "test-0"},
	}
	Test.StandardTest(t)
}

func Test_AcmeAccount_0_Get(t *testing.T) {
	Test := cliTest.Test{
		OutputJson: `
{
	"name": "test-0",
	"contact": [
		"a@nonexistantdomain.com",
		"b@nonexistantdomain.com",
		"c@nonexistantdomain.com",
		"d@nonexistantdomain.com"
	],
	"directory": "https://acme-staging-v02.api.letsencrypt.org/directory",
	"tos": true
}`,
		Args: []string{"-i", "get", "acmeaccount", "test-0"},
	}
	Test.StandardTest(t)
}

func Test_AcmeAccount_0_Delete(t *testing.T) {
	Test := cliTest.Test{
		Expected: "",
		ReqErr:   false,
		Args:     []string{"-i", "delete", "acmeaccount", "test-0"},
	}
	Test.StandardTest(t)
}
