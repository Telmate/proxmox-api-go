package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_NFS_0_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("nfs-test-0", t)
}

func Test_Storage_NFS_0_Create_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.NFSFull)
	storagesubtests.Create(s, "nfs-test-0", t)
}

func Test_Storage_NFS_0_Get_Full(t *testing.T) {
	storagesubtests.NFSGetFull("nfs-test-0", t)
}

func Test_Storage_NFS_0_Update_Empty(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := storagesubtests.CloneJson(storagesubtests.NFSEmpty)
	s.BackupRetention = &proxmox.ConfigStorageBackupRetention{}
	storagesubtests.Update(s, "nfs-test-0", t)
}

func Test_Storage_NFS_0_Get_Empty(t *testing.T) {
	storagesubtests.NFSGetEmpty("nfs-test-0", t)
}

func Test_Storage_NFS_0_Delete(t *testing.T) {
	storagesubtests.Delete("nfs-test-0", t)
}
