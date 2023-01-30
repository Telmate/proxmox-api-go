package cli_snapshot_test

import (
	"encoding/json"
	"testing"

	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Snapshot_0_GuestQemu_300_Cleanup(t *testing.T) {
	Test := cliTest.Test{
		ReqErr:      true,
		ErrContains: "300",
		Args:        []string{"-i", "delete", "guest", "300"},
	}
	Test.StandardTest(t)
}

func Test_Snapshot_0_GuestQemu_300_Create(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"name": "test-qemu300",
	"memory": 128,
	"os": "l26",
	"cores": 1,
	"sockets": 1
}`,
		Expected: "(300)",
		Contains: true,
		Args:     []string{"-i", "create", "guest", "qemu", "300", "pve"},
	}
	Test.StandardTest(t)
}

func Test_Snapshot_0_GuestQemu_300_Start(t *testing.T) {
	Test := cliTest.Test{
		Expected: "(300)",
		Contains: true,
		Args:     []string{"-i", "guest", "start", "300"},
	}
	Test.StandardTest(t)
}

// Create a snapshot with all settings populated
func Test_Snapshot_0_Create_Full(t *testing.T) {
	Test := cliTest.Test{
		Expected: "(snap00)",
		Contains: true,
		Args:     []string{"-i", "create", "snapshot", "300", "snap00", "description00", "--memory"},
	}
	Test.StandardTest(t)
}

// Check if the snapshot was made properly and the right json structure is returned (tree)
func Test_Snapshot_0_Get_Full(t *testing.T) {
	Test := cliTest.Test{
		Return: true,
		Args:   []string{"-i", "list", "snapshots", "300"},
	}
	var data []*snapshot
	require.NoError(t, json.Unmarshal(Test.StandardTest(t), &data))
	assert.Equal(t, "snap00", data[0].Name)
	assert.Equal(t, "description00", data[0].Description)
	assert.Equal(t, true, data[0].VmState)
	assert.GreaterOrEqual(t, data[0].SnapTime, uint(0))
}

// Remove the description of the snapshot
func Test_Snapshot_0_Update_Description_Empty(t *testing.T) {
	Test := cliTest.Test{
		Args: []string{"-i", "update", "snapshot", "300", "snap00", ""},
	}
	Test.StandardTest(t)
}

// Check if description is removed and the right json structure is returned (no tree)
func Test_Snapshot_0_Get_Description_Empty(t *testing.T) {
	Test := cliTest.Test{
		NotExpected: "description00",
		NotContains: true,
		Return:      true,
		Args:        []string{"-i", "list", "snapshots", "300", "--no-tree"},
	}
	var data []snapshot
	require.NoError(t, json.Unmarshal(Test.StandardTest(t), &data))
}

// Create a snapshot with the least settings populated
func Test_Snapshot_0_Create_Empty(t *testing.T) {
	// t.(time.Second*120, true)
	// time.Sleep(time.Second * 20)
	Test := cliTest.Test{
		Expected: "(snap01)",
		Contains: true,
		Args:     []string{"-i", "create", "snapshot", "300", "snap01"},
	}
	Test.StandardTest(t)
}

// Check if the snapshot was made properly
func Test_Snapshot_0_Get_Empty(t *testing.T) {
	// time.Sleep(time.Second * 5)
	Test := cliTest.Test{
		Return: true,
		Args:   []string{"-i", "list", "snapshots", "300"},
	}
	var data []*snapshot
	require.NoError(t, json.Unmarshal(Test.StandardTest(t), &data))
	assert.Equal(t, "snap01", data[0].Children[0].Name)
	assert.Equal(t, "", data[0].Children[0].Description)
	assert.Equal(t, false, data[0].Children[0].VmState)
	assert.GreaterOrEqual(t, data[0].Children[0].SnapTime, uint(0))
}

// Add the description to the snapshot
func Test_Snapshot_0_Update_Description_Full(t *testing.T) {
	Test := cliTest.Test{
		Args: []string{"-i", "update", "snapshot", "300", "snap01", "description01"},
	}
	Test.StandardTest(t)
}

// Check if description is added
func Test_Snapshot_0_Get_Description_Full(t *testing.T) {
	Test := cliTest.Test{
		Expected: "description01",
		Contains: true,
		Return:   true,
		Args:     []string{"-i", "list", "snapshots", "300"},
	}
	var data []*snapshot
	require.NoError(t, json.Unmarshal(Test.StandardTest(t), &data))
}

// rollback snapshot
func Test_Snapshot_0_Set_Rollback(t *testing.T) {
	Test := cliTest.Test{
		Expected: "(snap00)",
		Contains: true,
		Args:     []string{"-i", "guest", "rollback", "300", "snap00"},
	}
	Test.StandardTest(t)
}

// Check if the snapshot was rolled back
func Test_Snapshot_0_Get_Rollback(t *testing.T) {
	Test := cliTest.Test{
		Return: true,
		Args:   []string{"-i", "list", "snapshots", "300", "--no-tree"},
	}
	var data []*snapshot
	var nofail bool
	require.NoError(t, json.Unmarshal(Test.StandardTest(t), &data))
	for _, e := range data {
		if e.Name == "current" {
			assert.Equal(t, "You are here!", e.Description)
			assert.Equal(t, "snap00", e.Parent)
			nofail = true
			break
		}
	}
	assert.Equal(t, true, nofail)
}

// delete snapshot
func Test_Snapshot_0_Delete(t *testing.T) {
	Test := cliTest.Test{
		Expected: "(snap00)",
		Contains: true,
		Args:     []string{"-i", "delete", "snapshot", "300", "snap00"},
	}
	Test.StandardTest(t)
}

// Check if the snapshot was deleted
func Test_Snapshot_0_Get_Delete(t *testing.T) {
	Test := cliTest.Test{
		NotExpected: "snap00",
		NotContains: true,
		Args:        []string{"-i", "list", "snapshots", "300"},
	}
	Test.StandardTest(t)
}

func Test_Snapshot_0_GuestQemu_300_Delete(t *testing.T) {
	Test := cliTest.Test{
		ReqErr: false,
		Args:   []string{"-i", "delete", "guest", "300"},
	}
	Test.StandardTest(t)
}

type snapshot struct {
	Name        string      `json:"name"`
	SnapTime    uint        `json:"time,omitempty"`
	Description string      `json:"description,omitempty"`
	VmState     bool        `json:"ram,omitempty"`
	Children    []*snapshot `json:"children,omitempty"`
	Parent      string      `json:"parent,omitempty"`
}
