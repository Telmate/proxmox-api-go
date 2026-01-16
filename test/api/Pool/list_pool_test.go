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

func Test_Pool_List(t *testing.T) {

	pools := []pveSDK.ConfigPool{
		{Name: "Test_Pool_List_01"},
		{Name: "Test_Pool_List_02",
			Comment: util.Pointer("Test Comment")},
		{Name: "Test_Pool_List_03"}}
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Ensure pools do not exist`,
			test: func(t *testing.T) {
				for _, pool := range pools {
					deleted, err := c.Pool.Delete(ctx, pool.Name)
					require.NoError(t, err)
					require.False(t, deleted)
				}
			}},
		{name: `Create pools`,
			test: func(t *testing.T) {
				for _, pool := range pools {
					require.NoError(t, c.Pool.Create(ctx, pool))
				}
			}},
		{name: `List pools`,
			test: func(t *testing.T) {
				rawPools, err := c.Pool.List(ctx)
				require.NoError(t, err)
				require.GreaterOrEqual(t, rawPools.Len(), len(pools))
				for _, pool := range pools {
					poolMap := rawPools.AsMap()
					rawPool, exists := poolMap[pool.Name]
					require.True(t, exists, "pool %q not found in list", pool)
					name, comment := rawPool.Get()
					tmpPool := pveSDK.ConfigPool{
						Name: name,
					}
					if comment != "" {
						tmpPool.Comment = util.Pointer(comment)
					}
					require.Equal(t, pool, tmpPool)
				}
			}},
		{name: `Delete pools`,
			test: func(t *testing.T) {
				for _, pool := range pools {
					deleted, err := c.Pool.Delete(ctx, pool.Name)
					require.NoError(t, err)
					require.True(t, deleted)
				}
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
