package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Permission_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  Permission
		output error
	}{
		{"valid category", Permission{Category: PermissionCategory_Access}, nil},
		{"invalid category", Permission{Category: "abc"}, PermissionCategory("").Error()},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_PermissionCategory_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  PermissionCategory
		output error
	}{
		{"valid category", PermissionCategory_Access, nil},
		{"invalid category", "abc", PermissionCategory("").Error()},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
