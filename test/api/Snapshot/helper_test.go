package api_test

import (
	"context"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

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
