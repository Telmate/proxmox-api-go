package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_RBD_0_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("rbd-test-0", t)
}

func Test_Storage_RBD_0_Create_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.RBDFull)
	s.RBD.Keyring = proxmox.PointerString("keyringplaceholder")
	storagesubtests.Create(s, "rbd-test-0", t)
}

func Test_Storage_RBD_0_Get_Full(t *testing.T) {
	storagesubtests.RBDGetFull("rbd-test-0", t)
}

func Test_Storage_RBD_0_Update_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.RBDEmpty)
	storagesubtests.Update(s, "rbd-test-0", t)
}

func Test_Storage_RBD_0_Get_Empty(t *testing.T) {
	storagesubtests.RBDGetEmpty("rbd-test-0", t)
}

func Test_Storage_RBD_0_Delete(t *testing.T) {
	storagesubtests.Delete("rbd-test-0", t)
}
