package api_test

import (
	"testing"

	pxapi "github.com/Bluearchive/proxmox-api-go/proxmox"
	api_test "github.com/Bluearchive/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_Create_Lxc_Container(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	config := _create_lxc_spec(true)

	err := config.CreateLxc(_create_vmref(), Test.GetClient())
	require.NoError(t, err)
}

func Test_Lxc_Container_Is_Added(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	config, _ := pxapi.NewConfigLxcFromApi(_create_vmref(), Test.GetClient())

	require.Equal(t, "alpine", config.OsType)
}

func Test_Update_Lxc_Container(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	config, _ := pxapi.NewConfigLxcFromApi(_create_vmref(), Test.GetClient())

	config.Cores = 2

	err := config.UpdateConfig(_create_vmref(), Test.GetClient())

	require.NoError(t, err)
}

func Test_Lxc_Container_Is_Updated(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	config, _ := pxapi.NewConfigLxcFromApi(_create_vmref(), Test.GetClient())
	require.Equal(t, 2, config.Cores)
}

func Test_Remove_Lxc_Container(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	_, err := Test.GetClient().DeleteVm(_create_vmref())

	require.NoError(t, err)
}

func Test_Create_Template_Lxc_Container(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	config := _create_lxc_spec(true)

	vmRef := _create_vmref()
	err := config.CreateLxc(vmRef, Test.GetClient())
	require.NoError(t, err)

	err = Test.GetClient().CreateTemplate(vmRef)
	require.NoError(t, err)
}
