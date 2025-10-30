package acmeplugin

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
)

func Test_AcmePlugin_0(t *testing.T) {
	t.Run("0 ensure plugin is removed", func(t *testing.T) {
		cliTest.Test{Args: []string{"-i", "delete", "acmeplugin", "test-0"}}.StandardTest(t)
	})
	t.Run("1 create", func(t *testing.T) {
		cliTest.Test{
			InputJson: `
{
	"api": "aws",
	"data": "AWS_ACCESS_KEY_ID=DEMOACCESSKEYID\nAWS_SECRET_ACCESS_KEY=DEMOSECRETACCESSKEY\n",
	"enable": true,
	"validation-delay": 30
}`,
			Contains: []string{"(test-0)"},
			Args:     []string{"-i", "set", "acmeplugin", "test-0"},
		}.StandardTest(t)
	})
	t.Run("2 check exists", func(t *testing.T) {
		cliTest.Test{
			Contains: []string{"test-0"},
			Args:     []string{"-i", "list", "acmeplugins"},
		}.StandardTest(t)
	})
	t.Run("3 delete", func(t *testing.T) {
		cliTest.Test{Args: []string{"-i", "delete", "acmeplugin", "test-0"}}.StandardTest(t)
	})
	t.Run("4 check removed", func(t *testing.T) {
		cliTest.Test{
			NotContains: []string{"test-0"},
			Args:        []string{"-i", "list", "acmeplugins"},
		}.StandardTest(t)
	})
}
