package proxmox

import (
	"context"
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/mockServer"
	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_pool"
	"github.com/stretchr/testify/require"
)

func Test_poolClient_AddMembers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pool     PoolName
		guests   []GuestID
		storages []StorageName
		output   map[PoolName]string
		requests []mockServer.Request
		err      error
	}{
		{name: `No Add`,
			pool: "test_pool"},
		{name: `Add Storages Only, is version agnostic`,
			pool:     "test_pool",
			storages: []StorageName{"local", "nfs-1", "ceph-1"},
			requests: mockServer.Append(
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"storage": "local,nfs-1,ceph-1",
				}))},
		{name: `Add & remove Guests 8.0-`,
			pool:   "test_pool",
			guests: []GuestID{100, 200, 300, 400},
			requests: mockServer.Append(
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{
					"data": []map[string]any{
						{"vmid": float64(100), "pool": ""},
						{"vmid": float64(200), "pool": "original_pool"},
						{"vmid": float64(400), "pool": "test_pool"},
						{"vmid": float64(300), "pool": ""},
						{"vmid": float64(500), "pool": "original_pool"}}}),
				mockServer.RequestsPut("/pools/original_pool", map[string]any{
					"delete": "1",
					"vms":    "200"}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"vms": "100,200,300",
				}))},
		{name: `Add Guests Only 8.0-`,
			pool:   "test_pool",
			guests: []GuestID{100, 200, 300, 400},
			requests: mockServer.Append(
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{
					"data": []map[string]any{
						{"vmid": float64(100), "pool": ""},
						{"vmid": float64(200), "pool": ""},
						{"vmid": float64(400), "pool": "test_pool"},
						{"vmid": float64(300), "pool": ""},
						{"vmid": float64(500), "pool": "original_pool"},
					}}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"vms": "100,200,300",
				}))},
		{name: `Add Guests Only 8.0+`,
			pool:   "test_pool",
			guests: []GuestID{100, 200, 300},
			requests: mockServer.Append(
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"allow-move": "1",
					"vms":        "100,200,300",
				}))},
		{name: `Add Members 8.0-`,
			pool:     "test_pool",
			guests:   []GuestID{100, 200, 300, 400},
			storages: []StorageName{"local", "nfs-1"},
			requests: mockServer.Append(
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{
					"data": []map[string]any{
						{"vmid": float64(100), "pool": ""},
						{"vmid": float64(200), "pool": "original_pool"},
						{"vmid": float64(400), "pool": "test_pool"},
						{"vmid": float64(300), "pool": ""},
						{"vmid": float64(500), "pool": "original_pool"},
					}}),
				mockServer.RequestsPut("/pools/original_pool", map[string]any{
					"delete": "1",
					"vms":    "200",
				}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"vms":     "100,200,300",
					"storage": "local,nfs-1",
				}))},
		{name: `Add Members 8.0-, delete error`,
			pool:     "test_pool",
			guests:   []GuestID{100, 200, 300, 400},
			storages: []StorageName{"local", "nfs-1"},
			requests: mockServer.Append(
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{
					"data": []map[string]any{
						{"vmid": float64(100), "pool": ""},
						{"vmid": float64(200), "pool": "original_pool"},
						{"vmid": float64(400), "pool": "test_pool"},
						{"vmid": float64(300), "pool": ""},
						{"vmid": float64(500), "pool": "original_pool"},
					}}),
				mockServer.RequestsError("/pools/original_pool", mockServer.PUT, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `Add Members 8.0-, list error`,
			pool:     "test_pool",
			guests:   []GuestID{100, 200, 300, 400},
			storages: []StorageName{"local", "nfs-1"},
			requests: mockServer.Append(
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsError("/cluster/resources?type=vm", mockServer.GET, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `Add Members 8.0+`,
			pool:     "test_pool",
			guests:   []GuestID{100, 200, 300},
			storages: []StorageName{"local", "nfs-1"},
			requests: mockServer.Append(
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"allow-move": "1",
					"vms":        "100,200,300",
					"storage":    "local,nfs-1",
				}))},
		{name: `Add Members 8.0+, version error`,
			pool:     "test_pool",
			guests:   []GuestID{100, 200, 300},
			storages: []StorageName{"local", "nfs-1"},
			requests: mockServer.RequestsError("/version", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
		{name: `Validate error PoolName`,
			err: errors.New("PoolName cannot be empty")},
		{name: `Validate error GuestID`,
			pool:   "test",
			guests: []GuestID{99},
			err:    errors.New("guestID should be greater than 99")},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			c.clearVersion()
			server.Set(test.requests, t)
			err := c.New().Pool.AddMembers(context.Background(), test.pool, test.guests, test.storages)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_poolClient_Create(t *testing.T) {
	t.Parallel()
	const path = "/pools"
	tests := []struct {
		name     string
		pool     ConfigPool
		requests []mockServer.Request
		err      error
	}{
		{name: `Id Only`,
			pool: ConfigPool{Name: "test_pool"},
			requests: mockServer.RequestsPost(path, map[string]any{
				"poolid": "test_pool"})},
		{name: `With Comment empty`,
			pool: ConfigPool{
				Name:    "test_pool",
				Comment: util.Pointer("")},
			requests: mockServer.RequestsPost(path, map[string]any{
				"poolid": "test_pool"})},
		{name: `With Comment set`,
			pool: ConfigPool{
				Name:    "test_pool",
				Comment: util.Pointer("This is a test pool" + body.Symbols)},
			requests: mockServer.RequestsPost(path, map[string]any{
				"poolid":  "test_pool",
				"comment": "This is a test pool" + body.Symbols})},
		{name: `With Guests 8.0-`,
			pool: ConfigPool{
				Name:   "test_pool",
				Guests: &[]GuestID{100, 200, 300}},
			requests: mockServer.Append(
				mockServer.RequestsPost(path, map[string]any{
					"poolid": "test_pool"}),
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "pool": "original_pool"},
					map[string]any{"vmid": float64(200), "pool": "test_pool"},
					map[string]any{"vmid": float64(300), "pool": ""}}}),
				mockServer.RequestsPut("/pools/original_pool", map[string]any{
					"delete": "1",
					"vms":    "100"}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"vms": "100,300"}))},
		{name: `With Guests 8.0+`,
			pool: ConfigPool{
				Name:   "test_pool",
				Guests: &[]GuestID{100, 200, 300}},
			requests: mockServer.Append(
				mockServer.RequestsPost(path, map[string]any{
					"poolid": "test_pool"}),
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"allow-move": "1",
					"vms":        "100,200,300"}))},
		{name: `With Storages`,
			pool: ConfigPool{
				Name:     "test_pool",
				Storages: &[]StorageName{"local", "nfs-1", "ceph-1"}},
			requests: mockServer.Append(
				mockServer.RequestsPost(path, map[string]any{
					"poolid": "test_pool"}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"storage": "local,nfs-1,ceph-1"}))},
		{name: `Full Create 8.0-`,
			pool: ConfigPool{
				Name:     "test_pool",
				Comment:  util.Pointer("test Comment"),
				Guests:   &[]GuestID{100, 200, 300},
				Storages: &[]StorageName{"local", "nfs-1", "ceph-1"}},
			requests: mockServer.Append(
				mockServer.RequestsPost(path, map[string]any{
					"poolid":  "test_pool",
					"comment": "test Comment"}),
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "pool": "original_pool"},
					map[string]any{"vmid": float64(200), "pool": "test_pool"},
					map[string]any{"vmid": float64(300), "pool": ""}}}),
				mockServer.RequestsPut("/pools/original_pool", map[string]any{
					"delete": "1",
					"vms":    "100"}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"vms":     "100,300",
					"storage": "local,nfs-1,ceph-1"}))},
		{name: `Full Create 8.0+`,
			pool: ConfigPool{
				Name:     "test_pool",
				Comment:  util.Pointer("test Comment"),
				Guests:   &[]GuestID{100, 200, 300},
				Storages: &[]StorageName{"local", "nfs-1", "ceph-1"}},
			requests: mockServer.Append(
				mockServer.RequestsPost(path, map[string]any{
					"poolid":  "test_pool",
					"comment": "test Comment"}),
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"allow-move": "1",
					"vms":        "100,200,300",
					"storage":    "local,nfs-1,ceph-1"}))},
		{name: `Validate error`,
			err: errors.New("PoolName cannot be empty")},
		{name: `500 internal server error`,
			pool:     ConfigPool{Name: "test_pool"},
			requests: mockServer.RequestsError(path, mockServer.POST, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			c.clearVersion()
			server.Set(test.requests, t)
			err := c.New().Pool.Create(context.Background(), test.pool)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_poolClient_List(t *testing.T) {
	t.Parallel()
	const path = "/pools"
	tests := []struct {
		name     string
		output   map[PoolName]string
		requests []mockServer.Request
		err      error
	}{
		{name: `List`,
			output: map[PoolName]string{
				"pool1": "",
				"pool2": "This is pool 2",
				"pool3": ""},
			requests: mockServer.RequestsGetJson(path, map[string]any{
				"data": []map[string]any{
					{"poolid": "pool1"},
					{"poolid": "pool2", "comment": "This is pool 2"},
					{"poolid": "pool3", "comment": ""},
				}})},
		{name: `500 internal server error`,
			requests: mockServer.RequestsError(path, mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			raw, err := c.New().Pool.List(context.Background())
			require.Equal(t, test.err, err)
			if test.err == nil {
				poolMap := raw.AsMap()
				require.Len(t, poolMap, len(test.output))
				for k := range poolMap {
					v, ok := poolMap[k]
					if !ok {
						t.Fatalf("expected (%v) not found", k)
					}
					name, comment := v.Get()
					require.Equal(t, k, name)
					require.Equal(t, test.output[k], comment)
				}
			}
			server.Clear(t)
		})
	}
}

func Test_poolClient_Delete(t *testing.T) {
	t.Parallel()
	const path = "/pools/test_pool"
	tests := []struct {
		name     string
		pool     PoolName
		exists   bool
		requests []mockServer.Request
		err      error
	}{
		{name: `Delete exists`,
			pool:     "test_pool",
			exists:   true,
			requests: mockServer.RequestsDelete(path, map[string]any{})},
		{name: `Delete does not exist`,
			pool: "test_pool",
			requests: mockServer.RequestsErrorHandled(path, mockServer.DELETE, mockServer.HTTPerror{
				Code:    400,
				Message: `{"message":"delete pool failed: pool 'test_pool' does not exist\n"}`})},
		{name: `Delete with members`,
			pool:   "test_pool",
			exists: true,
			requests: mockServer.Append(
				mockServer.RequestsErrorHandled(path, mockServer.DELETE, mockServer.HTTPerror{
					Code:    400,
					Message: `{"message":"delete pool failed: pool 'test_pool' is not empty\n"}`}),
				mockServer.RequestsGetJson(path, map[string]any{"data": map[string]any{
					"members": []any{
						map[string]any{"type": "qemu", "vmid": 100},
						map[string]any{"type": "qemu", "vmid": 200},
						map[string]any{"type": "storage", "storage": "local"}}}}),
				mockServer.RequestsPut(path, map[string]any{
					"delete":  "1",
					"vms":     "100,200",
					"storage": "local"}),
				mockServer.RequestsDelete(path, map[string]any{}))},
		{name: `Delete, error deleting`,
			pool: "test_pool",
			requests: mockServer.Append(
				mockServer.RequestsErrorHandled(path, mockServer.DELETE, mockServer.HTTPerror{
					Code:    400,
					Message: `{"message":"delete pool failed: pool 'test_pool' is not empty\n"}`}),
				mockServer.RequestsGetJson(path, map[string]any{"data": map[string]any{
					"members": []any{
						map[string]any{"type": "qemu", "vmid": 100},
						map[string]any{"type": "qemu", "vmid": 200},
						map[string]any{"type": "storage", "storage": "local"}}}}),
				mockServer.RequestsPut(path, map[string]any{
					"delete":  "1",
					"vms":     "100,200",
					"storage": "local"}),
				mockServer.RequestsError(path, mockServer.DELETE, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `Delete, errors while reading members`,
			pool: "test_pool",
			requests: mockServer.Append(
				mockServer.RequestsErrorHandled(path, mockServer.DELETE, mockServer.HTTPerror{
					Code:    400,
					Message: `{"message":"delete pool failed: pool 'test_pool' is not empty\n"}`}),
				mockServer.RequestsError(path, mockServer.GET, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `Delete, errors while reading members, does not exist`,
			pool: "test_pool",
			requests: mockServer.Append(
				mockServer.RequestsErrorHandled(path, mockServer.DELETE, mockServer.HTTPerror{
					Code:    400,
					Message: `{"message":"delete pool failed: pool 'test_pool' is not empty\n"}`}),
				mockServer.RequestsErrorHandled(path, mockServer.GET, mockServer.HTTPerror{
					Code:    400,
					Message: `{"message":"pool 'test_pool' does not exist\n"}`})),
			err: &ApiError{
				Code:    "400",
				Message: `pool 'test_pool' does not exist`}},
		{name: `Validate error`,
			err: errors.New("PoolName cannot be empty")},
		{name: `500 internal server error`,
			pool:     "test_pool",
			requests: mockServer.RequestsError(path, mockServer.DELETE, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			exists, err := c.New().Pool.Delete(context.Background(), test.pool)
			require.Equal(t, test.err, err)
			require.Equal(t, test.exists, exists)
			server.Clear(t)
		})
	}
}

func Test_poolClient_Exists(t *testing.T) {
	t.Parallel()
	const path = "/pools/test_pool"
	tests := []struct {
		name     string
		pool     PoolName
		exists   bool
		requests []mockServer.Request
		err      error
	}{
		{name: `Exists false`,
			pool: "test_pool",
			requests: mockServer.RequestsErrorHandled(path, mockServer.GET, mockServer.HTTPerror{
				Code:    400,
				Message: `{"message":"pool 'test_pool' does not exist\n"}`,
			})},
		{name: `Exists true`,
			pool:     "test_pool",
			exists:   true,
			requests: mockServer.RequestsGetJson(path, map[string]any{"data": map[string]any{}})},
		{name: `Validate error`,
			err: errors.New("PoolName cannot be empty")},
		{name: `500 internal server error`,
			pool:     "test_pool",
			requests: mockServer.RequestsError(path, mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			exists, err := c.New().Pool.Exists(context.Background(), test.pool)
			require.Equal(t, test.err, err)
			require.Equal(t, test.exists, exists)
			server.Clear(t)
		})
	}
}

func Test_poolClient_Read(t *testing.T) {
	t.Parallel()
	const path = "/pools/test_pool"
	data := func() map[string]any {
		return map[string]any{
			"poolid":  "test_pool",
			"comment": "This is a test pool",
			"members": []any{
				map[string]any{"type": "qemu", "vmid": float64(100)},
				map[string]any{"type": "qemu", "vmid": float64(200)},
				map[string]any{"type": "qemu", "vmid": float64(300)},
				map[string]any{"type": "storage", "storage": "local"},
				map[string]any{"type": "storage", "storage": "nfs-1"},
			}}
	}
	tests := []struct {
		name     string
		pool     PoolName
		info     RawPoolInfo
		requests []mockServer.Request
		err      error
	}{
		{name: `Read`,
			pool:     "test_pool",
			info:     &rawPoolInfo{a: data()},
			requests: mockServer.RequestsGetJson(path, map[string]any{"data": data()})},
		{name: `Validate error`,
			err: errors.New("PoolName cannot be empty")},
		{name: `Pool does not exist`,
			pool: "test_pool",
			requests: mockServer.RequestsErrorHandled(path, mockServer.GET, mockServer.HTTPerror{
				Code:    400,
				Message: `{"message":"pool 'test_pool' does not exist\n"}`}),
			err: &ApiError{
				Code:    "400",
				Message: `pool 'test_pool' does not exist`}},
		{name: `500 internal server error`,
			pool:     "test_pool",
			requests: mockServer.RequestsError(path, mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			c.clearVersion()
			server.Set(test.requests, t)
			raw, err := c.New().Pool.Read(context.Background(), test.pool)
			require.Equal(t, test.err, err)
			require.Equal(t, test.info, raw)
			server.Clear(t)
		})
	}
}

func Test_poolClient_RemoveMembers(t *testing.T) {
	t.Parallel()
	const path = "/pools/test_pool"
	tests := []struct {
		name     string
		pool     PoolName
		guests   []GuestID
		storages []StorageName
		requests []mockServer.Request
		err      error
	}{
		{name: `Nothing`,
			pool: "test_pool"},
		{name: `Remove guest`,
			pool:   "test_pool",
			guests: []GuestID{200},
			requests: mockServer.RequestsPut(path, map[string]any{
				"delete": "1",
				"vms":    "200"})},
		{name: `Remove storage`,
			pool:     "test_pool",
			storages: []StorageName{"nfs-1"},
			requests: mockServer.RequestsPut(path, map[string]any{
				"delete":  "1",
				"storage": "nfs-1"})},
		{name: `Remove guest and storage`,
			pool:     "test_pool",
			guests:   []GuestID{200},
			storages: []StorageName{"nfs-1"},
			requests: mockServer.RequestsPut(path, map[string]any{
				"delete":  "1",
				"vms":     "200",
				"storage": "nfs-1"})},
		{name: `Validate error PoolName`,
			err: errors.New("PoolName cannot be empty")},
		{name: `Validate error GuestID`,
			pool:   "test_pool",
			guests: []GuestID{99},
			err:    errors.New("guestID should be greater than 99")},
		{name: `500 internal server error`,
			pool:     "test_pool",
			guests:   []GuestID{200},
			requests: mockServer.RequestsError(path, mockServer.PUT, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			c.clearVersion()
			server.Set(test.requests, t)
			err := c.New().Pool.RemoveMembers(context.Background(), test.pool, test.guests, test.storages)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_poolClient_Update(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pool     ConfigPool
		requests []mockServer.Request
		err      error
	}{
		{name: `Update nothing`,
			pool: ConfigPool{
				Name: "test_pool"}},
		{name: `Update Comment set`,
			pool: ConfigPool{
				Name:    "test_pool",
				Comment: util.Pointer("New Comment")},
			requests: mockServer.RequestsPut("/pools/test_pool", map[string]any{
				"comment": "New Comment"})},
		{name: `Update Comment empty`,
			pool: ConfigPool{
				Name:    "test_pool",
				Comment: util.Pointer("")},
			requests: mockServer.RequestsPut("/pools/test_pool", map[string]any{
				"comment": ""})},
		{name: `Update add storage and empty comment`,
			pool: ConfigPool{
				Name:     "test_pool",
				Comment:  util.Pointer(""),
				Storages: &[]StorageName{"local"}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid":  "test_pool",
					"comment": "Test Comment",
					"members": []any{}}}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"comment": "",
					"storage": "local"}))},
		{name: `Update no effect`,
			pool: ConfigPool{
				Name:     "test_pool",
				Comment:  util.Pointer("Current Comment"),
				Storages: &[]StorageName{},
				Guests:   &[]GuestID{}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid":  "test_pool",
					"comment": "Current Comment",
					"members": []any{}}}))},
		{name: `Update add guests, version error on get`,
			pool: ConfigPool{
				Name:   "test_pool",
				Guests: &[]GuestID{100, 200}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid": "test_pool",
					"members": []any{map[string]any{
						"type": "qemu",
						"vmid": 200}}}}),
				mockServer.RequestsError("/version", mockServer.GET, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `Update add guests, 8.0-`,
			pool: ConfigPool{
				Name:   "test_pool",
				Guests: &[]GuestID{100, 200}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid": "test_pool",
					"members": []any{map[string]any{
						"type": "qemu",
						"vmid": 200}}}}),
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{
					"data": []map[string]any{
						{"vmid": float64(100), "pool": "original_pool"},
						{"vmid": float64(200), "pool": "test_pool"},
						{"vmid": float64(300), "pool": ""}}}),
				mockServer.RequestsPut("/pools/original_pool", map[string]any{
					"delete": "1",
					"vms":    "100"}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"vms": "100"}))},
		{name: `Update add guests, 8.0+`,
			pool: ConfigPool{
				Name:   "test_pool",
				Guests: &[]GuestID{100, 200}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid": "test_pool",
					"members": []any{map[string]any{
						"type": "qemu",
						"vmid": 200}}}}),
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"allow-move": "1",
					"vms":        "100"}))},
		{name: `Update add guests, 8.0+, version error on put`,
			pool: ConfigPool{
				Name:   "test_pool",
				Guests: &[]GuestID{100, 200}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid": "test_pool",
					"members": []any{map[string]any{
						"type": "qemu",
						"vmid": 200}}}}),
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsError("/pools/test_pool", mockServer.PUT, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `Update remove storage and update comment`,
			pool: ConfigPool{
				Name:     "test_pool",
				Comment:  util.Pointer("test" + body.Symbols),
				Storages: &[]StorageName{"nfs"}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid": "test_pool",
					"members": []any{
						map[string]any{
							"type":    "storage",
							"storage": "local"},
						map[string]any{
							"type":    "storage",
							"storage": "nfs"}}}}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"delete":  "1",
					"comment": "test" + body.Symbols,
					"storage": "local"}))},
		{name: `Update remove guests 8.0-`,
			pool: ConfigPool{
				Name:   "test_pool",
				Guests: &[]GuestID{}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid":  "test_pool",
					"members": []any{map[string]any{"type": "qemu", "vmid": 100}}}}),
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{"delete": "1", "vms": "100"}))},
		{name: `Update remove guests 8.0+`,
			pool: ConfigPool{
				Name:   "test_pool",
				Guests: &[]GuestID{}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid":  "test_pool",
					"members": []any{map[string]any{"type": "qemu", "vmid": 100}}}}),
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{"delete": "1", "vms": "100"}))},
		{name: `Update add and remove guests and storages 8.0-`,
			pool: ConfigPool{
				Name:     "test_pool",
				Guests:   &[]GuestID{200, 300},
				Storages: &[]StorageName{"nfs-1", "ceph-1"}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid": "test_pool",
					"members": []any{
						map[string]any{"type": "qemu", "vmid": 100},
						map[string]any{"type": "qemu", "vmid": 200},
						map[string]any{"type": "storage", "storage": "local"},
						map[string]any{"type": "storage", "storage": "nfs-1"}}}}),
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"delete":  "1",
					"vms":     "100",
					"storage": "local"}),
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{
					"data": []map[string]any{
						{"vmid": float64(100), "pool": ""},
						{"vmid": float64(200), "pool": "test_pool"},
						{"vmid": float64(300), "pool": "original_pool"},
					}}),
				mockServer.RequestsPut("/pools/original_pool", map[string]any{
					"delete": "1",
					"vms":    "300"}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"vms":     "300",
					"storage": "ceph-1"}))},
		{name: `Update add and remove guests and storages 8.0+`,
			pool: ConfigPool{
				Name:     "test_pool",
				Guests:   &[]GuestID{200, 300},
				Storages: &[]StorageName{"nfs-1", "ceph-1"}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid": "test_pool",
					"members": []any{
						map[string]any{"type": "qemu", "vmid": 100},
						map[string]any{"type": "qemu", "vmid": 200},
						map[string]any{"type": "storage", "storage": "local"},
						map[string]any{"type": "storage", "storage": "nfs-1"}}}}),
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"delete":  "1",
					"vms":     "100",
					"storage": "local"}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"allow-move": "1",
					"vms":        "300",
					"storage":    "ceph-1"}))},
		{name: `Error nothing`,
			pool: ConfigPool{Name: ""},
			err:  errors.New("PoolName cannot be empty")},
		{name: `Error no such pool`,
			pool: ConfigPool{
				Name:   "test_pool",
				Guests: &[]GuestID{}},
			requests: mockServer.RequestsErrorHandled("/pools/test_pool", mockServer.GET, mockServer.HTTPerror{
				Code:    400,
				Message: `{"message":"pool 'test_pool' does not exist\n"}`}),
			err: &ApiError{
				Code:    "400",
				Message: `pool 'test_pool' does not exist`}},
		{name: `500 internal server error while reading`,
			pool: ConfigPool{
				Name:   "test_pool",
				Guests: &[]GuestID{}},
			requests: mockServer.RequestsError("/pools/test_pool", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			c.clearVersion()
			server.Set(test.requests, t)
			err := c.New().Pool.Update(context.Background(), test.pool)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_poolClient_Set(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pool     ConfigPool
		requests []mockServer.Request
		err      error
	}{
		{name: `Create`,
			pool: ConfigPool{
				Name:    "test_pool",
				Comment: util.Pointer("")},
			requests: mockServer.Append(
				mockServer.RequestsErrorHandled("/pools/test_pool", mockServer.GET, mockServer.HTTPerror{
					Code:    400,
					Message: `{"message":"pool 'test_pool' does not exist\n"}`}),
				mockServer.RequestsPost("/pools", map[string]any{
					"poolid": "test_pool"}))},
		{name: `Create with comment`,
			pool: ConfigPool{
				Name:    "test_pool",
				Comment: util.Pointer("This is a test pool" + body.Symbols)},
			requests: mockServer.Append(
				mockServer.RequestsErrorHandled("/pools/test_pool", mockServer.GET, mockServer.HTTPerror{
					Code:    400,
					Message: `{"message":"pool 'test_pool' does not exist\n"}`}),
				mockServer.RequestsPost("/pools", map[string]any{
					"poolid":  "test_pool",
					"comment": "This is a test pool" + body.Symbols}))},
		{name: `Create with guests 8-`,
			pool: ConfigPool{
				Name:   "test_pool",
				Guests: &[]GuestID{100}},
			requests: mockServer.Append(
				mockServer.RequestsErrorHandled("/pools/test_pool", mockServer.GET, mockServer.HTTPerror{
					Code:    400,
					Message: `{"message":"pool 'test_pool' does not exist\n"}`}),
				mockServer.RequestsPost("/pools", map[string]any{"poolid": "test_pool"}),
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "pool": ""}}}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"vms": "100"}))},
		{name: `Create with guests 8+`,
			pool: ConfigPool{
				Name:   "test_pool",
				Guests: &[]GuestID{100, 200, 300}},
			requests: mockServer.Append(
				mockServer.RequestsErrorHandled("/pools/test_pool", mockServer.GET, mockServer.HTTPerror{
					Code:    400,
					Message: `{"message":"pool 'test_pool' does not exist\n"}`}),
				mockServer.RequestsPost("/pools", map[string]any{"poolid": "test_pool"}),
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"allow-move": "1",
					"vms":        "100,200,300"}))},
		{name: `Create with storages`,
			pool: ConfigPool{
				Name:     "test_pool",
				Storages: &[]StorageName{"local", "nfs-1"}},
			requests: mockServer.Append(
				mockServer.RequestsErrorHandled("/pools/test_pool", mockServer.GET, mockServer.HTTPerror{
					Code:    400,
					Message: `{"message":"pool 'test_pool' does not exist\n"}`}),
				mockServer.RequestsPost("/pools", map[string]any{"poolid": "test_pool"}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"storage": "local,nfs-1"}))},
		{name: `Create full 8-`,
			pool: ConfigPool{
				Name:     "test_pool",
				Comment:  util.Pointer("test Comment"),
				Guests:   &[]GuestID{100},
				Storages: &[]StorageName{"local", "nfs-1"}},
			requests: mockServer.Append(
				mockServer.RequestsErrorHandled("/pools/test_pool", mockServer.GET, mockServer.HTTPerror{
					Code:    400,
					Message: `{"message":"pool 'test_pool' does not exist\n"}`}),
				mockServer.RequestsPost("/pools", map[string]any{
					"poolid":  "test_pool",
					"comment": "test Comment"}),
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "pool": "original_pool"}}}),
				mockServer.RequestsPut("/pools/original_pool", map[string]any{
					"delete": "1",
					"vms":    "100"}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"vms":     "100",
					"storage": "local,nfs-1"}))},
		{name: `Create full 8+`,
			pool: ConfigPool{
				Name:     "test_pool",
				Comment:  util.Pointer("test Comment"),
				Guests:   &[]GuestID{100},
				Storages: &[]StorageName{"local", "nfs-1"}},
			requests: mockServer.Append(
				mockServer.RequestsErrorHandled("/pools/test_pool", mockServer.GET, mockServer.HTTPerror{
					Code:    400,
					Message: `{"message":"pool 'test_pool' does not exist\n"}`}),
				mockServer.RequestsPost("/pools", map[string]any{
					"poolid":  "test_pool",
					"comment": "test Comment"}),
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"allow-move": "1",
					"vms":        "100",
					"storage":    "local,nfs-1"}))},
		{name: `Update Comment empty`,
			pool: ConfigPool{
				Name:    "test_pool",
				Comment: util.Pointer("")},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid":  "test_pool",
					"comment": "Old Comment"}}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"comment": ""}))},
		{name: `Update Comment no change`,
			pool: ConfigPool{
				Name:    "test_pool",
				Comment: util.Pointer("Current Comment")},
			requests: mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
				"poolid":  "test_pool",
				"comment": "Current Comment"}})},
		{name: `Update Comment set`,
			pool: ConfigPool{
				Name:    "test_pool",
				Comment: util.Pointer("New Comment")},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid":  "test_pool",
					"comment": "Old Comment"}}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"comment": "New Comment"}))},
		{name: `Update add and remove guests 8-`,
			pool: ConfigPool{
				Name:   "test_pool",
				Guests: &[]GuestID{200, 300}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid": "test_pool",
					"members": []any{
						map[string]any{"type": "qemu", "vmid": 100},
						map[string]any{"type": "qemu", "vmid": 200}}}}),
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"delete": "1",
					"vms":    "100"}),
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{
					"data": []map[string]any{
						{"vmid": float64(100), "pool": ""},
						{"vmid": float64(200), "pool": "test_pool"},
						{"vmid": float64(300), "pool": "original_pool"}}}),
				mockServer.RequestsPut("/pools/original_pool", map[string]any{
					"delete": "1",
					"vms":    "300"}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"vms": "300"}))},
		{name: `Update add and remove guests 8+`,
			pool: ConfigPool{
				Name:   "test_pool",
				Guests: &[]GuestID{200, 300}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid": "test_pool",
					"members": []any{
						map[string]any{"type": "qemu", "vmid": 100},
						map[string]any{"type": "qemu", "vmid": 200}}}}),
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"delete": "1",
					"vms":    "100"}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"allow-move": "1",
					"vms":        "300"}))},
		{name: `Update add and remove storages`,
			pool: ConfigPool{
				Name:     "test_pool",
				Storages: &[]StorageName{"nfs-1", "ceph-1"}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid": "test_pool",
					"members": []any{
						map[string]any{"type": "storage", "storage": "local"},
						map[string]any{"type": "storage", "storage": "nfs-1"}}}}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"delete":  "1",
					"storage": "local"}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"storage": "ceph-1"}))},
		{name: `Update full 8-`,
			pool: ConfigPool{
				Name:     "test_pool",
				Comment:  util.Pointer("test" + body.Symbols),
				Guests:   &[]GuestID{200, 300},
				Storages: &[]StorageName{"nfs-1", "ceph-1"}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid": "test_pool",
					"members": []any{
						map[string]any{"type": "qemu", "vmid": 100},
						map[string]any{"type": "qemu", "vmid": 200},
						map[string]any{"type": "storage", "storage": "local"},
						map[string]any{"type": "storage", "storage": "nfs-1"}}}}),
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"delete":  "1",
					"comment": "test" + body.Symbols,
					"vms":     "100",
					"storage": "local"}),
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{
					"data": []map[string]any{
						{"vmid": float64(100), "pool": ""},
						{"vmid": float64(200), "pool": "test_pool"},
						{"vmid": float64(300), "pool": "original_pool"}}}),
				mockServer.RequestsPut("/pools/original_pool", map[string]any{
					"delete": "1",
					"vms":    "300"}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"vms":     "300",
					"storage": "ceph-1"}))},
		{name: `Update full 8+`,
			pool: ConfigPool{
				Name:     "test_pool",
				Comment:  util.Pointer("test" + body.Symbols),
				Guests:   &[]GuestID{200, 300},
				Storages: &[]StorageName{"nfs-1", "ceph-1"}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/pools/test_pool", map[string]any{"data": map[string]any{
					"poolid": "test_pool",
					"members": []any{
						map[string]any{"type": "qemu", "vmid": 100},
						map[string]any{"type": "qemu", "vmid": 200},
						map[string]any{"type": "storage", "storage": "local"},
						map[string]any{"type": "storage", "storage": "nfs-1"}}}}),
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"delete":  "1",
					"comment": "test" + body.Symbols,
					"vms":     "100",
					"storage": "local"}),
				mockServer.RequestsPut("/pools/test_pool", map[string]any{
					"allow-move": "1",
					"vms":        "300",
					"storage":    "ceph-1"}))},
		{name: `Validate error`,
			err: errors.New("PoolName cannot be empty")},
		{name: `500 internal server error`,
			pool:     ConfigPool{Name: "test_pool"},
			requests: mockServer.RequestsError("/pools/test_pool", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			c.clearVersion()
			server.Set(test.requests, t)
			err := c.New().Pool.Set(context.Background(), test.pool)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_ConfigPool_mapToApi(t *testing.T) {
	t.Parallel()
	type testInput struct {
		new     ConfigPool
		current *ConfigPool
	}
	tests := []struct {
		name   string
		input  testInput
		output map[string]any
	}{
		{name: `Create Full`,
			input: testInput{
				new: ConfigPool{
					Name:    "test",
					Comment: util.Pointer("test-comment"),
					Guests:  &[]GuestID{100, 300, 200}}},
			output: map[string]any{
				"poolid":  "test",
				"comment": "test-comment"}},
		{name: `Create poolid`,
			input: testInput{
				new: ConfigPool{Name: "test"}},
			output: map[string]any{"poolid": "test"}},
		{name: `Create comment`,
			input: testInput{
				new: ConfigPool{Comment: util.Pointer("test-comment")}},
			output: map[string]any{
				"poolid":  "",
				"comment": "test-comment"}},
		{name: `Create members`,
			input: testInput{
				new: ConfigPool{Guests: &[]GuestID{100, 300, 200}}},
			output: map[string]any{"poolid": ""}},
		{name: `Update Full`,
			input: testInput{
				new: ConfigPool{
					Name:    "test",
					Comment: util.Pointer("test-comment"),
					Guests:  &[]GuestID{100, 300, 200}},
				current: &ConfigPool{
					Name:    "test",
					Comment: util.Pointer("old-comment"),
					Guests:  &[]GuestID{100, 300}}},
			output: map[string]any{
				"comment": "test-comment"}},
		{name: `Update poolid`,
			input: testInput{
				new:     ConfigPool{Name: "test"},
				current: &ConfigPool{Name: "old"}},
			output: map[string]any{}},
		{name: `Update comment`,
			input: testInput{
				new:     ConfigPool{Comment: util.Pointer("test-comment")},
				current: &ConfigPool{Comment: util.Pointer("old-comment")}},
			output: map[string]any{
				"comment": "test-comment"}},
		{name: `Update commen samet`,
			input: testInput{
				new:     ConfigPool{Comment: util.Pointer("test-comment")},
				current: &ConfigPool{Comment: util.Pointer("test-comment")}},
			output: map[string]any{}},
		{name: `Update members`,
			input: testInput{
				new:     ConfigPool{Guests: &[]GuestID{100, 300, 200}},
				current: &ConfigPool{Guests: &[]GuestID{100, 300}}},
			output: map[string]any{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.new.mapToApi(test.input.current))
		})
	}
}

func Test_RawConfigPool_Get(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   map[string]any
		poolPtr *PoolName
		pool    PoolName
		comment string
	}{
		{name: "Empty",
			input: map[string]any{}},
		{name: "pool name map",
			input: map[string]any{"poolid": "test"},
			pool:  "test"},
		{name: "pool name ptr",
			input:   map[string]any{},
			poolPtr: util.Pointer(PoolName("test")),
			pool:    "test"},
		{name: "comment only",
			input:   map[string]any{"comment": "test"},
			comment: "test"},
		{name: "all",
			input: map[string]any{
				"poolid":  "test",
				"comment": "my comment"},
			pool:    "test",
			comment: "my comment"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pool, comment := (&rawConfigPool{a: test.input, pool: test.poolPtr}).Get()
			require.Equal(t, test.pool, pool)
			require.Equal(t, test.comment, comment)
		})
	}
}

