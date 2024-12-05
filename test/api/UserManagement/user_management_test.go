package api_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	api_test "github.com/Telmate/proxmox-api-go/test/api"
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
	err := user.CreateUser(context.Background(), Test.GetClient())
	require.NoError(t, err)
}

func Test_User_Is_Added(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	users, _ := pxapi.ListUsers(context.Background(), Test.GetClient(), false)
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
	err := user.DeleteUser(context.Background(), Test.GetClient())
	require.NoError(t, err)
}

func Test_User_Is_Removed(t *testing.T) {
	Test := api_test.Test{}
	_ = Test.CreateTest()

	users, _ := pxapi.ListUsers(context.Background(), Test.GetClient(), false)
	var found = false
	for _, element := range *users {
		if element == user {
			found = true
		}
	}

	require.Equal(t, false, found)
}
