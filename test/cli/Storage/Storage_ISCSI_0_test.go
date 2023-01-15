package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_ISCSI_0_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("iscsi-test-0", t)
}

func Test_Storage_ISCSI_0_Create_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.IscsiFull)
	storagesubtests.Create(s, "iscsi-test-0", t)
}

func Test_Storage_ISCSI_0_Get_Full(t *testing.T) {
	storagesubtests.IscsiGetFull("iscsi-test-0", t)
}

func Test_Storage_ISCSI_0_Update_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.IscsiEmpty)
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	storagesubtests.Update(s, "iscsi-test-0", t)
}

func Test_Storage_ISCSI_0_Get_Empty(t *testing.T) {
	storagesubtests.IscsiGetEmpty("iscsi-test-0", t)
}

func Test_Storage_ISCSI_0_Delete(t *testing.T) {
	storagesubtests.Delete("iscsi-test-0", t)
}
