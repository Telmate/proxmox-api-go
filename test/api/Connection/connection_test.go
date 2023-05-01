package api_test

import (
	"testing"

	api_test "github.com/Telmate/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_Connection_Certificate_No_Validation(t *testing.T) {
	Test := api_test.Test{
		RequireSSL: false,
	}
	err := Test.CreateTest()
	require.NoError(t, err)
}

func Test_Connection_Certificate_Validation(t *testing.T) {
	Test := api_test.Test{
		RequireSSL: true,
	}
	err := Test.CreateTest()
	require.Error(t, err)
}
