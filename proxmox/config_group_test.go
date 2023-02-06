package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ConfigGroup_mapToStruct(t *testing.T) {
	testData := []struct {
		input  map[string]interface{}
		output *ConfigGroup
	}{
		{output: &ConfigGroup{}},
		{
			input: map[string]interface{}{
				"groupid": "group",
			},
			output: &ConfigGroup{Name: "group"},
		},
		{
			input: map[string]interface{}{
				"comment": "test Comment",
			},
			output: &ConfigGroup{Comment: "test Comment"},
		},
		{
			input: map[string]interface{}{
				"members": []interface{}{"user1@pam", "user2@pam", "user3@pam"},
			},
			output: &ConfigGroup{Members: &[]UserID{
				{Name: "user1", Realm: "pam"},
				{Name: "user2", Realm: "pam"},
				{Name: "user3", Realm: "pam"},
			}},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, ConfigGroup{}.mapToStruct(e.input))
	}
}

func Test_mapToStructConfigGroup(t *testing.T) {
	testMembers := &[]UserID{
		{Name: "user1", Realm: "pam"},
		{Name: "user2", Realm: "pam"},
		{Name: "user3", Realm: "pam"},
	}
	testData := []struct {
		input  map[string]interface{}
		output *ConfigGroup
	}{
		{
			input: map[string]interface{}{
				"comment": "test comment",
				"members": []interface{}{"user1@pam", "user2@pam", "user3@pam"},
			},
			output: &ConfigGroup{
				Comment: "test comment",
				Members: testMembers,
			},
		},
		{
			input: map[string]interface{}{
				"groupid": "testgroup",
				"members": []interface{}{"user1@pam", "user2@pam", "user3@pam"},
			},
			output: &ConfigGroup{
				Name:    "testgroup",
				Members: testMembers,
			},
		},
		{
			input: map[string]interface{}{
				"groupid": "testgroup",
				"comment": "test comment",
			},
			output: &ConfigGroup{
				Name:    "testgroup",
				Comment: "test comment",
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, ConfigGroup{}.mapToStruct(e.input))
	}
}
