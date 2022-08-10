package cli_storage_test

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
	storagesubtests "github.com/Telmate/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_PBS_0_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("pbs-test-0", t)
}

func Test_Storage_PBS_0_Create_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.PBSFull)
	s.PBS.Password = proxmox.PointerString("Enter123!")
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(pbs-test-0)",
		Contains:  true,
		Args:      []string{"-i", "create", "storage", "pbs-test-0"},
	}
	Test.StandardTest(t)
}

func Test_Storage_PBS_0_Get_Full(t *testing.T) {
	storagesubtests.PBSGetFull("pbs-test-0", t)
}

func Test_Storage_PBS_0_Update_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.PBSEmpty)
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(pbs-test-0)",
		Contains:  true,
		Args:      []string{"-i", "update", "storage", "pbs-test-0"},
	}
	Test.StandardTest(t)
}

func Test_Storage_PBS_0_Get_Empty(t *testing.T) {
	storagesubtests.PBSGetEmpty("pbs-test-0", t)
}

func Test_Storage_PBS_0_Delete(t *testing.T) {
	storagesubtests.Delete("pbs-test-0", t)
}
