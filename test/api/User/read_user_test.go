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

func Test_User_Read(t *testing.T) {
	userID := pveSDK.UserID{Name: "Test_User_Read", Realm: "pve"}
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
		{name: `Read non-existing user`,
			test: func(t *testing.T) {
				raw, err := c.User.Read(ctx, userID)
				require.Error(t, err)
				require.Nil(t, raw)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
