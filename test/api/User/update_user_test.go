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

func Test_User_Update(t *testing.T) {
	t.Parallel()
	base := func(user pveSDK.ConfigUser) pveSDK.ConfigUser {
		if user.Comment == nil {
			user.Comment = util.Pointer("")
		}
		if user.Email == nil {
			user.Email = util.Pointer("")
		}
		if user.Expire == nil {
			user.Expire = util.Pointer(uint(0))
		}
		if user.Enable == nil {
			user.Enable = util.Pointer(true)
		}
		if user.FirstName == nil {
			user.FirstName = util.Pointer("")
		}
		if user.LastName == nil {
			user.LastName = util.Pointer("")
		}
		if user.Groups == nil {
			user.Groups = util.Pointer([]pveSDK.GroupName{})
		}
		return user
	}
	userID := pveSDK.UserID{Name: "Test_User_Update", Realm: "pve"}
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	update := func(t *testing.T, config pveSDK.ConfigUser) {
		require.NoError(t, c.User.Update(ctx, config))
		raw, err := c.User.Read(ctx, config.User)
		require.NoError(t, err)
		require.NotNil(t, raw)
		require.Equal(t, &config, raw.Get())
	}
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
		{name: `Update user Comment`,
			test: func(t *testing.T) {
				update(t, base(pveSDK.ConfigUser{
					User:    userID,
					Comment: util.Pointer("Updated Comment"),
				}))
			}},
		{name: `Update user Email`,
			test: func(t *testing.T) {
				update(t, base(pveSDK.ConfigUser{
					User:  userID,
					Email: util.Pointer("test@example.com"),
				}))
			}},
		{name: `Update user Enable`,
			test: func(t *testing.T) {
				update(t, base(pveSDK.ConfigUser{
					User:   userID,
					Enable: util.Pointer(false),
				}))
			}},
		{name: `Update user Expire`,
			test: func(t *testing.T) {
				update(t, base(pveSDK.ConfigUser{
					User:   userID,
					Expire: util.Pointer(uint(123456789)),
				}))
			}},
		{name: `Update user FirstName`,
			test: func(t *testing.T) {
				update(t, base(pveSDK.ConfigUser{
					User:      userID,
					FirstName: util.Pointer("Updated FirstName"),
				}))
			}},
		{name: `Update user LastName`,
			test: func(t *testing.T) {
				update(t, base(pveSDK.ConfigUser{
					User:     userID,
					LastName: util.Pointer("Updated LastName"),
				}))
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
