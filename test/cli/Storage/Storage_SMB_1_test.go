package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_SMB_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("smb-test-1", t)
}

func Test_Storage_SMB_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.SMBEmpty)
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	storagesubtests.Create(s, "smb-test-1", t)
}

func Test_Storage_SMB_1_Get_Empty(t *testing.T) {
	storagesubtests.SMBGetEmpty("smb-test-1", t)
}

func Test_Storage_SMB_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.SMBFull)
	s.SMB.Password = proxmox.PointerString("Enter123!")
	storagesubtests.Update(s, "smb-test-1", t)
}

func Test_Storage_SMB_1_Get_Full(t *testing.T) {
	storagesubtests.SMBGetFull("smb-test-1", t)
}

func Test_Storage_SMB_1_Delete(t *testing.T) {
	storagesubtests.Delete("smb-test-1", t)
}
