package cli_storage_test

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
	storagesubtests "github.com/Telmate/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_ZFSoverISCSI_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("zfs-over-iscsi-test-0", t)
}

func Test_Storage_ZFSoverISCSI_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSIEmpty)
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(zfs-over-iscsi-test-1)",
		Contains:  true,
		Args:      []string{"-i", "create", "storage", "zfs-over-iscsi-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_ZFSoverISCSI_1_Get_Empty(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSIEmpty)
	s.ID = "zfs-over-iscsi-test-1"
	s.ZFSoverISCSI.Blocksize = proxmox.PointerString("4k")
	s.Content = &proxmox.ConfigStorageContent{
		DiskImage: proxmox.PointerBool(true),
	}
	Test := cliTest.Test{
		OutputJson: storagesubtests.InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", "zfs-over-iscsi-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_ZFSoverISCSI_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSIFull)
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(zfs-over-iscsi-test-1)",
		Contains:  true,
		Args:      []string{"-i", "update", "storage", "zfs-over-iscsi-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_ZFSoverISCSI_1_Get_Full(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSIFull)
	s.ID = "zfs-over-iscsi-test-1"
	s.ZFSoverISCSI.Blocksize = proxmox.PointerString("4k")
	Test := cliTest.Test{
		OutputJson: storagesubtests.InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", "zfs-over-iscsi-test-1"},
	}
	Test.StandardTest(t)
}

func Test_Storage_ZFSoverISCSI_1_Delete(t *testing.T) {
	storagesubtests.Delete("zfs-over-iscsi-test-1", t)
}
