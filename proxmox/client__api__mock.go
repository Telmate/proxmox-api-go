package proxmox

import (
	"context"
)

type mockClientAPI struct {
	getGuestConfigFunc     func(ctx context.Context, vmr *VmRef) (map[string]any, error)
	getPoolConfigFunc      func(ctx context.Context, pool PoolName) (map[string]any, error)
	getUserConfigFunc      func(ctx context.Context, userId UserID) (map[string]any, error)
	listGuestResourcesFunc func(ctx context.Context) ([]interface{}, error)
}

func (m mockClientAPI) new() clientApiInterface { return &m }

func (m *mockClientAPI) panic(field string) { panic(field + " not set in mockClientAPI") }

// Interface methods

func (m *mockClientAPI) getGuestConfig(ctx context.Context, vmr *VmRef) (vmConfig map[string]any, err error) {
	if m.getGuestConfigFunc == nil {
		m.panic("getGuestConfigFunc")
	}
	return m.getGuestConfigFunc(ctx, vmr)
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

func (m *mockClientAPI) listGuestResources(ctx context.Context) ([]interface{}, error) {
	if m.listGuestResourcesFunc == nil {
		m.panic("listGuestResourcesFunc")
	}
	return m.listGuestResourcesFunc(ctx)
}
