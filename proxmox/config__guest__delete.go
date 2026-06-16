package proxmox

import (
	"context"
	"errors"
	"strings"
)

func (c *guestClient) Delete(ctx context.Context, vmr VmRef) (bool, error) {
	if vmr.vmId == 0 {
		return false, errors.New(VmRef_Error_IDnotSet)
	}
	rawGuests, err := c.api.listGuestResources(ctx)
	if err != nil {
		return false, err
	}
	rawGuest, ok := rawGuests.selectID(vmr.vmId)
	if !ok {
		return false, nil
	}
	vmr.node = rawGuest.GetNode()
	vmr.vmType = rawGuest.GetType()

	var protection bool // Check if guest is protected
	switch vmr.vmType {
	case GuestLxc:
		var raw *rawConfigLXC
		if raw, err = guestGetLxcRawConfig_Unsafe(ctx, &vmr, c.api); err != nil {
			if apiErr, ok := err.(*ApiError); ok {
				if strings.HasSuffix(apiErr.Message, " does not exist") { // "Configuration file 'nodes/pve-9l/lxc/1000.conf' does not exist"
					return false, nil
				}
			}
			return false, err
		}
		protection = raw.GetProtection()
	case GuestQemu:
		var raw *rawConfigQemu
		if raw, err = guestGetRawQemuConfig_Unsafe(ctx, &vmr, c.api); err != nil {
			if apiErr, ok := err.(*ApiError); ok {
				if strings.HasSuffix(apiErr.Message, " does not exist") { // "Configuration file 'nodes/pve-9l/qemu-server/1023.conf' does not exist"
					return false, nil
				}
			}
			return false, err
		}
		protection = raw.GetProtection()
	}
	if protection {
		return false, errorMsg{}.guestIsProtectedCantDelete(vmr.vmId)
	}

	attempts := 2
	haEnabled := rawGuest.GetHaState() != nil
	if rawGuest.GetStatus() == PowerStateStopped {
		var deleted bool
		deleted, err = vmr.delete_Unsafe(ctx, c.api, haEnabled)
		if err == nil {
			return deleted, nil
		}
		apiErr, ok := err.(*ApiError)
		if !ok {
			return false, err
		}
		if !strings.HasSuffix(apiErr.Message, " is running") { // "unable to destroy CT 1000 - container is running"
			return false, err
		}
	} else {
		if haEnabled {
			if _, err = vmr.vmId.deleteHaResource(ctx, c.api); err != nil { // It's faster to delete HA resource first, instead of stopping via HA
				return false, err
			}
			haEnabled = false
		}
		attempts += 1
	}

	var version Version
	if version, err = c.oldClient.Version(ctx); err != nil {
		return false, err
	}
	for range attempts {
		if err = vmr.stopOverruleOpertunistic_Unsafe(ctx, c.api, version); err != nil {
			apiErr, ok := err.(*ApiError)
			if !ok {
				return false, err
			}
			if !strings.HasSuffix(apiErr.Message, " not running") { // "CT 1000 not running"
				return false, err
			}
		}
		var deleted bool
		if deleted, err = vmr.delete_Unsafe(ctx, c.api, haEnabled); err != nil {
			apiErr, ok := err.(*ApiError)
			if !ok {
				return false, err
			}
			if strings.HasSuffix(apiErr.Message, " is running") { // "unable to destroy CT 1000 - container is running"
				continue
			}
			return false, err
		}
		return deleted, nil
	}
	return false, errors.New("unable to delete guest in 3 attempts")
}

func (c *guestClient) DeleteNoCheck(ctx context.Context, vmr VmRef) (bool, error) {
	return vmr.delete_Unsafe(ctx, c.api, false)
}