func Test_ConfigPool_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  ConfigPool
		output error
	}{
		{name: "Valid PoolName",
			input: ConfigPool{Name: PoolName(test_data_pool.PoolName_Legal())}},
		{name: "Invalid PoolName Empty",
			input:  ConfigPool{Name: ""},
			output: errors.New(PoolName_Error_Empty)},
		{name: "Invalid PoolName Length",
			input:  ConfigPool{Name: PoolName(test_data_pool.PoolName_Max_Illegal())},
			output: errors.New(PoolName_Error_Length)},
		{name: "Invalid PoolName Characters",
			input:  ConfigPool{Name: PoolName(test_data_pool.PoolName_Error_Characters()[0])},
			output: errors.New(PoolName_Error_Characters)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_RawPools_AsArray(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawPools
		output []RawConfigPool
	}{
		{name: `Empty`,
			input:  rawPools{a: []any{}},
			output: []RawConfigPool{}},
		{name: `Single Pool`,
			input: rawPools{a: []any{
				map[string]any{"poolid": "pool1"}}},
			output: []RawConfigPool{
				&rawConfigPool{a: map[string]any{"poolid": "pool1"}}}},
		{name: `Single Pool with Comment`,
			input: rawPools{a: []any{
				map[string]any{"poolid": "pool1", "comment": "Test pool"}}},
			output: []RawConfigPool{
				&rawConfigPool{a: map[string]any{"poolid": "pool1", "comment": "Test pool"}}}},
		{name: `Multiple Pools`,
			input: rawPools{a: []any{
				map[string]any{"poolid": "pool1"},
				map[string]any{"poolid": "pool2", "comment": "Second pool"},
				map[string]any{"poolid": "pool3", "comment": ""}}},
			output: []RawConfigPool{
				&rawConfigPool{a: map[string]any{"poolid": "pool1"}},
				&rawConfigPool{a: map[string]any{"poolid": "pool2", "comment": "Second pool"}},
				&rawConfigPool{a: map[string]any{"poolid": "pool3", "comment": ""}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, RawPools(&test.input).AsArray())
		})
	}
}

func Test_RawPools_AsMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawPools
		output map[PoolName]RawConfigPool
	}{
		{name: `Empty`,
			input:  rawPools{a: []any{}},
			output: map[PoolName]RawConfigPool{}},
		{name: `Single Pool`,
			input: rawPools{a: []any{
				map[string]any{"poolid": "pool1"}}},
			output: map[PoolName]RawConfigPool{
				"pool1": &rawConfigPool{
					a:    map[string]any{"poolid": "pool1"},
					pool: util.Pointer(PoolName("pool1"))}}},
		{name: `Single Pool with Comment`,
			input: rawPools{a: []any{
				map[string]any{"poolid": "pool1", "comment": "Test pool"}}},
			output: map[PoolName]RawConfigPool{
				"pool1": &rawConfigPool{
					a:    map[string]any{"poolid": "pool1", "comment": "Test pool"},
					pool: util.Pointer(PoolName("pool1"))}}},
		{name: `Multiple Pools`,
			input: rawPools{a: []any{
				map[string]any{"poolid": "pool1"},
				map[string]any{"poolid": "pool2", "comment": "Second pool"},
				map[string]any{"poolid": "pool3", "comment": ""}}},
			output: map[PoolName]RawConfigPool{
				"pool1": &rawConfigPool{
					a:    map[string]any{"poolid": "pool1"},
					pool: util.Pointer(PoolName("pool1"))},
				"pool2": &rawConfigPool{
					a:    map[string]any{"poolid": "pool2", "comment": "Second pool"},
					pool: util.Pointer(PoolName("pool2"))},
				"pool3": &rawConfigPool{
					a:    map[string]any{"poolid": "pool3", "comment": ""},
					pool: util.Pointer(PoolName("pool3"))}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, RawPools(&test.input).AsMap())
		})
	}
}

