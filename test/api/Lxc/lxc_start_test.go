package api_test

import (
	"testing"

	api_test "github.com/Bluearchive/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_Start_Stop_Lxc_Container_Setup(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	config := _create_lxc_spec(false)
	config.CreateLxc(_create_vmref(), Test.GetClient())
}

func Test_Start_Lxc_Container(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	_, err := Test.GetClient().StartVm(_create_vmref())
	require.NoError(t, err)
}

func Test_Stop_Lxc_Container(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	_, err := Test.GetClient().StopVm(_create_vmref())
	require.NoError(t, err)
}

func Test_Start_Stop_Lxc_Container_Cleanup(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	Test.GetClient().DeleteVm(_create_vmref())
}
