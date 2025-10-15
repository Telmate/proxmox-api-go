package api_test

import (
	"context"
	"log"
	"testing"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	api_test "github.com/Telmate/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

const updatedMTU = 1337

func _create_sdn_zone_config() (pxapi.ConfigSDNZone) {
	config := pxapi.ConfigSDNZone{
		Type: "simple",
		Zone: "testzone",  // Less or equal than 8 char
	};

	return config
}

func Test_Create_Sdn_Zone(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	config := _create_sdn_zone_config()

	err := config.Create(context.Background(), config.Zone, Test.GetClient())
	require.NoError(
		t, err)
}

func Test_Sdn_Zone_Is_Added(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	config := _create_sdn_zone_config()

	_, err := Test.GetClient().GetSDNZone(context.Background(), config.Zone)
	require.NoError(t, err)
}

func Test_Update_Sdn_Zone(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	config := _create_sdn_zone_config()
	config.MTU = 1337

	err := config.UpdateWithValidate(context.Background(), config.Zone, Test.GetClient())
	require.NoError(t, err)
}

func Test_Sdn_Zone_Is_Updated(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	initialConfig := _create_sdn_zone_config()

	config, err := pxapi.NewConfigSDNZoneFromApi(context.Background(), initialConfig.Zone, Test.GetClient())
	require.NoError(t, err)
	log.Println(config)
	require.Equal(t, config.MTU, updatedMTU)
}

func Test_Delete_Sdn_Zone(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	config := _create_sdn_zone_config()

	err := Test.GetClient().DeleteSDNZone(context.Background(), config.Zone)
	require.NoError(t, err)
}

func Test_Sdn_Zone_Is_Deleted(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	config := _create_sdn_zone_config()

	_, err := Test.GetClient().GetSDNZone(context.Background(), config.Zone)
	require.Error(t, err)
}
