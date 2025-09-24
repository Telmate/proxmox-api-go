package proxmox

import "context"

type MockClient struct {
	// Guest
	GuestCheckPendingChangesFunc    func(ctx context.Context, vmr *VmRef) (bool, error)
	GuestCheckVmRefFunc             func(ctx context.Context, vmr *VmRef) error
	GuestGetLxcActiveRawConfigFunc  func(ctx context.Context, vmr *VmRef) (raw RawConfigLXC, pending bool, err error)
	GuestGetLxcRawConfigFunc        func(ctx context.Context, vmr *VmRef) (RawConfigLXC, error)
	GuestGetQemuActiveRawConfigFunc func(ctx context.Context, vmr *VmRef) (raw RawConfigQemu, pending bool, err error)
	GuestGetQemuRawConfigFunc       func(ctx context.Context, vmr *VmRef) (RawConfigQemu, error)
	GuestListResourcesFunc          func(ctx context.Context) (RawGuestResources, error)
	// HA
	HaCreateNodeAffinityRuleFunc            func(ctx context.Context, ha HaNodeAffinityRule) error
	HaCreateNodeAffinityRuleNoCheckFunc     func(ctx context.Context, ha HaNodeAffinityRule) error
	HaCreateResourceAffinityRuleFunc        func(ctx context.Context, ha HaResourceAffinityRule) error
	HaCreateResourceAffinityRuleNoCheckFunc func(ctx context.Context, ha HaResourceAffinityRule) error
	HaDeleteRuleFunc                        func(ctx context.Context, id HaRuleID) error
	HaDeleteRuleNoCheckFunc                 func(ctx context.Context, id HaRuleID) error
	HaGetRuleFunc                           func(ctx context.Context, id HaRuleID) (HaRule, error)
	HaListRulesFunc                         func(ctx context.Context) (HaRules, error)
	HaListRulesNoCheckFunc                  func(ctx context.Context) (HaRules, error)
	// Pool
	PoolGetRawConfigFunc        func(ctx context.Context, pool PoolName) (RawConfigPool, error)
	PoolGetRawConfigNoCheckFunc func(ctx context.Context, pool PoolName) (RawConfigPool, error)
	// User
	UserGetRawConfigFunc func(ctx context.Context, userID UserID) (RawConfigUser, error)
}

func (m MockClient) New() ClientNew { return &m }

func (m *MockClient) panic(field string) { panic(field + " not set in MockClient") }

func (m *MockClient) old() *Client { panic("old not implemented in MockClient") }

func (m *MockClient) apiGet() clientApiInterface { panic("apiGet not implemented in MockClient") }

func (m *MockClient) guestCheckPendingChanges(ctx context.Context, vmr *VmRef) (bool, error) {
	if m.GuestCheckPendingChangesFunc == nil {
		m.panic("GuestCheckPendingChangesFunc")
	}
	return m.GuestCheckPendingChangesFunc(ctx, vmr)
}

func (m *MockClient) guestCheckVmRef(ctx context.Context, vmr *VmRef) error {
	if m.GuestCheckVmRefFunc == nil {
		m.panic("GuestCheckVmRefFunc")
	}
	return m.GuestCheckVmRefFunc(ctx, vmr)
}

func (m *MockClient) guestGetLxcActiveRawConfig(ctx context.Context, vmr *VmRef) (raw RawConfigLXC, pending bool, err error) {
	if m.GuestGetLxcActiveRawConfigFunc == nil {
		m.panic("GuestGetLxcActiveRawConfigFunc")
	}
	return m.GuestGetLxcActiveRawConfigFunc(ctx, vmr)
}

func (m *MockClient) guestGetLxcRawConfig(ctx context.Context, vmr *VmRef) (RawConfigLXC, error) {
	if m.GuestGetLxcRawConfigFunc == nil {
		m.panic("GuestGetLxcRawConfigFunc")
	}
	return m.GuestGetLxcRawConfigFunc(ctx, vmr)
}

func (m *MockClient) guestGetQemuActiveRawConfig(ctx context.Context, vmr *VmRef) (raw RawConfigQemu, pending bool, err error) {
	if m.GuestGetQemuActiveRawConfigFunc == nil {
		m.panic("GuestGetQemuActiveRawConfigFunc")
	}
	return m.GuestGetQemuActiveRawConfigFunc(ctx, vmr)
}

