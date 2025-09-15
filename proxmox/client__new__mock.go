package proxmox

import "context"

type MockClient struct {
	// Guest
	GuestCheckPendingChangesFunc func(ctx context.Context, vmr *VmRef) (bool, error)
	GuestCheckVmRefFunc          func(ctx context.Context, vmr *VmRef) error
	GuestGetLxcRawConfigFunc     func(ctx context.Context, vmr *VmRef) (RawConfigLXC, error)
	GuestGetQemuRawConfigFunc    func(ctx context.Context, vmr *VmRef) (RawConfigQemu, error)
	GuestListResourcesFunc       func(ctx context.Context) (RawGuestResources, error)
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

func (m *MockClient) guestGetLxcRawConfig(ctx context.Context, vmr *VmRef) (RawConfigLXC, error) {
	if m.GuestGetLxcRawConfigFunc == nil {
		m.panic("GuestGetLxcRawConfigFunc")
	}
	return m.GuestGetLxcRawConfigFunc(ctx, vmr)
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
