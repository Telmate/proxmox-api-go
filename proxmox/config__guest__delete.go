package proxmox

import (
	"context"
	"errors"
	"strings"
)

func (c *guestClient) Delete(ctx context.Context, vmr VmRef) (bool, error) {
	ref := vmr.toGuestRef()
	rawGuest, err := ref.check(ctx, nil, c.api)
	if err != nil {
		if errors.Is(err, Error.GuestDoesNotExist()) {
			return false, nil
		}
		return false, err
	}

	var protection bool // Check if guest is protected
	switch ref.Type {
	case GuestLxc:
		var version Version
		version, err = c.oldClient.Version(ctx)
		if err != nil {
			return false, err
		}
		var raw *rawConfigLXC
		if raw, err = readLxcConfig(ctx, ref, version, c.api); err != nil {
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
		raw, err = readQemuConfig(ctx, ref, c.api)
		if err != nil {
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
		return false, errorMsg{}.guestIsProtectedCantDelete(ref.ID)
	}

	attempts := 2
	var haEnabled bool
	var state PowerState
	if rawGuest != nil {
		haEnabled = rawGuest.GetHaState() != nil
		state = rawGuest.GetState()
	} else {
		rawStatus, err := ref.status(ctx, c.api)
		if err != nil {
			return false, err
		}
		haEnabled = rawStatus.GetHaState() != nil
		state = rawStatus.GetState()
	}
	if state == PowerStateStopped {
		var deleted bool
		deleted, err = ref.delete(ctx, haEnabled, c.api)
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
			if _, err = ref.ID.deleteHaResource(ctx, c.api); err != nil { // It's faster to delete HA resource first, instead of stopping via HA
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
		if err = ref.stopOverruleOpertunistic(ctx, c.api, version); err != nil {
			apiErr, ok := err.(*ApiError)
			if !ok {
				return false, err
			}
			if !strings.HasSuffix(apiErr.Message, " not running") { // "CT 1000 not running"
				return false, err
			}
		}
		var deleted bool
		if deleted, err = ref.delete(ctx, haEnabled, c.api); err != nil {
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
	return vmr.toGuestRef().delete(ctx, false, c.api)
}
