package proxmox

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/mockServer"
	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_group"
	"github.com/stretchr/testify/require"
)

func Test_groupClient_AddMembers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		groups   []GroupName
		users    []UserID
		requests []mockServer.Request
		err      error
	}{
		{name: `Multiple User, Multiple Group`,
			groups: []GroupName{"group4", "group5"},
			users: []UserID{
				{Name: "test", Realm: "pve"},
				{Name: "root", Realm: "pam"}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/access/users/test@pve", map[string]any{"data": map[string]any{
					"groups": "group1,group2,group3",
					"userid": "test@pve"}}),
				mockServer.RequestsPutHandler("/access/users/test@pve", func(t *testing.T, v url.Values) {
					a := strings.Split(v.Get("groups"), ",")
					tmpMap := make(map[string]struct{})
					for i := range a {
						tmpMap[a[i]] = struct{}{}
					}
					require.Equal(t, map[string]struct{}{"group1": {}, "group2": {}, "group3": {}, "group4": {}, "group5": {}}, tmpMap)
				}),
				mockServer.RequestsGetJson("/access/users/root@pam", map[string]any{"data": map[string]any{
					"groups": "group3,group4,group6",
					"userid": "root@pam"}}),
				mockServer.RequestsPutHandler("/access/users/root@pam", func(t *testing.T, v url.Values) {
					a := strings.Split(v.Get("groups"), ",")
					tmpMap := make(map[string]struct{})
					for i := range a {
						tmpMap[a[i]] = struct{}{}
					}
					require.Equal(t, map[string]struct{}{"group3": {}, "group4": {}, "group5": {}, "group6": {}}, tmpMap)
				}))},
		{name: `handled error`,
			groups: []GroupName{"group4", "group5"},
			users: []UserID{
				{Name: "test", Realm: "pve"},
				{Name: "root", Realm: "pam"},
			},
			requests: mockServer.RequestsErrorHandled("/access/users/test@pve", mockServer.GET, mockServer.JsonError(404, map[string]any{
				"message": string("no such user ('test@pve')"),
			})),
			err: errors.New("user test@pve does not exist")},
		{name: `validate error empty groupname`,
			groups: []GroupName{""},
			err:    errors.New("variable of type (GroupName) may not be empty")},
		{name: `validate error empty username`,
			users: []UserID{{Realm: "pve"}},
			err:   errors.New("no username is specified")},
		{name: `API error`,
			groups:   []GroupName{"group4", "group5"},
			users:    []UserID{{Name: "test", Realm: "pve"}},
			requests: mockServer.RequestsError("/access/users/test@pve", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().Group.AddMembers(context.Background(), test.groups, test.users)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_groupClient_Create(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    ConfigGroup
		requests []mockServer.Request
		err      error
	}{
		{name: `members nil`,
			input: ConfigGroup{
				Comment: util.Pointer("Test Comment " + body.Symbols),
				Name:    "group1"},
			requests: mockServer.RequestsPost("/access/groups", map[string]any{
				"groupid": "group1",
				"comment": "Test Comment " + body.Symbols,
			})},
		{name: `members empty`,
			input: ConfigGroup{
				Name:    "group1",
				Members: &[]UserID{}},
			requests: mockServer.RequestsPost("/access/groups", map[string]any{
				"groupid": "group1"})},
		{name: `members single`,
			input: ConfigGroup{
				Name:    "group1",
				Members: &[]UserID{{Name: "user1", Realm: "pam"}}},
			requests: mockServer.Append(
				mockServer.RequestsPost("/access/groups", map[string]any{
					"groupid": "group1"}),
				mockServer.RequestsGetJson("/access/users/user1@pam", map[string]any{"data": map[string]any{
					"groups": "",
					"userid": "test@pve"}}),
				mockServer.RequestsPut("/access/users/user1@pam",
					map[string]any{"groups": "group1"}))},
		{name: `members single error`,
			input: ConfigGroup{
				Name:    "group1",
				Members: &[]UserID{{Name: "user1", Realm: "pam"}}},
			requests: mockServer.Append(
				mockServer.RequestsPost("/access/groups", map[string]any{
					"groupid": "group1"}),
				mockServer.RequestsError("/access/users/user1@pam", mockServer.GET, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `members multiple`,
			input: ConfigGroup{
				Name: "group1",
				Members: &[]UserID{
					{Name: "user1", Realm: "pam"},
					{Name: "user2", Realm: "pve"},
					{Name: "user3", Realm: "ldap"}}},
			requests: mockServer.Append(
				mockServer.RequestsPost("/access/groups", map[string]any{
					"groupid": "group1"}),

				mockServer.RequestsGetJson("/access/users/user1@pam", map[string]any{"data": map[string]any{
					"groups": "",
					"userid": "test1@pam"}}),
				mockServer.RequestsPut("/access/users/user1@pam",
					map[string]any{"groups": "group1"}),

				mockServer.RequestsGetJson("/access/users/user2@pve", map[string]any{"data": map[string]any{
					"groups": "group2,group7,group3",
					"userid": "test2@pve"}}),
				mockServer.RequestsPutHandler("/access/users/user2@pve", func(t *testing.T, v url.Values) {
					a := strings.Split(v.Get("groups"), ",")
					tmpMap := make(map[string]struct{})
					for i := range a {
						tmpMap[a[i]] = struct{}{}
					}
					require.Equal(t, map[string]struct{}{"group1": {}, "group2": {}, "group3": {}, "group7": {}}, tmpMap)
				}),

				mockServer.RequestsGetJson("/access/users/user3@ldap", map[string]any{"data": map[string]any{
					"groups": "group3",
					"userid": "test3@ldap"}}),
				mockServer.RequestsPutHandler("/access/users/user3@ldap", func(t *testing.T, v url.Values) {
					a := strings.Split(v.Get("groups"), ",")
					tmpMap := make(map[string]struct{})
					for i := range a {
						tmpMap[a[i]] = struct{}{}
					}
					require.Equal(t, map[string]struct{}{"group1": {}, "group3": {}}, tmpMap)
				}))},
		{name: `validate error empty groupname`,
			err: errors.New("variable of type (GroupName) may not be empty")},
		{name: `API error`,
			input:    ConfigGroup{Name: "group"},
			requests: mockServer.RequestsError("/access/groups", mockServer.POST, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().Group.Create(context.Background(), test.input)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_groupClient_Delete(t *testing.T) {
	t.Parallel()
	const path = "/access/groups/group1"
	tests := []struct {
		name     string
		input    GroupName
		exists   bool
		requests []mockServer.Request
		err      error
	}{
		{name: `exists`,
			input:    "group1",
			exists:   true,
			requests: mockServer.RequestsDelete(path, nil)},
		{name: `not exists`,
			input: "group1",
			requests: mockServer.RequestsErrorHandled(path, mockServer.DELETE,
				mockServer.JsonError(400, map[string]any{
					"message": string("delete group failed: group 'group1' does not exist"),
				}))},
		{name: `validate error empty groupname`,
			err: errors.New("variable of type (GroupName) may not be empty")},
		{name: `API error`,
			input:    "group1",
			requests: mockServer.RequestsError(path, mockServer.DELETE, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			exists, err := c.New().Group.Delete(context.Background(), test.input)
			require.Equal(t, test.err, err)
			require.Equal(t, test.exists, exists)
			server.Clear(t)
		})
	}
}

func Test_groupClient_Exists(t *testing.T) {
	t.Parallel()
	const path = "/access/groups/group1"
	tests := []struct {
		name     string
		input    GroupName
		exists   bool
		requests []mockServer.Request
		err      error
	}{
		{name: `exists`,
			input:  "group1",
			exists: true,
			requests: mockServer.RequestsGetJson(path, map[string]any{"data": map[string]any{
				"groupid": "group1",
			}})},
		{name: `not exists`,
			input: "group1",
			requests: mockServer.RequestsErrorHandled(path, mockServer.GET,
				mockServer.JsonError(400, map[string]any{
					"message": string("group 'group1' does not exist\n"),
				}))},
		{name: `validate error empty groupname`,
			err: errors.New("variable of type (GroupName) may not be empty")},
		{name: `API error`,
			input:    "group1",
			requests: mockServer.RequestsError(path, mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			exists, err := c.New().Group.Exists(context.Background(), test.input)
			require.Equal(t, test.err, err)
			require.Equal(t, test.exists, exists)
			server.Clear(t)
		})
	}
}

func Test_groupClient_List(t *testing.T) {
	t.Parallel()
	const path = "/access/groups"
	tests := []struct {
		name     string
		output   *rawGroups
		requests []mockServer.Request
		err      error
	}{
		{name: `list`,
			output: &rawGroups{a: []any{
				map[string]any{
					"groupid": "group1",
					"comment": "Test Comment",
				}}},
			requests: mockServer.RequestsGetJson(path, map[string]any{"data": []any{
				map[string]any{
					"groupid": "group1",
					"comment": "Test Comment",
				}}})},
		{name: `API error`,
			requests: mockServer.RequestsError(path, mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			raw, err := c.New().Group.List(context.Background())
			require.Equal(t, test.err, err)
			if test.output == nil {
				require.Nil(t, raw)
			} else {
				require.Equal(t, test.output, raw)
			}
			server.Clear(t)
		})
	}
}

func Test_groupClient_Read(t *testing.T) {
	t.Parallel()
	const path = "/access/groups/group1"
	tests := []struct {
		name     string
		input    GroupName
		output   *rawGroupConfig
		requests []mockServer.Request
		err      error
	}{
		{name: `exists`,
			input: "group1",
			output: &rawGroupConfig{
				group: util.Pointer(GroupName("group1")),
				a: map[string]any{
					"groupid": "group1",
					"comment": "Test Comment",
				}},
			requests: mockServer.RequestsGetJson(path, map[string]any{"data": map[string]any{
				"groupid": "group1",
				"comment": "Test Comment",
			}})},
		{name: `not exists`,
			input:  "group1",
			output: nil,
			requests: mockServer.RequestsErrorHandled(path, mockServer.GET, mockServer.JsonError(400,
				map[string]any{
					"message": string("group 'group1' does not exist\n"),
				})),
			err: errors.New("group does not exist")},
		{name: `validate error empty groupname`,
			err: errors.New("variable of type (GroupName) may not be empty")},
		{name: `API error`,
			input:    "group1",
			requests: mockServer.RequestsError(path, mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			raw, err := c.New().Group.Read(context.Background(), test.input)
			require.Equal(t, test.err, err)
			if test.output == nil {
				require.Nil(t, raw)
			} else {
				require.Equal(t, test.output, raw)
			}
			server.Clear(t)
		})
	}
}

func Test_groupClient_RemoveMembers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		groups   []GroupName
		users    []UserID
		requests []mockServer.Request
		err      error
	}{
		{name: `Multiple User, Multiple Group`,
			groups: []GroupName{"group2", "group3"},
			users: []UserID{
				{Name: "test", Realm: "pve"},
				{Name: "root", Realm: "pam"}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/access/users/test@pve", map[string]any{"data": map[string]any{
					"groups": "group1,group2,group3",
					"userid": "test@pve"}}),
				mockServer.RequestsPutHandler("/access/users/test@pve", func(t *testing.T, v url.Values) {
					a := strings.Split(v.Get("groups"), ",")
					tmpMap := make(map[string]struct{})
					for i := range a {
						tmpMap[a[i]] = struct{}{}
					}
					require.Equal(t, map[string]struct{}{"group1": {}}, tmpMap)
				}),
				mockServer.RequestsGetJson("/access/users/root@pam", map[string]any{"data": map[string]any{
					"groups": "group3,group4,group6",
					"userid": "root@pam"}}),
				mockServer.RequestsPutHandler("/access/users/root@pam", func(t *testing.T, v url.Values) {
					a := strings.Split(v.Get("groups"), ",")
					tmpMap := make(map[string]struct{})
					for i := range a {
						tmpMap[a[i]] = struct{}{}
					}
					require.Equal(t, map[string]struct{}{"group4": {}, "group6": {}}, tmpMap)
				}))},
		{name: `handled error`,
			groups: []GroupName{"group4", "group5"},
			users: []UserID{
				{Name: "test", Realm: "pve"},
				{Name: "root", Realm: "pam"},
			},
			requests: mockServer.RequestsErrorHandled("/access/users/test@pve", mockServer.GET, mockServer.JsonError(404, map[string]any{
				"message": string("no such user ('test@pve')"),
			})),
			err: errors.New("user test@pve does not exist")},
		{name: `validate error empty groupname`,
			groups: []GroupName{""},
			err:    errors.New("variable of type (GroupName) may not be empty")},
		{name: `validate error empty username`,
			users: []UserID{{Realm: "pve"}},
			err:   errors.New("no username is specified")},
		{name: `API error`,
			groups:   []GroupName{"group4", "group5"},
			users:    []UserID{{Name: "test", Realm: "pve"}},
			requests: mockServer.RequestsError("/access/users/test@pve", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().Group.RemoveMembers(context.Background(), test.groups, test.users)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_groupClient_Set(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    ConfigGroup
		requests []mockServer.Request
		err      error
	}{
		{name: `Create`,
			input: ConfigGroup{
				Name: "group1"},
			requests: mockServer.Append(
				mockServer.RequestsErrorHandled("/access/groups/group1", mockServer.GET, mockServer.JsonError(404, map[string]any{
					"message": string("group 'group1' does not exist\n"),
				})),
				mockServer.RequestsPost("/access/groups", map[string]any{
					"groupid": "group1"}))},
		{name: `Update`,
			input: ConfigGroup{
				Name:    "group1",
				Comment: util.Pointer("Test Comment " + body.Symbols)},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/access/groups/group1", map[string]any{"data": map[string]any{}}),
				mockServer.RequestsPut("/access/groups/group1", map[string]any{
					"comment": "Test Comment " + body.Symbols}))},
		{name: `validate error empty groupname`,
			err: errors.New("variable of type (GroupName) may not be empty")},
		{name: `API error`,
			input:    ConfigGroup{Name: "group1"},
			requests: mockServer.RequestsError("/access/groups/group1", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().Group.Set(context.Background(), test.input)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_groupClient_Update(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    ConfigGroup
		requests []mockServer.Request
		err      error
	}{
		{name: `members nil`,
			input: ConfigGroup{
				Comment: util.Pointer("Test Comment " + body.Symbols),
				Name:    "group1"},
			requests: mockServer.RequestsPut("/access/groups/group1", map[string]any{
				"comment": "Test Comment " + body.Symbols,
			})},
		{name: `members empty, comment same`,
			input: ConfigGroup{
				Name:    "group1",
				Comment: util.Pointer("Test Comment"),
				Members: &[]UserID{}},
			requests: mockServer.RequestsGetJson("/access/groups/group1", map[string]any{"data": map[string]any{
				"members": []any{},
				"comment": "Test Comment",
			}})},
		{name: `members empty error`,
			input: ConfigGroup{
				Name:    "group1",
				Members: &[]UserID{}},
			requests: mockServer.RequestsError("/access/groups/group1", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
		{name: `members empty error handled`,
			input: ConfigGroup{
				Name:    "group1",
				Members: &[]UserID{}},
			requests: mockServer.RequestsErrorHandled("/access/groups/group1", mockServer.GET, mockServer.JsonError(400, map[string]any{
				"message": string("group 'group1' does not exist\n"),
			})),
			err: errors.New("group does not exist")},
		{name: `members empty, comment different`,
			input: ConfigGroup{
				Name:    "group1",
				Comment: util.Pointer("Test Comment Different"),
				Members: &[]UserID{}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/access/groups/group1", map[string]any{"data": map[string]any{
					"members": []any{},
					"comment": "Test Comment",
				}}),
				mockServer.RequestsPut("/access/groups/group1", map[string]any{
					"comment": "Test Comment Different",
				}))},
		{name: `members empty to empty`,
			input: ConfigGroup{
				Name:    "group1",
				Members: &[]UserID{}},
			requests: mockServer.RequestsGetJson("/access/groups/group1", map[string]any{"data": map[string]any{
				"members": []any{},
				"comment": "",
			}})},
		{name: `members empty to set`,
			input: ConfigGroup{
				Name: "group1",
				Members: &[]UserID{
					{Name: "test", Realm: "pve"}}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/access/groups/group1", map[string]any{"data": map[string]any{
					"members": []any{},
					"comment": "",
				}}),

				mockServer.RequestsGetJson("/access/users/test@pve", map[string]any{"data": map[string]any{
					"groups": "group2,group3",
				}}),
				mockServer.RequestsPutHandler("/access/users/test@pve", func(t *testing.T, v url.Values) {
					a := strings.Split(v.Get("groups"), ",")
					tmpMap := make(map[string]struct{})
					for i := range a {
						tmpMap[a[i]] = struct{}{}
					}
					require.Equal(t, map[string]struct{}{"group1": {}, "group2": {}, "group3": {}}, tmpMap)
				}))},
		{name: `members empty to set error`,
			input: ConfigGroup{
				Name: "group1",
				Members: &[]UserID{
					{Name: "test", Realm: "pve"}}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/access/groups/group1", map[string]any{"data": map[string]any{
					"members": []any{},
					"comment": "",
				}}),

				mockServer.RequestsGetJson("/access/users/test@pve", map[string]any{"data": map[string]any{
					"groups": "group2,group3",
				}}),
				mockServer.RequestsError("/access/users/test@pve", mockServer.PUT, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `members set to empty`,
			input: ConfigGroup{
				Name:    "group1",
				Members: &[]UserID{}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/access/groups/group1", map[string]any{"data": map[string]any{
					"members": []any{string("test@pve")},
					"comment": "",
				}}),

				mockServer.RequestsGetJson("/access/users/test@pve", map[string]any{"data": map[string]any{
					"groups": "group1,group2,group3",
				}}),
				mockServer.RequestsPutHandler("/access/users/test@pve", func(t *testing.T, v url.Values) {
					a := strings.Split(v.Get("groups"), ",")
					tmpMap := make(map[string]struct{})
					for i := range a {
						tmpMap[a[i]] = struct{}{}
					}
					require.Equal(t, map[string]struct{}{"group2": {}, "group3": {}}, tmpMap)
				}))},
		{name: `members set to empty error`,
			input: ConfigGroup{
				Name:    "group1",
				Members: &[]UserID{}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/access/groups/group1", map[string]any{"data": map[string]any{
					"members": []any{string("test@pve")},
					"comment": "",
				}}),

				mockServer.RequestsGetJson("/access/users/test@pve", map[string]any{"data": map[string]any{
					"groups": "group1,group2,group3",
				}}),
				mockServer.RequestsError("/access/users/test@pve", mockServer.PUT, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `members set to set`,
			input: ConfigGroup{
				Name: "group1",
				Members: &[]UserID{
					{Name: "root", Realm: "pam"},
					{Name: "user1", Realm: "ldap"}}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/access/groups/group1", map[string]any{"data": map[string]any{
					"members": []any{string("test@pve"), string("root@pam"), string("user1@ldap")},
					"comment": "",
				}}),

				mockServer.RequestsGetJson("/access/users/test@pve", map[string]any{"data": map[string]any{
					"groups": "group1,group2,group3",
				}}),
				mockServer.RequestsPutHandler("/access/users/test@pve", func(t *testing.T, v url.Values) {
					a := strings.Split(v.Get("groups"), ",")
					tmpMap := make(map[string]struct{})
					for i := range a {
						tmpMap[a[i]] = struct{}{}
					}
					require.Equal(t, map[string]struct{}{"group2": {}, "group3": {}}, tmpMap)
				}))},
		{name: `validate error empty groupname`,
			err: errors.New("variable of type (GroupName) may not be empty")},
		{name: `API error`,
			input: ConfigGroup{Name: "group1",
				Comment: util.Pointer("Test")},
			requests: mockServer.RequestsError("/access/groups/group1", mockServer.PUT, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().Group.Update(context.Background(), test.input)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_ConfigGroup_mapToAPI(t *testing.T) {
	t.Parallel()
	type test struct {
		name    string
		input   ConfigGroup
		current *rawGroupConfig

		output map[string]string
	}
	tests := []struct {
		category     string
		create       []test
		createUpdate []test
		update       []test
	}{
		{category: `Name`,
			create: []test{
				{name: `set`,
					input:  ConfigGroup{Name: "group1"},
					output: map[string]string{"groupid": "group1"}}},
			update: []test{
				{name: `set`,
					input: ConfigGroup{Name: "group1"}}}},
		{category: `Comment`,
			create: []test{
				{name: `set`,
					input: ConfigGroup{Comment: util.Pointer("Test Comment " + body.Symbols)},
					output: map[string]string{
						"groupid": "",
						"comment": "Test Comment " + body.Symbols}},
				{name: `empty`,
					input: ConfigGroup{Comment: util.Pointer("")},
					output: map[string]string{
						"groupid": ""}}},
			update: []test{
				{name: `set`,
					input:  ConfigGroup{Comment: util.Pointer("Test Comment " + body.Symbols)},
					output: map[string]string{"comment": "Test Comment " + body.Symbols}},
				{name: `empty`,
					input:  ConfigGroup{Comment: util.Pointer("")},
					output: map[string]string{"comment": ""}},
				{name: `change`,
					input:   ConfigGroup{Comment: util.Pointer("test comment")},
					current: &rawGroupConfig{a: map[string]any{"comment": string("old comment")}},
					output:  map[string]string{"comment": "test comment"}},
				{name: `no change`,
					input:   ConfigGroup{Comment: util.Pointer("same comment")},
					current: &rawGroupConfig{a: map[string]any{"comment": string("same comment")}}}}},
	}
	for _, test := range tests {
		for _, subTest := range append(test.create, test.createUpdate...) {
			name := test.category + "/Create/" + subTest.name
			t.Run(name, func(*testing.T) {
				testParamsEqual(t, subTest.output, subTest.input.mapToApiCreate())
			})
		}
		for _, subTest := range append(test.update, test.createUpdate...) {
			name := test.category + "/Update/" + subTest.name
			t.Run(name, func(*testing.T) {
				testParamsEqual(t, subTest.output, subTest.input.mapToApiUpdate(subTest.current))
			})
		}
	}
}

func test_RawGroups_Array_Data() []struct {
	name   string
	input  rawGroups
	output []RawGroupConfig
} {
	return []struct {
		name   string
		input  rawGroups
		output []RawGroupConfig
	}{
		{name: `single`,
			input: rawGroups{
				a: []any{
					map[string]any{
						"comment": "Test Comment",
						"groupid": "group1"}}},
			output: []RawGroupConfig{
				&rawGroupConfig{a: map[string]any{
					"comment": "Test Comment",
					"groupid": "group1"}}}},
		{name: `multiple`,
			input: rawGroups{
				a: []any{
					map[string]any{
						"comment": "Test Comment",
						"groupid": "group1"},
					map[string]any{
						"comment": "",
						"members": []any{"user1@pam", "user2@pve"},
						"groupid": "group2"},
					map[string]any{
						"comment": "",
						"groupid": "group3"}}},
			output: []RawGroupConfig{
				&rawGroupConfig{a: map[string]any{
					"comment": "Test Comment",
					"groupid": "group1"}},
				&rawGroupConfig{a: map[string]any{
					"comment": "",
					"members": []any{"user1@pam", "user2@pve"},
					"groupid": "group2"}},
				&rawGroupConfig{a: map[string]any{
					"comment": "",
					"groupid": "group3"}}}},
		{name: `empty`,
			input:  rawGroups{a: []any{}},
			output: []RawGroupConfig{}},
	}
}

func Test_RawGroups_AsArray(t *testing.T) {
	t.Parallel()
	for _, test := range test_RawGroups_Array_Data() {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, (&test.input).AsArray())
		})
	}
}

