package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_LVM_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("lvm-test-1", t)
}

func Test_Storage_LVM_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.LVMEmpty)
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	storagesubtests.Create(s, "lvm-test-1", t)
}

func Test_Storage_LVM_1_Get_Empty(t *testing.T) {
	storagesubtests.LVMGetEmpty("lvm-test-1", t)
}

func Test_Storage_LVM_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.LVMFull)
	storagesubtests.Update(s, "lvm-test-1", t)
}

func Test_Storage_LVM_1_Get_Full(t *testing.T) {
	storagesubtests.LVMGetFull("lvm-test-1", t)
}

func Test_Storage_LVM_1_Delete(t *testing.T) {
	storagesubtests.Delete("lvm-test-1", t)
}
