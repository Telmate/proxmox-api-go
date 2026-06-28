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
	"github.com/Telmate/proxmox-api-go/test/api/lxc"
	"github.com/Telmate/proxmox-api-go/test/api/qemu"
	"github.com/stretchr/testify/require"
)

func Test_Snapshot_List_Qemu(t *testing.T) {
	t.Parallel()
	snapshots := []pveSDK.SnapshotInfo{
		{Name: "snap1",
			Description: "First snapshot" + body.Symbols,
			VmState:     new(false)},
		{Name: "mySnap",
			VmState: new(false),
			Parent:  new(pveSDK.SnapshotName("snap1"))},
		{Name: "test-snapshot",
			Description: "a",
			VmState:     new(false),
			Parent:      new(pveSDK.SnapshotName("mySnap"))}}
	const (
		guest = pveSDK.GuestID(802)
		name  = pveSDK.GuestName("Test-Snapshot-List-Qemu")
		node  = pveSDK.NodeName(test.FirstNode)
	)
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
				existed, err := c.Guest.Delete(ctx, *pveSDK.NewVmRef(guest))
				require.NoError(t, err)
				require.True(t, existed)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}

func Test_Snapshot_List_Lxc(t *testing.T) {
	t.Parallel()
	snapshots := []pveSDK.SnapshotInfo{
		{Name: "snap1",
			Description: "First snapshot" + body.Symbols},
		{Name: "mySnap",
			Parent: new(pveSDK.SnapshotName("snap1"))},
		{Name: "test-snapshot",
			Description: "a",
			Parent:      new(pveSDK.SnapshotName("mySnap"))}}
	const (
		guest   = pveSDK.GuestID(806)
		name    = pveSDK.GuestName("Test-Snapshot-List-Lxc")
		node    = pveSDK.NodeName(test.FirstNode)
		storage = pveSDK.StorageName(test.GuestStorage)
	)
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	set, _ := lxc.MinimumConfig(guest, node, storage, new(false), name)
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
				createLxc(t, ctx, cl, set)
			}},
		{name: `Create snapshot`,
			test: func(t *testing.T) {
				*currentTimePtr = time.Now()
				for _, snap := range snapshots {
					require.NoError(t, c.Snapshot.CreateLxc(ctx, *pveSDK.NewVmRef(guest), snap.Name, snap.Description))
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
				existed, err := c.Guest.Delete(ctx, *pveSDK.NewVmRef(guest))
				require.NoError(t, err)
				require.True(t, existed)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
