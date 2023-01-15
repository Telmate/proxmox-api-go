package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_LVMThin_0_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("lvm-thin-test-0", t)
}

func Test_Storage_LVMThin_0_Create_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.LVMThinFull)
	storagesubtests.Create(s, "lvm-thin-test-0", t)
}

func Test_Storage_LVMThin_0_Get_Full(t *testing.T) {
	storagesubtests.LVMThinGetFull("lvm-thin-test-0", t)
}

func Test_Storage_LVMThin_0_Update_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.LVMThinEmpty)
	storagesubtests.Update(s, "lvm-thin-test-0", t)
}

func Test_Storage_LVMThin_0_Get_Empty(t *testing.T) {
	storagesubtests.LVMThinGetEmpty("lvm-thin-test-0", t)
}

func Test_Storage_LVMThin_0_Delete(t *testing.T) {
	storagesubtests.Delete("lvm-thin-test-0", t)
}