func (m *MockClient) guestGetQemuRawConfig(ctx context.Context, vmr *VmRef) (RawConfigQemu, error) {
	if m.GuestGetQemuRawConfigFunc == nil {
		m.panic("GuestGetQemuRawConfigFunc")
	}
	return m.GuestGetQemuRawConfigFunc(ctx, vmr)
}

func (m *MockClient) guestListResources(ctx context.Context) (RawGuestResources, error) {
	if m.GuestListResourcesFunc == nil {
		m.panic("GuestListResourcesFunc")
	}
	return m.GuestListResourcesFunc(ctx)
}

func (m *MockClient) haCreateNodeAffinityRule(ctx context.Context, ha HaNodeAffinityRule) error {
	if m.HaCreateNodeAffinityRuleFunc == nil {
		m.panic("HaCreateNodeAffinityRuleFunc")
	}
	return m.HaCreateNodeAffinityRuleFunc(ctx, ha)
}

func (m *MockClient) haCreateNodeAffinityRuleNoCheck(ctx context.Context, ha HaNodeAffinityRule) error {
	if m.HaCreateNodeAffinityRuleNoCheckFunc == nil {
		m.panic("HaCreateNodeAffinityRuleNoCheckFunc")
	}
	return m.HaCreateNodeAffinityRuleNoCheckFunc(ctx, ha)
}

func (m *MockClient) haCreateResourceAffinityRule(ctx context.Context, ha HaResourceAffinityRule) error {
	if m.HaCreateResourceAffinityRuleFunc == nil {
		m.panic("HaCreateResourceAffinityRuleFunc")
	}
	return m.HaCreateResourceAffinityRuleFunc(ctx, ha)
}

func (m *MockClient) haCreateResourceAffinityRuleNoCheck(ctx context.Context, ha HaResourceAffinityRule) error {
	if m.HaCreateResourceAffinityRuleNoCheckFunc == nil {
		m.panic("HaCreateResourceAffinityRuleNoCheckFunc")
	}
	return m.HaCreateResourceAffinityRuleNoCheckFunc(ctx, ha)
}

func (m *MockClient) haDeleteRule(ctx context.Context, id HaRuleID) error {
	if m.HaDeleteRuleFunc == nil {
		m.panic("HaDeleteRuleFunc")
	}
	return m.HaDeleteRuleFunc(ctx, id)
}

func (m *MockClient) haDeleteRuleNoCheck(ctx context.Context, id HaRuleID) error {
	if m.HaDeleteRuleNoCheckFunc == nil {
		m.panic("HaDeleteRuleNoCheckFunc")
	}
	return m.HaDeleteRuleNoCheckFunc(ctx, id)
}

func (m *MockClient) haGetRule(ctx context.Context, id HaRuleID) (HaRule, error) {
	if m.HaGetRuleFunc == nil {
		m.panic("HaGetRuleFunc")
	}
	return m.HaGetRuleFunc(ctx, id)
}

func (m *MockClient) haListRules(ctx context.Context) (HaRules, error) {
	if m.HaListRulesFunc == nil {
		m.panic("HaListRulesFunc")
	}
	return m.HaListRulesFunc(ctx)
}

func (m *MockClient) haListRulesNoCheck(ctx context.Context) (HaRules, error) {
	if m.HaListRulesNoCheckFunc == nil {
		m.panic("HaListRulesNoCheckFunc")
	}
	return m.HaListRulesNoCheckFunc(ctx)
}

func (m *MockClient) poolGetRawConfig(ctx context.Context, pool PoolName) (RawConfigPool, error) {
	if m.PoolGetRawConfigFunc == nil {
		m.panic("PoolGetRawConfigFunc")
	}
	return m.PoolGetRawConfigFunc(ctx, pool)
}

func (m *MockClient) poolGetRawConfigNoCheck(ctx context.Context, pool PoolName) (RawConfigPool, error) {
	if m.PoolGetRawConfigNoCheckFunc == nil {
		m.panic("PoolGetRawConfigNoCheckFunc")
	}
	return m.PoolGetRawConfigNoCheckFunc(ctx, pool)
}

func (m *MockClient) userGetRawConfig(ctx context.Context, userID UserID) (RawConfigUser, error) {
	if m.UserGetRawConfigFunc == nil {
		m.panic("UserGetRawConfigFunc")
	}
	return m.UserGetRawConfigFunc(ctx, userID)
}
