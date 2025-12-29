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

func Test_Group_Delete(t *testing.T) {
	groupName := pveSDK.GroupName("Test_Group_Delete")
	group := pveSDK.ConfigGroup{
		Name: groupName,
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
		{name: `Create group`,
			test: func(t *testing.T) {
				require.NoError(t, c.Group.Create(ctx, group))
			}},
		{name: `Delete group`,
			test: func(t *testing.T) {
				exists, err := c.Group.Delete(ctx, groupName)
				require.NoError(t, err)
				require.True(t, exists)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
