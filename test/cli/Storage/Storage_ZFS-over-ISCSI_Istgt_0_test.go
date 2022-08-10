package cli_storage_test

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
	storagesubtests "github.com/Telmate/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_ZFSoverISCSI_Istgt_0_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("zfs-over-iscsi_istgt-test-0", t)
}

func Test_Storage_ZFSoverISCSI_Istgt_0_Create_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_IstgtFull)
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(zfs-over-iscsi_istgt-test-0)",
		Contains:  true,
		Args:      []string{"-i", "create", "storage", "zfs-over-iscsi_istgt-test-0"},
	}
	Test.StandardTest(t)
}

func Test_Storage_ZFSoverISCSI_Istgt_0_Get_Full(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_IstgtFull)
	s.ID = "zfs-over-iscsi_istgt-test-0"
	Test := cliTest.Test{
		OutputJson: storagesubtests.InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", "zfs-over-iscsi_istgt-test-0"},
	}
	Test.StandardTest(t)
}

func Test_Storage_ZFSoverISCSI_Istgt_0_Update_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_IstgtEmpty)
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(zfs-over-iscsi_istgt-test-0)",
		Contains:  true,
		Args:      []string{"-i", "update", "storage", "zfs-over-iscsi_istgt-test-0"},
	}
	Test.StandardTest(t)
}

func Test_Storage_ZFSoverISCSI_Istgt_0_Get_Empty(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := storagesubtests.CloneJson(storagesubtests.ZFSoverISCSI_IstgtEmpty)
	s.ID = "zfs-over-iscsi_istgt-test-0"
	s.ZFSoverISCSI.Blocksize = proxmox.PointerString("8k")
	s.Content = &proxmox.ConfigStorageContent{
		DiskImage: proxmox.PointerBool(true),
	}
	Test := cliTest.Test{
		OutputJson: storagesubtests.InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", "zfs-over-iscsi_istgt-test-0"},
	}
	Test.StandardTest(t)
}

func Test_Storage_ZFSoverISCSI_Istgt_0_Delete(t *testing.T) {
	storagesubtests.Delete("zfs-over-iscsi_istgt-test-0", t)
}
