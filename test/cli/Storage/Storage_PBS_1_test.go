package cli_storage_test

import (
	"testing"

	_ "github.com/Bluearchive/proxmox-api-go/cli/command/commands"
	"github.com/Bluearchive/proxmox-api-go/proxmox"
	storagesubtests "github.com/Bluearchive/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_PBS_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("pbs-test-1", t)
}

func Test_Storage_PBS_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.PBSEmpty)
	s.PBS.Password = proxmox.PointerString("Enter123!")
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	storagesubtests.Create(s, "pbs-test-1", t)
}

func Test_Storage_PBS_1_Get_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.PBSEmpty)
	s.ID = "pbs-test-1"
	s.PBS.Port = proxmox.PointerInt(8007)
	s.Content = &proxmox.ConfigStorageContent{
		Backup: proxmox.PointerBool(true),
	}
	storagesubtests.Get(s, s.ID, t)
}

func Test_Storage_PBS_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.PBSFull)
	storagesubtests.Update(s, "pbs-test-1", t)
}

func Test_Storage_PBS_1_Get_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.PBSFull)
	s.ID = "pbs-test-1"
	storagesubtests.Get(s, s.ID, t)
}

func Test_Storage_PBS_1_Delete(t *testing.T) {
	storagesubtests.Delete("pbs-test-1", t)
}
