package api_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	api_test "github.com/Telmate/proxmox-api-go/test/api"
)

func Test_List_Users(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	users, err := pxapi.ListUsers(context.Background(), Test.GetClient(), false)
	require.NoError(t, err)
	require.Equal(t, 1, len(*users))
}
