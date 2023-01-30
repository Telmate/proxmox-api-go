package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_LVMThin_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("lvm-thin-test-1", t)
}

func Test_Storage_LVMThin_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.LVMThinEmpty)
	storagesubtests.Create(s, "lvm-thin-test-1", t)
}

func Test_Storage_LVMThin_1_Get_Empty(t *testing.T) {
	storagesubtests.LVMThinGetEmpty("lvm-thin-test-1", t)
}

func Test_Storage_LVMThin_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.LVMThinFull)
	storagesubtests.Update(s, "lvm-thin-test-1", t)
}

func Test_Storage_LVMThin_1_Get_Full(t *testing.T) {
	storagesubtests.LVMThinGetFull("lvm-thin-test-1", t)
}

func Test_Storage_LVMThin_1_Delete(t *testing.T) {
	storagesubtests.Delete("lvm-thin-test-1", t)
}
