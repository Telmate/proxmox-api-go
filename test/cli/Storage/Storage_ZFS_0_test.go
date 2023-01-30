package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_ZFS_0_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("zfs-test-0", t)
}

func Test_Storage_ZFS_0_Create_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSFull)
	storagesubtests.Create(s, "zfs-test-0", t)
}

func Test_Storage_ZFS_0_Get_Full(t *testing.T) {
	storagesubtests.ZFSGetFull("zfs-test-0", t)
}

func Test_Storage_ZFS_0_Update_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSEmpty)
	storagesubtests.Update(s, "zfs-test-0", t)
}

func Test_Storage_ZFS_0_Get_Empty(t *testing.T) {
	storagesubtests.ZFSGetEmpty("zfs-test-0", t)
}

func Test_Storage_ZFS_0_Delete(t *testing.T) {
	storagesubtests.Delete("zfs-test-0", t)
}
