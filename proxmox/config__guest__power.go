package proxmox

import (
	"context"
	"errors"
)

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
