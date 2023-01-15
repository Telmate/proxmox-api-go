package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_ZFSoverISCSI_Comstar_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("zfs-over-iscsi_comstar-test-1", t)
}

func Test_Storage_ZFSoverISCSI_Comstar_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_ComstarEmpty)
	s.ZFSoverISCSI.Comstar = &proxmox.ConfigStorageZFSoverISCSI_Comstar{}
	storagesubtests.Create(s, "zfs-over-iscsi_comstar-test-1", t)
}

func Test_Storage_ZFSoverISCSI_Comstar_1_Get_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_ComstarEmpty)
	s.ID = "zfs-over-iscsi_comstar-test-1"
	s.ZFSoverISCSI.Blocksize = proxmox.PointerString("4k")
	s.ZFSoverISCSI.Comstar = &proxmox.ConfigStorageZFSoverISCSI_Comstar{
		TargetGroup: "",
		HostGroup:   "",
		Writecache:  true,
	}
	s.Content = &proxmox.ConfigStorageContent{
		DiskImage: proxmox.PointerBool(true),
	}
	storagesubtests.Get(s, s.ID, t)
}

func Test_Storage_ZFSoverISCSI_Comstar_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_ComstarFull)
	s.ZFSoverISCSI.Comstar.HostGroup = ""
	s.ZFSoverISCSI.Comstar.TargetGroup = ""
	storagesubtests.Update(s, "zfs-over-iscsi_comstar-test-1", t)
}

func Test_Storage_ZFSoverISCSI_Comstar_1_Get_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_ComstarFull)
	s.ID = "zfs-over-iscsi_comstar-test-1"
	s.ZFSoverISCSI.Blocksize = proxmox.PointerString("4k")
	s.ZFSoverISCSI.Comstar.HostGroup = ""
	s.ZFSoverISCSI.Comstar.TargetGroup = ""
	storagesubtests.Get(s, s.ID, t)
}

func Test_Storage_ZFSoverISCSI_Comstar_1_Delete(t *testing.T) {
	storagesubtests.Delete("zfs-over-iscsi_comstar-test-1", t)
}