func Test_RawPools_Iter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawPools
		output []RawConfigPool
	}{
		{name: `Empty`,
			input:  rawPools{a: []any{}},
			output: []RawConfigPool{}},
		{name: `Single Pool`,
			input: rawPools{a: []any{
				map[string]any{"poolid": "pool1"}}},
			output: []RawConfigPool{
				&rawConfigPool{a: map[string]any{"poolid": "pool1"}}}},
		{name: `Single Pool with Comment`,
			input: rawPools{a: []any{
				map[string]any{"poolid": "pool1", "comment": "Test pool"}}},
			output: []RawConfigPool{
				&rawConfigPool{a: map[string]any{"poolid": "pool1", "comment": "Test pool"}}}},
		{name: `Multiple Pools`,
			input: rawPools{a: []any{
				map[string]any{"poolid": "pool1"},
				map[string]any{"poolid": "pool2", "comment": "Second pool"},
				map[string]any{"poolid": "pool3", "comment": ""}}},
			output: []RawConfigPool{
				&rawConfigPool{a: map[string]any{"poolid": "pool1"}},
				&rawConfigPool{a: map[string]any{"poolid": "pool2", "comment": "Second pool"}},
				&rawConfigPool{a: map[string]any{"poolid": "pool3", "comment": ""}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Test iterating over all items
			var result []RawConfigPool
			for pool := range RawPools(&test.input).Iter() {
				result = append(result, pool)
			}
			require.Equal(t, len(test.output), len(result))
			for i := range result {
				name, comment := result[i].Get()
				expectedName, expectedComment := test.output[i].Get()
				require.Equal(t, expectedName, name)
				require.Equal(t, expectedComment, comment)
			}
			// Test early termination (break after first item)
			if len(test.output) > 0 {
				count := 0
				for range RawPools(&test.input).Iter() {
					count++
					break
				}
				require.Equal(t, 1, count)
			}
		})
	}
}

