package cli_storage_test

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
	storagesubtests "github.com/Telmate/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_ISCSI_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("iscsi-test-1", t)
}

func Test_Storage_ISCSI_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.IscsiEmpty)
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(iscsi-test-1)",
		Contains:  true,
		Args:      []string{"-i", "create", "storage", "iscsi-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_ISCSI_1_Get_Empty(t *testing.T) {
	storagesubtests.IscsiGetEmpty("iscsi-test-1", t)
}

func Test_Storage_ISCSI_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.IscsiFull)
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(iscsi-test-1)",
		Contains:  true,
		Args:      []string{"-i", "update", "storage", "iscsi-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_ISCSI_1_Get_Full(t *testing.T) {
	storagesubtests.IscsiGetFull("iscsi-test-1", t)
}

func Test_Storage_ISCSI_1_Delete(t *testing.T) {
	storagesubtests.Delete("iscsi-test-1", t)
}
