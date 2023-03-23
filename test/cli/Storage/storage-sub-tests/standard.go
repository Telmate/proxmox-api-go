package storagesubtests

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
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
		Contains: []string{name},
		Args:     []string{"-i", "delete", "storage", name},
	}
	Test.StandardTest(t)
}

func Get(s *proxmox.ConfigStorage, name string, t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", name},
	}
	Test.StandardTest(t)
}

func Create(s *proxmox.ConfigStorage, name string, t *testing.T) {
	createOrUpdate(s, name, "create", t)
}

func Update(s *proxmox.ConfigStorage, name string, t *testing.T) {
	createOrUpdate(s, name, "update", t)
}

func createOrUpdate(s *proxmox.ConfigStorage, name, command string, t *testing.T) {
	Test := cliTest.Test{
		InputJson: InlineMarshal(s),
		Contains:  []string{"(" + name + ")"},
		Args:      []string{"-i", command, "storage", name},
	}
	Test.StandardTest(t)
}
