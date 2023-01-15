package cli_storage_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	storagesubtests "github.com/perimeter-81/proxmox-api-go/test/cli/Storage/storage-sub-tests"
)

func Test_Storage_RBD_1_Cleanup(t *testing.T) {
	storagesubtests.Cleanup("rbd-test-1", t)
}

func Test_Storage_RBD_1_Create_Empty(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.RBDEmpty)
	s.RBD.Keyring = proxmox.PointerString("keyringplaceholder")
	storagesubtests.Create(s, "rbd-test-1", t)
}

func Test_Storage_RBD_1_Get_Empty(t *testing.T) {
	storagesubtests.RBDGetEmpty("rbd-test-1", t)
}

func Test_Storage_RBD_1_Update_Full(t *testing.T) {
	s := storagesubtests.CloneJson(storagesubtests.RBDFull)
	storagesubtests.Update(s, "rbd-test-1", t)
}

func Test_Storage_RBD_1_Get_Full(t *testing.T) {
	storagesubtests.RBDGetFull("rbd-test-1", t)
}

func Test_Storage_RBD_1_Delete(t *testing.T) {
	storagesubtests.Delete("rbd-test-1", t)
}
