package api_test

import (
	"context"
	"crypto/tls"
	"strconv"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/pad"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	"github.com/Telmate/proxmox-api-go/test/api/qemu"
	"github.com/stretchr/testify/require"
)

func Test_Snapshot_Delete(t *testing.T) {
	t.Parallel()
	const (
		guest    = pveSDK.GuestID(801)
		name     = pveSDK.GuestName("Test-Snapshot-Delete")
		node     = pveSDK.NodeName(test.FirstNode)
		snapName = pveSDK.SnapshotName("snap1")
	)
	snapshots := []pveSDK.SnapshotName{snapName}
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	set, _ := qemu.MinimumConfig(guest, node, name)
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Ensure guest does not exist`,
			test: func(t *testing.T) {
				existed, err := c.Guest.Delete(ctx, *pveSDK.NewVmRef(guest))
				require.NoError(t, err)
				require.False(t, existed)
			}},
		{name: `Create guest`,
			test: func(t *testing.T) {
				createQemu(t, ctx, c, set)
			}},
		{name: `Create snapshot`,
			test: func(t *testing.T) {
				require.NoError(t, c.Snapshot.CreateQemu(ctx, *pveSDK.NewVmRef(guest), snapName, "Test snapshot", false))
			}},
		{name: `List snapshots`,
			test: func(t *testing.T) {
				raw, err := c.Snapshot.List(ctx, *pveSDK.NewVmRef(guest))
				require.NoError(t, err)
				require.NotNil(t, raw)
				snapshotMap := raw.AsMap()
				for _, snapshot := range snapshots {
					_, exists := snapshotMap[snapshot]
					require.True(t, exists, "snapshot %q not found in list", snapshot)
				}
			}},
		{name: `Delete snapshot`,
			test: func(t *testing.T) {
				existed, err := c.Snapshot.Delete(ctx, *pveSDK.NewVmRef(guest), snapName)
				require.NoError(t, err)
				require.True(t, existed)
			}},
		{name: `Verify snapshot is deleted`,
			test: func(t *testing.T) {
				raw, err := c.Snapshot.List(ctx, *pveSDK.NewVmRef(guest))
				require.NoError(t, err)
				require.NotNil(t, raw)
				snapshotMap := raw.AsMap()
				for _, snapshot := range snapshots {
					_, exists := snapshotMap[snapshot]
					require.False(t, exists, "snapshot %q found in list after deletion", snapshot)
				}
			}},
		{name: `Delete non-existing snapshot`,
			test: func(t *testing.T) {
				existed, err := c.Snapshot.Delete(ctx, *pveSDK.NewVmRef(guest), snapName)
				require.NoError(t, err)
				require.False(t, existed)
			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				existed, err := c.Guest.Delete(ctx, *pveSDK.NewVmRef(guest))
				require.NoError(t, err)
				require.True(t, existed)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
