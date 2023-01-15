package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_ZFSoverISCSI_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("zfs-over-iscsi-test-0", t)
}

func Test_Storage_ZFSoverISCSI_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSIEmpty)
	storagesubtests.Create(s, "zfs-over-iscsi-test-1", t)
}

func Test_Storage_ZFSoverISCSI_1_Get_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSIEmpty)
	s.ID = "zfs-over-iscsi-test-1"
	s.ZFSoverISCSI.Blocksize = proxmox.PointerString("4k")
	s.Content = &proxmox.ConfigStorageContent{
		DiskImage: proxmox.PointerBool(true),
	}
	storagesubtests.Get(s, s.ID, t)
}

func Test_Storage_ZFSoverISCSI_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSIFull)
	storagesubtests.Update(s, "zfs-over-iscsi-test-1", t)
}

func Test_Storage_ZFSoverISCSI_1_Get_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSIFull)
	s.ID = "zfs-over-iscsi-test-1"
	s.ZFSoverISCSI.Blocksize = proxmox.PointerString("4k")
	storagesubtests.Get(s, s.ID, t)
}

func Test_Storage_ZFSoverISCSI_1_Delete(t *testing.T) {
	storagesubtests.Delete("zfs-over-iscsi-test-1", t)
}
