package api_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	//	"os"
	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	api_test "github.com/Telmate/proxmox-api-go/test/api"
)

var account = `
{
"contact": [
"a@nonexistantdomain.com",
"b@nonexistantdomain.com",
"c@nonexistantdomain.com",
"d@nonexistantdomain.com"
],
"directory": "https://acme-staging-v02.api.letsencrypt.org/directory",
"tos": true
}`

func Test_Create_Acme_Account(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	acmeAccount, _ := pxapi.NewConfigAcmeAccountFromJson([]byte(account))
	err := acmeAccount.CreateAcmeAccount(context.Background(), "test", Test.GetClient())
	require.NoError(t, err)
}

func Test_Acme_Account_Is_Added(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	_, err := pxapi.NewConfigAcmeAccountFromApi(context.Background(), "test", Test.GetClient())

	require.NoError(t, err)
}

func Test_Remove_Acme_Account(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	_, err := Test.GetClient().DeleteAcmeAccount(context.Background(), "test")

	require.NoError(t, err)
}

func Test_Acme_Account_Is_Removed(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	_, err := pxapi.NewConfigAcmeAccountFromApi(context.Background(), "test", Test.GetClient())

	require.Error(t, err)
}
