package api_test

import (
	"context"
	"crypto/tls"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Telmate/proxmox-api-go/internal/pad"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
)

func Test_Group_Create(t *testing.T) {
	groupName := pveSDK.GroupName("Test_Group_Create")
	users := []pveSDK.UserID{
		{Name: "Test_Group_Create_1", Realm: "pve"},
		{Name: "Test_Group_Create_2", Realm: "pve"},
	}
	group := pveSDK.ConfigGroup{
		Name:    groupName,
		Members: &users,
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
		{name: `Ensure group does not exist`,
			test: func(t *testing.T) {
				exists, err := c.Group.Delete(ctx, groupName)
				require.NoError(t, err)
				require.False(t, exists)
			}},
		{name: `Ensure users do not exist`,
			test: func(t *testing.T) {
				for _, user := range users {
					require.NoError(t, c.User.Delete(ctx, user))
				}
			}},
		{name: `Create users`,
			test: func(t *testing.T) {
				for _, user := range users {
					require.NoError(t, c.User.Create(ctx, pveSDK.ConfigUser{User: user}))
				}
			}},
		{name: `Create group`,
			test: func(t *testing.T) {
				require.NoError(t, c.Group.Create(ctx, group))
			}},
		{name: `Verify group exists`,
			test: func(t *testing.T) {
				raw, err := c.Group.List(ctx)
				require.NoError(t, err)
				require.NotNil(t, raw)
				require.GreaterOrEqual(t, raw.Len(), 1)
				_, exists := raw.AsMap()[groupName]
				require.True(t, exists)
			}},
		{name: `Delete group`,
			test: func(t *testing.T) {
				exists, err := c.Group.Delete(ctx, groupName)
				require.NoError(t, err)
				require.True(t, exists)
			}},
		{name: `Delete users`,
			test: func(t *testing.T) {
				for _, user := range users {
					require.NoError(t, c.User.Delete(ctx, user))
				}
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
