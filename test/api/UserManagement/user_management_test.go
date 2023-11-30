package api_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	pxapi "github.com/Bluearchive/proxmox-api-go/proxmox"
	api_test "github.com/Bluearchive/proxmox-api-go/test/api"
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

func Test_Create_User(t *testing.T) {
	Test := api_test.Test{}
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
	Test := api_test.Test{}
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