func Test_RawPools_Len(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawPools
		output int
	}{
		{name: `Empty`,
			input:  rawPools{a: []any{}},
			output: 0},
		{name: `Single Pool`,
			input: rawPools{a: []any{
				map[string]any{"poolid": "pool1"}}},
			output: 1},
		{name: `Multiple Pools`,
			input: rawPools{a: []any{
				map[string]any{"poolid": "pool1"},
				map[string]any{"poolid": "pool2"},
				map[string]any{"poolid": "pool3"}}},
			output: 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, RawPools(&test.input).Len())
		})
	}
}

func Test_RawPoolInfo_Get(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawPoolInfo
		output PoolInfo
	}{
		{name: `Empty Pool Info`,
			input: rawPoolInfo{a: map[string]any{}},
			output: PoolInfo{
				Members: RawPoolMembers(&rawPoolMembers{a: []any{}})}},
		{name: `Pool with no comment or members`,
			input: rawPoolInfo{a: map[string]any{"poolid": "test-pool"}},
			output: PoolInfo{
				Name:    "test-pool",
				Members: RawPoolMembers(&rawPoolMembers{a: []any{}})}},
		{name: `Pool with comment`,
			input: rawPoolInfo{a: map[string]any{"poolid": "prod-pool", "comment": "Production environment"}},
			output: PoolInfo{
				Name:    "prod-pool",
				Comment: "Production environment",
				Members: RawPoolMembers(&rawPoolMembers{a: []any{}})}},
		{name: `Pool with members but no comment`,
			input: rawPoolInfo{a: map[string]any{
				"poolid": "dev-pool",
				"members": []any{
					map[string]any{"type": "qemu", "vmid": float64(100)}}}},
			output: PoolInfo{
				Name:    "dev-pool",
				Members: RawPoolMembers(&rawPoolMembers{a: []any{map[string]any{"type": "qemu", "vmid": float64(100)}}})}},
		{name: `Pool with comment and members`,
			input: rawPoolInfo{a: map[string]any{
				"poolid":  "backup-pool",
				"comment": "Backup VMs",
				"members": []any{
					map[string]any{"type": "qemu", "vmid": float64(100)},
					map[string]any{"type": "storage", "storage": "backup-storage"}}}},
			output: PoolInfo{
				Name:    "backup-pool",
				Comment: "Backup VMs",
				Members: RawPoolMembers(&rawPoolMembers{a: []any{
					map[string]any{"type": "qemu", "vmid": float64(100)},
					map[string]any{"type": "storage", "storage": "backup-storage"}}})}},
		{name: `Pool with empty comment`,
			input: rawPoolInfo{a: map[string]any{"poolid": "empty-comment", "comment": ""}},
			output: PoolInfo{
				Name:    "empty-comment",
				Members: RawPoolMembers(&rawPoolMembers{a: []any{}})}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, RawPoolInfo(&test.input).Get())
		})
	}
}

