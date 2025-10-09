package proxmox

import (
	"context"
	"errors"
)

// The new implementation of the client

type ClientNew interface {
	// This interface is for mocking the client from the consumers perspective.
	// We should never call this interface from the SDK code.

	old() *Client               // TODO once we use `ClientNew` everywhere this function can be removed
	apiGet() clientApiInterface // TODO once we use `ClientNew` everywhere this function can be removed

	// Guest
	guestCheckPendingChanges(ctx context.Context, vmr *VmRef) (bool, error)
	guestCheckVmRef(ctx context.Context, vmr *VmRef) error
	guestGetLxcActiveRawConfig(ctx context.Context, vmr *VmRef) (raw RawConfigLXC, pending bool, err error)
	guestGetLxcRawConfig(ctx context.Context, vmr *VmRef) (RawConfigLXC, error)
	guestGetQemuActiveRawConfig(ctx context.Context, vmr *VmRef) (raw RawConfigQemu, pending bool, err error)
	guestGetQemuRawConfig(ctx context.Context, vmr *VmRef) (RawConfigQemu, error)
	guestListResources(ctx context.Context) (RawGuestResources, error)
	guestStop(ctx context.Context, vmr *VmRef) error
	guestStopForce(ctx context.Context, vmr *VmRef) error
	// HA
	haCreateNodeAffinityRule(ctx context.Context, ha HaNodeAffinityRule) error
	haCreateNodeAffinityRuleNoCheck(ctx context.Context, ha HaNodeAffinityRule) error
	haCreateResourceAffinityRule(ctx context.Context, ha HaResourceAffinityRule) error
	haCreateResourceAffinityRuleNoCheck(ctx context.Context, ha HaResourceAffinityRule) error
	haDeleteResource(ctx context.Context, id GuestID) error
	haDeleteRule(ctx context.Context, id HaRuleID) error
	haDeleteRuleNoCheck(ctx context.Context, id HaRuleID) error
	haGetRule(ctx context.Context, id HaRuleID) (HaRule, error)
	haListRules(ctx context.Context) (HaRules, error)
	haListRulesNoCheck(ctx context.Context) (HaRules, error)
	haUpdateNodeAffinityRule(ctx context.Context, ha HaNodeAffinityRule) error
	haUpdateNodeAffinityRuleNoCheck(ctx context.Context, ha HaNodeAffinityRule) error
	haUpdateResourceAffinityRule(ctx context.Context, ha HaResourceAffinityRule) error
	haUpdateResourceAffinityRuleNoCheck(ctx context.Context, ha HaResourceAffinityRule) error
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

func (c *clientNew) guestCheckVmRef(ctx context.Context, vmr *VmRef) error {
	if vmr == nil {
		return errors.New(VmRef_Error_Nil)
	}
	return c.guestCheckVmRef_Unsafe(ctx, vmr)
}

func (c *clientNew) guestCheckVmRef_Unsafe(ctx context.Context, vmr *VmRef) error {
	if vmr.node == "" || vmr.vmType == guestUnknown {
		raw, err := c.guestListResources(ctx)
		if err != nil {
			return err
		}
		rawGuest, err := raw.SelectID(vmr.vmId)
		if err != nil {
			return err
		}
		vmr.node = rawGuest.GetNode()
		vmr.vmType = rawGuest.GetType()
	}
	return nil
}
