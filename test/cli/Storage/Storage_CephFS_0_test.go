package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_CephFS_0_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("cephfs-test-0", t)
}

func Test_Storage_CephFS_0_Create_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.CephfsFull)
	storagesubtests.Create(s, "cephfs-test-0", t)
}

func Test_Storage_CephFS_0_Get_Full(t *testing.T) {
	storagesubtests.CephfsGetFull("cephfs-test-0", t)
}

func Test_Storage_CephFS_0_Update_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.CephfsEmpty)
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	storagesubtests.Update(s, "cephfs-test-0", t)
}

func Test_Storage_CephFS_0_Get_Empty(t *testing.T) {
	storagesubtests.CephfsGetEmpty("cephfs-test-0", t)
}

func Test_Storage_CephFS_0_Delete(t *testing.T) {
	storagesubtests.Delete("cephfs-test-0", t)
}
