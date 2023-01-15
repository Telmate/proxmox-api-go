package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_ZFSoverISCSI_Comstar_0_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("zfs-over-iscsi_comstar-test-0", t)
}

func Test_Storage_ZFSoverISCSI_Comstar_0_Create_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_ComstarFull)
	storagesubtests.Create(s, "zfs-over-iscsi_comstar-test-0", t)
}

func Test_Storage_ZFSoverISCSI_Comstar_0_Get_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_ComstarFull)
	s.ID = "zfs-over-iscsi_comstar-test-0"
	storagesubtests.Get(s, s.ID, t)
}

func Test_Storage_ZFSoverISCSI_Comstar_0_Update_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_ComstarEmpty)
	s.ZFSoverISCSI.Comstar.HostGroup = "h-group"
	s.ZFSoverISCSI.Comstar.TargetGroup = "t-group"
	storagesubtests.Update(s, "zfs-over-iscsi_comstar-test-0", t)
}

func Test_Storage_ZFSoverISCSI_Comstar_0_Get_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_ComstarEmpty)
	s.ID = "zfs-over-iscsi_comstar-test-0"
	s.ZFSoverISCSI.Blocksize = proxmox.PointerString("8k")
	s.ZFSoverISCSI.Comstar = &proxmox.ConfigStorageZFSoverISCSI_Comstar{
		TargetGroup: "t-group",
		HostGroup:   "h-group",
		Writecache:  false,
	}
	s.Content = &proxmox.ConfigStorageContent{
		DiskImage: proxmox.PointerBool(true),
	}
	storagesubtests.Get(s, s.ID, t)
}

func Test_Storage_ZFSoverISCSI_Comstar_0_Delete(t *testing.T) {
	storagesubtests.Delete("zfs-over-iscsi_comstar-test-0", t)
}
