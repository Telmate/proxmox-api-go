package cli_storage_test

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
	storagesubtests "github.com/Telmate/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_SMB_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("smb-test-1", t)
}

func Test_Storage_SMB_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.SMBEmpty)
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(smb-test-1)",
		Contains:  true,
		Args:      []string{"-i", "create", "storage", "smb-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_SMB_1_Get_Empty(t *testing.T) {
	storagesubtests.SMBGetEmpty("smb-test-1", t)
}

func Test_Storage_SMB_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.SMBFull)
	s.SMB.Password = proxmox.PointerString("Enter123!")
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(smb-test-1)",
		Contains:  true,
		Args:      []string{"-i", "update", "storage", "smb-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_SMB_1_Get_Full(t *testing.T) {
	storagesubtests.SMBGetFull("smb-test-1", t)
}

func Test_Storage_SMB_1_Delete(t *testing.T) {
	storagesubtests.Delete("smb-test-1", t)
}
