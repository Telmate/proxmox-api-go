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


func ListTest(t *testing.T, args []string, expected string) {
	cli.RootCmd.SetArgs(append(args))
	
	buffer := new(bytes.Buffer)

	cli.RootCmd.SetOut(buffer)
	err := cli.RootCmd.Execute()
	require.NoError(t, err)

	out, _ := ioutil.ReadAll(buffer)
	assert.Contains(t, string(out), expected)
}