func Test_RawGroups_Iter(t *testing.T) {
	t.Parallel()
	for _, test := range test_RawGroups_Array_Data() {
		t.Run(test.name, func(t *testing.T) {
			// Test iterating over all items
			var result []RawGroupConfig
			for group := range RawGroups(&test.input).Iter() {
				result = append(result, group)
			}
			require.Len(t, result, len(test.output))
			for i := range result {
				require.Equal(t, test.output[i].Get(), result[i].Get())
			}
			// Test early termination (break after first item)
			if len(test.output) > 0 {
				count := 0
				for range RawGroups(&test.input).Iter() {
					count++
					break
				}
				require.Equal(t, 1, count)
			}
		})
	}
}

func Test_RawGroups_AsMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawGroups
		output map[GroupName]RawGroupConfig
	}{
		{name: `single`,
			input: rawGroups{
				a: []any{
					map[string]any{
						"comment": "Test Comment",
						"groupid": "group1"}}},
			output: map[GroupName]RawGroupConfig{
				"group1": &rawGroupConfig{
					group: util.Pointer(GroupName("group1")),
					a: map[string]any{
						"comment": "Test Comment",
						"groupid": "group1",
					}}}},
		{name: `multiple`,
			input: rawGroups{
				a: []any{
					map[string]any{
						"comment": "Test Comment",
						"groupid": "group1"},
					map[string]any{
						"comment": "",
						"members": "user1@pam,user2@pve",
						"groupid": "group2"},
					map[string]any{
						"comment": "",
						"groupid": "group3"}}},
			output: map[GroupName]RawGroupConfig{
				"group1": &rawGroupConfig{
					group: util.Pointer(GroupName("group1")),
					a: map[string]any{
						"comment": "Test Comment",
						"groupid": "group1"}},
				"group2": &rawGroupConfig{
					group: util.Pointer(GroupName("group2")),
					a: map[string]any{
						"comment": "",
						"members": "user1@pam,user2@pve",
						"groupid": "group2"}},
				"group3": &rawGroupConfig{
					group: util.Pointer(GroupName("group3")),
					a: map[string]any{
						"comment": "",
						"groupid": "group3"}}}},
		{name: `empty`,
			input:  rawGroups{a: []any{}},
			output: map[GroupName]RawGroupConfig{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, (&test.input).AsMap())
		})
	}
}

