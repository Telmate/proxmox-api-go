package api_test

import (
	"context"
	"crypto/tls"
	"strconv"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/pad"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	"github.com/Telmate/proxmox-api-go/test/api/qemu"
	"github.com/stretchr/testify/require"
)

func Test_Snapshot_Update(t *testing.T) {
	t.Parallel()
	const (
		guest    = pveSDK.GuestID(803)
		name     = pveSDK.GuestName("Test-Snapshot-Update")
		node     = pveSDK.NodeName(test.FirstNode)
		snapName = pveSDK.SnapshotName("snap1")
	)
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
				require.NoError(t, c.Snapshot.CreateQemu(ctx, *pveSDK.NewVmRef(guest), snapName, "", false))
			}},
		{name: `Get empty description`,
			test: func(t *testing.T) {
				GetDescription(t, ctx, c, guest, snapName, "")
			}},
		{name: `Update snapshot description full`,
			test: func(t *testing.T) {
				require.NoError(t, c.Snapshot.Update(ctx, *pveSDK.NewVmRef(guest), snapName, "Test snapshot description"+body.Symbols))
			}},
		{name: `Get full description`,
			test: func(t *testing.T) {
				GetDescription(t, ctx, c, guest, snapName, "Test snapshot description"+body.Symbols)
			}},
		{name: `Update snapshot description empty`,
			test: func(t *testing.T) {
				require.NoError(t, c.Snapshot.Update(ctx, *pveSDK.NewVmRef(guest), snapName, ""))
			}},
		{name: `Get empty description again`,
			test: func(t *testing.T) {
				GetDescription(t, ctx, c, guest, snapName, "")
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

func GetDescription(t *testing.T, ctx context.Context, c pveSDK.ClientNew, guest pveSDK.GuestID, snap pveSDK.SnapshotName, expected string) {
	raw, err := c.Snapshot.List(ctx, *pveSDK.NewVmRef(guest))
	require.NoError(t, err)
	require.NotNil(t, raw)
	snapshotMap := raw.AsMap()
	rawSnap, exists := snapshotMap[snap]
	require.True(t, exists, "snapshot %q not found in list", snap)
	require.Equal(t, expected, rawSnap.GetDescription())

}
