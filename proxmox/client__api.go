package proxmox

import (
	"context"
	"errors"
	"strings"
	"time"
)

// in the future we might put the interface even lower, but for now this is sufficient
type clientApiInterface interface {
	createHaRule(ctx context.Context, params map[string]any) error
	deleteHaResource(ctx context.Context, id GuestID) error
	deleteHaRule(ctx context.Context, id HaRuleID) error
	getGuestConfig(ctx context.Context, vmr *VmRef) (map[string]any, error)
	getGuestPendingChanges(ctx context.Context, vmr *VmRef) ([]any, error)
	getGuestQemuAgent(ctx context.Context, vmr *VmRef) (map[string]any, GuestAgentState, error)
	getHaRule(ctx context.Context, id HaRuleID) (map[string]any, error)
	getPoolConfig(ctx context.Context, pool PoolName) (map[string]any, error)
	getUserConfig(ctx context.Context, userId UserID) (map[string]any, bool, error)
	listGuestResources(ctx context.Context) ([]any, error)
	listHaRules(ctx context.Context) ([]any, error)
	updateGuestStatus(ctx context.Context, vmr *VmRef, setStatus string, params map[string]interface{}) error
	updateHaRule(ctx context.Context, id HaRuleID, params map[string]any) error
}

type clientAPI struct {
	session     *Session
	url         string
	user        UserID
	taskTimeout time.Duration
	timeUnit    time.Duration
}

// Interface methods

func (c *clientAPI) createHaRule(ctx context.Context, params map[string]any) error {
	_, err := c.post(ctx, "/cluster/ha/rules", params)
	return err
}

func (c *clientAPI) deleteHaResource(ctx context.Context, id GuestID) error {
	_, err := c.delete(ctx, "/cluster/ha/resources/"+id.String())
	return err
}

func (c *clientAPI) deleteHaRule(ctx context.Context, id HaRuleID) error {
	_, err := c.delete(ctx, "/cluster/ha/rules/"+id.String())
	return err
}

func (c *clientAPI) getGuestConfig(ctx context.Context, vmr *VmRef) (vmConfig map[string]any, err error) {
	return c.getMap(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/config", "vm", "CONFIG")
}

func (c *clientAPI) getGuestPendingChanges(ctx context.Context, vmr *VmRef) ([]any, error) {
	return c.getList(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/pending", "Guest", "PENDING CONFIG")
}

func (c *clientAPI) getGuestQemuAgent(ctx context.Context, vmr *VmRef) (map[string]any, GuestAgentState, error) {
	guestID := vmr.vmId.String()
	out, err := c.getMap(ctx, "/nodes/"+vmr.node.String()+"/qemu/"+guestID+"/agent/network-get-interfaces", "guest agent", "data")
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		if strings.HasPrefix(apiErr.Message, "QEMU guest agent is not running") {
			return out, GuestAgentStateNotRunning, nil
		}
		if strings.HasPrefix(apiErr.Message, "VM "+guestID+" is not running") {
			return out, GuestAgentStateVmNotRunning, nil
		}
	}
	return out, GuestAgentStateUnknown, err
}

func (c *clientAPI) getHaRule(ctx context.Context, id HaRuleID) (haRule map[string]any, err error) {
	out, err := c.getMap(ctx, "/cluster/ha/rules/"+id.String(), "ha rule", "CONFIG")
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		if strings.HasPrefix(apiErr.Message, "no such ha rule ") {
			return out, nil
		}
	}
	return out, err
}

func (c *clientAPI) getPoolConfig(ctx context.Context, pool PoolName) (poolConfig map[string]any, err error) {
	return c.getMap(ctx, "/pools/"+string(pool), "pool", "CONFIG")
}

func (c *clientAPI) getUserConfig(ctx context.Context, userID UserID) (map[string]any, bool, error) {
	config, err := c.getMap(ctx, "/access/users/"+userID.String(), "user", "CONFIG")
	if err == nil {
		return config, true, nil
	}
	var apiErr *ApiError
	if ok := errors.As(err, &apiErr); ok {
		if strings.HasPrefix(apiErr.Message, "no such user ") {
			return config, false, nil
		}
	}
	return config, false, err
}

func (c *clientAPI) listGuestResources(ctx context.Context) ([]any, error) {
	return c.getResourceList(ctx, resourceListGuest)
}

func (c *clientAPI) listHaRules(ctx context.Context) ([]any, error) {
	return c.getList(ctx, "/cluster/ha/rules", "ha rules", "CONFIG")
}

func (c *clientAPI) updateGuestStatus(ctx context.Context, vmr *VmRef, setStatus string, params map[string]interface{}) error {
	_, err := c.postTask(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/status/"+setStatus, params)
	return err
}

func (c *clientAPI) updateHaRule(ctx context.Context, id HaRuleID, params map[string]any) error {
	_, err := c.put(ctx, "/cluster/ha/rules/"+id.String(), params)
	return err
}
