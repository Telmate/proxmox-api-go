package cli_storage_test

import (
	"testing"

	_ "github.com/Bluearchive/proxmox-api-go/cli/command/commands"
	storagesubtests "github.com/Bluearchive/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_ZFS_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("zfs-test-1", t)
}

func Test_Storage_ZFS_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSEmpty)
	storagesubtests.Create(s, "zfs-test-1", t)
}

func Test_Storage_ZFS_1_Get_Empty(t *testing.T) {
	storagesubtests.ZFSGetEmpty("zfs-test-1", t)
}

func Test_Storage_ZFS_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSFull)
	storagesubtests.Update(s, "zfs-test-1", t)
}

func Test_Storage_ZFS_1_Get_Full(t *testing.T) {
	storagesubtests.ZFSGetFull("zfs-test-1", t)
}

func Test_Storage_ZFS_1_Delete(t *testing.T) {
	storagesubtests.Delete("zfs-test-1", t)
}
