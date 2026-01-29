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

func Test_Snapshot_CreateQemu(t *testing.T) {
	t.Parallel()
	const snap1 = pveSDK.SnapshotName("snap1")
	const guest = pveSDK.GuestID(800)
	const node = pveSDK.NodeName(test.FirstNode)
	snapshot := pveSDK.SnapshotInfo{
		Name:        snap1,
		Description: "Test snapshot" + body.Symbols,
		VmState:     util.Pointer(false),
	}
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000)
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
				guestCreate(t, ctx, cl, guest, node, "Test-Snapshot-CreateQemu")
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
				require.NoError(t, pveSDK.NewVmRef(guest).Delete(ctx, cl))
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
