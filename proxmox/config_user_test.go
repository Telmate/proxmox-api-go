package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ConfigUser_mapToAPI(t *testing.T) {
	True := 1
	False := 0
	TrueAndFalse := 2
	minimumOutput := func() map[string]interface{} {
		return map[string]interface{}{
			"comment":   "",
			"email":     "",
			"enable":    false,
			"expire":    uint(0),
			"firstname": "",
			"groups":    "",
			"keys":      "",
			"lastname":  "",
		}
	}
	setKeys := func(key []string, value []any, params map[string]interface{}) map[string]interface{} {
		for i := range key {
			params[key[i]] = value[i]
		}
		return params
	}
	testData := []struct {
		input  ConfigUser
		create int
		output map[string]interface{}
	}{
		{
			input:  ConfigUser{Comment: "test comment"},
			create: TrueAndFalse,
			output: setKeys([]string{"comment"}, []any{"test comment"}, minimumOutput()),
		},
		{
			input:  ConfigUser{Email: "test@example.com"},
			create: TrueAndFalse,
			output: setKeys([]string{"email"}, []any{"test@example.com"}, minimumOutput()),
		},
		{
			input:  ConfigUser{Enable: true},
			create: TrueAndFalse,
			output: setKeys([]string{"enable"}, []any{true}, minimumOutput()),
		},
		{
			input:  ConfigUser{Expire: 784873474},
			create: TrueAndFalse,
			output: setKeys([]string{"expire"}, []any{uint(784873474)}, minimumOutput()),
		},
		{
			input:  ConfigUser{FirstName: "Tony"},
			create: TrueAndFalse,
			output: setKeys([]string{"firstname"}, []any{"Tony"}, minimumOutput()),
		},
		{
			input:  ConfigUser{Groups: []string{"admin", "user"}},
			create: TrueAndFalse,
			output: setKeys([]string{"groups"}, []any{"admin,user"}, minimumOutput()),
		},
		{
			input:  ConfigUser{Keys: "aaaa"},
			create: TrueAndFalse,
			output: setKeys([]string{"keys"}, []any{"aaaa"}, minimumOutput()),
		},
		{
			input:  ConfigUser{LastName: "Stark"},
			create: TrueAndFalse,
			output: setKeys([]string{"lastname"}, []any{"Stark"}, minimumOutput()),
		},
		{
			input:  ConfigUser{Password: "Enter123!"},
			create: True,
			output: setKeys([]string{"password", "userid"}, []any{UserPassword("Enter123!"), ""}, minimumOutput()),
		},
		{
			input: ConfigUser{User: UserID{
				Name:  "TStark",
				Realm: "pam",
			}},
			create: True,
			output: setKeys([]string{"password", "userid"}, []any{UserPassword(""), "TStark@pam"}, minimumOutput()),
		},
		{
			input: ConfigUser{
				Password: "Enter123!",
				User: UserID{
					Name:  "TStark",
					Realm: "pam",
				},
			},
			create: False,
			output: minimumOutput(),
		},
	}
	for _, e := range testData {
		switch e.create {
		case TrueAndFalse:
			require.Equal(t, e.output, e.input.mapToAPI(false))
			// for create add a empty "password" and "userid" keys
			require.Equal(t, setKeys([]string{"password", "userid"}, []any{UserPassword(""), ""}, e.output), e.input.mapToAPI(true))
		case True:
			require.Equal(t, e.output, e.input.mapToAPI(true))
		case False:
			require.Equal(t, e.output, e.input.mapToAPI(false))
		}
	}
}

// TODO improve test when a validation function for the UserID exists
func Test_ConfigUser_Validate(t *testing.T) {
	testData := []struct {
		input ConfigUser
		err   bool
	}{
		{
			input: ConfigUser{},
		},
		{
			input: ConfigUser{
				Password: "aaa",
			},
			err: true,
		},
		{
			input: ConfigUser{
				Password: "aaaaa",
			},
		},
	}
	for _, e := range testData {
		err := e.input.Validate()
		if e.err {
			require.Error(t, err)
		}
	}
}

// TODO improve test when a validation function for the UserID exists
func Test_UserID_ToString(t *testing.T) {
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
		require.Equal(t, e.Output, e.input.ToString())
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

func Test_mapToStruct(t *testing.T) {
	testData := []struct {
		input struct {
			id     UserID
			params map[string]interface{}
		}
		output *ConfigUser
	}{
		{
			input: struct {
				id     UserID
				params map[string]interface{}
			}{
				id: UserID{Name: "username", Realm: "pam"},
				params: map[string]interface{}{
					"comment":   "test comment",
					"email":     "test@example.com",
					"enable":    float64(1),
					"expire":    float64(123456789),
					"firstname": "testFirstName",
					"groups":    []interface{}{"group1", "group2", "group3"},
					"keys":      "2fa",
					"lastname":  "testLastName",
				},
			},
			output: &ConfigUser{
				User:      UserID{Name: "username", Realm: "pam"},
				Comment:   "test comment",
				Email:     "test@example.com",
				Enable:    true,
				Expire:    123456789,
				FirstName: "testFirstName",
				Groups:    []string{"group1", "group2", "group3"},
				Keys:      "2fa",
				LastName:  "testLastName",
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, mapToStructConfigUser(e.input.id, e.input.params))
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
