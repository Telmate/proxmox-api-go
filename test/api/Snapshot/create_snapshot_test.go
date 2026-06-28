package api_test

import (
	"context"
	"crypto/tls"
	"strconv"
	"testing"
	"time"

	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/pad"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	"github.com/Telmate/proxmox-api-go/test/api/qemu"
	"github.com/stretchr/testify/require"
)

func Test_Snapshot_CreateQemu(t *testing.T) {
	t.Parallel()
	const (
		guest    = pveSDK.GuestID(800)
		name     = pveSDK.GuestName("Test-Snapshot-CreateQemu")
		node     = pveSDK.NodeName(test.FirstNode)
		snapName = pveSDK.SnapshotName("snap1")
	)
	snapshot := pveSDK.SnapshotInfo{
		Name:        snapName,
		Description: "Test snapshot" + body.Symbols,
		VmState:     new(false),
	}
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	set, _ := qemu.MinimumConfig(guest, node, name)
	var currentTime time.Time
	currentTimePtr := &currentTime
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
				*currentTimePtr = time.Now()
				require.NoError(t, c.Snapshot.CreateQemu(ctx, *pveSDK.NewVmRef(guest), snapshot.Name, snapshot.Description, false))
			}},
		{name: `List snapshots`,
			test: func(t *testing.T) {
				raw, err := c.Snapshot.List(ctx, *pveSDK.NewVmRef(guest))
				require.NoError(t, err)
				require.NotNil(t, raw)
				snapshotMap := raw.AsMap()
				rawSnap, exists := snapshotMap[snapshot.Name]
				require.True(t, exists, "snapshot %q not found in list", snapshot.Name)
				require.Equal(t, snapshot.Description, rawSnap.GetDescription())
				require.Equal(t, snapshot.Name, rawSnap.GetName())

				require.WithinDuration(t, *currentTimePtr, *rawSnap.GetTime(), time.Minute)
				require.Equal(t, snapshot.Parent, rawSnap.GetParent())
				require.Equal(t, snapshot.VmState, rawSnap.GetVmState())
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
