package proxmox

import "context"

// The new implementation of the client

type ClientNew interface {
	// This interface is for mocking the client from the consumers perspective.
	// We should never call this interface from the SDK code.

	old() *Client               // TODO once we use `ClientNew` everywhere this function can be removed
	apiGet() clientApiInterface // TODO once we use `ClientNew` everywhere this function can be removed

	// Guest
	guestGetLxcRawConfig(ctx context.Context, vmr *VmRef) (RawConfigLXC, error)
	guestGetQemuRawConfig(ctx context.Context, vmr *VmRef) (RawConfigQemu, error)
	guestListResources(ctx context.Context) (RawGuestResources, error)
	// Pool
	poolGetRawConfig(ctx context.Context, pool PoolName) (RawConfigPool, error)
	poolGetRawConfigNoCheck(ctx context.Context, pool PoolName) (RawConfigPool, error)
	// User
	userGetRawConfig(ctx context.Context, userID UserID) (RawConfigUser, error)
}

type clientNew struct {
	api       clientApiInterface
	oldClient *Client
}

func (c *clientNew) old() *Client { return c.oldClient }

func (c *clientNew) apiGet() clientApiInterface { return c.api }
