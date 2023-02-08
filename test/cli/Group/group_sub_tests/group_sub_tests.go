package group_sub_tests

import (
	"encoding/json"
	"testing"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

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
