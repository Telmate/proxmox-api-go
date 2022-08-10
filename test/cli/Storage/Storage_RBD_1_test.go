package cli_storage_test

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
	storagesubtests "github.com/Telmate/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_RBD_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("rbd-test-1", t)
}

func Test_Storage_RBD_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.RBDEmpty)
	s.RBD.Keyring = proxmox.PointerString("keyringplaceholder")
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(rbd-test-1)",
		Contains:  true,
		Args:      []string{"-i", "create", "storage", "rbd-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_RBD_1_Get_Empty(t *testing.T) {
	storagesubtests.RBDGetEmpty("rbd-test-1", t)
}

func Test_Storage_RBD_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.RBDFull)
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(rbd-test-1)",
		Contains:  true,
		Args:      []string{"-i", "update", "storage", "rbd-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_RBD_1_Get_Full(t *testing.T) {
	storagesubtests.RBDGetFull("rbd-test-1", t)
}

func Test_Storage_RBD_1_Delete(t *testing.T) {
	storagesubtests.Delete("rbd-test-1", t)
}
