package api_test

import (
	"testing"

	api_test "github.com/Telmate/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_Start_Stop_Qemu_VM_Setup(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	config := _create_vm_spec(false)
	config.CreateVm(_create_vmref(), Test.GetClient())
}

func Test_Start_Qemu_VM(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	_, err := Test.GetClient().StartVm(_create_vmref())
	require.NoError(t, err)
}

func Test_Stop_Qemu_VM(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	_, err := Test.GetClient().StopVm(_create_vmref())
	require.NoError(t, err)
}

func Test_Start_Stop_Qemu_VM_Cleanup(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	Test.GetClient().DeleteVm(_create_vmref())
}
