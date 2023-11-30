package api_test

import (
	"os"
	"testing"

	api_test "github.com/Bluearchive/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_Root_Login_Correct_Password(t *testing.T) {
	api_test.SetEnvironmentVariables()
	Test := api_test.Test{
		UserID:   os.Getenv("PM_USER"),
		Password: os.Getenv("PM_PASS"),
	}
	err := Test.Login()
	require.NoError(t, err)
}

func Test_Root_Login_Incorrect_Password(t *testing.T) {
	api_test.SetEnvironmentVariables()
	Test := api_test.Test{
		UserID:   os.Getenv("PM_USER"),
		Password: "xx",
	}
	err := Test.Login()
	require.Error(t, err)
}

func Test_Login_Incorrect_Username(t *testing.T) {
	api_test.SetEnvironmentVariables()
	Test := api_test.Test{
		UserID:   "xx",
		Password: "xx",
	}
	err := Test.Login()
	require.Error(t, err)
}
