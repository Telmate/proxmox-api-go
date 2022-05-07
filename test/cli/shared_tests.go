package test

import (
	"testing"
	"bytes"
	"io/ioutil"
	"strings"

	"github.com/Telmate/proxmox-api-go/cli"
	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Test struct {
	InputJson string //the inputted json
	OutputJson string //the outputted json

	Expected string //the output that is expected
	Contains bool //if the output contains (expected) or qeuals it

	NotExpected string //the output that is notexpected
	NotContains bool //if the output contains (notexpected) or qeuals it

	ReqErr bool //if an error is expected as output
	ErrContains string //the string the error should contain

	Args []string //cli arguments
}

func ListTest(t *testing.T, args []string, expected string) {
	cli.RootCmd.SetArgs(append(args))
	
	buffer := new(bytes.Buffer)

	cli.RootCmd.SetOut(buffer)
	err := cli.RootCmd.Execute()
	require.NoError(t, err)

	out, _ := ioutil.ReadAll(buffer)
	assert.Contains(t, string(out), expected)
}

func (test *Test) StandardTest(t *testing.T) {
	SetEnvironmentVariables()
	cli.RootCmd.SetArgs(test.Args)
	buffer := new(bytes.Buffer)
	cli.RootCmd.SetOut(buffer)
	cli.RootCmd.SetIn(strings.NewReader(test.InputJson))
	err := cli.RootCmd.Execute()

	if test.ReqErr {
		require.Error(t, err)
		if test.ErrContains != "" {
			assert.Contains(t, err.Error(), test.ErrContains)
		}
	} else {
		require.NoError(t, err)
	}
	if test.Expected != "" {
		out, _ := ioutil.ReadAll(buffer)
		if test.Contains {
			assert.Contains(t, string(out), test.Expected)
		} else {
			assert.Equal(t, string(out), test.Expected)
		}
	}
	if test.NotExpected != "" {
		out, _ := ioutil.ReadAll(buffer)
		if test.NotContains {
			assert.NotContains(t, string(out), test.NotExpected)
		} else {
			assert.NotEqual(t, string(out), test.NotExpected)
		}
	}
	if test.OutputJson != "" {
		out, _ := ioutil.ReadAll(buffer)
		require.JSONEq(t, test.OutputJson ,string(out))
	}
}
