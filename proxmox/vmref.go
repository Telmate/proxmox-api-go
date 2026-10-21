package proxmox

import (
	"context"
	"errors"
)

const (
	clone_Error_MutuallyExclusive string = "linked and full clone are mutually exclusive"
	clone_Error_NoneSet           string = "either linked or full clone must be set"
)

func (vmr VmRef) toGuestRef() guestRef {
	return guestRef{ID: vmr.vmId, Node: vmr.node, Type: vmr.vmType}
}

// Deprecated: use LxcGuestInterface.Clone() instead.
// CloneLxc clones a new LXC container by cloning current container
func (vmr *VmRef) CloneLxc(ctx context.Context, settings CloneLxcTarget, c *Client) (*VmRef, error) {
	return c.New().LxcGuest.Clone(ctx, *vmr, settings)
}

// Deprecated: use LxcGuestInterface.CloneNoCheck() instead.
// CloneLxcNoCheck creates a new LXC container by cloning the current container, without input validation.
func (vmr *VmRef) CloneLxcNoCheck(ctx context.Context, settings CloneLxcTarget, c *Client) (*VmRef, error) {
	return c.New().LxcGuest.CloneNoCheck(ctx, *vmr, settings)
}

// Deprecated: use QemuGuestInterface.Clone() instead.
// CloneQemu creates a new Qemu VM by cloning the current VM.
func (vmr *VmRef) CloneQemu(ctx context.Context, settings CloneQemuTarget, c *Client) (*VmRef, error) {
	if vmr == nil {
		return nil, errors.New(VmRef_Error_Nil)
	}
	return c.New().QemuGuest.Clone(ctx, *vmr, settings)
}

// Deprecated: use QemuGuestInterface.CloneNoCheck() instead.
// CloneQemuNoCheck creates a new VM by cloning the current VM, without input validation.
func (vmr *VmRef) CloneQemuNoCheck(ctx context.Context, settings CloneQemuTarget, c *Client) (*VmRef, error) {
	return c.New().QemuGuest.CloneNoCheck(ctx, *vmr, settings)
}

// Deprecated: use GuestInterface.Delete() instead.
func (vmr VmRef) Delete(ctx context.Context, c *Client) error {
	_, err := c.New().Guest.Delete(ctx, vmr)
	return err
}

// Deprecated: use GuestInterface.DeleteNoCheck() instead.
func (vmr VmRef) DeleteNoCheck(ctx context.Context, c *Client) error {
	if err := c.checkInitialized(); err != nil {
		return err
	}
	_, err := c.New().Guest.Delete(ctx, vmr)
	return err
}

func (c *clientAPI) deleteGuest(ctx context.Context, vmr *VmRef, purge bool) error {
	url := "/nodes/" + vmr.node.String() + "/" + vmr.vmType.String() + "/" + vmr.vmId.String()
	if purge {
		url += "?purge=1"
	}
	return c.deleteTask(ctx, url)
}

// Deprecated: use GuestInterface.StopOverrule() instead.
// ForceStop stops the guest immediately without a graceful shutdown and cancels any stop/shutdown operations in progress.
// This function requires Proxmox VE 8.0 or later.
func (vmr *VmRef) ForceStop(ctx context.Context, c *Client) error {
	return c.New().Guest.StopOverrule(ctx, *vmr)
}

// Deprecated: use guestClient.Status() instead.
func (vmr *VmRef) GetRawGuestStatus(ctx context.Context, c *Client) (RawGuestStatus, error) {
	if err := c.checkInitialized(); err != nil {
		return nil, err
	}
	err := c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	return c.New().Guest.Status(ctx, *vmr)
}

// Deprecated: use guestRef.status() instead.
func (vmr *VmRef) getRawGuestStatus_Unsafe(ctx context.Context, c *clientAPI) (*rawGuestStatus, error) {
	raw, err := c.getMap(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/status/current", "vm", "STATE")
	if err != nil {
		return nil, err
	}
	return &rawGuestStatus{a: raw}, nil
}

// Deprecated: use LxcGuestInterface.Migrate() and QemuGuestInterface.Migrate() instead
func (vmr *VmRef) Migrate(ctx context.Context, c *Client, newNode NodeName, LiveMigrate bool) error {
	if vmr == nil {
		return errors.New(VmRef_Error_Nil)
	}
	if err := c.checkInitialized(); err != nil {
		return err
	}
	if err := newNode.Validate(); err != nil {
		return err
	}
	return vmr.toGuestRef().migrate(ctx, c.api(), newNode, LiveMigrate)
}

// Deprecated: use LxcGuestInterface.MigrateCoCheck() and QemuGuestInterface.MigrateNoCheck() instead
func (vmr *VmRef) MigrateNoCheck(ctx context.Context, c *Client, newNode NodeName, LiveMigrate bool) error {
	if vmr == nil {
		return errors.New(VmRef_Error_Nil)
	}
	if err := c.checkInitialized(); err != nil {
		return err
	}
	return vmr.toGuestRef().migrate(ctx, c.api(), newNode, LiveMigrate)
}

// Deprecated: use GuestInterface.PendingChanges() instead.
func (vmr *VmRef) PendingChanges(ctx context.Context, c *Client) (bool, error) {
	return c.New().Guest.PendingChanges(ctx, *vmr)
}

// Deprecated: use GuestInterface.Stop() instead.
func (vmr *VmRef) Stop(ctx context.Context, c *Client) error { return c.new().guestStop(ctx, vmr) }

// TODO Remove this when VmRef.Stop() is removed.
func (c *clientNewTest) guestStop(ctx context.Context, vmr *VmRef) error {
	if err := c.oldClient.CheckVmRef(ctx, vmr); err != nil {
		return err
	}
	return c.oldClient.New().Guest.Stop(ctx, *vmr, false)
}
