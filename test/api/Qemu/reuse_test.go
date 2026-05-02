package api_test

import (
	"context"
	"testing"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

func CheckConfig(t *testing.T, ctx context.Context, c *pveSDK.Client, guestID pveSDK.GuestID, expected *pveSDK.ConfigQemu) {
	vmr := pveSDK.NewVmRef(pveSDK.GuestID(guestID))
	raw, err := pveSDK.NewRawConfigQemuFromApi(ctx, vmr, c)
	require.NoError(t, err)
	require.NotNil(t, raw)
	config, err := raw.Get(*vmr)
	require.NoError(t, err)
	require.NotNil(t, config)
	config.Smbios1 = ""                           // This field is unique and cannot be predicted, so we ignore it for the comparison.
	config.QemuUnusedDisks = pveSDK.QemuDevices{} // TODO include this field in test once reworked
	require.EqualExportedValues(t, expected, config)
}

func DeleteGuest(t *testing.T, ctx context.Context, c *pveSDK.Client, guestID pveSDK.GuestID) {
	vmr := pveSDK.NewVmRef(pveSDK.GuestID(guestID))
	raw, err := pveSDK.NewRawConfigQemuFromApi(ctx, vmr, c)
	require.NoError(t, err)
	require.NotNil(t, raw)
	require.NoError(t, vmr.Delete(ctx, c))
}
