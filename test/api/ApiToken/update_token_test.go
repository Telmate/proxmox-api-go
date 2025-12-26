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

func Test_Token_Update(t *testing.T) {
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
	tokenID := pveSDK.ApiTokenID{
		User:      pveSDK.UserID{Name: "Test_Token_Update", Realm: "pve"},
		TokenName: "testToken"}
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
				secret, err := c.ApiToken.Create(ctx, tokenID.User, pveSDK.ApiTokenConfig{Name: tokenID.TokenName})
				require.NoError(t, err)
				require.NotEmpty(t, secret)
			}},
		{name: `Read token`,
			test: func(t *testing.T) {
				raw, err := c.ApiToken.Read(ctx, tokenID)
				require.NoError(t, err)
				require.NotNil(t, raw)
				require.Equal(t, base(pveSDK.ApiTokenConfig{
					Name: tokenID.TokenName,
				}), raw.Get())
			}},
		{name: `Update token`,
			test: func(t *testing.T) {
				require.NoError(t, c.ApiToken.Update(ctx, tokenID.User, pveSDK.ApiTokenConfig{
					Name:                tokenID.TokenName,
					Comment:             util.Pointer("Updated comment"),
					Expiration:          util.Pointer(uint(1766444400)),
					PrivilegeSeparation: util.Pointer(false),
				}))
			}},
		{name: `Read token full details`,
			test: func(t *testing.T) {
				raw, err := c.ApiToken.Read(ctx, tokenID)
				require.NoError(t, err)
				require.NotNil(t, raw)
				require.Equal(t, base(pveSDK.ApiTokenConfig{
					Name:                tokenID.TokenName,
					Comment:             util.Pointer("Updated comment"),
					Expiration:          util.Pointer(uint(1766444400)),
					PrivilegeSeparation: util.Pointer(false),
				}), raw.Get())
			}},
		{name: `Update token Comment`,
			test: func(t *testing.T) {
				require.NoError(t, c.ApiToken.Update(ctx, tokenID.User, pveSDK.ApiTokenConfig{
					Name:    tokenID.TokenName,
					Comment: util.Pointer("My comment with symbols " + body.Symbols),
				}))
			}},
		{name: `Read token updated Comment`,
			test: func(t *testing.T) {
				raw, err := c.ApiToken.Read(ctx, tokenID)
				require.NoError(t, err)
				require.NotNil(t, raw)
				require.Equal(t, base(pveSDK.ApiTokenConfig{
					Name:                tokenID.TokenName,
					Comment:             util.Pointer("My comment with symbols " + body.Symbols),
					Expiration:          util.Pointer(uint(1766444400)),
					PrivilegeSeparation: util.Pointer(false),
				}), raw.Get())
			}},
		{name: `Clear token`,
			test: func(t *testing.T) {
				require.NoError(t, c.ApiToken.Update(ctx, tokenID.User, pveSDK.ApiTokenConfig{
					Name:                tokenID.TokenName,
					Comment:             util.Pointer(""),
					Expiration:          util.Pointer(uint(0)),
					PrivilegeSeparation: util.Pointer(true),
				}))
			}},
		{name: `Read token cleared`,
			test: func(t *testing.T) {
				raw, err := c.ApiToken.Read(ctx, tokenID)
				require.NoError(t, err)
				require.NotNil(t, raw)
				require.Equal(t, base(pveSDK.ApiTokenConfig{
					Name: tokenID.TokenName,
				}), raw.Get())
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
