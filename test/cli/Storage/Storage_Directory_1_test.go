package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_Directory_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("directory-test-1", t)
}

func Test_Storage_Directory_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.DirectoryEmpty)
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	storagesubtests.Create(s, "directory-test-1", t)
}

func Test_Storage_Directory_1_Get_Empty(t *testing.T) {
	storagesubtests.DirectoryGetEmpty("directory-test-1", t)
}

func Test_Storage_Directory_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.DirectoryFull)
	storagesubtests.Update(s, "directory-test-1", t)
}

func Test_Storage_Directory_1_Get_Full(t *testing.T) {
	storagesubtests.DirectoryGetFull("directory-test-1", t)
}

func Test_Storage_Directory_1_Delete(t *testing.T) {
	storagesubtests.Delete("directory-test-1", t)
}
