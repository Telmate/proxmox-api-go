package api_test

import (
	"context"
	"testing"

	api_test "github.com/Telmate/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_Start_Stop_Lxc_Container_Setup(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	config := _create_lxc_spec(false)
	config.CreateLxc(context.Background(), _create_vmref(), Test.GetClient())
}

func Test_Start_Lxc_Container(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	_, err := Test.GetClient().StartVm(context.Background(), _create_vmref())
	require.NoError(t, err)
}

func Test_Stop_Lxc_Container(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	err := _create_vmref().Stop(context.Background(), Test.GetClient())
	require.NoError(t, err)
}

func Test_Start_Stop_Lxc_Container_Cleanup(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	_ = _create_vmref().Delete(context.Background(), Test.GetClient())
}
