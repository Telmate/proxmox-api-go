package api_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Telmate/proxmox-api-go/internal/pad"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	api_test "github.com/Telmate/proxmox-api-go/test/api"
)

func Test_Group_List(t *testing.T) {
	t.Parallel()
	groups := []pveSDK.GroupName{
		pveSDK.GroupName("Test_Group_List_1"),
		pveSDK.GroupName("Test_Group_List_2"),
		pveSDK.GroupName("Test_Group_List_3"),
	}
	cl, err := api_test.NewClient()
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Ensure groups do not exist`,
			test: func(t *testing.T) {
				for _, groupName := range groups {
					exists, err := c.Group.Delete(ctx, groupName)
					require.NoError(t, err)
					require.False(t, exists)
				}
			}},
		{name: `Create groups`,
			test: func(t *testing.T) {
				for _, groupName := range groups {
					require.NoError(t, c.Group.Create(ctx, pveSDK.ConfigGroup{Name: groupName}))
				}
			}},
		{name: `List groups`,
			test: func(t *testing.T) {
				raw, err := c.Group.List(ctx)
				require.NoError(t, err)
				require.GreaterOrEqual(t, raw.Len(), len(groups))
				_, exists := raw.AsMap()[groups[0]]
				require.True(t, exists)
			}},
		{name: `Delete groups`,
			test: func(t *testing.T) {
				for _, groupName := range groups {
					exists, err := c.Group.Delete(ctx, groupName)
					require.NoError(t, err)
					require.True(t, exists)
				}
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
