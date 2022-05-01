package test

import (
	"testing"
	"bytes"
	"io/ioutil"
	"github.com/Telmate/proxmox-api-go/cli"
	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Test struct {
	Name string
	Expected string
	NotExpected string
	ReqErr bool
	Contains bool
	NotContains bool
	Args []string
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
	err := cli.RootCmd.Execute()

	if test.ReqErr {
		require.Error(t, err)
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
}
