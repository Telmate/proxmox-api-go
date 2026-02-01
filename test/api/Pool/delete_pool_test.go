package api_test

import (
	"context"
	"crypto/tls"
	"strconv"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/pad"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	"github.com/stretchr/testify/require"
)

func Test_Pool_Delete(t *testing.T) {
	t.Parallel()
	pool := pveSDK.ConfigPool{
		Name: "Test_Pool_Delete",
	}
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Ensure pool does not exist`,
			test: func(t *testing.T) {
				deleted, err := c.Pool.Delete(ctx, pool.Name)
				require.NoError(t, err)
				require.False(t, deleted)
			}},
		{name: `Create pool`,
			test: func(t *testing.T) {
				require.NoError(t, c.Pool.Create(ctx, pool))
			}},
		{name: `Read pool`,
			test: func(t *testing.T) {
				poolRead(t, ctx, c.Pool, pool.Name, "", []pveSDK.GuestID{}, []pveSDK.StorageName{})
			}},
		{name: `Delete pool`,
			test: func(t *testing.T) {
				deleted, err := c.Pool.Delete(ctx, pool.Name)
				require.NoError(t, err)
				require.True(t, deleted)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
