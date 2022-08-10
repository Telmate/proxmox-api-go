package cli_storage_test

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
	storagesubtests "github.com/Telmate/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_PBS_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("pbs-test-1", t)
}

func Test_Storage_PBS_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.PBSEmpty)
	s.PBS.Password = proxmox.PointerString("Enter123!")
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(pbs-test-1)",
		Contains:  true,
		Args:      []string{"-i", "create", "storage", "pbs-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_PBS_1_Get_Empty(t *testing.T) {
	storagesubtests.PBSGetEmpty("pbs-test-1", t)
}

func Test_Storage_PBS_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.PBSFull)
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(pbs-test-1)",
		Contains:  true,
		Args:      []string{"-i", "update", "storage", "pbs-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_PBS_1_Get_Full(t *testing.T) {
	storagesubtests.PBSGetFull("pbs-test-1", t)
}

func Test_Storage_PBS_1_Delete(t *testing.T) {
	storagesubtests.Delete("pbs-test-1", t)
}
