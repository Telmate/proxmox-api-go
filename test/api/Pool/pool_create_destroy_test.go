package api_test

import (
	"testing"

	api_test "github.com/Bluearchive/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_Pool_Create(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	Test.GetClient().CreatePool("test-pool", "Test pool")
}

func Test_Pool_Is_Created(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	_, err := Test.GetClient().GetPoolInfo("test-pool")
	require.NoError(t, err)
}

func Test_Pool_Delete(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	Test.GetClient().DeletePool("test-pool")
}

func Test_Pool_Is_Deleted(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	_, err := Test.GetClient().GetPoolInfo("test-pool")
	require.Error(t, err)
}
