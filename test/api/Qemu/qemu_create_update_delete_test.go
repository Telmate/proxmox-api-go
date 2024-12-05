package api_test

import (
	"context"
	"testing"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	api_test "github.com/Telmate/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_Create_Qemu_VM(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	config := _create_vm_spec(true)

	err := config.Create(context.Background(), _create_vmref(), Test.GetClient())
	require.NoError(t, err)
}

func Test_Qemu_VM_Is_Added(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	config, _ := pxapi.NewConfigQemuFromApi(context.Background(), _create_vmref(), Test.GetClient())

	require.Equal(t, "order=ide2;net0", config.Boot)
}

func Test_Update_Qemu_VM(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	config := _create_vm_spec(true)

	config.Boot = "order=net0;ide2"

	_, err := config.Update(context.Background(), true, _create_vmref(), Test.GetClient())

	require.NoError(t, err)
}

func Test_Qemu_VM_Is_Updated(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	config, _ := pxapi.NewConfigQemuFromApi(context.Background(), _create_vmref(), Test.GetClient())
	require.Equal(t, "order=net0;ide2", config.Boot)
}

func Test_Remove_Qemu_VM(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	_, err := Test.GetClient().DeleteVm(context.Background(), _create_vmref())

	require.NoError(t, err)
}
