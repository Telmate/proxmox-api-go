package proxmox

import (
	"context"
	"time"
)

// in the future we might put the interface even lower, but for now this is sufficient
type clientApiInterface interface {
	deleteHaRule(ctx context.Context, id HaRuleID) error
	getGuestConfig(ctx context.Context, vmr *VmRef) (map[string]any, error)
	getGuestPendingChanges(ctx context.Context, vmr *VmRef) ([]any, error)
	getPoolConfig(ctx context.Context, pool PoolName) (map[string]any, error)
	getUserConfig(ctx context.Context, userId UserID) (map[string]any, error)
	listGuestResources(ctx context.Context) ([]any, error)
	listHaRules(ctx context.Context) ([]any, error)
}

type clientAPI struct {
	session     *Session
	url         string
	user        UserID
	taskTimeout time.Duration
}

// Interface methods

func (c *clientAPI) deleteHaRule(ctx context.Context, id HaRuleID) error {
	return c.delete(ctx, "/cluster/ha/rules/"+id.String())
}

func (c *clientAPI) getGuestConfig(ctx context.Context, vmr *VmRef) (vmConfig map[string]any, err error) {
	return c.getMap(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/config", "vm", "CONFIG", nil)
}

func (c *clientAPI) getGuestPendingChanges(ctx context.Context, vmr *VmRef) ([]any, error) {
	return c.getList(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/pending", "Guest", "PENDING CONFIG", nil)
}

func (c *clientAPI) getPoolConfig(ctx context.Context, pool PoolName) (poolConfig map[string]any, err error) {
	return c.getMap(ctx, "/pools/"+string(pool), "pool", "CONFIG", nil)
}

func (c *clientAPI) getUserConfig(ctx context.Context, userID UserID) (userConfig map[string]any, err error) {
	return c.getMap(ctx, "/access/users/"+userID.String(), "user", "CONFIG", nil)
}

func (c *clientAPI) listGuestResources(ctx context.Context) ([]any, error) {
	return c.getResourceList(ctx, resourceListGuest)
}

func (c *clientAPI) listHaRules(ctx context.Context) ([]any, error) {
	return c.getList(ctx, "/cluster/ha/rules", "ha rules", "CONFIG", nil)
}
