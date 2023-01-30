package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_ZFSoverISCSI_Istgt_0_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("zfs-over-iscsi_istgt-test-0", t)
}

func Test_Storage_ZFSoverISCSI_Istgt_0_Create_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_IstgtFull)
	storagesubtests.Create(s, "zfs-over-iscsi_istgt-test-0", t)
}

func Test_Storage_ZFSoverISCSI_Istgt_0_Get_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_IstgtFull)
	s.ID = "zfs-over-iscsi_istgt-test-0"
	storagesubtests.Get(s, s.ID, t)
}

func Test_Storage_ZFSoverISCSI_Istgt_0_Update_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_IstgtEmpty)
	storagesubtests.Update(s, "zfs-over-iscsi_istgt-test-0", t)
}

func Test_Storage_ZFSoverISCSI_Istgt_0_Get_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_IstgtEmpty)
	s.ID = "zfs-over-iscsi_istgt-test-0"
	s.ZFSoverISCSI.Blocksize = proxmox.PointerString("8k")
	s.Content = &proxmox.ConfigStorageContent{
		DiskImage: proxmox.PointerBool(true),
	}
	storagesubtests.Get(s, s.ID, t)
}

func Test_Storage_ZFSoverISCSI_Istgt_0_Delete(t *testing.T) {
	storagesubtests.Delete("zfs-over-iscsi_istgt-test-0", t)
}
