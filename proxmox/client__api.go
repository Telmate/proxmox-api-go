package proxmox

import (
	"context"
	"time"
)

// in the future we might put the interface even lower, but for now this is sufficient
type clientApiInterface interface {
	getGuestConfig(ctx context.Context, vmr *VmRef) (map[string]any, error)
	getPoolConfig(ctx context.Context, pool PoolName) (map[string]any, error)
	listGuestResources(ctx context.Context) ([]any, error)
}

type clientAPI struct {
	session     *Session
	url         string
	user        UserID
	taskTimeout time.Duration
}

// Interface methods

func (c *clientAPI) getGuestConfig(ctx context.Context, vmr *VmRef) (vmConfig map[string]any, err error) {
	return c.getMap(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+vmr.vmId.String()+"/config", "vm", "CONFIG", nil)
}

func (c *clientAPI) getPoolConfig(ctx context.Context, pool PoolName) (poolConfig map[string]any, err error) {
	return c.getMap(ctx, "/pools/"+string(pool), "pool", "CONFIG", nil)
}

func (c *clientAPI) listGuestResources(ctx context.Context) ([]any, error) {
	return c.getResourceList(ctx, resourceListGuest)
}
