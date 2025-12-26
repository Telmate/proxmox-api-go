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

func Test_User_Exists(t *testing.T) {
	userID := pveSDK.UserID{Name: "Test_User_Exists", Realm: "pve"}
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Ensure user does not exist`,
			test: func(t *testing.T) {
				require.NoError(t, c.User.Delete(ctx, userID))
			}},
		{name: `Create user`,
			test: func(t *testing.T) {
				require.NoError(t, c.User.Create(ctx, pveSDK.ConfigUser{
					User: userID,
				}))
			}},
		{name: `Verify user exists`,
			test: func(t *testing.T) {
				exists, err := c.User.Exists(ctx, userID)
				require.NoError(t, err)
				require.Equal(t, true, exists)
			}},
		{name: `Delete user`,
			test: func(t *testing.T) {
				require.NoError(t, c.User.Delete(ctx, userID))
			}},
		{name: `Verify user deleted`,
			test: func(t *testing.T) {
				exists, err := c.User.Exists(ctx, userID)
				require.NoError(t, err)
				require.False(t, exists)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
