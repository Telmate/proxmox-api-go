package api_test

import (
	"testing"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	api_test "github.com/Telmate/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func _create_clone_vmref() (ref *pxapi.VmRef) {
	ref = pxapi.NewVmRef(101)
	ref.SetNode("pve")
	ref.SetVmType("qemu")
	return ref
}

func Test_Clone_Qemu_VM(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	config := _create_vm_spec(false)

	config.Create(_create_vmref(), Test.GetClient())

	cloneConfig := _create_vm_spec(false)

	fullClone := 1

	cloneConfig.Name = "test-qemu02"
	cloneConfig.FullClone = &fullClone

	err := cloneConfig.CloneVm(_create_vmref(), _create_clone_vmref(), Test.GetClient())

	require.NoError(t, err)

}

func Test_Qemu_VM_Is_Cloned(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	config, _ := pxapi.NewConfigQemuFromApi(_create_clone_vmref(), Test.GetClient())

	require.Equal(t, "order=ide2;net0", config.Boot)
}

func Test_Clone_Qemu_VM_Cleanup(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	Test.GetClient().DeleteVm(_create_clone_vmref())
	Test.GetClient().DeleteVm(_create_vmref())
}
