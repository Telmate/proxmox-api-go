package storagesubtests

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
)

func Cleanup(name string, t *testing.T) {
	Test := cliTest.Test{
		ReqErr:      true,
		ErrContains: name,
		Args:        []string{"-i", "delete", "storage", name},
	}
	Test.StandardTest(t)
}

func Delete(name string, t *testing.T) {
	Test := cliTest.Test{
		Expected: name,
		Contains: true,
		ReqErr:   false,
		Args:     []string{"-i", "delete", "storage", name},
	}
	Test.StandardTest(t)
}
