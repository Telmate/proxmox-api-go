package api_test

import (
	"context"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

func poolRead(t *testing.T, ctx context.Context, c pveSDK.PoolInterface,
	name pveSDK.PoolName, comment string,
	guests []pveSDK.GuestID, storages []pveSDK.StorageName,
) {
	raw, err := c.Read(ctx, name)
	require.NoError(t, err)
	require.NotNil(t, raw)
	require.Equal(t, comment, raw.GetComment())
	require.Equal(t, name, raw.GetName())
	members := raw.GetMembers()
	require.NotNil(t, members)
	rawGuests, rawStorages := members.AsArrays()
	require.Len(t, rawGuests, len(guests))
	require.Len(t, rawStorages, len(storages))
	tmpGuests := make([]pveSDK.GuestID, len(rawGuests))
	for i := range rawGuests {
		tmpGuests[i] = rawGuests[i].GetID()
	}
	require.ElementsMatch(t, guests, tmpGuests)
	tmpStorages := make([]pveSDK.StorageName, len(rawStorages))
	for i := range rawStorages {
		tmpStorages[i] = rawStorages[i].GetName()
	}
	require.ElementsMatch(t, storages, tmpStorages)
}

func guestCreate(t *testing.T, ctx context.Context, cl *pveSDK.Client, guest pveSDK.GuestID, node pveSDK.NodeName, name pveSDK.GuestName) {
	config := pveSDK.ConfigQemu{
		CPU:    &pveSDK.QemuCPU{Cores: util.Pointer(pveSDK.QemuCpuCores(1))},
		ID:     &guest,
		Memory: &pveSDK.QemuMemory{CapacityMiB: util.Pointer(pveSDK.QemuMemoryCapacity(16))},
		Name:   util.Pointer(name),
		Node:   util.Pointer(node),
	}
	vmRef, err := config.Create(ctx, cl)
	require.NoError(t, err)
	require.NotNil(t, vmRef)
}
