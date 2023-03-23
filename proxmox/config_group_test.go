package proxmox

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/test/data/test_data_group"
	"github.com/stretchr/testify/require"
)

func Test_ConfigGroup_mapToApiValues(t *testing.T) {
	testData := []struct {
		input  ConfigGroup
		create bool
		output map[string]interface{}
	}{
		{
			input: ConfigGroup{
				Name:    "testGroup",
				Comment: "test comment",
				Members: &[]UserID{
					{Name: "userA", Realm: "pam"},
					{Name: "userB", Realm: "pam"},
				},
			},
			create: true,
			output: map[string]interface{}{
				"comment": "test comment",
				"groupid": "testGroup",
			},
		},
		{
			input: ConfigGroup{
				Name:    "testGroup",
				Comment: "test comment",
				Members: &[]UserID{
					{Name: "userA", Realm: "pam"},
					{Name: "userB", Realm: "pam"},
				},
			},
			create: false,
			output: map[string]interface{}{
				"comment": "test comment",
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.input.mapToApiValues(e.create))
	}
}

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
		require.Equal(t, e.output, e.base.mapToStruct(e.input))
	}
}

func Test_ConfigGroup_nilCheck(t *testing.T) {
	testData := []struct {
		input *ConfigGroup
		err   bool
	}{
		{input: &ConfigGroup{}},
		{err: true},
	}
	for _, e := range testData {
		if e.err {
			require.Error(t, e.input.nilCheck())
		} else {
			require.NoError(t, e.input.nilCheck())
		}
	}
}

// TODO improve when Name and Realm have their own types
func Test_ConfigGroup_Validate(t *testing.T) {
	validGroupName := GroupName("groupName")
	False := 0
	TrueAndFalse := 1
	True := 2
	testData := []struct {
		input  *ConfigGroup
		err    bool
		create int
	}{
		// GroupName
		{
			err:    true,
			create: TrueAndFalse,
		},
		{
			input:  &ConfigGroup{},
			err:    true,
			create: True,
		},
		{input: &ConfigGroup{}},
		{
			input:  &ConfigGroup{Name: GroupName(test_data_group.GroupName_Max_Legal())},
			create: True,
		},
		{
			input:  &ConfigGroup{Name: GroupName(test_data_group.GroupName_Max_Illegal())},
			err:    true,
			create: True,
		},
		// GroupMembers
		{
			input: &ConfigGroup{
				Name: validGroupName,
				Members: &[]UserID{
					{Name: "user1"},
				}},
			err:    true,
			create: TrueAndFalse,
		},
		{
			input: &ConfigGroup{
				Name:    validGroupName,
				Members: &[]UserID{{Name: "user1", Realm: "pam"}}},
			create: TrueAndFalse,
		},
		{
			input: &ConfigGroup{
				Name: validGroupName,
				Members: &[]UserID{
					{Name: "user1", Realm: "pam"},
					{Name: "user2", Realm: "pam"},
					{Name: "user3", Realm: "pam"},
				}},
			create: TrueAndFalse,
		},
	}
	for _, e := range testData {
		if e.create < True {
			if e.err {
				require.Error(t, e.input.Validate(false))
			} else {
				require.NoError(t, e.input.Validate(false))
			}
		}
		if e.create > False {
			if e.err {
				require.Error(t, e.input.Validate(true))
			} else {
				require.NoError(t, e.input.Validate(true))
			}
		}
	}
}

func Test_GroupName_arrayToCsv(t *testing.T) {
	testData := []struct {
		input  *[]GroupName
		output string
	}{
		{},
		{
			input:  &[]GroupName{},
			output: "",
		},
		{
			input:  &[]GroupName{"group1"},
			output: "group1",
		},
		{
			input:  &[]GroupName{"group1", "group2", "group3"},
			output: "group1,group2,group3",
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, GroupName("").arrayToCsv(e.input))
	}
}

func Test_GroupName_csvToArray(t *testing.T) {
	testData := []struct {
		input  string
		output []GroupName
	}{
		{output: []GroupName{}},
		{
			input:  "group1",
			output: []GroupName{"group1"},
		},
		{
			input:  "group1,group2,group3",
			output: []GroupName{"group1", "group2", "group3"},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, GroupName("").csvToArray(e.input))
	}
}

