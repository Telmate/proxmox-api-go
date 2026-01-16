package api_test

import (
	"context"
	"crypto/tls"
	"strconv"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/pad"
	"github.com/Telmate/proxmox-api-go/internal/util"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	"github.com/stretchr/testify/require"
)

func Test_Pool_Create(t *testing.T) {
	pool := pveSDK.ConfigPool{
		Name:    "Test_Pool_Create",
		Comment: util.Pointer("Test Comment" + body.Symbols),
		Guests:  &[]pveSDK.GuestID{510, 511, 512},
	}
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
				for _, guest := range *pool.Guests {
					require.Error(t, pveSDK.NewVmRef(guest).Delete(ctx, cl))
				}
			}},
		{name: `Ensure pool does not exist`,
			test: func(t *testing.T) {
				deleted, err := c.Pool.Delete(ctx, pool.Name)
				require.NoError(t, err)
				require.False(t, deleted)
			}},
		{name: `Create guests`,
			test: func(t *testing.T) {
				for _, guest := range *pool.Guests {
					guestCreate(t, ctx, cl, guest, node, pveSDK.GuestName("Test-Pool-Create-VM"))
				}
			}},
		{name: `Create pool`,
			test: func(t *testing.T) {
				require.NoError(t, c.Pool.Create(ctx, pool))
			}},
		{name: `Read pool`,
			test: func(t *testing.T) {
				poolRead(t, ctx, c.Pool, pool.Name, *pool.Comment, *pool.Guests, []pveSDK.StorageName{})
			}},
		{name: `Delete pool`,
			test: func(t *testing.T) {
				deleted, err := c.Pool.Delete(ctx, pool.Name)
				require.NoError(t, err)
				require.True(t, deleted)
			}},
		{name: `Delete guests`,
			test: func(t *testing.T) {
				for _, guest := range *pool.Guests {
					require.NoError(t, pveSDK.NewVmRef(guest).Delete(ctx, cl))
				}
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
