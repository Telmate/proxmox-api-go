package proxmox

import (
	"context"
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
	if raw != nil && raw.GetState() == PowerStateStopped {
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
	if raw != nil && raw.GetState() == PowerStateStopped {
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
	if raw != nil && raw.GetState() == PowerStateRunning {
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
	if raw != nil && raw.GetState() == PowerStateStopped {
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
		return newVersionError("overrule stop", Version{Major: 9})
	}
	raw, err := vmr.check_unsafe(ctx, c.api)
	if err != nil {
		return err
	}
	if raw != nil && raw.GetState() == PowerStateStopped {
		return nil
	}
	return c.StopOverruleNoCheck(ctx, vmr)
}

func (c *guestClient) StopOverruleNoCheck(ctx context.Context, vmr VmRef) error {
	return vmr.stopOverrule_Unsafe(ctx, c.api)
}
