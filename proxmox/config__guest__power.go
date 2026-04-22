package proxmox

import (
	"context"
	"errors"
)

type (
	GuestInterface interface {
		Reboot(context.Context, VmRef) error
		RebootNoCheck(context.Context, VmRef) error

		Shutdown(context.Context, VmRef) error
		ShutdownNoCheck(context.Context, VmRef) error

		ShutdownForce(context.Context, VmRef) error
		ShutdownForceNoCheck(context.Context, VmRef) error

		Start(context.Context, VmRef) error
		StartNoCheck(context.Context, VmRef) error

		// Stop will stop the guest.
		// The overrule flag will opportunistically set the overrule-shutdown parameter if the Proxmox VE version is 8.0 or higher.
		// If the version is lower, overrule will be ignored and the normal stop command will be executed, as overrule is not supported on versions lower than 8.0.
		Stop(ctx context.Context, vmr VmRef, overrule bool) error
		StopNoCheck(context.Context, VmRef) error

		// StopOverrule is not supported on Proxmox VE versions lower than 8.0.
		// On unsupported versions, this method will return an error.
		// On supported versions, this method will stop the guest and overrule any shutdown hooks or timeouts.
		StopOverrule(context.Context, VmRef) error
		StopOverruleNoCheck(context.Context, VmRef) error
	}

	guestClient struct {
		api       *clientAPI
		oldClient *Client
	}
)

var _ GuestInterface = (*guestClient)(nil)

func (c *guestClient) Reboot(ctx context.Context, vmr VmRef) error {
	if _, err := vmr.check_unsafe(ctx, c.api); err != nil {
		return err
	}
	return c.RebootNoCheck(ctx, vmr)
}

func (c *guestClient) RebootNoCheck(ctx context.Context, vmr VmRef) error {
	return vmr.reboot_Unsafe(ctx, c.api)
}

func (c *guestClient) Shutdown(ctx context.Context, vmr VmRef) error {
	raw, err := vmr.check_unsafe(ctx, c.api)
	if err != nil {
		return err
	}
	if raw != nil && raw.GetStatus() == PowerStateStopped {
		return nil
	}
	return c.ShutdownNoCheck(ctx, vmr)
}

func (c *guestClient) ShutdownNoCheck(ctx context.Context, vmr VmRef) error {
	return vmr.shutdown_Unsafe(ctx, c.api)
}

func (c *guestClient) ShutdownForce(ctx context.Context, vmr VmRef) error {
	raw, err := vmr.check_unsafe(ctx, c.api)
	if err != nil {
		return err
	}
	if raw != nil && raw.GetStatus() == PowerStateStopped {
		return nil
	}
	return c.ShutdownForceNoCheck(ctx, vmr)
}

func (c *guestClient) ShutdownForceNoCheck(ctx context.Context, vmr VmRef) error {
	return vmr.shutdownForce_Unsafe(ctx, c.api)
}

func (c *guestClient) Start(ctx context.Context, vmr VmRef) error {
	raw, err := vmr.check_unsafe(ctx, c.api)
	if err != nil {
		return err
	}
	if raw != nil && raw.GetStatus() == PowerStateRunning {
		return nil
	}
	return c.StartNoCheck(ctx, vmr)
}

func (c *guestClient) StartNoCheck(ctx context.Context, vmr VmRef) error {
	return vmr.start_Unsafe(ctx, c.api)
}

func (c *guestClient) Stop(ctx context.Context, vmr VmRef, overrule bool) error {
	raw, err := vmr.check_unsafe(ctx, c.api)
	if err != nil {
		return err
	}
	if raw != nil && raw.GetStatus() == PowerStateStopped {
		return nil
	}
	if overrule {
		version, err := c.oldClient.Version(ctx)
		if err != nil {
			return err
		}
		if version.Major >= 8 {
			return c.StopOverruleNoCheck(ctx, vmr)
		}
	}
	return c.StopNoCheck(ctx, vmr)
}

func (c *guestClient) StopNoCheck(ctx context.Context, vmr VmRef) error {
	return vmr.stop_Unsafe(ctx, c.api)
}

func (c *guestClient) StopOverrule(ctx context.Context, vmr VmRef) error {
	version, err := c.oldClient.Version(ctx)
	if err != nil {
		return err
	}
	if version.Major < 8 {
		return errors.New("overrule stop is only supported on Proxmox VE 8.0 and higher")
	}
	raw, err := vmr.check_unsafe(ctx, c.api)
	if err != nil {
		return err
	}
	if raw != nil && raw.GetStatus() == PowerStateStopped {
		return nil
	}
	return c.StopOverruleNoCheck(ctx, vmr)
}

func (c *guestClient) StopOverruleNoCheck(ctx context.Context, vmr VmRef) error {
	return vmr.stopOverrule_Unsafe(ctx, c.api)
}