func Test_RawPoolMembers_AsArray(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  RawPoolInfo
		output []RawPoolMember
	}{
		{name: `Empty members`,
			input:  &rawPoolInfo{a: map[string]any{"poolid": "test"}},
			output: []RawPoolMember{}},
		{name: `Single guest member`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "qemu", "vmid": float64(100)}}}},
			output: []RawPoolMember{
				&rawPoolMember{a: map[string]any{"type": "qemu", "vmid": float64(100)}}}},
		{name: `Single storage member`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "storage", "storage": "local"}}}},
			output: []RawPoolMember{
				&rawPoolMember{a: map[string]any{"type": "storage", "storage": "local"}}}},
		{name: `Multiple guests`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "qemu", "vmid": float64(100)},
					map[string]any{"type": "qemu", "vmid": float64(200)},
					map[string]any{"type": "lxc", "vmid": float64(300)}}}},
			output: []RawPoolMember{
				&rawPoolMember{a: map[string]any{"type": "qemu", "vmid": float64(100)}},
				&rawPoolMember{a: map[string]any{"type": "qemu", "vmid": float64(200)}},
				&rawPoolMember{a: map[string]any{"type": "lxc", "vmid": float64(300)}}}},
		{name: `Multiple storages`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "storage", "storage": "local"},
					map[string]any{"type": "storage", "storage": "nfs-1"},
					map[string]any{"type": "storage", "storage": "ceph-1"}}}},
			output: []RawPoolMember{
				&rawPoolMember{a: map[string]any{"type": "storage", "storage": "local"}},
				&rawPoolMember{a: map[string]any{"type": "storage", "storage": "nfs-1"}},
				&rawPoolMember{a: map[string]any{"type": "storage", "storage": "ceph-1"}}}},
		{name: `Mixed guests and storages`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "qemu", "vmid": float64(100)},
					map[string]any{"type": "storage", "storage": "local"},
					map[string]any{"type": "qemu", "vmid": float64(200)},
					map[string]any{"type": "storage", "storage": "nfs-1"},
					map[string]any{"type": "lxc", "vmid": float64(300)}}}},
			output: []RawPoolMember{
				&rawPoolMember{a: map[string]any{"type": "qemu", "vmid": float64(100)}},
				&rawPoolMember{a: map[string]any{"type": "storage", "storage": "local"}},
				&rawPoolMember{a: map[string]any{"type": "qemu", "vmid": float64(200)}},
				&rawPoolMember{a: map[string]any{"type": "storage", "storage": "nfs-1"}},
				&rawPoolMember{a: map[string]any{"type": "lxc", "vmid": float64(300)}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.input.GetMembers().AsArray()
			require.Equal(t, len(test.output), len(result))
			for i := range result {
				require.Equal(t, test.output[i].Type(), result[i].Type())
				// Verify guest members
				if expectedGuest, ok := test.output[i].AsGuest(); ok {
					actualGuest, ok := result[i].AsGuest()
					require.True(t, ok)
					require.Equal(t, expectedGuest.GetID(), actualGuest.GetID())
				}
				// Verify storage members
				if expectedStorage, ok := test.output[i].AsStorage(); ok {
					actualStorage, ok := result[i].AsStorage()
					require.True(t, ok)
					require.Equal(t, expectedStorage.GetName(), actualStorage.GetName())
				}
			}
		})
	}
}

