package cli_storage_test

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
	storagesubtests "github.com/Telmate/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_LVMThin_0_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("lvm-thin-test-0", t)
}

func Test_Storage_LVMThin_0_Create_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.LVMThinFull)
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(lvm-thin-test-0)",
		Contains:  true,
		Args:      []string{"-i", "create", "storage", "lvm-thin-test-0"},
	}
	Test.StandardTest(t)
}

func Test_Storage_LVMThin_0_Get_Full(t *testing.T) {
	storagesubtests.LVMThinGetFull("lvm-thin-test-0", t)
}

func Test_Storage_LVMThin_0_Update_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.LVMThinEmpty)
	Test := cliTest.Test{
		InputJson: storagesubtests.InlineMarshal(s),
		Expected:  "(lvm-thin-test-0)",
		Contains:  true,
		Args:      []string{"-i", "update", "storage", "lvm-thin-test-0"},
	}
	Test.StandardTest(t)
}

func Test_Storage_LVMThin_0_Get_Empty(t *testing.T) {
	storagesubtests.LVMThinGetEmpty("lvm-thin-test-0", t)
}

func Test_Storage_LVMThin_0_Delete(t *testing.T) {
	storagesubtests.Delete("lvm-thin-test-0", t)
}
