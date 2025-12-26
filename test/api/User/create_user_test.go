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

func Test_User_Create(t *testing.T) {
	userID := pveSDK.UserID{Name: "Test_User_Create", Realm: "pve"}
	user := pveSDK.ConfigUser{
		Comment:   util.Pointer("My Comment" + body.Symbols),
		Email:     util.Pointer("test@example.com"),
		Enable:    util.Pointer(false),
		Groups:    util.Pointer([]pveSDK.GroupName{}),
		Expire:    util.Pointer(uint(123456789)),
		FirstName: util.Pointer("First" + body.Symbols),
		LastName:  util.Pointer("Last" + body.Symbols),
		User:      userID,
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
		{name: `Ensure user does not exist`,
			test: func(t *testing.T) {
				require.NoError(t, c.User.Delete(ctx, userID))
			}},
		{name: `Create user`,
			test: func(t *testing.T) {
				require.NoError(t, c.User.Create(ctx, user))
			}},
		{name: `Verify user exists`,
			test: func(t *testing.T) {
				raw, err := c.User.Read(ctx, userID)
				require.NoError(t, err)
				require.NotNil(t, raw)
				require.Equal(t, &user, raw.Get())
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