func Test_RawPoolMembers_AsArrays(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		input         RawPoolInfo
		outputGuest   []RawPoolGuest
		outputStorage []RawPoolStorage
	}{
		{name: `Empty members`,
			input:         &rawPoolInfo{a: map[string]any{"poolid": "test"}},
			outputGuest:   []RawPoolGuest{},
			outputStorage: []RawPoolStorage{}},
		{name: `Single guest member`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "qemu", "vmid": float64(100)}}}},
			outputGuest: []RawPoolGuest{
				&rawPoolGuest{a: map[string]any{"type": "qemu", "vmid": float64(100)}}},
			outputStorage: []RawPoolStorage{}},
		{name: `Single storage member`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "storage", "storage": "local"}}}},
			outputGuest: []RawPoolGuest{},
			outputStorage: []RawPoolStorage{
				&rawPoolStorage{a: map[string]any{"type": "storage", "storage": "local"}}}},
		{name: `Multiple guests`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "qemu", "vmid": float64(100)},
					map[string]any{"type": "qemu", "vmid": float64(200)},
					map[string]any{"type": "lxc", "vmid": float64(300)}}}},
			outputGuest: []RawPoolGuest{
				&rawPoolGuest{a: map[string]any{"type": "qemu", "vmid": float64(100)}},
				&rawPoolGuest{a: map[string]any{"type": "qemu", "vmid": float64(200)}},
				&rawPoolGuest{a: map[string]any{"type": "lxc", "vmid": float64(300)}}},
			outputStorage: []RawPoolStorage{}},
		{name: `Multiple storages`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "storage", "storage": "local"},
					map[string]any{"type": "storage", "storage": "nfs-1"},
					map[string]any{"type": "storage", "storage": "ceph-1"}}}},
			outputGuest: []RawPoolGuest{},
			outputStorage: []RawPoolStorage{
				&rawPoolStorage{a: map[string]any{"type": "storage", "storage": "local"}},
				&rawPoolStorage{a: map[string]any{"type": "storage", "storage": "nfs-1"}},
				&rawPoolStorage{a: map[string]any{"type": "storage", "storage": "ceph-1"}}}},
		{name: `Mixed guests and storages`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "qemu", "vmid": float64(100)},
					map[string]any{"type": "storage", "storage": "local"},
					map[string]any{"type": "qemu", "vmid": float64(200)},
					map[string]any{"type": "storage", "storage": "nfs-1"},
					map[string]any{"type": "lxc", "vmid": float64(300)}}}},
			outputGuest: []RawPoolGuest{
				&rawPoolGuest{a: map[string]any{"type": "qemu", "vmid": float64(100)}},
				&rawPoolGuest{a: map[string]any{"type": "qemu", "vmid": float64(200)}},
				&rawPoolGuest{a: map[string]any{"type": "lxc", "vmid": float64(300)}}},
			outputStorage: []RawPoolStorage{
				&rawPoolStorage{a: map[string]any{"type": "storage", "storage": "local"}},
				&rawPoolStorage{a: map[string]any{"type": "storage", "storage": "nfs-1"}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			guests, storages := test.input.GetMembers().AsArrays()
			require.Equal(t, len(test.outputGuest), len(guests))
			require.Equal(t, len(test.outputStorage), len(storages))
			// Verify guests
			for i := range guests {
				require.Equal(t, test.outputGuest[i].GetID(), guests[i].GetID())
			}
			// Verify storages
			for i := range storages {
				require.Equal(t, test.outputStorage[i].GetName(), storages[i].GetName())
			}
		})
	}
}

