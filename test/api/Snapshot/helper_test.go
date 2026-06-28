package api_test

import (
	"context"
	"testing"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

func createQemu(t *testing.T, ctx context.Context, c pveSDK.ClientNew, config pveSDK.ConfigQemu) {
	vmRef, err := c.QemuGuest.Create(ctx, config)
	require.NoError(t, err)
	require.NotNil(t, vmRef)
}

func createLxc(t *testing.T, ctx context.Context, c *pveSDK.Client, config pveSDK.ConfigLXC) {
	vmRef, err := config.Create(ctx, c)
	require.NoError(t, err)
	require.NotNil(t, vmRef)
}
