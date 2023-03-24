package proxmox

import (
	"strconv"
)

// All code LXC and Qemu have in common should be placed here.

// Check if there are any pending changes that require a reboot to be applied.
func GuestHasPendingChanges(vmr *VmRef, client *Client) (bool, error) {
	params, err := pendingGuestConfigFromApi(vmr, client)
	if err != nil {
		return false, err
	}
	return keyExists(params, "pending"), nil
}

// Reboot the specified guest
func GuestReboot(vmr *VmRef, client *Client) (err error) {
	_, err = client.ShutdownVm(vmr)
	if err != nil {
		return
	}
	_, err = client.StartVm(vmr)
	return
}

func pendingGuestConfigFromApi(vmr *VmRef, client *Client) ([]interface{}, error) {
	err := vmr.nilCheck()
	if err != nil {
		return nil, err
	}
	if err = client.CheckVmRef(vmr); err != nil {
		return nil, err
	}
	return client.GetItemConfigInterfaceArray("/nodes/"+vmr.node+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/pending", "Guest", "PENDING CONFIG")
}
