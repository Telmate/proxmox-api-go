package api_test

import (
	"context"
	"crypto/tls"
	"strconv"
	"testing"
	"time"

	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/pad"
	"github.com/Telmate/proxmox-api-go/internal/util"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	"github.com/stretchr/testify/require"
)

func Test_Snapshot_List(t *testing.T) {
	t.Parallel()
	snapshots := []pveSDK.SnapshotInfo{
		{Name: "snap1",
			Description: "First snapshot" + body.Symbols,
			VmState:     util.Pointer(false)},
		{Name: "mySnap",
			VmState: util.Pointer(false),
			Parent:  util.Pointer(pveSDK.SnapshotName("snap1"))},
		{Name: "test-snapshot",
			Description: "a",
			VmState:     util.Pointer(false),
			Parent:      util.Pointer(pveSDK.SnapshotName("mySnap"))}}
	const guest = pveSDK.GuestID(802)
	const node = pveSDK.NodeName(test.FirstNode)
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	var currentTime time.Time
	currentTimePtr := &currentTime
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Ensure guest does not exist`,
			test: func(t *testing.T) {
				require.Error(t, pveSDK.NewVmRef(guest).Delete(ctx, cl))
			}},
		{name: `Create guest`,
			test: func(t *testing.T) {
				guestCreate(t, ctx, cl, guest, node, "Test-Snapshot-List")
			}},
		{name: `Create snapshot`,
			test: func(t *testing.T) {
				*currentTimePtr = time.Now()
				for _, snap := range snapshots {
					require.NoError(t, c.Snapshot.CreateQemu(ctx, *pveSDK.NewVmRef(guest), snap.Name, snap.Description, *snap.VmState))
				}
			}},
		{name: `List snapshots`,
			test: func(t *testing.T) {
				raw, err := c.Snapshot.List(ctx, *pveSDK.NewVmRef(guest))
				require.NoError(t, err)
				require.NotNil(t, raw)
				snapshotMap := raw.AsMap()
				for _, snapshot := range snapshots {
					raw, exists := snapshotMap[snapshot.Name]
					require.True(t, exists, "snapshot %q not found in list", snapshot.Name)
					require.NotNil(t, raw)
					snapshotInfo := raw.Get()
					require.Equal(t, snapshot.Name, snapshotInfo.Name)
					require.Equal(t, snapshot.Description, snapshotInfo.Description)
					require.Equal(t, snapshot.VmState, snapshotInfo.VmState)
					require.Equal(t, snapshot.Parent, snapshotInfo.Parent)
					require.WithinDuration(t, *currentTimePtr, *snapshotInfo.Time, time.Minute)
				}
			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				require.NoError(t, pveSDK.NewVmRef(guest).Delete(ctx, cl))
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
