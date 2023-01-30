package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_Directory_0_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("directory-test-0", t)
}

func Test_Storage_Directory_0_Create_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.DirectoryFull)
	storagesubtests.Create(s, "directory-test-0", t)
}

func Test_Storage_Directory_0_Get_Full(t *testing.T) {
	storagesubtests.DirectoryGetFull("directory-test-0", t)
}

func Test_Storage_Directory_0_Update_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.DirectoryEmpty)
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	storagesubtests.Update(s, "directory-test-0", t)
}

func Test_Storage_Directory_0_Get_Empty(t *testing.T) {
	storagesubtests.DirectoryGetEmpty("directory-test-0", t)
}

func Test_Storage_Directory_0_Delete(t *testing.T) {
	storagesubtests.Delete("directory-test-0", t)
}
