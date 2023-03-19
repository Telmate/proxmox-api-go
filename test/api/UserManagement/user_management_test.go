package api_test

import (
	"github.com/stretchr/testify/require"
	"testing"
//	"os"
	"github.com/Telmate/proxmox-api-go/test/api"
	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
)

var user = pxapi.ConfigUser{
	User: pxapi.UserID{
		Name:  "Bob",
		Realm: "pve",
		},
		Comment:   "",
		Email:     "bob@example.com",
		Enable:    true,
		FirstName: "Bob",
		LastName:  "Bobson",
}

func Test_List_Users(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()
	users, err := pxapi.ListUsers(Test.GetClient(), false)
	require.NoError(t, err)
	require.Equal(t, 1, len(*users))
}

func Test_Create_User(t *testing.T) {
	Test := api_test.Test {}
	_ = Test.CreateTest()
    err := user.CreateUser(Test.GetClient())
	require.NoError(t, err)
}

func Test_User_Is_Added(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	users, _ := pxapi.ListUsers(Test.GetClient(), false)
	var found = false
	for _, element := range *users {
		if element == user {
			found = true
		}
	}

	require.Equal(t, true, found)
}

func Test_Remove_User(t *testing.T) {
	Test := api_test.Test {}
	_ = Test.CreateTest()
	err := user.DeleteUser(Test.GetClient())
	require.NoError(t, err)
}

func Test_User_Is_Removed(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	users, _ := pxapi.ListUsers(Test.GetClient(), false)
	var found = false
	for _, element := range *users {
		if element == user {
			found = true
		}
	}

	require.Equal(t, false, found)
}
