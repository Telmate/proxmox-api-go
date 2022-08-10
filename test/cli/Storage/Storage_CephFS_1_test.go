package cli_storage_test

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
	storagesubtests "github.com/Telmate/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_CephFS_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("cephfs-test-1", t)
}

func Test_Storage_CephFS_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.CephfsEmpty)
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(cephfs-test-1)",
		Contains:  true,
		Args:      []string{"-i", "create", "storage", "cephfs-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_CephFS_1_Get_Empty(t *testing.T) {
	storagesubtests.CephfsGetEmpty("cephfs-test-1", t)
}

func Test_Storage_CephFS_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.CephfsFull)
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(cephfs-test-1)",
		Contains:  true,
		Args:      []string{"-i", "update", "storage", "cephfs-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_CephFS_1_Get_Full(t *testing.T) {
	storagesubtests.CephfsGetFull("cephfs-test-1", t)
}

func Test_Storage_CephFS_1_Delete(t *testing.T) {
	storagesubtests.Delete("cephfs-test-1", t)
}
