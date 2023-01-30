package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_GlusterFS_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("glusterfs-test-1", t)
}

func Test_Storage_GlusterFS_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.GlusterfsEmpty)
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	storagesubtests.Create(s, "glusterfs-test-1", t)
}

func Test_Storage_GlusterFS_1_Get_Empty(t *testing.T) {
	storagesubtests.GlusterfsGetEmpty("glusterfs-test-1", t)
}

func Test_Storage_GlusterFS_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.GlusterfsFull)
	storagesubtests.Update(s, "glusterfs-test-1", t)
}

func Test_Storage_GlusterFS_1_Get_Full(t *testing.T) {
	storagesubtests.GlusterfsGetFull("glusterfs-test-1", t)
}

func Test_Storage_GlusterFS_1_Delete(t *testing.T) {
	storagesubtests.Delete("glusterfs-test-1", t)
}
