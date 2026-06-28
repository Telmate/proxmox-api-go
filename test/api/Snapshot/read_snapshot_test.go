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

func Test_Snapshot_ReadQemu(t *testing.T) {
	t.Parallel()
	const (
		guest    = pveSDK.GuestID(804)
		name     = pveSDK.GuestName("Test-Snapshot-ReadQemu")
		node     = pveSDK.NodeName(test.FirstNode)
		snapName = pveSDK.SnapshotName("snap1")
	)
	snapshot := pveSDK.SnapshotInfo{
		Name:        snapName,
		Description: "Test snapshot" + body.Symbols,
		VmState:     new(false)}
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	set, expected := qemu.MinimumConfig(guest, node, name)
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
		{name: `Read snapshot`,
			test: func(t *testing.T) {
				raw, err := c.Snapshot.ReadQemu(ctx, *pveSDK.NewVmRef(guest), snapshot.Name)
				require.NoError(t, err)
				require.NotNil(t, raw)
				config, _ := raw.Get(pveSDK.VmRef{})
				config.Smbios1 = ""
				expected.Description = &snapshot.Description
				require.Equal(t, expected, config)
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
func Test_Snapshot_ReadLxc(t *testing.T) {
	t.Parallel()
	const (
		guest    = pveSDK.GuestID(805)
		name     = pveSDK.GuestName("Test-Snapshot-ReadLxc")
		node     = pveSDK.NodeName(test.FirstNode)
		snapName = pveSDK.SnapshotName("snap1")
		storage  = pveSDK.StorageName(test.GuestStorage)
	)
	snapshot := pveSDK.SnapshotInfo{
		Name:        snapName,
		Description: "Test snapshot" + body.Symbols,
		VmState:     new(false)}
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	set, expected := lxc.MinimumConfig(guest, node, storage, new(false), name)
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
				require.NoError(t, c.Snapshot.CreateLxc(ctx, *pveSDK.NewVmRef(guest), snapshot.Name, snapshot.Description))
			}},
		{name: `Read snapshot`,
			test: func(t *testing.T) {
				raw, err := c.Snapshot.ReadLxc(ctx, *pveSDK.NewVmRef(guest), snapshot.Name)
				require.NoError(t, err)
				require.NotNil(t, raw)
				config := raw.Get(pveSDK.VmRef{}, pveSDK.PowerStateRunning)
				config.State = nil
				expected.Description = &snapshot.Description
				require.EqualExportedValues(t, &expected, config)
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
