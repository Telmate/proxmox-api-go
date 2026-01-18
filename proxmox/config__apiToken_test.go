package proxmox

import (
	"context"
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/mockServer"
	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_apiTokenClient_Create(t *testing.T) {
	const path = "/access/users/test@pve/token/testToken"
	tests := []struct {
		name     string
		userID   UserID
		apiToken ApiTokenConfig
		secret   ApiTokenSecret
		requests []mockServer.Request
		err      error
	}{
		{name: `Create`,
			userID: UserID{Name: "test", Realm: "pve"},
			secret: "FAKE_SECRET",
			apiToken: ApiTokenConfig{
				Comment:             util.Pointer("this is a test token"),
				Expiration:          util.Pointer(uint(5)),
				Name:                "testToken",
				PrivilegeSeparation: util.Pointer(true)},
			requests: mockServer.RequestsPostResponse(path, map[string]any{
				"privsep": "1",
				"comment": "this is a test token",
				"expire":  "5",
			}, []byte(`{"data":{"value":"FAKE_SECRET"}}`))},
		{name: `validate error empty username`,
			userID: UserID{Realm: "pve"},
			err:    errors.New("no username is specified")},
		{name: `validate error empty username`,
			userID: UserID{Realm: "pve", Name: "test"},
			err:    errors.New(`api token name must match the following regex: ^[A-Za-z][A-Za-z0-9\.-_]{1,127}$`)},
		{name: `500 internal server error`,
			userID: UserID{
				Name:  "test",
				Realm: "pve"},
			apiToken: ApiTokenConfig{
				Name: "testToken"},
			requests: mockServer.RequestsError(path, mockServer.POST, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			secret, err := c.New().ApiToken.Create(context.Background(), test.userID, test.apiToken)
			require.Equal(t, test.err, err)
			require.Equal(t, test.secret, secret)
			server.Clear(t)
		})
	}
}

func Test_apiTokenClient_Delete(t *testing.T) {
	const path = "/access/users/test@pve/token/testToken"
	tests := []struct {
		name     string
		deleted  bool
		apiToken ApiTokenID
		requests []mockServer.Request
		err      error
	}{
		{name: `Delete token exists`,
			deleted: true,
			apiToken: ApiTokenID{
				User:      UserID{Name: "test", Realm: "pve"},
				TokenName: "testToken"},
			requests: mockServer.RequestsDelete(path, nil)},
		{name: `Delete token doesn't exists`,
			apiToken: ApiTokenID{
				User:      UserID{Name: "test", Realm: "pve"},
				TokenName: "testToken"},
			requests: mockServer.RequestsErrorHandled(path, mockServer.DELETE, mockServer.HTTPerror{
				Message: `{"message":"no such token 'testToken' for user 'test@pve'\n"}`,
				Code:    500})},
		{name: `validate error empty username`,
			apiToken: ApiTokenID{
				User: UserID{Name: "test"}},
			err: errors.New("no realm is specified")},
		{name: `500 internal server error`,
			apiToken: ApiTokenID{
				User:      UserID{Name: "test", Realm: "pve"},
				TokenName: "testToken"},
			requests: mockServer.RequestsError(path, mockServer.DELETE, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			deleted, err := c.New().ApiToken.Delete(context.Background(), test.apiToken)
			require.Equal(t, test.err, err)
			require.Equal(t, test.deleted, deleted)
			server.Clear(t)
		})
	}
}

func Test_apiTokenClient_Exists(t *testing.T) {
	const path = "/access/users/test@pve/token/testToken"
	tests := []struct {
		name     string
		apiToken ApiTokenID
		exists   bool
		requests []mockServer.Request
		err      error
	}{
		{name: `Exits true`,
			exists: true,
			apiToken: ApiTokenID{
				User:      UserID{Name: "test", Realm: "pve"},
				TokenName: "testToken"},
			requests: mockServer.RequestsGetJson(path, map[string]any{
				"data": map[string]any{},
			})},
		{name: `Exits false`,
			apiToken: ApiTokenID{
				User:      UserID{Name: "test", Realm: "pve"},
				TokenName: "testToken"},
			requests: mockServer.RequestsErrorHandled(path, mockServer.GET, mockServer.HTTPerror{
				Message: `{"message":"no such token 'testtoken' for user 'Test_Token_Create@pve'\n"}`,
				Code:    500})},
		{name: `validate error empty username`,
			apiToken: ApiTokenID{User: UserID{Realm: "pve"}},
			err:      errors.New("no username is specified")},
		{name: `500 internal server error`,
			apiToken: ApiTokenID{
				User:      UserID{Name: "test", Realm: "pve"},
				TokenName: "testToken"},
			requests: mockServer.RequestsError(path, mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			exists, err := c.New().ApiToken.Exists(context.Background(), test.apiToken)
			require.Equal(t, test.err, err)
			require.Equal(t, exists, exists)
			server.Clear(t)
		})
	}
}

func Test_apiTokenClient_List(t *testing.T) {
	const path = "/access/users/test@pve/token"
	base := func(token ApiTokenConfig) ApiTokenConfig {
		if token.Comment == nil {
			token.Comment = util.Pointer("")
		}
		if token.Expiration == nil {
			token.Expiration = util.Pointer(uint(0))
		}
		if token.PrivilegeSeparation == nil {
			token.PrivilegeSeparation = util.Pointer(true)
		}
		return token
	}

	tests := []struct {
		name     string
		user     UserID
		output   map[ApiTokenName]ApiTokenConfig
		requests []mockServer.Request
		err      error
	}{
		{name: `List`,
			user: UserID{Name: "test", Realm: "pve"},
			output: map[ApiTokenName]ApiTokenConfig{
				"token1": base(ApiTokenConfig{
					Name:                "token1",
					PrivilegeSeparation: util.Pointer(false)}),
				"token2": base(ApiTokenConfig{
					Name:       "token2",
					Expiration: util.Pointer(uint(123456))}),
				"token3": base(ApiTokenConfig{
					Name:    "token3",
					Comment: util.Pointer("test comment")}),
			},
			requests: mockServer.RequestsGetJson(path, map[string]any{
				"data": []map[string]any{
					{"tokenid": "token1",
						"comment": "",
						"expire":  0,
						"privsep": 0},
					{"tokenid": "token2",
						"comment": "",
						"expire":  123456,
						"privsep": 1},
					{"tokenid": "token3",
						"comment": "test comment",
						"privsep": 1,
						"expire":  0},
				},
			})},
		{name: `validate error empty username`,
			user: UserID{Realm: "pve"},
			err:  errors.New("no username is specified")},
		{name: `500 internal server error`,
			user:     UserID{Name: "test", Realm: "pve"},
			requests: mockServer.RequestsError(path, mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			raw, err := c.New().ApiToken.List(context.Background(), test.user)
			require.Equal(t, test.err, err)
			if test.err == nil {
				testCompareRawMap(t, test.output, raw.AsMap())
				require.Equal(t, len(test.output), raw.Len())
			}
			server.Clear(t)
		})
	}
}

func Test_apiTokenClient_Read(t *testing.T) {
	const token = "token1"
	const path = "/access/users/test@pve/token/" + token
	tests := []struct {
		name     string
		token    ApiTokenID
		output   ApiTokenConfig
		requests []mockServer.Request
		err      error
	}{
		{name: `Read exists`,
			token: ApiTokenID{User: UserID{Name: "test", Realm: "pve"}, TokenName: token},
			output: ApiTokenConfig{
				Name:                token,
				Comment:             util.Pointer(""),
				Expiration:          util.Pointer(uint(0)),
				PrivilegeSeparation: util.Pointer(false)},
			requests: mockServer.RequestsGetJson(path, map[string]any{
				"data": map[string]any{
					"name":    token,
					"comment": "",
					"expire":  0,
					"privsep": 0,
				}})},
		{name: `Read not exists`,
			token: ApiTokenID{User: UserID{Name: "test", Realm: "pve"}, TokenName: token},
			err:   errors.New("api token does not exist"),
			requests: mockServer.RequestsErrorHandled(path, mockServer.GET, mockServer.HTTPerror{
				Message: `{"message":"no such token 'token1' for user 'test@pve'\n"}`,
				Code:    500})},
		{name: `validate error empty username`,
			token: ApiTokenID{User: UserID{Realm: "pve"}, TokenName: token},
			err:   errors.New("no username is specified")},
		{name: `500 internal server error`,
			token:    ApiTokenID{User: UserID{Name: "test", Realm: "pve"}, TokenName: "token1"},
			requests: mockServer.RequestsError(path, mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			raw, err := c.New().ApiToken.Read(context.Background(), test.token)
			require.Equal(t, test.err, err)
			if test.err == nil {
				require.Equal(t, test.output, raw.Get())
			}
			server.Clear(t)
		})
	}
}

func Test_apiTokenClient_Update(t *testing.T) {
	const token = "token1"
	const path = "/access/users/test@pve/token/" + token
	tests := []struct {
		name     string
		token    ApiTokenConfig
		user     UserID
		requests []mockServer.Request
		err      error
	}{
		{name: `update`,
			user: UserID{Name: "test", Realm: "pve"},
			token: ApiTokenConfig{
				Name:                token,
				Comment:             util.Pointer("test comment"),
				Expiration:          util.Pointer(uint(123456)),
				PrivilegeSeparation: util.Pointer(true)},
			requests: mockServer.RequestsPut(path, map[string]any{
				"comment": "test comment",
				"expire":  "123456",
				"privsep": "1",
			})},
		{name: `do nothing`,
			user:  UserID{Name: "test", Realm: "pve"},
			token: ApiTokenConfig{Name: token}},
		{name: `validate error empty username`,
			user: UserID{Realm: "pve"},
			err:  errors.New("no username is specified")},
		{name: `validate error invalid token name`,
			user: UserID{Name: "test", Realm: "pve"},
			token: ApiTokenConfig{
				Name: "!nVAlid"},
			err: errors.New(`api token name must match the following regex: ^[A-Za-z][A-Za-z0-9\.-_]{1,127}$`)},
		{name: `500 internal server error`,
			user: UserID{Name: "test", Realm: "pve"},
			token: ApiTokenConfig{
				Name:                token,
				Comment:             util.Pointer("test comment"),
				Expiration:          util.Pointer(uint(123456)),
				PrivilegeSeparation: util.Pointer(true)},
			requests: mockServer.RequestsError(path, mockServer.PUT, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().ApiToken.Update(context.Background(), test.user, test.token)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_ApiTokenConfig_mapToAPI(t *testing.T) {
	type test struct {
		name   string
		config ApiTokenConfig
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
					config: ApiTokenConfig{Comment: util.Pointer("")}}},
			createUpdate: []test{
				{name: `set`,
					config: ApiTokenConfig{Comment: util.Pointer("My comment with symbols !@#$%^&*()=@+")},
					output: map[string]string{"comment": "My comment with symbols !@#$%^&*()=@+"}}},
			update: []test{
				{name: `empty`, // Bug in Proxmox API: setting empty comment via API leaves previous comment unchanged, prefixed spaces are trimmed from the comment.
					config: ApiTokenConfig{Comment: util.Pointer("")},
					output: map[string]string{"comment": " "}}}},
		{category: `Expiration`,
			create: []test{
				{name: `empty`,
					config: ApiTokenConfig{Expiration: util.Pointer(uint(0))}}},
			createUpdate: []test{
				{name: `set`,
					config: ApiTokenConfig{Expiration: util.Pointer(uint(1672531199))},
					output: map[string]string{"expire": "1672531199"}}},
			update: []test{
				{name: `empty`,
					config: ApiTokenConfig{Expiration: util.Pointer(uint(0))},
					output: map[string]string{"expire": "0"}}}},
		{category: `PrivilegeSeparation`,
			createUpdate: []test{
				{name: `true`,
					config: ApiTokenConfig{PrivilegeSeparation: util.Pointer(true)},
					output: map[string]string{"privsep": "1"}},
				{name: `false`,
					config: ApiTokenConfig{PrivilegeSeparation: util.Pointer(false)},
					output: map[string]string{"privsep": "0"}}}},
		{category: `Name`,
			createUpdate: []test{
				{config: ApiTokenConfig{Name: "myToken"}}}},
		{category: `all`,
			createUpdate: []test{
				{name: `full`,
					config: ApiTokenConfig{
						Name:                "fullToken",
						Comment:             util.Pointer("full comment"),
						Expiration:          util.Pointer(uint(1672531199)),
						PrivilegeSeparation: util.Pointer(true)},
					output: map[string]string{
						"comment": "full comment",
						"expire":  "1672531199",
						"privsep": "1"}}}},
	}
	for _, test := range tests {
		for _, subTest := range append(test.create, test.createUpdate...) {
			name := test.category + "/Create/" + subTest.name
			t.Run(name, func(*testing.T) {
				testParamsEqual(t, subTest.output, subTest.config.mapToApiCreate())
			})
		}
		for _, subTest := range append(test.update, test.createUpdate...) {
			name := test.category + "/Update/" + subTest.name
			t.Run(name, func(*testing.T) {
				testParamsEqual(t, subTest.output, subTest.config.mapToApiUpdate())
			})
		}
	}
}

func Test_rawApiTokens_AsArray(t *testing.T) {
	tests := []struct {
		name   string
		input  rawApiTokens
		output []RawApiTokenConfig
	}{
		{name: `No token`,
			input: rawApiTokens{
				a: []any{}},
			output: []RawApiTokenConfig{}},
		{name: `Single token`,
			input: rawApiTokens{
				a: []any{
					map[string]any{"tokenid": "token1"}}},
			output: []RawApiTokenConfig{
				&rawApiTokenConfig{
					a: map[string]any{"tokenid": "token1"}}}},
		{name: `Multiple tokens`,
			input: rawApiTokens{
				a: []any{
					map[string]any{
						"tokenid": "token1",
						"comment": "comment2"},
					map[string]any{
						"expire":  float64(1000),
						"tokenid": "token2"},
					map[string]any{
						"privsep": float64(1),
						"tokenid": "token3"}}},
			output: []RawApiTokenConfig{
				&rawApiTokenConfig{
					a: map[string]any{
						"comment": "comment2",
						"tokenid": "token1"}},
				&rawApiTokenConfig{
					a: map[string]any{
						"expire":  float64(1000),
						"tokenid": "token2"}},
				&rawApiTokenConfig{
					a: map[string]any{
						"privsep": float64(1),
						"tokenid": "token3"}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.ElementsMatch(t, test.output, RawApiTokens(&test.input).AsArray())
		})
	}
}

func Test_rawApiTokenConfig_Get(t *testing.T) {
	base := func(config ApiTokenConfig) ApiTokenConfig {
		if config.Comment == nil {
			config.Comment = util.Pointer("")
		}
		if config.Expiration == nil {
			config.Expiration = util.Pointer(uint(0))
		}
		if config.PrivilegeSeparation == nil {
			config.PrivilegeSeparation = util.Pointer(false)
		}
		return config
	}
	tests := []struct {
		name   string
		input  rawApiTokenConfig
		output ApiTokenConfig
	}{
		{name: `Name`,
			input: rawApiTokenConfig{a: map[string]any{
				"tokenid": "test",
			}},
			output: base(ApiTokenConfig{Name: "test"})},
		{name: `Name pointer`,
			input:  rawApiTokenConfig{name: util.Pointer(ApiTokenName("test"))},
			output: base(ApiTokenConfig{Name: "test"})},
		{name: `Comment`,
			input: rawApiTokenConfig{
				a: map[string]any{"comment": string("a comment")}},
			output: base(ApiTokenConfig{Comment: util.Pointer("a comment")})},
		{name: `Expiration`,
			input: rawApiTokenConfig{
				a: map[string]any{"expire": float64(123456)}},
			output: base(ApiTokenConfig{Expiration: util.Pointer(uint(123456))})},
		{name: `PrivilegeSeparation true`,
			input: rawApiTokenConfig{
				a: map[string]any{"privsep": float64(1)}},
			output: base(ApiTokenConfig{PrivilegeSeparation: util.Pointer(true)})},
		{name: `PrivilegeSeparation false`,
			input: rawApiTokenConfig{
				a: map[string]any{"privsep": float64(0)}},
			output: base(ApiTokenConfig{PrivilegeSeparation: util.Pointer(false)})},
		{name: `all`,
			input: rawApiTokenConfig{
				name: util.Pointer(ApiTokenName("testName")),
				a: map[string]any{
					"comment": string("test comment"),
					"expire":  float64(654321),
					"privsep": float64(1)}},
			output: ApiTokenConfig{
				Comment:             util.Pointer("test comment"),
				Expiration:          util.Pointer(uint(654321)),
				Name:                "testName",
				PrivilegeSeparation: util.Pointer(true)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, RawApiTokenConfig(&test.input).Get())
		})
	}
}

func Test_ApiTokenID_Parse(t *testing.T) {
	err := errors.New("api token ID must be in the format user@realm!tokenname")
	tests := []struct {
		name   string
		input  string
		output ApiTokenID
		err    error
	}{
		{name: `Invalid empty`,
			input: "",
			err:   err},
		{name: `Invalid missing @`,
			input: "userRealm!token",
			err:   err},
		{name: `Invalid no room between @ and !`,
			input: "user@!token",
			err:   err},
		{name: `Invalid missing !`,
			input: "user@RealmToken",
			err:   err},
		{name: `Valid`,
			input: "user@Realm!tokenName",
			output: ApiTokenID{
				User:      UserID{Name: "user", Realm: "Realm"},
				TokenName: "tokenName"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			var result ApiTokenID
			err := result.Parse(test.input)
			require.Equal(t, test.err, err)
			require.Equal(t, test.output, result)
		})
	}
}

func Test_ApiTokenID_String(t *testing.T) {
	require.Equal(t, string("test@pve!TokeName"), ApiTokenID{
		User:      UserID{Name: "test", Realm: "pve"},
		TokenName: "TokeName"}.String())
}

func Test_ApiTokenID_Validate(t *testing.T) {
	tests := []struct {
		name  string
		input ApiTokenID
		err   error
	}{
		{name: `Invalid empty username`,
			err: errors.New("no username is specified")},
		{name: `Invalid empty realm`,
			input: ApiTokenID{
				User: UserID{Name: "test"}},
			err: errors.New("no realm is specified")},
		{name: `Invalid empty token name`,
			input: ApiTokenID{
				User: UserID{Name: "test", Realm: "pve"}},
			err: errors.New(`api token name must match the following regex: ^[A-Za-z][A-Za-z0-9\.-_]{1,127}$`)},
		{name: `Invalid token name too long`,
			input: ApiTokenID{
				User:      UserID{Name: "test", Realm: "pve"},
				TokenName: ApiTokenName("thisTokenNameIsWayTooLongBecauseItExceedsTheMaximumLengthOfOneHundredAndTwentyEightCharactersWhichIsNotAllowedAaaaaaaaaaaaaaaaaaa")},
			err: errors.New(`api token name must match the following regex: ^[A-Za-z][A-Za-z0-9\.-_]{1,127}$`)},
		{name: `Invalid token name invalid characters`,
			input: ApiTokenID{
				User:      UserID{Name: "test", Realm: "pve"},
				TokenName: ApiTokenName("invalid*name")},
			err: errors.New(`api token name must match the following regex: ^[A-Za-z][A-Za-z0-9\.-_]{1,127}$`)},
		{name: `Valid`,
			input: ApiTokenID{
				User:      UserID{Name: "test", Realm: "pve"},
				TokenName: ApiTokenName("valid-name")}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.err, test.input.Validate())
		})
	}
}

func Test_ApiTokenSecret_String(t *testing.T) {
	require.Equal(t, string("secretValue"), ApiTokenSecret("secretValue").String())
}
