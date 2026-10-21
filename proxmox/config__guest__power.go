package proxmox

import (
	"context"
)

func (c *guestClient) Reboot(ctx context.Context, vmr VmRef) error {
	ref := vmr.toGuestRef()
	if _, err := ref.check(ctx, nil, c.api); err != nil {
		return err
	}
	return c.RebootNoCheck(ctx, ref.legacy())
}

func (c *guestClient) RebootNoCheck(ctx context.Context, vmr VmRef) error {
	return vmr.toGuestRef().reboot(ctx, c.api)
}

func (c *guestClient) Shutdown(ctx context.Context, vmr VmRef) error {
	ref := vmr.toGuestRef()
	raw, err := ref.check(ctx, nil, c.api)
	if err != nil {
		return err
	}
	if raw != nil && raw.GetState() == PowerStateStopped {
		return nil
	}
	return c.ShutdownNoCheck(ctx, ref.legacy())
}

func (c *guestClient) ShutdownNoCheck(ctx context.Context, vmr VmRef) error {
	return vmr.toGuestRef().shutdown(ctx, c.api)
}

func (c *guestClient) ShutdownForce(ctx context.Context, vmr VmRef) error {
	ref := vmr.toGuestRef()
	raw, err := ref.check(ctx, nil, c.api)
	if err != nil {
		return err
	}
	if raw != nil && raw.GetState() == PowerStateStopped {
		return nil
	}
	return c.ShutdownForceNoCheck(ctx, ref.legacy())
}

func (c *guestClient) ShutdownForceNoCheck(ctx context.Context, vmr VmRef) error {
	return vmr.toGuestRef().shutdownForce(ctx, c.api)
}

func (c *guestClient) Start(ctx context.Context, vmr VmRef) error {
	ref := vmr.toGuestRef()
	raw, err := ref.check(ctx, nil, c.api)
	if err != nil {
		return err
	}
	if raw != nil && raw.GetState() == PowerStateRunning {
		return nil
	}
	return c.StartNoCheck(ctx, ref.legacy())
}

func (c *guestClient) StartNoCheck(ctx context.Context, vmr VmRef) error {
	return vmr.toGuestRef().start(ctx, c.api)
}

func (c *guestClient) Stop(ctx context.Context, vmr VmRef, overrule bool) error {
	ref := vmr.toGuestRef()
	raw, err := ref.check(ctx, nil, c.api)
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
			return c.StopOverruleNoCheck(ctx, ref.legacy())
		}
	}
	return c.StopNoCheck(ctx, ref.legacy())
}

func (c *guestClient) StopNoCheck(ctx context.Context, vmr VmRef) error {
	return vmr.toGuestRef().stop(ctx, c.api)
}

func (c *guestClient) StopOverrule(ctx context.Context, vmr VmRef) error {
	version, err := c.oldClient.Version(ctx)
	if err != nil {
		return err
	}
	if version.Major < 8 {
		return newVersionError("overrule stop", Version{Major: 9})
	}
	ref := vmr.toGuestRef()
	raw, err := ref.check(ctx, nil, c.api)
	if err != nil {
		return err
	}
	if raw != nil && raw.GetState() == PowerStateStopped {
		return nil
	}
	return c.StopOverruleNoCheck(ctx, ref.legacy())
}

func (c *guestClient) StopOverruleNoCheck(ctx context.Context, vmr VmRef) error {
	return vmr.toGuestRef().stopOverrule(ctx, c.api)
}
