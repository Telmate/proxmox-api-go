package api_test

import (
	"context"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
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

	err := configNetwork.CreateNetwork(context.Background(), Test.GetClient())
	require.NoError(t, err)
	_, err = Test.GetClient().ApplyNetwork(context.Background(), "pve")
	require.NoError(t, err)

	config.Disks.VirtIO.Disk_0.Disk = &pxapi.QemuVirtIODisk{
		ImportFrom: "/tmp/jammy-server-cloudimg-amd64.img",
		Storage:    "local"}

	config.Name = util.Pointer(pxapi.GuestName("Base-Image"))

	_, err = Test.GetClient().New().QemuGuest.Create(context.Background(), config)
	require.NoError(t, err)
	config.Boot = &pxapi.BootOrder{
		pxapi.QemuVirtIoID0.BootDevice(),
		pxapi.QemuIdeID2.BootDevice(),
		pxapi.QemuNetworkInterfaceID0.BootDevice()}
	config.CloudInit = &pxapi.CloudInit{
		NetworkInterfaces: pxapi.CloudInitNetworkInterfaces{
			pxapi.QemuNetworkInterfaceID0: pxapi.CloudInitNetworkConfig{
				IPv4: &pxapi.CloudInitIPv4Config{
					Address: util.Pointer(pxapi.IPv4CIDR("10.0.0.2/24")),
					Gateway: util.Pointer(pxapi.IPv4Address("10.0.0.1"))}}}}
	err = Test.GetClient().New().QemuGuest.Update(context.Background(), *vmref, true, true, config)
	require.NoError(t, err)

	testConfig, _ := pxapi.NewConfigQemuFromApi(context.Background(), vmref, Test.GetClient())

	require.Equal(t, testConfig.CloudInit.NetworkInterfaces[pxapi.QemuNetworkInterfaceID0],
		pxapi.CloudInitNetworkConfig{
			IPv4: &pxapi.CloudInitIPv4Config{
				Address: util.Pointer(pxapi.IPv4CIDR("10.0.0.2/24")),
				Gateway: util.Pointer(pxapi.IPv4Address("10.0.0.1"))}})

	existed, err := Test.GetClient().New().Guest.Delete(context.Background(), *vmref)
	require.NoError(t, err)
	require.True(t, existed)

	_, err = Test.GetClient().DeleteNetwork(context.Background(), "pve", "vmbr0")
	require.NoError(t, err)
	_, err = Test.GetClient().ApplyNetwork(context.Background(), "pve")
	require.NoError(t, err)
}
