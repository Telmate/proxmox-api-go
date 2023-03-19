package api_test

import (
	"github.com/stretchr/testify/require"
	"testing"
	"github.com/Telmate/proxmox-api-go/test/api"
)

func Test_Connection_Certificate_No_Validation(t *testing.T) {
	api_test.SetEnvironmentVariables()
	Test := api_test.Test{
		RequireSSL: false,
	}
	err := Test.CreateTest()
	require.NoError(t, err)
}

func Test_Connection_Certificate_Validation(t *testing.T) {
	api_test.SetEnvironmentVariables()
	Test := api_test.Test{
		RequireSSL: true,
	}
	err := Test.CreateTest()
	require.Error(t, err)
}