func Test_GroupName_inArray(t *testing.T) {
	testData := []struct {
		group  GroupName
		groups []GroupName
		output bool
	}{
		{},
		{group: "group1"},
		{groups: []GroupName{"group1"}},
		{
			group:  "group1",
			groups: []GroupName{"group1"},
			output: true,
		},
		{
			group:  "group1",
			groups: []GroupName{"group2"},
		},
		{
			group:  "group2",
			groups: []GroupName{"group1", "group2", "group3"},
			output: true,
		},
		{
			group:  "group4",
			groups: []GroupName{"group1", "group2", "group3"},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.group.inArray(e.groups))
	}
}

func Test_GroupName_mapToArray(t *testing.T) {
	testData := []struct {
		input  any
		output *[]GroupName
	}{
		{output: &[]GroupName{}},
		// []interface{} Type
		{
			input:  []interface{}{},
			output: &[]GroupName{},
		},
		{
			input:  []interface{}{""},
			output: &[]GroupName{},
		},
		{
			input:  []interface{}{"group1"},
			output: &[]GroupName{"group1"},
		},
		{
			input:  []interface{}{"group1", "group2", "group3"},
			output: &[]GroupName{"group1", "group2", "group3"},
		},
		// string Type
		{
			input:  "",
			output: &[]GroupName{},
		},
		{
			input:  "group1",
			output: &[]GroupName{"group1"},
		},
		{
			input:  "group1,group2,group3",
			output: &[]GroupName{"group1", "group2", "group3"},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, GroupName("").mapToArray(e.input))
	}
}

