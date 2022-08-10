package cli_storage_test

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
	storagesubtests "github.com/Telmate/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_NFS_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("nfs-test-1", t)
}

func Test_Storage_NFS_1_Create_Empty(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := storagesubtests.CloneJson(storagesubtests.NFSEmpty)
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(nfs-test-1)",
		Contains:  true,
		Args:      []string{"-i", "create", "storage", "nfs-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_NFS_1_Get_Empty(t *testing.T) {
	storagesubtests.NFSGetEmpty("nfs-test-1", t)
}

func Test_Storage_NFS_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.NFSFull)
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(nfs-test-1)",
		Contains:  true,
		Args:      []string{"-i", "update", "storage", "nfs-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_NFS_1_Get_Full(t *testing.T) {
	storagesubtests.NFSGetFull("nfs-test-1", t)
}

func Test_Storage_NFS_1_Delete(t *testing.T) {
	storagesubtests.Delete("nfs-test-1", t)
}
