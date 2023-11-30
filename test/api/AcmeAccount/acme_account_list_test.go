package api_test

import (
	"testing"

	api_test "github.com/Bluearchive/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_List_Acme_Accounts(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	_, err := Test.GetClient().GetAcmeAccountList()
	require.NoError(t, err)
}
