package proxmox

import (
	"errors"
	"testing"

	"github.com/perimeter-81/proxmox-api-go/test/data/test_data_group"
	"github.com/stretchr/testify/require"
)

func Test_ConfigUser_mapToApiValues(t *testing.T) {
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
			input:  ConfigUser{Groups: &[]GroupName{"admin", "user"}},
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
			require.Equal(t, e.output, e.input.mapToApiValues(false))
			// for create add a empty "password" and "userid" keys
			require.Equal(t, setKeys([]string{"password", "userid"}, []any{UserPassword(""), ""}, e.output), e.input.mapToApiValues(true))
		case True:
			require.Equal(t, e.output, e.input.mapToApiValues(true))
		case False:
			require.Equal(t, e.output, e.input.mapToApiValues(false))
		}
	}
}

func Test_ConfigUser_mapToArray(t *testing.T) {
	testData := []struct {
		input  []interface{}
		Output *[]ConfigUser
	}{
		{
			input: []interface{}{
				map[string]interface{}{
					"comment":   "test comment",
					"email":     "test@example.com",
					"expire":    float64(123456789),
					"firstname": "testFirstName",
					"keys":      "2fa",
					"lastname":  "testLastName",
				},
				map[string]interface{}{
					"userid":    "username@pam",
					"email":     "test@example.com",
					"enable":    float64(1),
					"firstname": "testFirstName",
					"groups":    []interface{}{"group1", "group2", "group3"},
					"lastname":  "testLastName",
				},
				map[string]interface{}{
					"userid":  "username@pam",
					"comment": "test comment",
					"enable":  float64(1),
					"expire":  float64(123456789),
					"groups":  []interface{}{"group1", "group2", "group3"},
					"keys":    "2fa",
				},
			},
			Output: &[]ConfigUser{
				{
					Comment:   "test comment",
					Email:     "test@example.com",
					Expire:    123456789,
					FirstName: "testFirstName",
					Keys:      "2fa",
					LastName:  "testLastName",
				},
				{
					User:      UserID{Name: "username", Realm: "pam"},
					Email:     "test@example.com",
					Enable:    true,
					FirstName: "testFirstName",
					Groups:    &[]GroupName{"group1", "group2", "group3"},
					LastName:  "testLastName",
				},
				{
					User:    UserID{Name: "username", Realm: "pam"},
					Comment: "test comment",
					Enable:  true,
					Expire:  123456789,
					Groups:  &[]GroupName{"group1", "group2", "group3"},
					Keys:    "2fa",
				},
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.Output, ConfigUser{}.mapToArray(e.input))
	}
}

func Test_ConfigUser_mapToStruct(t *testing.T) {
	testData := []struct {
		base   ConfigUser
		input  map[string]interface{}
		output *ConfigUser
	}{
		{
			input: map[string]interface{}{
				"comment":   "test comment",
				"email":     "test@example.com",
				"enable":    float64(1),
				"expire":    float64(123456789),
				"firstname": "testFirstName",
				"groups":    "group1,group2,group3",
				"keys":      "2fa",
				"lastname":  "testLastName",
			},
			output: &ConfigUser{
				Comment:   "test comment",
				Email:     "test@example.com",
				Enable:    true,
				Expire:    123456789,
				FirstName: "testFirstName",
				Groups:    &[]GroupName{"group1", "group2", "group3"},
				Keys:      "2fa",
				LastName:  "testLastName",
			},
		},
		// Only User
		{
			input:  map[string]interface{}{"userid": "username@pam"},
			output: &ConfigUser{User: UserID{Name: "username", Realm: "pam"}},
		},
		{
			base:   ConfigUser{User: UserID{Name: "username1", Realm: "pve"}},
			output: &ConfigUser{User: UserID{Name: "username1", Realm: "pve"}},
		},
		{
			base:   ConfigUser{User: UserID{Name: "username1", Realm: "pve"}},
			input:  map[string]interface{}{"userid": "username@pam"},
			output: &ConfigUser{User: UserID{Name: "username", Realm: "pam"}},
		},
		// Only Comment
		{
			input:  map[string]interface{}{"comment": "test comment"},
			output: &ConfigUser{Comment: "test comment"},
		},
		{
			base:   ConfigUser{Comment: "Comment 1"},
			output: &ConfigUser{Comment: "Comment 1"},
		},
		{
			base:   ConfigUser{Comment: "Comment 1"},
			input:  map[string]interface{}{"comment": "test comment"},
			output: &ConfigUser{Comment: "test comment"},
		},
		// Only Email
		{
			input:  map[string]interface{}{"email": "test@example.com"},
			output: &ConfigUser{Email: "test@example.com"},
		},
		{
			base:   ConfigUser{Email: "test@proxmox.com"},
			output: &ConfigUser{Email: "test@proxmox.com"},
		},
		{
			base:   ConfigUser{Email: "test@proxmox.com"},
			input:  map[string]interface{}{"email": "test@example.com"},
			output: &ConfigUser{Email: "test@example.com"},
		},
		// Only Enable
		{
			input:  map[string]interface{}{"enable": float64(1)},
			output: &ConfigUser{Enable: true},
		},
		{
			base:   ConfigUser{Enable: true},
			output: &ConfigUser{Enable: true},
		},
		{
			base:   ConfigUser{Enable: true},
			input:  map[string]interface{}{"enable": float64(0)},
			output: &ConfigUser{Enable: false},
		},
		// Only Expire
		{
			input:  map[string]interface{}{"expire": float64(123456789)},
			output: &ConfigUser{Expire: 123456789},
		},
		{
			base:   ConfigUser{Expire: 10},
			output: &ConfigUser{Expire: 10},
		},
		{
			base:   ConfigUser{Expire: 10},
			input:  map[string]interface{}{"expire": float64(123456789)},
			output: &ConfigUser{Expire: 123456789},
		},
		// Only FirstName
		{
			input:  map[string]interface{}{"firstname": "testFirstName"},
			output: &ConfigUser{FirstName: "testFirstName"},
		},
		{
			base:   ConfigUser{FirstName: "TestName"},
			output: &ConfigUser{FirstName: "TestName"},
		},
		{
			base:   ConfigUser{FirstName: "TestName"},
			input:  map[string]interface{}{"firstname": "testFirstName"},
			output: &ConfigUser{FirstName: "testFirstName"},
		},
		// Only Groups
		{
			input:  map[string]interface{}{"groups": "group1,group2,group3"},
			output: &ConfigUser{Groups: &[]GroupName{"group1", "group2", "group3"}},
		},
		{
			base:   ConfigUser{Groups: &[]GroupName{"group4", "group5", "group6"}},
			output: &ConfigUser{Groups: &[]GroupName{"group4", "group5", "group6"}},
		},
		{
			base:   ConfigUser{Groups: &[]GroupName{"group4", "group5", "group6"}},
			input:  map[string]interface{}{"groups": "group1,group2,group3"},
			output: &ConfigUser{Groups: &[]GroupName{"group1", "group2", "group3"}},
		},
		// Group Empty List
		{
			input:  map[string]interface{}{"groups": ""},
			output: &ConfigUser{Groups: &[]GroupName{}},
		},
		// Groups as interface
		{
			input:  map[string]interface{}{"groups": []interface{}{"group1", "group2", "group3"}},
			output: &ConfigUser{Groups: &[]GroupName{"group1", "group2", "group3"}},
		},
		{
			input:  map[string]interface{}{"groups": []interface{}{}},
			output: &ConfigUser{Groups: &[]GroupName{}},
		},
		// Only Keys
		{
			input:  map[string]interface{}{"keys": "2fa"},
			output: &ConfigUser{Keys: "2fa"},
		},
		{
			base:   ConfigUser{Keys: "testKey"},
			output: &ConfigUser{Keys: "testKey"},
		},
		{
			base:   ConfigUser{Keys: "testKey"},
			input:  map[string]interface{}{"keys": "2fa"},
			output: &ConfigUser{Keys: "2fa"},
		},
		// Only LastName
		{
			input:  map[string]interface{}{"lastname": "testLastName"},
			output: &ConfigUser{LastName: "testLastName"},
		},
		{
			base:   ConfigUser{LastName: "Name"},
			output: &ConfigUser{LastName: "Name"},
		},
		{
			base:   ConfigUser{LastName: "Name"},
			input:  map[string]interface{}{"lastname": "testLastName"},
			output: &ConfigUser{LastName: "testLastName"},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.base.mapToStruct(e.input))
	}
}

// TODO improve when Name and Realm have their own types
func Test_ConfigUser_Validate(t *testing.T) {
	userId := UserID{Name: "user", Realm: "pam"}
	testData := []struct {
		input ConfigUser
		err   bool
	}{
		// Empty
		{
			input: ConfigUser{},
			err:   true,
		},
		// UserID
		{
			input: ConfigUser{User: UserID{Name: "user"}},
			err:   true,
		},
		// Groups
		{
			input: ConfigUser{
				User:   userId,
				Groups: &[]GroupName{"group1", "group2", "group3"},
			},
		},
		{
			input: ConfigUser{
				User:   userId,
				Groups: &[]GroupName{GroupName(test_data_group.GroupName_Max_Illegal())},
			},
			err: true,
		},
		// Password
		{
			input: ConfigUser{User: userId, Password: "aaa"},
			err:   true,
		},
		{input: ConfigUser{User: userId, Password: "aaaaa"}},
	}
	for _, e := range testData {
		err := e.input.Validate()
		if e.err {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func Test_configUserShort_mapToApiValues(t *testing.T) {
	testData := []struct {
		input  configUserShort
		output map[string]interface{}
	}{
		{
			input:  configUserShort{},
			output: map[string]interface{}{"groups": ""},
		},
		{
			input:  configUserShort{Groups: &[]GroupName{"group1"}},
			output: map[string]interface{}{"groups": "group1"},
		},
		{
			input:  configUserShort{Groups: &[]GroupName{"group1", "group2", "group3"}},
			output: map[string]interface{}{"groups": "group1,group2,group3"},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.input.mapToApiValues())
	}
}

func Test_UserID_mapToArray(t *testing.T) {
	testData := []struct {
		input  []interface{}
		Output *[]UserID
	}{
		{
			input:  []interface{}{},
			Output: &[]UserID{},
		},
		{
			input:  []interface{}{"user1realm"},
			Output: &[]UserID{{}},
		},
		{
			input: []interface{}{"user1realm", "", "user3@pve"},
			Output: &[]UserID{
				{},
				{},
				{Name: "user3", Realm: "pve"},
			},
		},
		{
			input:  []interface{}{"user1@realm"},
			Output: &[]UserID{{Name: "user1", Realm: "realm"}},
		},
		{
			input: []interface{}{"user1@realm", "user2@pam", "user3@pve"},
			Output: &[]UserID{
				{Name: "user1", Realm: "realm"},
				{Name: "user2", Realm: "pam"},
				{Name: "user3", Realm: "pve"},
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.Output, UserID{}.mapToArray(e.input))
	}
}

func Test_UserID_mapToStruct(t *testing.T) {
	testData := []struct {
		input  string
		output UserID
	}{
		{},
		{input: "user"},
		{input: "@realm"},
		{
			input:  "user@realm",
			output: UserID{Name: "user", Realm: "realm"},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, UserID{}.mapToStruct(e.input))
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
