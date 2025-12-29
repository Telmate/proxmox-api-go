package api_test

import (
	"context"
	"crypto/tls"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Telmate/proxmox-api-go/internal/pad"
	"github.com/Telmate/proxmox-api-go/internal/util"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
)

func Test_Group_Set(t *testing.T) {
	groupName := pveSDK.GroupName("Test_Group_Set")
	users := []pveSDK.UserID{
		{Name: "Test_Group_Set_1", Realm: "pve"},
		{Name: "Test_Group_Set_2", Realm: "pve"},
	}
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000)
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
				require.NoError(t, c.Group.Set(ctx, pveSDK.ConfigGroup{
					Name:    groupName,
					Members: &users,
					Comment: util.Pointer("Test comment"),
				}))
			}},
		{name: `Check group Created`,
			test: func(t *testing.T) {
				raw, err := c.Group.Read(ctx, groupName)
				require.NoError(t, err)
				require.NotNil(t, raw)

				require.Equal(t, groupName, raw.GetName())
				require.Equal(t, "Test comment", raw.GetComment())
				require.Len(t, raw.GetMembers(), len(users))

				usersMap := make(map[pveSDK.UserID]struct{})
				for _, user := range users {
					usersMap[user] = struct{}{}
				}
				expectedMap := make(map[pveSDK.UserID]struct{})
				for _, member := range raw.GetMembers() {
					expectedMap[member] = struct{}{}
				}
				require.Equal(t, usersMap, expectedMap)
			}},
		{name: `Update group`,
			test: func(t *testing.T) {
				require.NoError(t, c.Group.Set(ctx, pveSDK.ConfigGroup{
					Name:    groupName,
					Members: &[]pveSDK.UserID{users[0]},
					Comment: util.Pointer(""),
				}))
			}},
		{name: `Check group updated`,
			test: func(t *testing.T) {
				raw, err := c.Group.Read(ctx, groupName)
				require.NoError(t, err)
				require.NotNil(t, raw)

				require.Equal(t, groupName, raw.GetName())
				require.Equal(t, "", raw.GetComment())
				require.Len(t, raw.GetMembers(), 1)
				require.Equal(t, []pveSDK.UserID{users[0]}, raw.GetMembers())
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