func Test_GroupName_removeAllUsersFromGroup(t *testing.T) {
	testData := []struct {
		group  GroupName
		users  []interface{}
		output *[]configUserShort
	}{
		// group empty
		{
			users:  test_data_group.UserMap(),
			output: &[]configUserShort{},
		},
		// users empty
		{
			group:  "group1",
			output: &[]configUserShort{},
		},
		// good result
		{
			group: "group1",
			users: test_data_group.UserMap(),
			output: &[]configUserShort{
				{User: UserID{Name: "user2", Realm: "pve"}, Groups: &[]GroupName{}},
				{User: UserID{Name: "user3", Realm: "pve"}, Groups: &[]GroupName{}},
				{User: UserID{Name: "user4", Realm: "pve"}, Groups: &[]GroupName{"group2"}},
				{User: UserID{Name: "user5", Realm: "pve"}, Groups: &[]GroupName{"group2", "group3"}},
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.group.removeAllUsersFromGroup(e.users))
	}
}

func Test_GroupName_removeAllUsersFromGroupExcept(t *testing.T) {
	outputNoMembers := &[]configUserShort{
		{
			User:   UserID{Name: "user2", Realm: "pve"},
			Groups: &[]GroupName{},
		},
		{
			User:   UserID{Name: "user3", Realm: "pve"},
			Groups: &[]GroupName{},
		},
		{
			User:   UserID{Name: "user4", Realm: "pve"},
			Groups: &[]GroupName{"group2"},
		},
		{
			User:   UserID{Name: "user5", Realm: "pve"},
			Groups: &[]GroupName{"group2", "group3"},
		},
	}
	testData := []struct {
		group   GroupName
		members *[]UserID
		users   []interface{}
		output  *[]configUserShort
	}{
		// group empty
		{
			members: &[]UserID{{Name: "user1", Realm: "pve"}},
			users:   test_data_group.UserMap(),
		},
		// members nil
		{
			group:  "group1",
			users:  test_data_group.UserMap(),
			output: outputNoMembers,
		},
		// members empty
		{
			group:   "group1",
			members: &[]UserID{},
			users:   test_data_group.UserMap(),
			output:  outputNoMembers,
		},
		// users empty
		{
			group:   "group1",
			members: &[]UserID{{Name: "user1", Realm: "pve"}},
			output:  &[]configUserShort{},
		},
		// good result
		{
			group: "group1",
			members: &[]UserID{
				{Name: "user3", Realm: "pve"},
				{Name: "user5", Realm: "pve"},
			},
			users: test_data_group.UserMap(),
			output: &[]configUserShort{
				{
					User:   UserID{Name: "user2", Realm: "pve"},
					Groups: &[]GroupName{},
				},
				{
					User:   UserID{Name: "user4", Realm: "pve"},
					Groups: &[]GroupName{"group2"},
				},
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.group.removeAllUsersFromGroupExcept(e.users, e.members))
	}
}

func Test_GroupName_removeFromArray(t *testing.T) {
	testData := []struct {
		group  GroupName
		groups []GroupName
		output []GroupName
	}{
		// not fully populated
		{output: []GroupName{}},
		{
			group:  "group1",
			output: []GroupName{},
		},
		{
			groups: []GroupName{"group1"},
			output: []GroupName{"group1"},
		},
		// includes
		{
			group:  "group1",
			groups: []GroupName{"group1"},
			output: []GroupName{},
		},
		{
			group:  "group1",
			groups: []GroupName{"group1", "group2", "group3"},
			output: []GroupName{"group2", "group3"},
		},
		// does not include
		{
			group:  "group1",
			groups: []GroupName{"group0"},
			output: []GroupName{"group0"},
		},
		{
			group:  "group1",
			groups: []GroupName{"group0", "group2", "group3"},
			output: []GroupName{"group0", "group2", "group3"},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.group.removeFromArray(e.groups))
	}
}

func Test_GroupName_usersToAddToGroup(t *testing.T) {
	testData := []struct {
		group   GroupName
		members *[]UserID
		users   []interface{}
		output  *[]configUserShort
	}{
		// group empty
		{
			members: &[]UserID{{Name: "user1", Realm: "pve"}},
			users:   test_data_group.UserMap(),
		},
		// members nil
		{
			group: "group1",
			users: test_data_group.UserMap(),
		},
		// members empty
		{
			group:   "group1",
			members: &[]UserID{},
			users:   test_data_group.UserMap(),
			output:  &[]configUserShort{},
		},
		// users empty
		{
			group:   "group1",
			members: &[]UserID{{Name: "user1", Realm: "pve"}},
			output:  &[]configUserShort{},
		},
		// good result
		{
			group: "group1",
			members: &[]UserID{
				{Name: "user1", Realm: "pve"},
				{Name: "user2", Realm: "pve"},
				{Name: "user4", Realm: "pve"},
				{Name: "user5", Realm: "pve"},
				{Name: "user6", Realm: "pve"},
				{Name: "user7", Realm: "pve"},
			},
			users: test_data_group.UserMap(),
			output: &[]configUserShort{
				{
					User:   UserID{Name: "user1", Realm: "pve"},
					Groups: &[]GroupName{"group1"},
				},
				{
					User:   UserID{Name: "user6", Realm: "pve"},
					Groups: &[]GroupName{"group2", "group3", "group1"},
				},
				{
					User:   UserID{Name: "user7", Realm: "pve"},
					Groups: &[]GroupName{"group1"},
				},
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.group.usersToAddToGroup(e.users, e.members))
	}
}

func Test_GroupName_usersToRemoveFromGroup(t *testing.T) {
	testData := []struct {
		group   GroupName
		members *[]UserID
		users   []interface{}
		output  *[]configUserShort
	}{
		// group empty
		{
			members: &[]UserID{{Name: "user1", Realm: "pve"}},
			users:   test_data_group.UserMap(),
		},
		// members nil
		{
			group: "group1",
			users: test_data_group.UserMap(),
		},
		// members empty
		{
			group:   "group1",
			members: &[]UserID{},
			users:   test_data_group.UserMap(),
			output:  &[]configUserShort{},
		},
		// users empty
		{
			group:   "group1",
			members: &[]UserID{{Name: "user1", Realm: "pve"}},
			output:  &[]configUserShort{},
		},
		// good result
		{
			group: "group1",
			members: &[]UserID{
				{Name: "user1", Realm: "pve"},
				{Name: "user2", Realm: "pve"},
				{Name: "user5", Realm: "pve"},
				{Name: "user6", Realm: "pve"},
			},
			users: test_data_group.UserMap(),
			output: &[]configUserShort{
				{
					User:   UserID{Name: "user2", Realm: "pve"},
					Groups: &[]GroupName{},
				},
				{
					User:   UserID{Name: "user5", Realm: "pve"},
					Groups: &[]GroupName{"group2", "group3"},
				},
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.group.usersToRemoveFromGroup(e.users, e.members))
	}
}

func Test_GroupName_Validate(t *testing.T) {
	testRunes := struct {
		legal   []string
		illegal []string
	}{
		legal:   test_data_group.GroupName_Legal(),
		illegal: test_data_group.GroupName_Illegal(),
	}
	for _, e := range testRunes.legal {
		require.NoError(t, GroupName(e).Validate())
	}
	for _, e := range testRunes.illegal {
		require.Error(t, GroupName(e).Validate())
	}
}
