package user_sub_tests

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
)

// Default CLEANUP test for User
func Cleanup(t *testing.T, user proxmox.UserID) {
	Test := cliTest.Test{
		ReqErr:      true,
		ErrContains: user.ToString(),
		Args:        []string{"-i", "delete", "user", user.ToString()},
	}
	Test.StandardTest(t)
}

// Default DELETE test for User
func Delete(t *testing.T, user proxmox.UserID) {
	Test := cliTest.Test{
		Contains: []string{user.ToString()},
		Args:     []string{"-i", "delete", "user", user.ToString()},
	}
	Test.StandardTest(t)
}

// Default SET test for User
func Set(t *testing.T, user proxmox.ConfigUser) {
	userID := user.User.ToString()
	user.User = proxmox.UserID{}
	Test := cliTest.Test{
		InputJson: user,
		Contains:  []string{userID},
		Args:      []string{"-i", "set", "user", userID},
	}
	Test.StandardTest(t)
}
