package proxmox

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Client_checkInitialized(t *testing.T) {
	tests := []struct {
		name   string
		input  *Client
		output error
	}{
		{name: `nil`,
			output: errors.New(Client_Error_Nil)},
		{name: `session nil`,
			input:  &Client{},
			output: errors.New(Client_Error_NotInitialized)},
		{name: `Bypass checkInitialized`,
			input: fakeClient()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.checkInitialized())
		})
	}
}

func Test_Client_CheckPermissions(t *testing.T) {
	type input struct {
		client *Client
		perms  []Permission
	}
	tests := []struct {
		name   string
		input  input
		output error
	}{
		{"nil client", input{nil, []Permission{}}, errors.New(Client_Error_Nil)},
		{"user root@pam", input{&Client{Username: "root@pam"}, []Permission{}}, nil},
		{name: "direct permissions",
			input: input{
				client: &Client{permissions: map[permissionPath]privileges{
					"/access/pve": {UserModify: privilegeTrue}}},
				perms: []Permission{
					{Category: PermissionCategory_Access, Item: "pve", Privileges: Privileges{UserModify: true}}}}},
		{name: "propagate permissions",
			input: input{
				client: &Client{permissions: map[permissionPath]privileges{
					"/access": {UserModify: privilegePropagate}}},
				perms: []Permission{
					{Category: PermissionCategory_Access, Item: "pve", Privileges: Privileges{UserModify: true}}}}},
		{name: "missing permissions",
			input: input{
				client: &Client{permissions: map[permissionPath]privileges{
					"/": {UserModify: privilegeTrue}}},
				perms: []Permission{
					{Category: PermissionCategory_Root, Privileges: Privileges{PoolAllocate: true}}}},
			output: Permission{Category: PermissionCategory_Root, Privileges: Privileges{PoolAllocate: true}}.error()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.client.CheckPermissions(context.Background(), test.input.perms))
		})
	}
}

func Test_Version_Greater(t *testing.T) {
	type input struct {
		a Version
		b Version
	}
	tests := []struct {
		name   string
		input  input
		output bool
	}{
		{"a > b 0", input{Version{1, 0, 0}, Version{0, 0, 0}}, true},
		{"a > b 1", input{Version{0, 1, 0}, Version{0, 0, 255}}, true},
		{"a > b 2", input{Version{1, 0, 0}, Version{0, 255, 255}}, true},
		{"a < b 0", input{Version{7, 4, 1}, Version{7, 4, 2}}, false},
		{"a < b 1", input{Version{0, 0, 255}, Version{0, 1, 0}}, false},
		{"a < b 2", input{Version{0, 255, 255}, Version{1, 0, 0}}, false},
		{"a = b", input{Version{0, 0, 0}, Version{0, 0, 0}}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.a.Greater(test.input.b))
		})
	}
}

func Test_Version_mapToSDK(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]any
		output Version
		err    error
	}{
		{name: "unset",
			input: map[string]any{},
			err:   errors.New(Client_Error_UnableVersion)},
		{name: "full",
			input:  map[string]any{"version": "1.2.3"},
			output: Version{1, 2, 3}},
		{name: "invalid",
			input: map[string]any{"version": ""}},
		{name: "major",
			input:  map[string]any{"version": "1"},
			output: Version{1, 0, 0}},
		{name: "partial",
			input:  map[string]any{"version": "1.2"},
			output: Version{1, 2, 0}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, err := Version{}.mapToSDK(test.input)
			require.Equal(t, test.output, v)
			require.Equal(t, test.err, err)
		})
	}
}

func Test_Version_max(t *testing.T) {
	tests := []struct {
		name   string
		input  Version
		output Version
	}{
		{name: `max`,
			input:  Version{1, 5, 7},
			output: Version{1, 5, 7}},
		{name: `max Major, Minor, Patch`,
			input:  Version{0, 0, 0},
			output: Version{255, 255, 255}},
		{name: `max Major, Patch`,
			input:  Version{0, 5, 0},
			output: Version{255, 5, 255}},
		{name: `max Minor`,
			input:  Version{1, 0, 7},
			output: Version{1, 255, 7}},
		{name: `max Minor, Patch`,
			input:  Version{1, 0, 0},
			output: Version{1, 255, 255}},
		{name: `max Patch`,
			input:  Version{1, 5, 0},
			output: Version{1, 5, 255}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.max())
		})
	}
}

func Test_Version_Smaller(t *testing.T) {
	type input struct {
		a Version
		b Version
	}
	tests := []struct {
		name   string
		input  input
		output bool
	}{
		{"a > b 0", input{Version{1, 0, 0}, Version{0, 0, 0}}, false},
		{"a > b 1", input{Version{0, 1, 0}, Version{0, 0, 255}}, false},
		{"a > b 2", input{Version{1, 0, 0}, Version{0, 255, 255}}, false},
		{"a < b 0", input{Version{7, 4, 1}, Version{7, 4, 2}}, true},
		{"a < b 1", input{Version{0, 0, 255}, Version{0, 1, 0}}, true},
		{"a < b 2", input{Version{0, 255, 255}, Version{1, 0, 0}}, true},
		{"a = b", input{Version{0, 0, 0}, Version{0, 0, 0}}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.a.Smaller(test.input.b))
		})
	}
}

func Test_EncodedVersion_const(t *testing.T) {
	tests := []struct {
		input  Version
		output EncodedVersion
	}{
		{input: Version{Major: 9}, output: version_9_0_0},
		{input: Version{Major: 8}, output: version_8_0_0},
	}
	for _, test := range tests {
		t.Run(test.input.String(), func(t *testing.T) {
			require.Equal(t, test.output, test.input.Encode())
		})
	}
}
