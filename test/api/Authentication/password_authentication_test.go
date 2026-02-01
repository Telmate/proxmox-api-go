package api_test

import (
	"context"
	"crypto/tls"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/pad"
	"github.com/Telmate/proxmox-api-go/internal/util"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
)

func Test_Authenticate_Password(t *testing.T) {
	t.Parallel()
	userID := pveSDK.UserID{Name: "Test_Authenticate_Password", Realm: "pve"}
	password := pveSDK.UserPassword("Enter123!" + body.Symbols)
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
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
					User:     userID,
					Password: util.Pointer(password),
				}))
			}},
		{name: `Login in with incorrect password`,
			test: func(t *testing.T) {
				cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
				require.NoError(t, err)
				ctx := context.Background()
				require.Error(t, cl.Login(ctx, userID.String(), "incorrect", ""))
			}},
		{name: `Login in with correct password`,
			test: func(t *testing.T) {
				cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
				require.NoError(t, err)
				ctx := context.Background()
				require.NoError(t, cl.Login(ctx, userID.String(), password.String(), ""))
				version, err := cl.Version(ctx)
				require.NoError(t, err)
				require.NotEmpty(t, version.String())
			}},
		{name: `Delete user`,
			test: func(t *testing.T) {
				require.NoError(t, c.User.Delete(ctx, userID))
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
