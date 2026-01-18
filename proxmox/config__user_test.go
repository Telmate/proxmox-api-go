package proxmox

import (
	"context"
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/mockServer"
	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_group"
	"github.com/stretchr/testify/require"
)

func Test_userClient_Create(t *testing.T) {
	const path = "/access/users"
	tests := []struct {
		name     string
		input    ConfigUser
		userid   UserID
		requests []mockServer.Request
		err      error
	}{
		{name: `Create without password`,
			input: ConfigUser{
				Comment:   util.Pointer("test"),
				Email:     util.Pointer("user@example.com"),
				Expire:    util.Pointer(uint(12345678)),
				FirstName: util.Pointer("acme"),
				Groups:    util.Pointer([]GroupName{"group1", "group2", "group3"}),
				LastName:  util.Pointer("corp"),
				User:      UserID{Name: "test", Realm: "pve"}},
			requests: mockServer.RequestsPost(path, map[string]any{
				"comment":   "test",
				"email":     "user@example.com",
				"expire":    "12345678",
				"firstname": "acme",
				"groups":    "group1,group2,group3",
				"lastname":  "corp",
				"userid":    "test@pve"})},
		{name: `Create with password`,
			input: ConfigUser{
				Comment:   util.Pointer("test"),
				Email:     util.Pointer("user@example.com"),
				Expire:    util.Pointer(uint(12345678)),
				FirstName: util.Pointer("acme"),
				Groups:    util.Pointer([]GroupName{"group1", "group2", "group3"}),
				LastName:  util.Pointer("corp"),
				User:      UserID{Name: "test", Realm: "pve"},
				Password:  util.Pointer(UserPassword("secret123!"))},
			requests: mockServer.Append(
				mockServer.RequestsPost(path, map[string]any{
					"comment":   "test",
					"email":     "user@example.com",
					"expire":    "12345678",
					"firstname": "acme",
					"groups":    "group1,group2,group3",
					"lastname":  "corp",
					"userid":    "test@pve"}),
				mockServer.RequestsPut("/access/password", map[string]any{
					"userid":   "test@pve",
					"password": "secret123!",
				}))},
		{name: `validate error empty username`,
			input: ConfigUser{User: UserID{Realm: "pve"}},
			err:   errors.New("no username is specified")},
		{name: `user already exists`,
			input: ConfigUser{User: UserID{Name: "test", Realm: "pve"}},
			requests: mockServer.RequestsErrorHandled(path, mockServer.POST, mockServer.JsonError(500, map[string]any{
				"message": string("no such user ('test@pve')")})),
			err: errors.New(`error creating User: api error: code: 500 message: no such user ('test@pve')`)},
		{name: `API error create`,
			input:    ConfigUser{User: UserID{Name: "test", Realm: "pve"}},
			requests: mockServer.RequestsError(path, mockServer.POST, 500, 3),
			err:      errors.New("error creating User: " + mockServer.InternalServerError)},
		{name: `API error password`,
			input: ConfigUser{
				User:     UserID{Name: "test", Realm: "pve"},
				Password: util.Pointer(UserPassword("Enter123!"))},
			requests: mockServer.Append(
				mockServer.RequestsPost(path, map[string]any{"userid": "test@pve"}),
				mockServer.RequestsError("/access/password", mockServer.PUT, 500, 3)),
			err: errors.New("error setting password: " + mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().User.Create(context.Background(), test.input)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_userClient_Delete(t *testing.T) {
	const path = "/access/users/test@pve"
	tests := []struct {
		name     string
		input    UserID
		requests []mockServer.Request
		err      error
	}{
		{name: `Delete`,
			input:    UserID{Name: "test", Realm: "pve"},
			requests: mockServer.RequestsDelete(path, map[string]any{})},
		{name: `validate error empty username`,
			input: UserID{Realm: "pve"},
			err:   errors.New("no username is specified")},
		{name: `API error`,
			input:    UserID{Name: "test", Realm: "pve"},
			requests: mockServer.RequestsError(path, mockServer.DELETE, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().User.Delete(context.Background(), test.input)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_userClient_Exists(t *testing.T) {
	const path = "/access/users/test@pve"
	tests := []struct {
		name     string
		input    UserID
		exists   bool
		requests []mockServer.Request
		err      error
	}{
		{name: `Exists true`,
			exists: true,
			input:  UserID{Name: "test", Realm: "pve"},
			requests: mockServer.RequestsGetJson(path, map[string]any{
				"data": map[string]any{"userid": "test@pve"}})},
		{name: `Exists false`,
			input: UserID{Name: "test", Realm: "pve"},
			requests: mockServer.RequestsErrorHandled(path, mockServer.GET, mockServer.HTTPerror{
				Message: `{"message":"no such user ('test@pve')\n"}`,
				Code:    500})},
		{name: `validate error empty username`,
			input: UserID{Realm: "pve"},
			err:   errors.New("no username is specified")},
		{name: `API error`,
			input:    UserID{Name: "test", Realm: "pve"},
			requests: mockServer.RequestsError(path, mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			exists, err := c.New().User.Exists(context.Background(), test.input)
			require.Equal(t, test.err, err)
			require.Equal(t, test.exists, exists)
			server.Clear(t)
		})
	}
}

func Test_userClient_List(t *testing.T) {
	const path = "/access/users?full=1"
	baseUser := func(user ConfigUser) ConfigUser {
		if user.Comment == nil {
			user.Comment = util.Pointer("")
		}
		if user.Email == nil {
			user.Email = util.Pointer("")
		}
		if user.Enable == nil {
			user.Enable = util.Pointer(true)
		}
		if user.Expire == nil {
			user.Expire = util.Pointer(uint(0))
		}
		if user.FirstName == nil {
			user.FirstName = util.Pointer("")
		}
		if user.LastName == nil {
			user.LastName = util.Pointer("")
		}
		return user
	}
	baseToken := func(token ApiTokenConfig) ApiTokenConfig {
		if token.Comment == nil {
			token.Comment = util.Pointer("")
		}
		if token.Expiration == nil {
			token.Expiration = util.Pointer(uint(0))
		}
		if token.PrivilegeSeparation == nil {
			token.PrivilegeSeparation = util.Pointer(false)
		}
		return token
	}

	tests := []struct {
		name     string
		output   map[UserID]UserInfo
		requests []mockServer.Request
		err      error
	}{
		{name: `List full`,
			output: map[UserID]UserInfo{
				{Name: "user1", Realm: "pve"}: {
					Config: baseUser(ConfigUser{
						User:    UserID{Name: "user1", Realm: "pve"},
						Comment: util.Pointer("First user"),
						Groups:  util.Pointer([]GroupName{"group1", "group2", "group3"})}),
					Tokens: &[]ApiTokenConfig{}},
				{Name: "user2", Realm: "pam"}: {
					Config: baseUser(ConfigUser{
						User:   UserID{Name: "user2", Realm: "pam"},
						Enable: util.Pointer(false),
						Expire: util.Pointer(uint(1625097600)),
						Groups: util.Pointer([]GroupName{"group2"})}),
					Tokens: &[]ApiTokenConfig{baseToken(ApiTokenConfig{Name: "token1"})}},
				{Name: "root", Realm: "pam"}: {
					Config: baseUser(ConfigUser{
						User:   UserID{Name: "root", Realm: "pam"},
						Groups: util.Pointer([]GroupName{})}),
					Tokens: &[]ApiTokenConfig{
						baseToken(ApiTokenConfig{Name: "token1",
							Comment: util.Pointer("test comment")}),
						baseToken(ApiTokenConfig{Name: "token2",
							Expiration: util.Pointer(uint(123456789))}),
						baseToken(ApiTokenConfig{Name: "token3",
							PrivilegeSeparation: util.Pointer(false)})}}},
			requests: mockServer.RequestsGetJson(path, map[string]any{
				"data": []map[string]any{
					{"userid": "user1@pve",
						"comment":    "First user",
						"email":      "",
						"enable":     float64(1),
						"expire":     float64(0),
						"groups":     []any{"group1", "group2", "group3"},
						"realm-type": "pve"},
					{"userid": "user2@pam",
						"enable":     float64(0),
						"expire":     float64(1625097600),
						"groups":     []any{"group2"},
						"realm-type": "pam",
						"tokens": []map[string]any{
							{"tokenid": "token1"},
						}},
					{"userid": "root@pam",
						"enable":     float64(1),
						"expire":     float64(0),
						"realm-type": "pam",
						"tokens": []map[string]any{
							{"tokenid": "token1",
								"comment": "test comment"},
							{"tokenid": "token2",
								"expire": float64(123456789)},
							{"tokenid": "token3",
								"privsep": float64(0)},
						}}}})},
		{name: `API error`,
			requests: mockServer.RequestsError(path, mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			raw, err := c.New().User.List(context.Background())
			require.Equal(t, test.err, err)
			if err == nil {
				testCompareRawMap(t, test.output, raw.AsMap())
			}
			server.Clear(t)
		})
	}
}

func Test_userClient_ListPartial(t *testing.T) {
	const path = "/access/users"
	base := func(info UserInfo) UserInfo {
		if info.Config.Comment == nil {
			info.Config.Comment = util.Pointer("")
		}
		if info.Config.Email == nil {
			info.Config.Email = util.Pointer("")
		}
		if info.Config.Enable == nil {
			info.Config.Enable = util.Pointer(true)
		}
		if info.Config.Expire == nil {
			info.Config.Expire = util.Pointer(uint(0))
		}
		if info.Config.FirstName == nil {
			info.Config.FirstName = util.Pointer("")
		}
		if info.Config.LastName == nil {
			info.Config.LastName = util.Pointer("")
		}
		return info
	}

	tests := []struct {
		name     string
		output   map[UserID]UserInfo
		requests []mockServer.Request
		err      error
	}{
		{name: `List partial`,
			output: map[UserID]UserInfo{
				{Name: "user1", Realm: "pve"}: base(UserInfo{Config: ConfigUser{
					User:    UserID{Name: "user1", Realm: "pve"},
					Comment: util.Pointer("First user"),
					Enable:  util.Pointer(true)}}),
				{Name: "user2", Realm: "pam"}: base(UserInfo{Config: ConfigUser{
					User:   UserID{Name: "user2", Realm: "pam"},
					Enable: util.Pointer(false),
					Expire: util.Pointer(uint(1625097600))}}),
				{Name: "root", Realm: "pam"}: base(UserInfo{Config: ConfigUser{
					User: UserID{Name: "root", Realm: "pam"}}})},
			requests: mockServer.RequestsGetJson(path, map[string]any{
				"data": []map[string]any{
					{"userid": "user1@pve",
						"comment":    "First user",
						"email":      "",
						"enable":     float64(1),
						"expire":     float64(0),
						"realm-type": "pve"},
					{"userid": "user2@pam",
						"enable":     float64(0),
						"expire":     float64(1625097600),
						"realm-type": "pam"},
					{"userid": "root@pam",
						"enable":     float64(1),
						"expire":     float64(0),
						"realm-type": "pam"}}})},
		{name: `API error`,
			requests: mockServer.RequestsError(path, mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			raw, err := c.New().User.ListPartial(context.Background())
			require.Equal(t, test.err, err)
			if err == nil {
				testCompareRawMap(t, test.output, raw.AsMap())
			}
			server.Clear(t)
		})
	}
}

func Test_userClient_Read(t *testing.T) {
	const path = "/access/users/test@pve"
	tests := []struct {
		name     string
		userID   UserID
		requests []mockServer.Request
		output   *ConfigUser
		err      error
	}{
		{name: `Get existing user`,
			userID: UserID{Name: "test", Realm: "pve"},
			requests: mockServer.RequestsGetJson(path, map[string]any{"data": map[string]any{
				"comment": "Test User",
				"enable":  1,
				"userid":  "test@pve"}}),
			output: &ConfigUser{
				Comment:   util.Pointer("Test User"),
				Email:     util.Pointer(""),
				Enable:    util.Pointer(true),
				Expire:    util.Pointer(uint(0)),
				FirstName: util.Pointer(""),
				Groups:    &[]GroupName{},
				LastName:  util.Pointer(""),
				User:      UserID{Name: "test", Realm: "pve"}}},
		{name: `User does not exist`,
			userID: UserID{Name: "test", Realm: "pve"},
			requests: mockServer.RequestsErrorHandled(path, mockServer.GET, mockServer.HTTPerror{
				Message: `{"message":"no such user ('test@pve')\n"}`,
				Code:    500}),
			err: errors.New(`user test@pve does not exist`)},
		{name: `API error`,
			userID:   UserID{Name: "test", Realm: "pve"},
			requests: mockServer.RequestsError(path, mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
		{name: `Validation error`,
			userID: UserID{Name: "", Realm: "pve"},
			err:    errors.New("no username is specified")},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			raw, err := c.New().User.Read(context.Background(), test.userID)
			require.Equal(t, test.err, err)
			if err == nil {
				require.Equal(t, test.output, raw.Get())
			}
			server.Clear(t)
		})
	}
}

func Test_userClient_Set(t *testing.T) {
	const path = "/access/users"
	noSuchUser := func(user string) []mockServer.Request {
		return mockServer.RequestsErrorHandled(mockServer.Path(path+"/"+user), mockServer.GET, mockServer.HTTPerror{
			Message: `{"message":"no such user ('` + user + `')\n"}`,
			Code:    500})
	}
	tests := []struct {
		name     string
		input    ConfigUser
		requests []mockServer.Request
		err      error
	}{
		{name: `Create without password`,
			input: ConfigUser{
				Comment: util.Pointer("a comment"),
				Email:   util.Pointer("test@example.com"),
				User:    UserID{Name: "test", Realm: "pve"}},
			requests: mockServer.Append(
				noSuchUser("test@pve"),
				mockServer.RequestsPost(path, map[string]any{
					"comment": "a comment",
					"email":   "test@example.com",
					"userid":  "test@pve"}))},
		{name: `Create with password`,
			input: ConfigUser{
				Comment:  util.Pointer("a comment"),
				Email:    util.Pointer("test@example.com"),
				Password: util.Pointer(UserPassword("Enter123!")),
				User:     UserID{Name: "test", Realm: "pve"}},
			requests: mockServer.Append(
				noSuchUser("test@pve"),
				mockServer.RequestsPost(path, map[string]any{
					"comment": "a comment",
					"email":   "test@example.com",
					"userid":  "test@pve"}),
				mockServer.RequestsPut("/access/password", map[string]any{
					"userid":   "test@pve",
					"password": "Enter123!"}))},
		{name: `Update without password`,
			input: ConfigUser{
				Comment: util.Pointer("a comment"),
				Email:   util.Pointer("test@example.com"),
				User:    UserID{Name: "test", Realm: "pve"}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson(path+"/test@pve", map[string]any{"data": map[string]any{
					"userid": "test@pve"}}),
				mockServer.RequestsPut(path+"/test@pve", map[string]any{
					"comment": "a comment",
					"email":   "test@example.com"}))},
		{name: `Update with password`,
			input: ConfigUser{
				Comment:  util.Pointer("a comment"),
				Email:    util.Pointer("test@example.com"),
				Password: util.Pointer(UserPassword("Enter123!")),
				User:     UserID{Name: "test", Realm: "pve"}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson(path+"/test@pve", map[string]any{"data": map[string]any{
					"userid": "test@pve"}}),
				mockServer.RequestsPut(path+"/test@pve", map[string]any{
					"comment": "a comment",
					"email":   "test@example.com"}),
				mockServer.RequestsPut("/access/password", map[string]any{
					"userid":   "test@pve",
					"password": "Enter123!"}))},
		{name: `Update password`,
			input: ConfigUser{
				Password: util.Pointer(UserPassword("Enter123!")),
				User:     UserID{Name: "test", Realm: "pve"}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson(path+"/test@pve", map[string]any{"data": map[string]any{
					"userid": "test@pve"}}),
				mockServer.RequestsPut("/access/password", map[string]any{
					"userid":   "test@pve",
					"password": "Enter123!"}))},
		{name: `Validation error`,
			input: ConfigUser{
				User: UserID{Name: "", Realm: "pve"}},
			err: errors.New("no username is specified")},
		{name: `API error exists`,
			input: ConfigUser{
				User: UserID{Name: "test", Realm: "pve"}},
			requests: mockServer.RequestsError(path+"/test@pve", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().User.Set(context.Background(), test.input)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_userClient_Update(t *testing.T) {
	const path = "/access/users/test@pve"
	tests := []struct {
		name     string
		input    ConfigUser
		userID   UserID
		requests []mockServer.Request
		err      error
	}{
		{name: `Update without password`,
			input: ConfigUser{
				Comment:   util.Pointer("test"),
				Email:     util.Pointer("user@example.com"),
				Expire:    util.Pointer(uint(12345678)),
				FirstName: util.Pointer("acme"),
				Groups:    util.Pointer([]GroupName{"group1", "group2", "group3"}),
				LastName:  util.Pointer("corp"),
				User:      UserID{Name: "test", Realm: "pve"}},
			requests: mockServer.RequestsPut(path, map[string]any{
				"comment":   "test",
				"email":     "user@example.com",
				"expire":    "12345678",
				"firstname": "acme",
				"groups":    "group1,group2,group3",
				"lastname":  "corp"})},
		{name: `Update with password`,
			input: ConfigUser{
				Comment:   util.Pointer("test"),
				Email:     util.Pointer("user@example.com"),
				Expire:    util.Pointer(uint(12345678)),
				FirstName: util.Pointer("acme"),
				Groups:    util.Pointer([]GroupName{"group1", "group2", "group3"}),
				LastName:  util.Pointer("corp"),
				User:      UserID{Name: "test", Realm: "pve"},
				Password:  util.Pointer(UserPassword("secret123!"))},
			requests: mockServer.Append(
				mockServer.RequestsPut(path, map[string]any{
					"comment":   "test",
					"email":     "user@example.com",
					"expire":    "12345678",
					"firstname": "acme",
					"groups":    "group1,group2,group3",
					"lastname":  "corp"}),
				mockServer.RequestsPut("/access/password", map[string]any{
					"userid":   "test@pve",
					"password": "secret123!",
				}))},
		{name: `validate error empty username`,
			input: ConfigUser{User: UserID{Realm: "pve"}},
			err:   errors.New("no username is specified")},
		{name: `user doesn't exists`,
			input: ConfigUser{
				Comment: util.Pointer("test comment"),
				User:    UserID{Name: "test", Realm: "pve"}},
			requests: mockServer.RequestsErrorHandled(path, mockServer.PUT, mockServer.HTTPerror{
				Message: `{"message":"update user failed: no such user ('test@pve')\n"}`,
				Code:    500}),
			err: &ApiError{
				Message: "update user failed: no such user ('test@pve')",
				Code:    "500"}},
		{name: `API error Update`,
			input: ConfigUser{
				Comment: util.Pointer("test comment"),
				User:    UserID{Name: "test", Realm: "pve"}},
			requests: mockServer.RequestsError(path, mockServer.PUT, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
		{name: `API error password`,
			input: ConfigUser{
				User:     UserID{Name: "test", Realm: "pve"},
				Password: util.Pointer(UserPassword("Enter123!@#"))},
			requests: mockServer.Append(
				mockServer.RequestsError("/access/password", mockServer.PUT, 500, 3)),
			err: errors.New("error setting password: " + mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().User.Update(context.Background(), test.input)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_ConfigUser_mapToAPI(t *testing.T) {
	type test struct {
		name   string
		input  ConfigUser
		output map[string]string
	}
	tests := []struct {
		category     string
		create       []test
		createUpdate []test
		update       []test
	}{
		{category: `Comment`,
			create: []test{
				{name: `empty`,
					input:  ConfigUser{Comment: util.Pointer("")},
					output: map[string]string{"userid": ""}},
				{name: `set`,
					input: ConfigUser{Comment: util.Pointer("Test comment " + body.Symbols)},
					output: map[string]string{
						"comment": "Test comment " + body.Symbols,
						"userid":  ""}}},
			update: []test{
				{name: `empty`,
					input:  ConfigUser{Comment: util.Pointer("")},
					output: map[string]string{"comment": ""}},
				{name: `set`,
					input:  ConfigUser{Comment: util.Pointer("Test comment " + body.Symbols)},
					output: map[string]string{"comment": "Test comment " + body.Symbols}}}},
		{category: `Email`,
			create: []test{
				{name: `empty`,
					input:  ConfigUser{Email: util.Pointer("")},
					output: map[string]string{"userid": ""}},
				{name: `set`,
					input: ConfigUser{Email: util.Pointer("tony@stark-industries.com")},
					output: map[string]string{
						"email":  "tony@stark-industries.com",
						"userid": ""}}},
			update: []test{
				{name: `empty`,
					input:  ConfigUser{Email: util.Pointer("")},
					output: map[string]string{"email": ""}},
				{name: `set`,
					input:  ConfigUser{Email: util.Pointer("tony@stark-industries.com")},
					output: map[string]string{"email": "tony@stark-industries.com"}}}},
		{category: `Enable`,
			create: []test{
				{name: `true`,
					input: ConfigUser{Enable: util.Pointer(true)},
					output: map[string]string{
						"userid": ""}},
				{name: `false`,
					input: ConfigUser{Enable: util.Pointer(false)},
					output: map[string]string{
						"enable": "0",
						"userid": ""}}},
			update: []test{
				{name: `true`,
					input:  ConfigUser{Enable: util.Pointer(true)},
					output: map[string]string{"enable": "1"}},
				{name: `false`,
					input:  ConfigUser{Enable: util.Pointer(false)},
					output: map[string]string{"enable": "0"}}}},
		{category: `Expire`,
			create: []test{
				{name: `zero`,
					input:  ConfigUser{Expire: util.Pointer(uint(0))},
					output: map[string]string{"userid": ""}},
				{name: `set`,
					input: ConfigUser{Expire: util.Pointer(uint(784873474))},
					output: map[string]string{
						"expire": "784873474",
						"userid": ""}}},
			update: []test{
				{name: `zero`,
					input:  ConfigUser{Expire: util.Pointer(uint(0))},
					output: map[string]string{"expire": "0"}},
				{name: `set`,
					input:  ConfigUser{Expire: util.Pointer(uint(784873474))},
					output: map[string]string{"expire": "784873474"}}}},
		{category: `FirstName`,
			create: []test{
				{name: `empty`,
					input:  ConfigUser{FirstName: util.Pointer("")},
					output: map[string]string{"userid": ""}},
				{name: `set`,
					input: ConfigUser{FirstName: util.Pointer("Tony")},
					output: map[string]string{
						"firstname": "Tony",
						"userid":    ""}}},
			update: []test{
				{name: `empty`,
					input:  ConfigUser{FirstName: util.Pointer("")},
					output: map[string]string{"firstname": ""}},
				{name: `set`,
					input:  ConfigUser{FirstName: util.Pointer("Tony")},
					output: map[string]string{"firstname": "Tony"}}}},
		{category: `Groups`,
			create: []test{
				{name: `empty`,
					input:  ConfigUser{Groups: &[]GroupName{}},
					output: map[string]string{"userid": ""}},
				{name: `set`,
					input: ConfigUser{Groups: &[]GroupName{"admin", "user"}},
					output: map[string]string{
						"groups": "admin,user",
						"userid": ""}}},
			update: []test{
				{name: `empty`,
					input:  ConfigUser{Groups: &[]GroupName{}},
					output: map[string]string{"groups": ""}},
				{name: `set`,
					input:  ConfigUser{Groups: &[]GroupName{"admin", "user"}},
					output: map[string]string{"groups": "admin,user"}}}},
		{category: `Keys`,
			create: []test{
				{name: `empty`,
					input:  ConfigUser{Keys: util.Pointer("")},
					output: map[string]string{"userid": ""}},
				{name: `set`,
					input: ConfigUser{Keys: util.Pointer("aaaa")},
					output: map[string]string{
						"keys":   "aaaa",
						"userid": ""}}},
			update: []test{
				{name: `empty`,
					input:  ConfigUser{Keys: util.Pointer("")},
					output: map[string]string{"keys": ""}},
				{name: `set`,
					input:  ConfigUser{Keys: util.Pointer("aaaa")},
					output: map[string]string{"keys": "aaaa"}}}},
		{category: `LastName`,
			create: []test{
				{name: `empty`,
					input:  ConfigUser{LastName: util.Pointer("")},
					output: map[string]string{"userid": ""}},
				{name: `set`,
					input: ConfigUser{LastName: util.Pointer("Stark")},
					output: map[string]string{
						"lastname": "Stark",
						"userid":   ""}}},
			update: []test{
				{name: `empty`,
					input:  ConfigUser{LastName: util.Pointer("")},
					output: map[string]string{"lastname": ""}},
				{name: `set`,
					input:  ConfigUser{LastName: util.Pointer("Stark")},
					output: map[string]string{"lastname": "Stark"}}}},
		{category: `Password`,
			create: []test{
				{name: `set`,
					input:  ConfigUser{Password: util.Pointer(UserPassword("Enter123!"))},
					output: map[string]string{"userid": ""}}},
			update: []test{
				{name: `set`,
					input:  ConfigUser{Password: util.Pointer(UserPassword("Enter123!"))},
					output: nil}}},
		{category: `UserID`,
			create: []test{
				{name: `set`,
					input: ConfigUser{
						User: UserID{Name: "tony", Realm: "pve"}},
					output: map[string]string{
						"userid": "tony@pve"}}},
			update: []test{
				{name: `set`,
					input: ConfigUser{
						User: UserID{Name: "tony", Realm: "pve"}},
					output: nil}}},
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
				testParamsEqual(t, subTest.output, subTest.input.mapToApiUpdate())
			})
		}
	}
}

func Test_ConfigUser_mapToArray(t *testing.T) {
	base := func(user ConfigUser) ConfigUser {
		if user.Comment == nil {
			user.Comment = util.Pointer("")
		}
		if user.Email == nil {
			user.Email = util.Pointer("")
		}
		if user.Expire == nil {
			user.Expire = util.Pointer(uint(0))
		}
		if user.Enable == nil {
			user.Enable = util.Pointer(false)
		}
		if user.FirstName == nil {
			user.FirstName = util.Pointer("")
		}
		if user.LastName == nil {
			user.LastName = util.Pointer("")
		}
		return user
	}
	testData := []struct {
		input  []interface{}
		Output *[]ConfigUser
	}{
		{
			input: []any{
				map[string]any{
					"comment":   "test comment",
					"email":     "test@example.com",
					"expire":    float64(123456789),
					"firstname": "testFirstName",
					"keys":      "2fa",
					"lastname":  "testLastName"},
				map[string]any{
					"userid":    "username@pam",
					"email":     "test@example.com",
					"enable":    float64(1),
					"firstname": "testFirstName",
					"groups":    []any{"group1", "group2", "group3"},
					"lastname":  "testLastName"},
				map[string]any{
					"userid":  "username@pam",
					"comment": "test comment",
					"enable":  float64(1),
					"expire":  float64(123456789),
					"groups":  []any{"group1", "group2", "group3"},
					"keys":    "2fa"}},
			Output: &[]ConfigUser{
				base(ConfigUser{
					Comment:   util.Pointer("test comment"),
					Email:     util.Pointer("test@example.com"),
					Enable:    util.Pointer(false),
					Expire:    util.Pointer(uint(123456789)),
					FirstName: util.Pointer("testFirstName"),
					Groups:    &[]GroupName{},
					Keys:      util.Pointer("2fa"),
					LastName:  util.Pointer("testLastName")}),
				base(ConfigUser{
					User:      UserID{Name: "username", Realm: "pam"},
					Email:     util.Pointer("test@example.com"),
					Enable:    util.Pointer(true),
					FirstName: util.Pointer("testFirstName"),
					Groups:    &[]GroupName{"group1", "group2", "group3"},
					LastName:  util.Pointer("testLastName")}),
				base(ConfigUser{
					User:    UserID{Name: "username", Realm: "pam"},
					Comment: util.Pointer("test comment"),
					Enable:  util.Pointer(true),
					Expire:  util.Pointer(uint(123456789)),
					Groups:  &[]GroupName{"group1", "group2", "group3"},
					Keys:    util.Pointer("2fa")})}},
	}
	for _, e := range testData {
		require.Equal(t, e.Output, ConfigUser{}.mapToArray(e.input))
	}
}

func Test_rawConfigUser_Get(t *testing.T) {
	base := func(user ConfigUser) *ConfigUser {
		if user.Comment == nil {
			user.Comment = util.Pointer("")
		}
		if user.Email == nil {
			user.Email = util.Pointer("")
		}
		if user.Expire == nil {
			user.Expire = util.Pointer(uint(0))
		}
		if user.Enable == nil {
			user.Enable = util.Pointer(false)
		}
		if user.FirstName == nil {
			user.FirstName = util.Pointer("")
		}
		if user.Groups == nil {
			user.Groups = &[]GroupName{}
		}
		if user.LastName == nil {
			user.LastName = util.Pointer("")
		}
		return &user
	}
	type input struct {
		a    map[string]any
		user *UserID
	}
	tests := []struct {
		name   string
		input  input
		output *ConfigUser
	}{
		{name: `all fields`,
			input: input{a: map[string]any{
				"comment":   string("test comment"),
				"email":     string("test@example.com"),
				"enable":    float64(1),
				"expire":    float64(123456789),
				"firstname": string("testFirstName"),
				"groups":    string("group1,group2,group3"),
				"keys":      string("2fa"),
				"lastname":  string("testLastName")}},
			output: &ConfigUser{
				Comment:   util.Pointer("test comment"),
				Email:     util.Pointer("test@example.com"),
				Enable:    util.Pointer(true),
				Expire:    util.Pointer(uint(123456789)),
				FirstName: util.Pointer("testFirstName"),
				Groups:    &[]GroupName{"group1", "group2", "group3"},
				Keys:      util.Pointer("2fa"),
				LastName:  util.Pointer("testLastName")}},
		{name: `User list users`,
			input: input{a: map[string]any{"userid": string("username@pam")}},
			output: base(ConfigUser{
				User: UserID{Name: "username", Realm: "pam"}})},
		{name: `User get user`,
			input: input{a: map[string]any{}, user: util.Pointer(UserID{Name: "test-user", Realm: "pve"})},
			output: base(ConfigUser{
				User: UserID{Name: "test-user", Realm: "pve"}})},
		{name: `Comment`,
			input: input{a: map[string]any{"comment": "test comment"}},
			output: base(ConfigUser{
				Comment: util.Pointer("test comment")})},
		{name: `Email`,
			input: input{a: map[string]any{"email": string("test@example.com")}},
			output: base(ConfigUser{
				Email: util.Pointer("test@example.com")})},
		{name: `Enable`,
			input: input{a: map[string]any{"enable": float64(1)}},
			output: base(ConfigUser{
				Enable: util.Pointer(true)})},
		{name: `Expire`,
			input: input{a: map[string]any{"expire": float64(123456789)}},
			output: base(ConfigUser{
				Expire: util.Pointer(uint(123456789))})},
		{name: `FirstName`,
			input: input{a: map[string]any{"firstname": string("testFirstName")}},
			output: base(ConfigUser{
				FirstName: util.Pointer("testFirstName")})},
		{name: `Groups`,
			input: input{a: map[string]any{"groups": string("group1,group2,group3")}},
			output: base(ConfigUser{
				Groups: &[]GroupName{"group1", "group2", "group3"}})},
		{name: `Groups empty`,
			input: input{a: map[string]any{"groups": string("")}},
			output: base(ConfigUser{
				Groups: &[]GroupName{}})},
		{name: `Keys`,
			input: input{a: map[string]any{"keys": string("2fa")}},
			output: base(ConfigUser{
				Keys: util.Pointer("2fa")})},
		{name: `LastName`,
			input: input{a: map[string]any{"lastname": string("testLastName")}},
			output: base(ConfigUser{
				LastName: util.Pointer("testLastName")})},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, (&rawConfigUser{a: test.input.a, user: test.input.user}).Get())
		})
	}
}

// TODO improve when Name and Realm have their own types
func Test_ConfigUser_Validate(t *testing.T) {
	userId := UserID{Name: "user", Realm: "pam"}
	tests := []struct {
		name  string
		input ConfigUser
		err   error
	}{
		{name: `Invalid Empty`,
			input: ConfigUser{},
			err:   errors.New(`no username is specified`)},
		{name: `Invalid UserID`,
			input: ConfigUser{User: UserID{Name: "user"}},
			err:   errors.New(`no realm is specified`)},
		{name: `Valid Groups`,
			input: ConfigUser{
				User:   userId,
				Groups: &[]GroupName{"group1", "group2", "group3"}}},
		{name: `Invalid Groups Illegal`,
			input: ConfigUser{
				User:   userId,
				Groups: &[]GroupName{GroupName(test_data_group.GroupName_Max_Illegal())}},
			err: errors.New(`variable of type (GroupName) may not be more than 1000 characters long`)},
		{name: `Invalid Password too short`,
			input: ConfigUser{User: userId, Password: util.Pointer(UserPassword("aaaaaaa"))},
			err:   errors.New(`the minimum password length is 8 characters`)},
		{name: `Valid Password`,
			input: ConfigUser{User: userId, Password: util.Pointer(UserPassword("aaaaaaaa"))}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.err, test.input.Validate())
		})
	}
}

func Test_UserID_Parse(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output UserID
		err    error
	}{
		{name: `Valid`,
			input:  "username@pam",
			output: UserID{Name: "username", Realm: "pam"}},
		{name: `Invalid no realm`,
			input: "username@",
			err:   errors.New(Error_NewUserID)},
		{name: `Invalid no separator`,
			input: "usernamerealm",
			err:   errors.New(Error_NewUserID)},
		{name: `Invalid no name`,
			input: "@realm",
			err:   errors.New(Error_NewUserID)},
		{name: `Invalid empty`,
			input: "",
			err:   errors.New(Error_NewUserID)},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			var tmpID UserID
			err := tmpID.Parse(test.input)
			require.Equal(t, test.output, tmpID)
			require.Equal(t, test.err, err)
		})
	}
}

// TODO improve test when a validation function for the UserID exists
func Test_UserID_String(t *testing.T) {
	testData := []struct {
		input  UserID
		Output string
	}{
		{
			input: UserID{
				Name:  "username",
				Realm: "realm",
			},
			Output: "username@realm",
		},
		{input: UserID{Realm: "realm"}},
		{input: UserID{Name: "username"}},
		{input: UserID{}},
	}
	for _, e := range testData {
		require.Equal(t, e.Output, e.input.String())
	}
}

// TODO improve when Name and Realm have their own types
func Test_UserID_Validate(t *testing.T) {
	testData := []struct {
		input UserID
		err   bool
	}{
		{
			input: UserID{},
			err:   true,
		},
		{
			input: UserID{Name: "username"},
			err:   true,
		},
		{input: UserID{Name: "username", Realm: "pam"}},
	}
	for _, e := range testData {
		if e.err {
			require.Error(t, e.input.Validate())
		} else {
			require.NoError(t, e.input.Validate())
		}
	}
}

func Test_UserPassword_Validate(t *testing.T) {
	testData := []struct {
		input UserPassword
		err   bool
	}{
		{},
		{
			input: "1",
			err:   true,
		},
		{
			input: "12",
			err:   true,
		},
		{
			input: "123",
			err:   true,
		},
		{
			input: "1234",
			err:   true,
		},
		{
			input: "12345",
		},
		{
			input: "123456",
		},
	}
	for _, e := range testData {
		err := e.input.Validate()
		if e.err {
			require.Error(t, err)
		}
	}
}

func Test_rawUsersInfo_AsArray(t *testing.T) {
	tests := []struct {
		name   string
		input  rawUsersInfo
		output []RawUserInfo
	}{
		{name: `Single Partial`,
			input: rawUsersInfo{
				a: []any{map[string]any{"userid": "user1@pam"}}},
			output: []RawUserInfo{
				RawUserInfo(&rawUserInfo{a: map[string]any{"userid": "user1@pam"}})}},
		{name: `Single Full`,
			input: rawUsersInfo{
				full: true,
				a:    []any{map[string]any{"userid": "user1@pam"}}},
			output: []RawUserInfo{
				RawUserInfo(&rawUserInfo{
					a:    map[string]any{"userid": "user1@pam"},
					full: true})}},
		{name: `Multiple Partial`,
			input: rawUsersInfo{
				a: []any{
					map[string]any{"userid": "user1@pam"},
					map[string]any{"userid": "user2@pve"},
					map[string]any{"userid": "user3@ldap"}}},
			output: []RawUserInfo{
				RawUserInfo(&rawUserInfo{a: map[string]any{"userid": "user1@pam"}}),
				RawUserInfo(&rawUserInfo{a: map[string]any{"userid": "user2@pve"}}),
				RawUserInfo(&rawUserInfo{a: map[string]any{"userid": "user3@ldap"}})}},
		{name: `Multiple Full`,
			input: rawUsersInfo{
				full: true,
				a: []any{
					map[string]any{"userid": "user1@pam"},
					map[string]any{"userid": "user2@pve"},
					map[string]any{"userid": "user3@ldap"}}},
			output: []RawUserInfo{
				RawUserInfo(&rawUserInfo{
					a:    map[string]any{"userid": "user1@pam"},
					full: true}),
				RawUserInfo(&rawUserInfo{
					a:    map[string]any{"userid": "user2@pve"},
					full: true}),
				RawUserInfo(&rawUserInfo{
					a:    map[string]any{"userid": "user3@ldap"},
					full: true})}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, test.input.AsArray())
		})
	}
}

func Test_rawUsersInfo_AsMap(t *testing.T) {
	tests := []struct {
		name   string
		input  rawUsersInfo
		output map[UserID]RawUserInfo
	}{
		{name: `Single Partial`,
			input: rawUsersInfo{
				a: []any{map[string]any{"userid": "user1@pam"}}},
			output: map[UserID]RawUserInfo{
				{Name: "user1", Realm: "pam"}: &rawUserInfo{
					user: util.Pointer(UserID{Name: "user1", Realm: "pam"}),
					a:    map[string]any{"userid": "user1@pam"}}}},
		{name: `Single Full`,
			input: rawUsersInfo{
				full: true,
				a:    []any{map[string]any{"userid": "user1@pam"}}},
			output: map[UserID]RawUserInfo{
				{Name: "user1", Realm: "pam"}: &rawUserInfo{
					user: util.Pointer(UserID{Name: "user1", Realm: "pam"}),
					a:    map[string]any{"userid": "user1@pam"},
					full: true}}},
		{name: `Multiple Partial`,
			input: rawUsersInfo{
				a: []any{
					map[string]any{"userid": "user1@pam"},
					map[string]any{"userid": "user2@pve"},
					map[string]any{"userid": "user3@ldap"}}},
			output: map[UserID]RawUserInfo{
				{Name: "user1", Realm: "pam"}: &rawUserInfo{
					user: util.Pointer(UserID{Name: "user1", Realm: "pam"}),
					a:    map[string]any{"userid": "user1@pam"}},
				{Name: "user2", Realm: "pve"}: &rawUserInfo{
					user: util.Pointer(UserID{Name: "user2", Realm: "pve"}),
					a:    map[string]any{"userid": "user2@pve"}},
				{Name: "user3", Realm: "ldap"}: &rawUserInfo{
					user: util.Pointer(UserID{Name: "user3", Realm: "ldap"}),
					a:    map[string]any{"userid": "user3@ldap"}}}},
		{name: `Multiple Full`,
			input: rawUsersInfo{
				full: true,
				a: []any{
					map[string]any{"userid": "user1@pam"},
					map[string]any{"userid": "user2@pve"},
					map[string]any{"userid": "user3@ldap"}}},
			output: map[UserID]RawUserInfo{
				{Name: "user1", Realm: "pam"}: &rawUserInfo{
					user: util.Pointer(UserID{Name: "user1", Realm: "pam"}),
					a:    map[string]any{"userid": "user1@pam"},
					full: true},
				{Name: "user2", Realm: "pve"}: &rawUserInfo{
					user: util.Pointer(UserID{Name: "user2", Realm: "pve"}),
					a:    map[string]any{"userid": "user2@pve"},
					full: true},
				{Name: "user3", Realm: "ldap"}: &rawUserInfo{
					user: util.Pointer(UserID{Name: "user3", Realm: "ldap"}),
					a:    map[string]any{"userid": "user3@ldap"},
					full: true}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, test.input.AsMap())
		})
	}
}

func Test_rawUsersInfo_Len(t *testing.T) {
	tests := []struct {
		name   string
		input  rawUsersInfo
		output int
	}{
		{name: `Single`,
			input: rawUsersInfo{
				a: []any{map[string]any{"userid": "user1@pam"}}},
			output: 1},
		{name: `Multiple`,
			input: rawUsersInfo{
				a: []any{
					map[string]any{"userid": "user1@pam"},
					map[string]any{"userid": "user2@pve"},
					map[string]any{"userid": "user3@ldap"}}},
			output: 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, test.input.Len())
		})
	}
}

func Test_rawUsersInfo_SelectUser(t *testing.T) {
	tests := []struct {
		name   string
		user   UserID
		input  rawUsersInfo
		output RawUserInfo
		exists bool
	}{
		{name: `Single Nonexistent`,
			user: UserID{Name: "user1", Realm: "pam"},
			input: rawUsersInfo{
				a: []any{map[string]any{"userid": "user2@pam"}}}},
		{name: `Single Partial`,
			user: UserID{Name: "user1", Realm: "pam"},
			input: rawUsersInfo{
				a: []any{map[string]any{"userid": "user1@pam"}}},
			output: RawUserInfo(&rawUserInfo{
				user: util.Pointer(UserID{Name: "user1", Realm: "pam"}),
				a:    map[string]any{"userid": "user1@pam"}}),
			exists: true},
		{name: `Single Full`,
			user: UserID{Name: "user1", Realm: "pam"},
			input: rawUsersInfo{
				full: true,
				a:    []any{map[string]any{"userid": "user1@pam"}}},
			output: RawUserInfo(&rawUserInfo{
				user: util.Pointer(UserID{Name: "user1", Realm: "pam"}),
				a:    map[string]any{"userid": "user1@pam"},
				full: true}),
			exists: true},
		{name: `Multiple Nonexistent`,
			user: UserID{Name: "user2", Realm: "pam"},
			input: rawUsersInfo{
				a: []any{
					map[string]any{"userid": "user1@pam"},
					map[string]any{"userid": "user2@pve"},
					map[string]any{"userid": "user3@ldap"}}}},
		{name: `Multiple Partial`,
			user: UserID{Name: "user2", Realm: "pve"},
			input: rawUsersInfo{
				a: []any{
					map[string]any{"userid": "user1@pam"},
					map[string]any{"userid": "user2@pve"},
					map[string]any{"userid": "user3@ldap"}}},
			output: RawUserInfo(&rawUserInfo{
				user: util.Pointer(UserID{Name: "user2", Realm: "pve"}),
				a:    map[string]any{"userid": "user2@pve"}}),
			exists: true},
		{name: `Multiple Full`,
			user: UserID{Name: "user3", Realm: "ldap"},
			input: rawUsersInfo{
				full: true,
				a: []any{
					map[string]any{"userid": "user1@pam"},
					map[string]any{"userid": "user2@pve"},
					map[string]any{"userid": "user3@ldap"}}},
			output: RawUserInfo(&rawUserInfo{
				user: util.Pointer(UserID{Name: "user3", Realm: "ldap"}),
				a:    map[string]any{"userid": "user3@ldap"},
				full: true}),
			exists: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			raw, exists := test.input.SelectUser(test.user)
			require.Equal(t, test.exists, exists)
			require.Equal(t, test.output, raw)
		})
	}
}

func Test_rawUserInfo_Get(t *testing.T) {
	base := func(user UserInfo) UserInfo {
		if user.Config.Comment == nil {
			user.Config.Comment = util.Pointer("")
		}
		if user.Config.Email == nil {
			user.Config.Email = util.Pointer("")
		}
		if user.Config.Expire == nil {
			user.Config.Expire = util.Pointer(uint(0))
		}
		if user.Config.Enable == nil {
			user.Config.Enable = util.Pointer(false)
		}
		if user.Config.FirstName == nil {
			user.Config.FirstName = util.Pointer("")
		}
		if user.Config.LastName == nil {
			user.Config.LastName = util.Pointer("")
		}
		if user.Config.Groups == nil {
			user.Config.Groups = &[]GroupName{}
		}
		if user.Tokens == nil {
			user.Tokens = &[]ApiTokenConfig{}
		}
		return user
	}
	baseToken := func(token ApiTokenConfig) ApiTokenConfig {
		if token.Comment == nil {
			token.Comment = util.Pointer("")
		}
		if token.Expiration == nil {
			token.Expiration = util.Pointer(uint(0))
		}
		if token.PrivilegeSeparation == nil {
			token.PrivilegeSeparation = util.Pointer(false)
		}
		return token
	}
	removeGroupsAndTokens := func(user UserInfo) UserInfo {
		user.Config.Groups = nil
		user.Tokens = nil
		return user
	}
	tests := []struct {
		name   string
		input  map[string]any
		output UserInfo
	}{
		{name: `User`,
			input: map[string]any{"userid": "username@pam"},
			output: base(UserInfo{
				Config: ConfigUser{User: UserID{Name: "username", Realm: "pam"}}})},
		{name: `Comment`,
			input: map[string]any{"comment": "test comment"},
			output: base(UserInfo{
				Config: ConfigUser{Comment: util.Pointer("test comment")}})},
		{name: `Email`,
			input: map[string]any{"email": "test@example.com"},
			output: base(UserInfo{
				Config: ConfigUser{Email: util.Pointer("test@example.com")}})},
		{name: `Enable true`,
			input: map[string]any{"enable": float64(1)},
			output: base(UserInfo{
				Config: ConfigUser{Enable: util.Pointer(true)}})},
		{name: `Enable false`,
			input: map[string]any{"enable": float64(0)},
			output: base(UserInfo{
				Config: ConfigUser{Enable: util.Pointer(false)}})},
		{name: `Expire`,
			input: map[string]any{"expire": float64(123456789)},
			output: base(UserInfo{
				Config: ConfigUser{Expire: util.Pointer(uint(123456789))}})},
		{name: `FirstName`,
			input: map[string]any{"firstname": "testFirstName"},
			output: base(UserInfo{
				Config: ConfigUser{FirstName: util.Pointer("testFirstName")}})},
		{name: `Groups single`,
			input: map[string]any{"groups": "group1"},
			output: base(UserInfo{
				Config: ConfigUser{Groups: &[]GroupName{"group1"}}})},
		{name: `Groups multiple`,
			input: map[string]any{"groups": "group1,group2,group3"},
			output: base(UserInfo{
				Config: ConfigUser{Groups: &[]GroupName{"group1", "group2", "group3"}}})},
		{name: `LastName`,
			input: map[string]any{"lastname": "testLastName"},
			output: base(UserInfo{
				Config: ConfigUser{LastName: util.Pointer("testLastName")}})},
		{name: `Tokens`,
			input: map[string]any{"tokens": []any{
				map[string]any{"tokenid": "tokenName"},
				map[string]any{"comment": "token comment"},
				map[string]any{"expire": float64(123456789)},
				map[string]any{"privsep": float64(1)}}},
			output: base(UserInfo{
				Tokens: &[]ApiTokenConfig{
					baseToken(ApiTokenConfig{Name: "tokenName"}),
					baseToken(ApiTokenConfig{Comment: util.Pointer("token comment")}),
					baseToken(ApiTokenConfig{Expiration: util.Pointer(uint(123456789))}),
					baseToken(ApiTokenConfig{PrivilegeSeparation: util.Pointer(true)})}})},
	}
	for _, test := range tests {
		t.Run("Full/"+test.name, func(*testing.T) {
			require.Equal(t, test.output, RawUserInfo(&rawUserInfo{
				a:    test.input,
				full: true,
			}).Get())
		})
		t.Run("Partial/"+test.name, func(*testing.T) {
			require.Equal(t, removeGroupsAndTokens(test.output), RawUserInfo(&rawUserInfo{
				a:    test.input,
				full: false,
			}).Get())
		})
	}
}

func Test_NewUserID(t *testing.T) {
	testData := []struct {
		input  string
		output struct {
			id  UserID
			err error
		}
	}{
		{
			input: "username@pam",
			output: struct {
				id  UserID
				err error
			}{
				id: UserID{
					Name:  "username",
					Realm: "pam",
				},
			},
		},
		{
			input: "username@",
			output: struct {
				id  UserID
				err error
			}{
				err: errors.New(Error_NewUserID),
			},
		},
		{
			input: "usernamerealm",
			output: struct {
				id  UserID
				err error
			}{
				err: errors.New(Error_NewUserID),
			},
		},
		{
			input: "@realm",
			output: struct {
				id  UserID
				err error
			}{
				err: errors.New(Error_NewUserID),
			},
		},
	}
	for _, e := range testData {
		id, err := NewUserID(e.input)
		require.Equal(t, e.output.id, id)
		require.Equal(t, e.output.err, err)
	}
}

func Test_NewUserIDs(t *testing.T) {
	testData := []struct {
		input  string
		output *[]UserID
		err    bool
	}{
		// Valid
		{
			input:  "username@pam",
			output: &[]UserID{{Name: "username", Realm: "pam"}},
		},
		{
			input: "username@pam,root@pve,test@pam",
			output: &[]UserID{
				{Name: "username", Realm: "pam"},
				{Name: "root", Realm: "pve"},
				{Name: "test", Realm: "pam"},
			},
		},
		// Invalid
		{
			input: "username@",
			err:   true,
		},
		{
			input: "usernamerealm",
			err:   true,
		},
		{
			input: "@realm",
			err:   true,
		},
		{
			input: "username@pam,rootpve,test@pam",
			err:   true,
		},
	}
	for _, e := range testData {
		iDs, err := NewUserIDs(e.input)
		if e.err {
			require.Error(t, err)
		} else {
			require.Equal(t, e.output, iDs)
		}
	}
}
