package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ConfigGroup_mapToStruct(t *testing.T) {
	testMembers0_Output := &[]UserID{
		{Name: "user1", Realm: "pam"},
		{Name: "user2", Realm: "pam"},
		{Name: "user3", Realm: "pam"},
	}
	testMembers1_Output := &[]UserID{
		{Name: "user4", Realm: "pam"},
		{Name: "user5", Realm: "pam"},
		{Name: "user6", Realm: "pam"},
	}
	testData := []struct {
		base   ConfigGroup
		input  map[string]interface{}
		output *ConfigGroup
	}{
		// Empty
		{output: &ConfigGroup{}},
		// Only group Name
		{
			input:  map[string]interface{}{"groupid": "group0"},
			output: &ConfigGroup{Name: "group0"},
		},
		{
			base:   ConfigGroup{Name: "group1"},
			output: &ConfigGroup{Name: "group1"},
		},
		{
			base:   ConfigGroup{Name: "group1"},
			input:  map[string]interface{}{"groupid": "group0"},
			output: &ConfigGroup{Name: "group0"},
		},
		// Only group Comment
		{
			input:  map[string]interface{}{"comment": "test Comment"},
			output: &ConfigGroup{Comment: "test Comment"},
		},
		{
			base:   ConfigGroup{Comment: "Comment1"},
			output: &ConfigGroup{Comment: "Comment1"},
		},
		{
			base:   ConfigGroup{Comment: "Comment1"},
			input:  map[string]interface{}{"comment": "test Comment"},
			output: &ConfigGroup{Comment: "test Comment"},
		},
		// Only group Members
		{
			input:  map[string]interface{}{"members": []interface{}{"user1@pam", "user2@pam", "user3@pam"}},
			output: &ConfigGroup{Members: testMembers0_Output},
		},
		{
			base:   ConfigGroup{Members: testMembers1_Output},
			output: &ConfigGroup{Members: testMembers1_Output},
		},
		{
			base:   ConfigGroup{Members: testMembers1_Output},
			input:  map[string]interface{}{"members": []interface{}{"user1@pam", "user2@pam", "user3@pam"}},
			output: &ConfigGroup{Members: testMembers0_Output},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, ConfigGroup{}.mapToStruct(e.input))
	}
}
