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

func Test_Token_List(t *testing.T) {
	base := func(in pveSDK.ApiTokenConfig) pveSDK.ApiTokenConfig {
		if in.Comment == nil {
			in.Comment = util.Pointer("")
		}
		if in.Expiration == nil {
			in.Expiration = util.Pointer(uint(0))
		}
		if in.PrivilegeSeparation == nil {
			in.PrivilegeSeparation = util.Pointer(true)
		}
		return in
	}
	tokens := []struct {
		input  pveSDK.ApiTokenConfig
		output pveSDK.ApiTokenConfig
	}{
		{
			input: pveSDK.ApiTokenConfig{
				Name:    "token1",
				Comment: util.Pointer("test comment"),
			},
			output: base(pveSDK.ApiTokenConfig{
				Name:    "token1",
				Comment: util.Pointer("test comment"),
			})},
		{
			input: pveSDK.ApiTokenConfig{
				Name:       "token2",
				Expiration: util.Pointer(uint(12345)),
			},
			output: base(pveSDK.ApiTokenConfig{
				Name:       "token2",
				Expiration: util.Pointer(uint(12345)),
			})},
		{
			input: pveSDK.ApiTokenConfig{
				Name:                "token3",
				PrivilegeSeparation: util.Pointer(false),
			},
			output: base(pveSDK.ApiTokenConfig{
				Name:                "token3",
				PrivilegeSeparation: util.Pointer(false),
			})},
	}
	user := pveSDK.UserID{Name: "Test_Token_List", Realm: "pve"}
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
				require.NoError(t, c.User.Delete(ctx, user))
			}},
		{name: `Create user`,
			test: func(t *testing.T) {
				require.NoError(t, c.User.Create(ctx, pveSDK.ConfigUser{User: user}))
			}},
		{name: `Create tokens`,
			test: func(t *testing.T) {
				for i := range tokens {
					secret, err := c.ApiToken.Create(ctx, user, tokens[i].input)
					require.NoError(t, err)
					require.NotEmpty(t, secret)
				}
			}},
		{name: `List tokens`,
			test: func(t *testing.T) {
				raw, err := c.ApiToken.List(ctx, user)
				require.NoError(t, err)
				rawMap := raw.AsMap()
				for i := range tokens {
					rawToken, exists := rawMap[tokens[i].output.Name]
					require.True(t, exists)
					require.Equal(t, tokens[i].output, rawToken.Get())
				}
			}},
		{name: `Delete user`,
			test: func(t *testing.T) {
				require.NoError(t, c.User.Delete(ctx, user))
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
