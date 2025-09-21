package proxmox

import (
	"context"
)

type mockClientAPI struct {
	deleteHaRuleFunc           func(ctx context.Context, id HaRuleID) error
	getGuestConfigFunc         func(ctx context.Context, vmr *VmRef) (map[string]any, error)
	getGuestPendingChangesFunc func(ctx context.Context, vmr *VmRef) ([]any, error)
	getHaRuleFunc              func(ctx context.Context, id HaRuleID) (map[string]any, error)
	getPoolConfigFunc          func(ctx context.Context, pool PoolName) (map[string]any, error)
	getUserConfigFunc          func(ctx context.Context, userId UserID) (map[string]any, error)
	listGuestResourcesFunc     func(ctx context.Context) ([]any, error)
	listHaRulesFunc            func(ctx context.Context) ([]any, error)
}

func (m mockClientAPI) new() clientApiInterface { return &m }

func (m *mockClientAPI) panic(field string) { panic(field + " not set in mockClientAPI") }

// Interface methods

func (m *mockClientAPI) deleteHaRule(ctx context.Context, id HaRuleID) error {
	if m.deleteHaRuleFunc == nil {
		m.panic("deleteHaRuleFunc")
	}
	return m.deleteHaRuleFunc(ctx, id)
}

func (m *mockClientAPI) getGuestConfig(ctx context.Context, vmr *VmRef) (vmConfig map[string]any, err error) {
	if m.getGuestConfigFunc == nil {
		m.panic("getGuestConfigFunc")
	}
	return m.getGuestConfigFunc(ctx, vmr)
}

func (m *mockClientAPI) getGuestPendingChanges(ctx context.Context, vmr *VmRef) ([]any, error) {
	if m.getGuestPendingChangesFunc == nil {
		m.panic("getGuestPendingChangesFunc")
	}
	return m.getGuestPendingChangesFunc(ctx, vmr)
}

func (m *mockClientAPI) getHaRule(ctx context.Context, id HaRuleID) (haRule map[string]any, err error) {
	if m.getHaRuleFunc == nil {
		m.panic("getHaRuleFunc")
	}
	return m.getHaRuleFunc(ctx, id)
}

func (m *mockClientAPI) getPoolConfig(ctx context.Context, pool PoolName) (poolConfig map[string]any, err error) {
	if m.getPoolConfigFunc == nil {
		m.panic("getPoolConfigFunc")
	}
	return m.getPoolConfigFunc(ctx, pool)
}

func (m *mockClientAPI) getUserConfig(ctx context.Context, userId UserID) (userConfig map[string]any, err error) {
	if m.getUserConfigFunc == nil {
		m.panic("getUserConfigFunc")
	}
	return m.getUserConfigFunc(ctx, userId)
}

func (m *mockClientAPI) listGuestResources(ctx context.Context) ([]any, error) {
	if m.listGuestResourcesFunc == nil {
		m.panic("listGuestResourcesFunc")
	}
	return m.listGuestResourcesFunc(ctx)
}

func (m *mockClientAPI) listHaRules(ctx context.Context) ([]any, error) {
	if m.listHaRulesFunc == nil {
		m.panic("ListHaRulesFunc")
	}
	return m.listHaRulesFunc(ctx)
}
