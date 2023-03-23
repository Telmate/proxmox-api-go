package test

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/perimeter-81/proxmox-api-go/cli"
	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Test struct {
	InputJson  any //the inputted json
	OutputJson any //the outputted json

	Expected string   //matches the output exactly
	Contains []string //the output contains all of the strings

	NotExpected string   //the output that is not expected
	NotContains []string //the output may not contain any of these strings

	ReqErr      bool   //if an error is expected as output
	ErrContains string //the string the error should contain

	// TODO remove is obsolete
	Return bool //if the output should be read and returned for more advanced processing

	Args []string //cli arguments
}

func ListTest(t *testing.T, args []string, expected string) {
	cli.RootCmd.SetArgs(args)

	buffer := new(bytes.Buffer)

	cli.RootCmd.SetOut(buffer)
	err := cli.RootCmd.Execute()
	require.NoError(t, err)

	out, _ := io.ReadAll(buffer)
	assert.Contains(t, string(out), expected)
}

func (test *Test) StandardTest(t *testing.T) (out []byte) {
	SetEnvironmentVariables()
	cli.RootCmd.SetArgs(test.Args)
	buffer := new(bytes.Buffer)
	cli.RootCmd.SetOut(buffer)

	switch InputJson := test.InputJson.(type) {
	case string:
		if InputJson != "" {
			cli.RootCmd.SetIn(strings.NewReader(InputJson))
		}
	default:
		if InputJson != nil {
			tmpJson, err := json.Marshal(InputJson)
			require.NoError(t, err)
			cli.RootCmd.SetIn(strings.NewReader(string(tmpJson)))
		}
	}

	err := cli.RootCmd.Execute()

	out, _ = io.ReadAll(buffer)

	if test.ReqErr {
		require.Error(t, err)
		if test.ErrContains != "" {
			assert.Contains(t, err.Error(), test.ErrContains)
		}
	} else {
		require.NoError(t, err)
	}
	if test.Expected != "" {
		assert.Equal(t, string(out), test.Expected)
	}
	if len(test.Contains) != 0 {
		for _, e := range test.Contains {
			assert.Contains(t, string(out), e)
		}
	}
	if test.NotExpected != "" {
		assert.NotEqual(t, string(out), test.NotExpected)
	}
	if len(test.NotContains) != 0 {
		for _, e := range test.NotContains {
			assert.NotContains(t, string(out), e)
		}
	}
	switch outputJson := test.OutputJson.(type) {
	case string:
		if outputJson != "" {
			require.JSONEq(t, outputJson, string(out))
		}
	default:
		if outputJson != nil {
			tmpJson, err := json.Marshal(outputJson)
			require.NoError(t, err)
			require.JSONEq(t, string(tmpJson), string(out))
		}
	}
	return
}

type LoginTest struct {
	APIurl      string
	UserID      string
	Password    string
	OTP         string
	HttpHeaders string
	ReqErr      bool //if an error is expected as output
}

func (test *LoginTest) Login(t *testing.T) {
	_, err := cli.Client(test.APIurl, test.UserID, test.Password, test.OTP, test.HttpHeaders)
	if test.ReqErr {
		require.Error(t, err)
	} else {
		require.NoError(t, err)
	}
}
