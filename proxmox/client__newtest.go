package proxmox

import (
	"context"
	"errors"
)

// The new implementation of the client

type ClientNewTest interface {
	// This interface is for mocking the client from the consumers perspective.
	// We should never call this interface from the SDK code.

	old() *Client               // TODO once we use `ClientNew` everywhere this function can be removed
	apiGet() clientApiInterface // TODO once we use `ClientNew` everywhere this function can be removed
	apiRaw() *clientAPI

	// Guest
	guestCheckVmRef(ctx context.Context, vmr *VmRef) error
	guestStop(ctx context.Context, vmr *VmRef) error
	// HA
	haCreateNodeAffinityRule(ctx context.Context, ha HaNodeAffinityRule) error
	haCreateNodeAffinityRuleNoCheck(ctx context.Context, ha HaNodeAffinityRule) error
	haCreateResourceAffinityRule(ctx context.Context, ha HaResourceAffinityRule) error
	haCreateResourceAffinityRuleNoCheck(ctx context.Context, ha HaResourceAffinityRule) error
	haDeleteResource(ctx context.Context, id GuestID) (bool, error)
	haDeleteRule(ctx context.Context, id HaRuleID) error
	haDeleteRuleNoCheck(ctx context.Context, id HaRuleID) error
	haGetRule(ctx context.Context, id HaRuleID) (HaRule, error)
	haListRules(ctx context.Context) (HaRules, error)
	haListRulesNoCheck(ctx context.Context) (HaRules, error)
	haUpdateNodeAffinityRule(ctx context.Context, ha HaNodeAffinityRule) error
	haUpdateNodeAffinityRuleNoCheck(ctx context.Context, ha HaNodeAffinityRule) error
	haUpdateResourceAffinityRule(ctx context.Context, ha HaResourceAffinityRule) error
	haUpdateResourceAffinityRuleNoCheck(ctx context.Context, ha HaResourceAffinityRule) error
	// User
	userGetRawConfig(ctx context.Context, userID UserID) (RawConfigUser, error)
}

type clientNewTest struct {
	api       clientApiInterface
	oldClient *Client
}

func (c *clientNewTest) old() *Client { return c.oldClient }

func (c *clientNewTest) apiGet() clientApiInterface { return c.api }

func (c *clientNewTest) apiRaw() *clientAPI { return c.api.(*clientAPI) }

func (c *clientNewTest) guestCheckVmRef(ctx context.Context, vmr *VmRef) error {
	if vmr == nil {
		return errors.New(VmRef_Error_Nil)
	}
	return c.guestCheckVmRef_Unsafe(ctx, vmr)
}

func (c *clientNewTest) guestCheckVmRef_Unsafe(ctx context.Context, vmr *VmRef) error {
	if vmr.node == "" || vmr.vmType == guestUnknown {
		raw, err := c.api.listGuestResources(ctx)
		if err != nil {
			return err
		}
		rawGuest, ok := raw.selectID(vmr.vmId)
		if !ok {
			return errorMsg{}.guestDoesNotExist(vmr.vmId)
		}
		vmr.node = rawGuest.GetNode()
		vmr.vmType = rawGuest.GetType()
	}
	return nil
}