func Test_RawGroups_Len(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawGroups
		output int
	}{
		{name: `single`,
			input: rawGroups{
				a: []any{
					map[string]any{
						"comment": "Test Comment",
						"groupid": "group1"}}},
			output: 1},
		{name: `multiple`,
			input: rawGroups{
				a: []any{
					map[string]any{
						"comment": "Test Comment",
						"groupid": "group1"},
					map[string]any{
						"comment": "",
						"members": "user1@pam,user2@pve",
						"groupid": "group2"},
					map[string]any{
						"comment": "",
						"groupid": "group3"}}},
			output: 3},
		{name: `empty`,
			input: rawGroups{a: []any{}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, (&test.input).Len())
		})
	}
}

func Test_RawGroupConfig_Get(t *testing.T) {
	t.Parallel()
	base := func(group ConfigGroup) ConfigGroup {
		if group.Comment == nil {
			group.Comment = util.Pointer("")
		}
		if group.Members == nil {
			group.Members = &[]UserID{}
		}
		return group
	}
	tests := []struct {
		name   string
		input  rawGroupConfig
		output ConfigGroup
	}{
		{name: `Name from body`,
			input:  rawGroupConfig{a: map[string]any{"groupid": string("test")}},
			output: base(ConfigGroup{Name: "test"})},
		{name: `Name from pointer`,
			input:  rawGroupConfig{group: util.Pointer(GroupName("test"))},
			output: base(ConfigGroup{Name: "test"})},
		{name: `Comment`,
			input:  rawGroupConfig{a: map[string]any{"comment": string("test comment")}},
			output: base(ConfigGroup{Comment: util.Pointer("test comment")})},
		{name: `Members empty`,
			input:  rawGroupConfig{a: map[string]any{"members": []any{}}},
			output: base(ConfigGroup{Members: &[]UserID{}})},
		{name: `Members set`,
			input: rawGroupConfig{a: map[string]any{"members": []any{string("user1@pam"), string("user2@pve"), string("user3@ldap")}}},
			output: base(ConfigGroup{Members: &[]UserID{
				{Name: "user1", Realm: "pam"},
				{Name: "user2", Realm: "pve"},
				{Name: "user3", Realm: "ldap"}}})},
		{name: `all fields`,
			input: rawGroupConfig{a: map[string]any{
				"groupid": "group1",
				"comment": "test comment",
				"members": []any{string("user1@pam"), string("user2@pve"), string("user3@ldap")}}},
			output: base(ConfigGroup{
				Name:    "group1",
				Comment: util.Pointer("test comment"),
				Members: &[]UserID{
					{Name: "user1", Realm: "pam"},
					{Name: "user2", Realm: "pve"},
					{Name: "user3", Realm: "ldap"}}})},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, (&test.input).Get())
		})
	}
}

