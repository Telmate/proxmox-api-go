package api_test

import (
	"github.com/stretchr/testify/require"
	"testing"
	"github.com/Telmate/proxmox-api-go/test/api"
)

func Test_List_Acme_Accounts(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	_, err := Test.GetClient().GetAcmeAccountList()
	require.NoError(t, err)
}