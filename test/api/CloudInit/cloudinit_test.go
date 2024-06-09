package api_test

import (
	"testing"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	api_test "github.com/Telmate/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_Cloud_Init_VM(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	config := _create_vm_spec(true)
	vmref := _create_vmref()

	// Create network
	configNetwork := _create_network_spec()

	err := configNetwork.CreateNetwork(Test.GetClient())
	require.NoError(t, err)
	_, err = Test.GetClient().ApplyNetwork("pve")
	require.NoError(t, err)

	disk := make(map[string]interface{})
	disk["import-from"] = "/tmp/jammy-server-cloudimg-amd64.img"
	disk["type"] = "virtio"
	disk["storage"] = "local"

	config.QemuDisks[0] = disk
	config.Name = "Base-Image"

	err = config.Create(vmref, Test.GetClient())
	require.NoError(t, err)

	config.Boot = "order=virtio0;ide2;net0"

	config.CloudInit = &pxapi.CloudInit{
		NetworkInterfaces: pxapi.CloudInitNetworkInterfaces{
			pxapi.QemuNetworkInterfaceID0: "gw=10.0.0.1,ip=10.0.0.2/24"}}

	_, err = config.Update(true, vmref, Test.GetClient())
	require.NoError(t, err)

	testConfig, _ := pxapi.NewConfigQemuFromApi(vmref, Test.GetClient())

	require.Equal(t, testConfig.CloudInit.NetworkInterfaces[pxapi.QemuNetworkInterfaceID0], "gw=10.0.0.1,ip=10.0.0.2/24")

	_, err = Test.GetClient().DeleteVm(vmref)
	require.NoError(t, err)

	_, err = Test.GetClient().DeleteNetwork("pve", "vmbr0")
	require.NoError(t, err)
	_, err = Test.GetClient().ApplyNetwork("pve")
	require.NoError(t, err)
}
