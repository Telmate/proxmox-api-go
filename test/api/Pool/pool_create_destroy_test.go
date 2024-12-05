package api_test

import (
	"context"
	"testing"

	"github.com/Telmate/proxmox-api-go/proxmox"
	api_test "github.com/Telmate/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_Pool_Create(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	proxmox.ConfigPool{Name: "test-pool"}.Create(context.Background(), Test.GetClient())
}

func Test_Pool_Is_Created(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	_, err := Test.GetClient().GetPoolInfo(context.Background(), "test-pool")
	require.NoError(t, err)
}

func Test_Pool_Delete(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	proxmox.ConfigPool{Name: "test-pool"}.Create(context.Background(), Test.GetClient())
}

func Test_Pool_Is_Deleted(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	_, err := Test.GetClient().GetPoolInfo(context.Background(), "test-pool")
	require.Error(t, err)
}
