package api_test

import (
	"testing"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	api_test "github.com/Telmate/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_Create_Token(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	user, _ := pxapi.NewConfigUserFromApi(pxapi.UserID{Name: "root", Realm: "pam"}, Test.GetClient())

	_, err := user.CreateApiToken(Test.GetClient(), pxapi.ApiToken{TokenId: "testing", Comment: "This is a test", Expire: 1679404904, Privsep: true})
	require.NoError(t, err)
}

func Test_Token_Is_Created(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	user, _ := pxapi.NewConfigUserFromApi(pxapi.UserID{Name: "root", Realm: "pam"}, Test.GetClient())

	tokens, _ := user.ListApiTokens(Test.GetClient())

	listoftokens := *tokens

	t.Log(listoftokens[0].TokenId)
	require.Equal(t, "testing", listoftokens[0].TokenId)
}

func Test_Update_Token(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	user, _ := pxapi.NewConfigUserFromApi(pxapi.UserID{Name: "root", Realm: "pam"}, Test.GetClient())

	tokens, _ := user.ListApiTokens(Test.GetClient())

	listoftokens := *tokens

	listoftokens[0].Comment = "New Comment"

	err := user.UpdateApiToken(Test.GetClient(), listoftokens[0])
	require.NoError(t, err)
}

func Test_Token_Is_Updated(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	user, _ := pxapi.NewConfigUserFromApi(pxapi.UserID{Name: "root", Realm: "pam"}, Test.GetClient())

	tokens, _ := user.ListApiTokens(Test.GetClient())

	listoftokens := *tokens

	require.Equal(t, "New Comment", listoftokens[0].Comment)
}

func Test_Delete_Token(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	user, _ := pxapi.NewConfigUserFromApi(pxapi.UserID{Name: "root", Realm: "pam"}, Test.GetClient())

	tokens, _ := user.ListApiTokens(Test.GetClient())

	listoftokens := *tokens

	err := user.DeleteApiToken(Test.GetClient(), listoftokens[0])
	require.NoError(t, err)
}

func Test_Token_Is_Deleted(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	user, _ := pxapi.NewConfigUserFromApi(pxapi.UserID{Name: "root", Realm: "pam"}, Test.GetClient())

	tokens, _ := user.ListApiTokens(Test.GetClient())

	require.Equal(t, 0, len(*tokens))
}
