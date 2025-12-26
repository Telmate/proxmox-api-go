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

func Test_Token_Delete(t *testing.T) {
	tokenID := pveSDK.ApiTokenID{
		User:      pveSDK.UserID{Name: "Test_Token_Delete", Realm: "pve"},
		TokenName: "testToken"}
	secret := util.Pointer(pveSDK.ApiTokenSecret(""))
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
				require.NoError(t, c.User.Delete(ctx, tokenID.User))
			}},
		{name: `Create user`,
			test: func(t *testing.T) {
				require.NoError(t, c.User.Create(ctx, pveSDK.ConfigUser{User: tokenID.User}))
			}},
		{name: `Create token`,
			test: func(t *testing.T) {
				var err error
				*secret, err = c.ApiToken.Create(ctx, tokenID.User, pveSDK.ApiTokenConfig{
					Name:                tokenID.TokenName,
					Comment:             util.Pointer("This is a test token"),
					PrivilegeSeparation: util.Pointer(true)})
				require.NoError(t, err)
				require.NotEmpty(t, *secret)
			}},
		{name: `Delete token`,
			test: func(t *testing.T) {
				deleted, err := c.ApiToken.Delete(ctx, tokenID)
				require.NoError(t, err)
				require.True(t, deleted)
			}},
		{name: `Delete token no effect`,
			test: func(t *testing.T) {
				deleted, err := c.ApiToken.Delete(ctx, tokenID)
				require.NoError(t, err)
				require.False(t, deleted)
			}},
		{name: `Delete user`,
			test: func(t *testing.T) {
				require.NoError(t, c.User.Delete(ctx, tokenID.User))
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
