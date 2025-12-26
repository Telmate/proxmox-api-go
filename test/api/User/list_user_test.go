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

func Test_User_List(t *testing.T) {
	base := func(info pveSDK.UserInfo) pveSDK.UserInfo {
		if info.Config.Comment == nil {
			info.Config.Comment = util.Pointer("")
		}
		if info.Config.Email == nil {
			info.Config.Email = util.Pointer("")
		}
		if info.Config.Expire == nil {
			info.Config.Expire = util.Pointer(uint(0))
		}
		if info.Config.Enable == nil {
			info.Config.Enable = util.Pointer(true)
		}
		if info.Config.FirstName == nil {
			info.Config.FirstName = util.Pointer("")
		}
		if info.Config.LastName == nil {
			info.Config.LastName = util.Pointer("")
		}
		if info.Config.Groups == nil {
			info.Config.Groups = util.Pointer([]pveSDK.GroupName{})
		}
		if info.Tokens == nil {
			info.Tokens = &[]pveSDK.ApiTokenConfig{}
		}
		for i := range *info.Tokens {
			if (*info.Tokens)[i].Comment == nil {
				(*info.Tokens)[i].Comment = util.Pointer("")
			}
			if (*info.Tokens)[i].Expiration == nil {
				(*info.Tokens)[i].Expiration = util.Pointer(uint(0))
			}
			if (*info.Tokens)[i].PrivilegeSeparation == nil {
				(*info.Tokens)[i].PrivilegeSeparation = util.Pointer(true)
			}
		}
		return info
	}
	stripGroupAndTokens := func(info pveSDK.UserInfo) pveSDK.UserInfo {
		info.Config.Groups = nil
		info.Tokens = nil
		return info
	}
	users := []pveSDK.UserInfo{
		{Config: pveSDK.ConfigUser{
			User:      pveSDK.UserID{Name: "Test_User_List_0", Realm: "pve"},
			Email:     util.Pointer("Bruce@Wayne-industries.com"),
			FirstName: util.Pointer("Bruce"),
			LastName:  util.Pointer("Wayne")}},
		{Config: pveSDK.ConfigUser{
			User:   pveSDK.UserID{Name: "Test_User_List_1", Realm: "pve"},
			Enable: util.Pointer(false)},
			Tokens: &[]pveSDK.ApiTokenConfig{{Name: "token1"}}},
		{Config: pveSDK.ConfigUser{
			User:   pveSDK.UserID{Name: "Test_User_List_2", Realm: "pve"},
			Expire: util.Pointer(uint(987654321))},
			Tokens: &[]pveSDK.ApiTokenConfig{
				{Name: "token1",
					Comment: util.Pointer("Token 1 for user 0")},
				{Name: "token2",
					Expiration: util.Pointer(uint(123456789))},
				{Name: "token3",
					PrivilegeSeparation: util.Pointer(false)}}}}
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Ensure users do not exist`,
			test: func(t *testing.T) {
				for _, user := range users {
					require.NoError(t, c.User.Delete(ctx, user.Config.User))
				}
			}},
		{name: `Create users`,
			test: func(t *testing.T) {
				for _, user := range users {
					require.NoError(t, c.User.Create(ctx, user.Config))
				}
			}},
		{name: `Create Tokes`,
			test: func(t *testing.T) {
				for _, user := range users {
					if user.Tokens == nil {
						continue
					}
					for _, tokenConfig := range *user.Tokens {
						secret, err := c.ApiToken.Create(ctx, user.Config.User, tokenConfig)
						require.NoError(t, err)
						require.NotEmpty(t, secret)
					}
				}
			}},
		{name: `Get user info full`,
			test: func(t *testing.T) {
				raw, err := c.User.List(ctx)
				require.NoError(t, err)
				require.GreaterOrEqual(t, raw.Len(), len(users))
				for _, user := range users {
					rawUser, exists := raw.SelectUser(user.Config.User)
					require.True(t, exists)
					require.Equal(t, base(user), rawUser.Get())
				}
			}},
		{name: `Get user info partial`,
			test: func(t *testing.T) {
				raw, err := c.User.ListPartial(ctx)
				require.NoError(t, err)
				require.GreaterOrEqual(t, raw.Len(), len(users))
				for _, user := range users {
					rawUser, exists := raw.SelectUser(user.Config.User)
					require.True(t, exists)
					require.Equal(t, stripGroupAndTokens(base(user)), rawUser.Get())
				}
			}},
		{name: `Delete user`,
			test: func(t *testing.T) {
				for _, user := range users {
					require.NoError(t, c.User.Delete(ctx, user.Config.User))
				}
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
