package api_test

import (
	"testing"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	api_test "github.com/Telmate/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_Root_Login_Correct_Api_Key(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	user, _ := pxapi.NewConfigUserFromApi(pxapi.UserID{Name: "root", Realm: "pam"}, Test.GetClient())

	token := pxapi.ApiToken{TokenId: "testing", Comment: "This is a test", Expire: 0, Privsep: false}

	value, err := user.CreateApiToken(Test.GetClient(), token)

	NewTest := api_test.Test{}
	NewTest.CreateClient()
	NewTest.GetClient().SetAPIToken("root@pam!testing", value)

	_, err = NewTest.GetClient().GetVersion()
	require.NoError(t, err)

	user.DeleteApiToken(Test.GetClient(), token)
}
