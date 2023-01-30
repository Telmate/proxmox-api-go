package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_CephFS_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("cephfs-test-1", t)
}

func Test_Storage_CephFS_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.CephfsEmpty)
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	storagesubtests.Create(s, "cephfs-test-1", t)
}

func Test_Storage_CephFS_1_Get_Empty(t *testing.T) {
	storagesubtests.CephfsGetEmpty("cephfs-test-1", t)
}

func Test_Storage_CephFS_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.CephfsFull)
	storagesubtests.Update(s, "cephfs-test-1", t)
}

func Test_Storage_CephFS_1_Get_Full(t *testing.T) {
	storagesubtests.CephfsGetFull("cephfs-test-1", t)
}

func Test_Storage_CephFS_1_Delete(t *testing.T) {
	storagesubtests.Delete("cephfs-test-1", t)
}
