package cli_storage_test

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
	storagesubtests "github.com/Telmate/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_RBD_0_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("rbd-test-0", t)
}

func Test_Storage_RBD_0_Create_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.RBDFull)
	s.RBD.Keyring = proxmox.PointerString("keyringplaceholder")
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(rbd-test-0)",
		Contains:  true,
		Args:      []string{"-i", "create", "storage", "rbd-test-0"},
	}
	Test.StandardTest(t)
}

func Test_Storage_RBD_0_Get_Full(t *testing.T) {
	storagesubtests.RBDGetFull("rbd-test-0", t)
}

func Test_Storage_RBD_0_Update_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.RBDEmpty)
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(rbd-test-0)",
		Contains:  true,
		Args:      []string{"-i", "update", "storage", "rbd-test-0"},
	}
	Test.StandardTest(t)
}

func Test_Storage_RBD_0_Get_Empty(t *testing.T) {
	storagesubtests.RBDGetEmpty("rbd-test-0", t)
}

func Test_Storage_RBD_0_Delete(t *testing.T) {
	storagesubtests.Delete("rbd-test-0", t)
}
