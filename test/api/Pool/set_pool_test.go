package api_test

import (
	"context"
	"crypto/tls"
	"strconv"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/pad"
	"github.com/Telmate/proxmox-api-go/internal/util"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	"github.com/stretchr/testify/require"
)

func Test_Pool_Set(t *testing.T) {
	guests := []pveSDK.GuestID{520, 521, 522}
	pool := pveSDK.PoolName("Test_Pool_Set")
	const node = pveSDK.NodeName(test.FirstNode)
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Ensure guests do not exist`,
			test: func(t *testing.T) {
				for _, guest := range guests {
					require.Error(t, pveSDK.NewVmRef(guest).Delete(ctx, cl))
				}
			}},
		{name: `Ensure pool does not exist`,
			test: func(t *testing.T) {
				deleted, err := c.Pool.Delete(ctx, pool)
				require.NoError(t, err)
				require.False(t, deleted)
			}},
		{name: `Create guests`,
			test: func(t *testing.T) {
				for _, guest := range guests {
					guestCreate(t, ctx, cl, guest, node, "Test-Pool-Set-VM")
				}
			}},
		{name: `Create pool`,
			test: func(t *testing.T) {
				require.NoError(t, c.Pool.Set(ctx, pveSDK.ConfigPool{
					Name:    pool,
					Comment: util.Pointer("Test Pool Set Comment"),
					Guests:  &guests,
				}))
			}},
		{name: `Read pool 1`,
			test: func(t *testing.T) {
				poolRead(t, ctx, c.Pool, pool, "Test Pool Set Comment", guests, []pveSDK.StorageName{})
			}},
		{name: `Update pool`,
			test: func(t *testing.T) {
				require.NoError(t, c.Pool.Set(ctx, pveSDK.ConfigPool{
					Name:    pool,
					Comment: util.Pointer(""),
					Guests:  &[]pveSDK.GuestID{521},
				}))
			}},
		{name: `Read pool 2`,
			test: func(t *testing.T) {
				poolRead(t, ctx, c.Pool, pool, "", []pveSDK.GuestID{521}, []pveSDK.StorageName{})
			}},
		{name: `Update pool new comment`,
			test: func(t *testing.T) {
				require.NoError(t, c.Pool.Set(ctx, pveSDK.ConfigPool{
					Name:    pool,
					Comment: util.Pointer("New comment"),
				}))
			}},
		{name: `Read pool 3`,
			test: func(t *testing.T) {
				poolRead(t, ctx, c.Pool, pool, "New comment", []pveSDK.GuestID{521}, []pveSDK.StorageName{})
			}},
		{name: `Update pool guests`,
			test: func(t *testing.T) {
				require.NoError(t, c.Pool.Set(ctx, pveSDK.ConfigPool{
					Name:   pool,
					Guests: &[]pveSDK.GuestID{520, 522},
				}))
			}},
		{name: `Read pool 4`,
			test: func(t *testing.T) {
				poolRead(t, ctx, c.Pool, pool, "New comment", []pveSDK.GuestID{520, 522}, []pveSDK.StorageName{})
			}},
		{name: `Delete guests`,
			test: func(t *testing.T) {
				for _, guest := range guests {
					require.NoError(t, pveSDK.NewVmRef(guest).Delete(ctx, cl))
				}
			}},
		{name: `Delete pool`,
			test: func(t *testing.T) {
				deleted, err := c.Pool.Delete(ctx, pool)
				require.NoError(t, err)
				require.True(t, deleted)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