func Test_RawPoolMembers_AsMaps(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		input         RawPoolInfo
		outputGuest   map[GuestID]RawPoolGuest
		outputStorage map[StorageName]RawPoolStorage
	}{
		{name: `Empty members`,
			input:         &rawPoolInfo{a: map[string]any{"poolid": "test"}},
			outputGuest:   map[GuestID]RawPoolGuest{},
			outputStorage: map[StorageName]RawPoolStorage{}},
		{name: `Single guest member`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "qemu", "vmid": float64(100)}}}},
			outputGuest: map[GuestID]RawPoolGuest{
				100: &rawPoolGuest{
					a:  map[string]any{"type": "qemu", "vmid": float64(100)},
					id: util.Pointer(GuestID(100))}},
			outputStorage: map[StorageName]RawPoolStorage{}},
		{name: `Single storage member`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "storage", "storage": "local"}}}},
			outputGuest: map[GuestID]RawPoolGuest{},
			outputStorage: map[StorageName]RawPoolStorage{
				"local": &rawPoolStorage{
					a:    map[string]any{"type": "storage", "storage": "local"},
					name: util.Pointer(StorageName("local"))}}},
		{name: `Multiple guests`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "qemu", "vmid": float64(100)},
					map[string]any{"type": "qemu", "vmid": float64(200)},
					map[string]any{"type": "lxc", "vmid": float64(300)}}}},
			outputGuest: map[GuestID]RawPoolGuest{
				100: &rawPoolGuest{
					a:  map[string]any{"type": "qemu", "vmid": float64(100)},
					id: util.Pointer(GuestID(100))},
				200: &rawPoolGuest{
					a:  map[string]any{"type": "qemu", "vmid": float64(200)},
					id: util.Pointer(GuestID(200))},
				300: &rawPoolGuest{
					a:  map[string]any{"type": "lxc", "vmid": float64(300)},
					id: util.Pointer(GuestID(300))}},
			outputStorage: map[StorageName]RawPoolStorage{}},
		{name: `Multiple storages`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "storage", "storage": "local"},
					map[string]any{"type": "storage", "storage": "nfs-1"},
					map[string]any{"type": "storage", "storage": "ceph-1"}}}},
			outputGuest: map[GuestID]RawPoolGuest{},
			outputStorage: map[StorageName]RawPoolStorage{
				"local": &rawPoolStorage{
					a:    map[string]any{"type": "storage", "storage": "local"},
					name: util.Pointer(StorageName("local"))},
				"nfs-1": &rawPoolStorage{
					a:    map[string]any{"type": "storage", "storage": "nfs-1"},
					name: util.Pointer(StorageName("nfs-1"))},
				"ceph-1": &rawPoolStorage{
					a:    map[string]any{"type": "storage", "storage": "ceph-1"},
					name: util.Pointer(StorageName("ceph-1"))}}},
		{name: `Mixed guests and storages`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "qemu", "vmid": float64(100)},
					map[string]any{"type": "storage", "storage": "local"},
					map[string]any{"type": "qemu", "vmid": float64(200)},
					map[string]any{"type": "storage", "storage": "nfs-1"},
					map[string]any{"type": "lxc", "vmid": float64(300)}}}},
			outputGuest: map[GuestID]RawPoolGuest{
				100: &rawPoolGuest{
					a:  map[string]any{"type": "qemu", "vmid": float64(100)},
					id: util.Pointer(GuestID(100))},
				200: &rawPoolGuest{
					a:  map[string]any{"type": "qemu", "vmid": float64(200)},
					id: util.Pointer(GuestID(200))},
				300: &rawPoolGuest{
					a:  map[string]any{"type": "lxc", "vmid": float64(300)},
					id: util.Pointer(GuestID(300))}},
			outputStorage: map[StorageName]RawPoolStorage{
				"local": &rawPoolStorage{
					a:    map[string]any{"type": "storage", "storage": "local"},
					name: util.Pointer(StorageName("local"))},
				"nfs-1": &rawPoolStorage{
					a:    map[string]any{"type": "storage", "storage": "nfs-1"},
					name: util.Pointer(StorageName("nfs-1"))}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			guests, storages := test.input.GetMembers().AsMaps()
			require.Len(t, guests, len(test.outputGuest))
			require.Len(t, storages, len(test.outputStorage))
			require.Equal(t, test.outputGuest, guests)
			require.Equal(t, test.outputStorage, storages)
		})
	}
}

func Test_RawPoolMembers_Iter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  RawPoolInfo
		output []RawPoolMember
	}{
		{name: `Empty members`,
			input:  &rawPoolInfo{a: map[string]any{"poolid": "test"}},
			output: []RawPoolMember{}},
		{name: `Single guest member`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "qemu", "vmid": float64(100)}}}},
			output: []RawPoolMember{
				&rawPoolMember{a: map[string]any{"type": "qemu", "vmid": float64(100)}}}},
		{name: `Single storage member`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "storage", "storage": "local"}}}},
			output: []RawPoolMember{
				&rawPoolMember{a: map[string]any{"type": "storage", "storage": "local"}}}},
		{name: `Multiple guests`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "qemu", "vmid": float64(100)},
					map[string]any{"type": "qemu", "vmid": float64(200)},
					map[string]any{"type": "lxc", "vmid": float64(300)}}}},
			output: []RawPoolMember{
				&rawPoolMember{a: map[string]any{"type": "qemu", "vmid": float64(100)}},
				&rawPoolMember{a: map[string]any{"type": "qemu", "vmid": float64(200)}},
				&rawPoolMember{a: map[string]any{"type": "lxc", "vmid": float64(300)}}}},
		{name: `Multiple storages`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "storage", "storage": "local"},
					map[string]any{"type": "storage", "storage": "nfs-1"},
					map[string]any{"type": "storage", "storage": "ceph-1"}}}},
			output: []RawPoolMember{
				&rawPoolMember{a: map[string]any{"type": "storage", "storage": "local"}},
				&rawPoolMember{a: map[string]any{"type": "storage", "storage": "nfs-1"}},
				&rawPoolMember{a: map[string]any{"type": "storage", "storage": "ceph-1"}}}},
		{name: `Mixed guests and storages`,
			input: &rawPoolInfo{a: map[string]any{
				"poolid": "test",
				"members": []any{
					map[string]any{"type": "qemu", "vmid": float64(100)},
					map[string]any{"type": "storage", "storage": "local"},
					map[string]any{"type": "qemu", "vmid": float64(200)},
					map[string]any{"type": "storage", "storage": "nfs-1"},
					map[string]any{"type": "lxc", "vmid": float64(300)}}}},
			output: []RawPoolMember{
				&rawPoolMember{a: map[string]any{"type": "qemu", "vmid": float64(100)}},
				&rawPoolMember{a: map[string]any{"type": "storage", "storage": "local"}},
				&rawPoolMember{a: map[string]any{"type": "qemu", "vmid": float64(200)}},
				&rawPoolMember{a: map[string]any{"type": "storage", "storage": "nfs-1"}},
				&rawPoolMember{a: map[string]any{"type": "lxc", "vmid": float64(300)}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Test full iteration
			var members []RawPoolMember
			for member := range test.input.GetMembers().Iter() {
				members = append(members, member)
			}
			require.Equal(t, len(test.output), len(members))
			for i := range members {
				require.Equal(t, test.output[i].Type(), members[i].Type())
			}
			// Test early termination
			if len(test.output) > 1 {
				count := 0
				for range test.input.GetMembers().Iter() {
					count++
					if count == 1 {
						break
					}
				}
				require.Equal(t, 1, count)
			}
		})
	}
}