// TODO improve when Name and Realm have their own types
func Test_ConfigGroup_Validate(t *testing.T) {
	t.Parallel()
	validGroupName := GroupName("groupName")
	False := 0
	TrueAndFalse := 1
	True := 2
	testData := []struct {
		input    ConfigGroup
		hasError bool
		err      error

		create int
	}{
		// GroupName
		{
			input:  ConfigGroup{},
			err:    errors.New(GroupName_Error_Empty),
			create: TrueAndFalse},
		{
			input:  ConfigGroup{Name: GroupName(test_data_group.GroupName_Max_Legal())},
			create: TrueAndFalse},
		{
			input:  ConfigGroup{Name: GroupName(test_data_group.GroupName_Max_Illegal())},
			err:    errors.New(GroupName_Error_MaxLength),
			create: TrueAndFalse},
		// GroupMembers
		{
			input: ConfigGroup{
				Name: validGroupName,
				Members: &[]UserID{
					{Name: "user1"},
				}},
			hasError: true,
			create:   TrueAndFalse,
		},
		{
			input: ConfigGroup{
				Name:    validGroupName,
				Members: &[]UserID{{Name: "user1", Realm: "pam"}}},
			create: TrueAndFalse,
		},
		{
			input: ConfigGroup{
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
			if e.err != nil {
				require.Equal(t, e.err, e.input.Validate())
			} else if e.hasError {
				require.Error(t, e.input.Validate())
			} else {
				require.NoError(t, e.input.Validate())
			}
		}
		if e.create > False {
			if e.err != nil {
				require.Equal(t, e.err, e.input.Validate())
			} else if e.hasError {
				require.Error(t, e.input.Validate())
			} else {
				require.NoError(t, e.input.Validate())
			}
		}
	}
}

func Test_GroupName_csvToArray(t *testing.T) {
	t.Parallel()
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

func Test_GroupName_mapToArray(t *testing.T) {
	t.Parallel()
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

func Test_GroupName_Validate(t *testing.T) {
	t.Parallel()
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
