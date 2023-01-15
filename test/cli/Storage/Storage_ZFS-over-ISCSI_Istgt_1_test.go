package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_ZFSoverISCSI_Istgt_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("zfs-over-iscsi_istgt-test-1", t)
}

func Test_Storage_ZFSoverISCSI_Istgt_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_IstgtEmpty)
	s.ZFSoverISCSI.Comstar = &proxmox.ConfigStorageZFSoverISCSI_Comstar{}
	storagesubtests.Create(s, "zfs-over-iscsi_istgt-test-1", t)
}

func Test_Storage_ZFSoverISCSI_Istgt_1_Get_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_IstgtEmpty)
	s.ID = "zfs-over-iscsi_istgt-test-1"
	s.ZFSoverISCSI.Blocksize = proxmox.PointerString("4k")
	s.Content = &proxmox.ConfigStorageContent{
		DiskImage: proxmox.PointerBool(true),
	}
	storagesubtests.Get(s, s.ID, t)
}

func Test_Storage_ZFSoverISCSI_Istgt_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_IstgtFull)
	storagesubtests.Update(s, "zfs-over-iscsi_istgt-test-1", t)
}

func Test_Storage_ZFSoverISCSI_Istgt_1_Get_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_IstgtFull)
	s.ID = "zfs-over-iscsi_istgt-test-1"
	s.ZFSoverISCSI.Blocksize = proxmox.PointerString("4k")
	storagesubtests.Get(s, s.ID, t)
}

func Test_Storage_ZFSoverISCSI_Istgt_1_Delete(t *testing.T) {
	storagesubtests.Delete("zfs-over-iscsi_istgt-test-1", t)
}