func Test_RawPoolMembers_Len(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawPoolMembers
		output int
	}{
		{name: `Empty members`,
			input:  rawPoolMembers{a: []any{}},
			output: 0},
		{name: `Single member`,
			input: rawPoolMembers{a: []any{
				map[string]any{"type": "qemu", "vmid": float64(100)}}},
			output: 1},
		{name: `Multiple members`,
			input: rawPoolMembers{a: []any{
				map[string]any{"type": "qemu", "vmid": float64(100)},
				map[string]any{"type": "storage", "storage": "local"},
				map[string]any{"type": "lxc", "vmid": float64(200)}}},
			output: 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, RawPoolMembers(&test.input).Len())
		})
	}
}

func Test_RawPoolGuest_GetID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawPoolGuest
		output GuestID
	}{
		{name: `not set`},
		{name: `parse map`,
			input:  rawPoolGuest{a: map[string]any{"vmid": float64(100)}},
			output: 100},
		{name: `parse pointer`,
			input: rawPoolGuest{
				a:  map[string]any{},
				id: util.Pointer(GuestID(200))},
			output: 200},
		{name: `prefer pointer over map`,
			input: rawPoolGuest{
				a:  map[string]any{"vmid": float64(100)},
				id: util.Pointer(GuestID(200))},
			output: 200},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, RawPoolGuest(&test.input).GetID())
		})
	}
}

func Test_RawPoolStorage_GetName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawPoolStorage
		output StorageName
	}{
		{name: `not set`},
		{name: `parse map`,
			input:  rawPoolStorage{a: map[string]any{"storage": "local"}},
			output: "local"},
		{name: `parse pointer`,
			input: rawPoolStorage{
				a:    map[string]any{},
				name: util.Pointer(StorageName("nfs-1"))},
			output: "nfs-1"},
		{name: `prefer pointer over map`,
			input: rawPoolStorage{
				a:    map[string]any{"storage": "local"},
				name: util.Pointer(StorageName("nfs-1"))},
			output: "nfs-1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, RawPoolStorage(&test.input).GetName())
		})
	}
}

func Test_PoolName_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  []string
		output error
	}{
		{name: `Valid PoolName`,
			input: test_data_pool.PoolName_Legals()},
		{name: `Invalid Empty`,
			output: errors.New(PoolName_Error_Empty)},
		{name: `Invalid Length`,
			input:  []string{test_data_pool.PoolName_Max_Illegal()},
			output: errors.New(PoolName_Error_Length)},
		{name: `Invalid Characters`,
			input:  test_data_pool.PoolName_Error_Characters(),
			output: errors.New(PoolName_Error_Characters)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, input := range test.input {
				require.Equal(t, test.output, PoolName(input).Validate())
			}
		})
	}
}

func test_guestsToAddAndRemoveFromPools_data() []struct {
	name        string
	pool        PoolName
	guests      RawGuestResources
	guestsToAdd []GuestID
	remove      map[PoolName][]GuestID
	add         []GuestID
} {
	set := func(raw []rawGuestResource) RawGuestResources {
		interfaces := make([]RawGuestResource, len(raw))
		for i := range raw {
			interfaces[i] = &raw[i]
		}
		return interfaces
	}
	return []struct {
		name        string
		pool        PoolName
		guests      RawGuestResources
		guestsToAdd []GuestID
		remove      map[PoolName][]GuestID
		add         []GuestID
	}{
		{name: `'guestsToAdd' Not in 'guests'`,
			pool: "poolB",
			guests: set([]rawGuestResource{
				{a: map[string]any{"vmid": float64(100), "pool": "test"}},
				{a: map[string]any{"vmid": float64(200), "pool": "poolA"}},
				{a: map[string]any{"vmid": float64(300), "pool": "test"}}}),
			guestsToAdd: []GuestID{700, 800, 900},
			add:         []GuestID{700, 800, 900},
			remove:      map[PoolName][]GuestID{}},
		{name: `Empty`,
			pool:   "poolB",
			remove: map[PoolName][]GuestID{}},
		{name: `Empty 'guests'`,
			pool:        "poolB",
			guestsToAdd: []GuestID{100, 300, 200},
			add:         []GuestID{100, 300, 200},
			remove:      map[PoolName][]GuestID{}},
		{name: `Empty 'guestsToAdd'`,
			pool: "poolB",
			guests: set([]rawGuestResource{
				{a: map[string]any{"vmid": float64(100), "pool": "test"}},
				{a: map[string]any{"vmid": float64(200), "pool": "poolA"}},
				{a: map[string]any{"vmid": float64(300), "pool": "test"}}}),
			add:    []GuestID{},
			remove: map[PoolName][]GuestID{}},
		{name: `Guest already in target pool`,
			pool: "test",
			guests: set([]rawGuestResource{
				{a: map[string]any{"vmid": float64(700), "pool": "test"}}}),
			guestsToAdd: []GuestID{700},
			add:         []GuestID{},
			remove:      map[PoolName][]GuestID{}},
		{name: `Full`,
			pool: "poolX",
			guests: set([]rawGuestResource{
				{a: map[string]any{"vmid": float64(100), "pool": "test"}},
				{a: map[string]any{"vmid": float64(200), "pool": "poolA"}},
				{a: map[string]any{"vmid": float64(300), "pool": "test"}},
				{a: map[string]any{"vmid": float64(400), "pool": "poolB"}},
				{a: map[string]any{"vmid": float64(500), "pool": "poolC"}},
				{a: map[string]any{"vmid": float64(600), "pool": ""}},
				{a: map[string]any{"vmid": float64(700), "pool": "poolC"}},
				{a: map[string]any{"vmid": float64(800), "pool": "poolB"}},
				{a: map[string]any{"vmid": float64(900), "pool": ""}},
				{a: map[string]any{"vmid": float64(1000), "pool": "test"}},
				{a: map[string]any{"vmid": float64(1100), "pool": "poolA"}},
				{a: map[string]any{"vmid": float64(1200), "pool": "poolX"}},
			}),
			guestsToAdd: []GuestID{100, 300, 200, 500, 700, 900, 1100, 1200},
			add:         []GuestID{100, 300, 200, 500, 700, 900, 1100},
			remove: map[PoolName][]GuestID{
				"test":  {100, 300},
				"poolA": {200, 1100},
				"poolC": {500, 700},
			}}}
}

func Test_guestsToAddAndRemoveFromPools(t *testing.T) {
	t.Parallel()
	for _, test := range test_guestsToAddAndRemoveFromPools_data() {
		t.Run(test.name, func(t *testing.T) {
			add, remove := guestsToAddAndRemoveFromPools(test.guests, test.guestsToAdd, test.pool)
			require.ElementsMatch(t, test.add, add)
			require.Equal(t, test.remove, remove)
		})
	}
}

func Benchmark_guestsToAddAndRemoveFromPools(b *testing.B) {
	tests := test_guestsToAddAndRemoveFromPools_data()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			guestsToAddAndRemoveFromPools(test.guests, test.guestsToAdd, test.pool)
		}
	}
}
