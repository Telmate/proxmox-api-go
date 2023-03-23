package group_sub_tests

import (
	"encoding/json"
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
	"github.com/stretchr/testify/require"
)

// Default CLEANUP test for Group
func Cleanup(t *testing.T, group proxmox.GroupName) {
	Test := &cliTest.Test{
		ReqErr:      true,
		ErrContains: string(group),
		Args:        []string{"-i", "delete", "group", string(group)},
	}
	Test.StandardTest(t)
}

// Default DELETE test for Group
func Delete(t *testing.T, group proxmox.GroupName) {
	Test := &cliTest.Test{
		Contains: []string{string(group)},
		Args:     []string{"-i", "delete", "group", string(group)},
	}
	Test.StandardTest(t)
}

// Default GET test for Group
func Get(t *testing.T, group proxmox.ConfigGroup) {
	Test := &cliTest.Test{
		Args: []string{"-i", "get", "group", string(group.Name)},
	}
	Get_Test(t, group, Test.StandardTest(t))
}

// Custom test as require.JSONEq() wont work here due to *[]UserID being an unordered list
func Get_Test(t *testing.T, expected proxmox.ConfigGroup, actualRaw []byte) {
	actual := proxmox.ConfigGroup{}
	require.NoError(t, json.Unmarshal(actualRaw, &actual))
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Comment, actual.Comment)
	if expected.Members != nil && actual.Members != nil {
		require.ElementsMatch(t, *expected.Members, *actual.Members)
	} else {
		t.FailNow()
	}
}
