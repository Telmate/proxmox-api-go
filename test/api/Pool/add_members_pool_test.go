package api_test

import (
	"context"
	"crypto/tls"
	"strconv"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/pad"
	"github.com/Telmate/proxmox-api-go/internal/util"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	"github.com/stretchr/testify/require"
)

func Test_Pool_AddMembers_Empty(t *testing.T) {
	pool := pveSDK.ConfigPool{
		Name:    "Test_Pool_AddMembers_Empty",
		Comment: util.Pointer(""),
	}
	pool_AddMembers(t, pool, []pveSDK.GuestID{}, []pveSDK.GuestID{500, 501, 502})
}

func Test_Pool_AddMembers_NonEmpty(t *testing.T) {
	initialGuests := []pveSDK.GuestID{503, 504, 505}
	pool := pveSDK.ConfigPool{
		Name:    "Test_Pool_AddMembers_NonEmpty",
		Comment: util.Pointer(""),
		Guests:  &initialGuests,
	}
	pool_AddMembers(t, pool, initialGuests, []pveSDK.GuestID{506, 507, 508})
}

func pool_AddMembers(t *testing.T, pool pveSDK.ConfigPool, initialGuests, additionalGuests []pveSDK.GuestID) {
	allGuests := append(initialGuests, additionalGuests...)
	const node = pveSDK.NodeName(test.FirstNode)
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Ensure guests do not exist`,
			test: func(t *testing.T) {
				for _, guest := range allGuests {
					require.Error(t, pveSDK.NewVmRef(guest).Delete(ctx, cl))
				}
			}},
		{name: `Ensure pool does not exist`,
			test: func(t *testing.T) {
				deleted, err := c.Pool.Delete(ctx, pool.Name)
				require.NoError(t, err)
				require.False(t, deleted)
			}},
		{name: `Create guests`,
			test: func(t *testing.T) {
				for _, guest := range allGuests {
					config := pveSDK.ConfigQemu{
						CPU:    &pveSDK.QemuCPU{Cores: util.Pointer(pveSDK.QemuCpuCores(1))},
						ID:     &guest,
						Memory: &pveSDK.QemuMemory{CapacityMiB: util.Pointer(pveSDK.QemuMemoryCapacity(16))},
						Name:   util.Pointer(pveSDK.GuestName("Test-Pool-AddMembers-VM")),
						Node:   util.Pointer(node)}
					vmRef, err := config.Create(ctx, cl)
					require.NoError(t, err)
					require.NotNil(t, vmRef)
				}
			}},
		{name: `Create pool`,
			test: func(t *testing.T) {
				require.NoError(t, c.Pool.Create(ctx, pool))
			}},
		{name: `Read pool`,
			test: func(t *testing.T) {
				raw, err := c.Pool.Read(ctx, pool.Name)
				require.NoError(t, err)
				require.NotNil(t, raw)
				require.Equal(t, *pool.Comment, raw.GetComment())
				require.Equal(t, pool.Name, raw.GetName())
				members := raw.GetMembers()
				require.NotNil(t, members)
				rawGuests, storages := members.AsArrays()
				require.Len(t, rawGuests, len(initialGuests))
				require.Equal(t, 0, len(storages))
				guests := make([]pveSDK.GuestID, len(rawGuests))
				for i := range rawGuests {
					guests[i] = rawGuests[i].GetID()
				}
				require.ElementsMatch(t, initialGuests, guests)
			}},
		{name: `Add members to pool`,
			test: func(t *testing.T) {
				require.NoError(t, c.Pool.AddMembers(ctx, pool.Name, additionalGuests, nil))
			}},
		{name: `Read pool again`,
			test: func(t *testing.T) {
				raw, err := c.Pool.Read(ctx, pool.Name)
				require.NoError(t, err)
				require.NotNil(t, raw)
				require.Equal(t, *pool.Comment, raw.GetComment())
				require.Equal(t, pool.Name, raw.GetName())
				members := raw.GetMembers()
				require.NotNil(t, members)
				rawGuests, storages := members.AsArrays()
				require.Len(t, rawGuests, len(allGuests))
				require.Equal(t, 0, len(storages))
				guests := make([]pveSDK.GuestID, len(rawGuests))
				for i := range rawGuests {
					guests[i] = rawGuests[i].GetID()
				}
				require.ElementsMatch(t, allGuests, guests)
			}},
		{name: `Delete guests`,
			test: func(t *testing.T) {
				for _, guest := range allGuests {
					require.NoError(t, pveSDK.NewVmRef(guest).Delete(ctx, cl))
				}
			}},
		{name: `Delete pool`,
			test: func(t *testing.T) {
				deleted, err := c.Pool.Delete(ctx, pool.Name)
				require.NoError(t, err)
				require.True(t, deleted)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}

func Test_Pool_AddMembers_Move(t *testing.T) {
	guest := pveSDK.GuestID(509)
	poolOriginal := pveSDK.ConfigPool{
		Name:   "Test_Pool_AddMembers_Move_Original",
		Guests: &[]pveSDK.GuestID{guest},
	}
	poolTarget := pveSDK.ConfigPool{
		Name: "Test_Pool_AddMembers_Move_Target",
	}
	const node = pveSDK.NodeName(test.FirstNode)
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Ensure pool does not exist`,
			test: func(t *testing.T) {
				for _, pool := range []pveSDK.ConfigPool{poolOriginal, poolTarget} {
					deleted, err := c.Pool.Delete(ctx, pool.Name)
					require.NoError(t, err)
					require.False(t, deleted)
				}
			}},
		{name: `Create guest`,
			test: func(t *testing.T) {
				guestCreate(t, ctx, cl, guest, node, "Test-Pool-AddMembers-VM")
			}},
		{name: `Create pools`,
			test: func(t *testing.T) {
				for _, pool := range []pveSDK.ConfigPool{poolOriginal, poolTarget} {
					require.NoError(t, c.Pool.Create(ctx, pool))
				}
			}},
		{name: `Read pools`,
			test: func(t *testing.T) {
				for _, pool := range []pveSDK.ConfigPool{
					{Name: poolOriginal.Name,
						Comment:  util.Pointer(""),
						Guests:   &[]pveSDK.GuestID{509},
						Storages: &[]pveSDK.StorageName{}},
					{Name: poolTarget.Name,
						Comment:  util.Pointer(""),
						Guests:   &[]pveSDK.GuestID{},
						Storages: &[]pveSDK.StorageName{}},
				} {
					poolRead(t, ctx, c.Pool, pool.Name, *pool.Comment, *pool.Guests, *pool.Storages)
				}
			}},
		{name: `Add member to target pool`,
			test: func(t *testing.T) {
				require.NoError(t, c.Pool.AddMembers(ctx, poolTarget.Name, []pveSDK.GuestID{guest}, nil))
			}},
		{name: `Read pool again`,
			test: func(t *testing.T) {
				for _, pool := range []pveSDK.ConfigPool{
					{Name: poolOriginal.Name,
						Comment:  util.Pointer(""),
						Guests:   &[]pveSDK.GuestID{},
						Storages: &[]pveSDK.StorageName{}},
					{Name: poolTarget.Name,
						Comment:  util.Pointer(""),
						Guests:   &[]pveSDK.GuestID{509},
						Storages: &[]pveSDK.StorageName{}},
				} {
					poolRead(t, ctx, c.Pool, pool.Name, *pool.Comment, *pool.Guests, *pool.Storages)
				}
			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				require.NoError(t, pveSDK.NewVmRef(guest).Delete(ctx, cl))
			}},
		{name: `Delete pools`,
			test: func(t *testing.T) {
				for _, pool := range []pveSDK.ConfigPool{poolOriginal, poolTarget} {
					deleted, err := c.Pool.Delete(ctx, pool.Name)
					require.NoError(t, err)
					require.True(t, deleted)
				}
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
