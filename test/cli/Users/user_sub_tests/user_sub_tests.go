package user_sub_tests

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
)

// Default CLEANUP test for User
func Cleanup(t *testing.T, user proxmox.UserID) {
	Test := cliTest.Test{
		ReqErr:      true,
		ErrContains: user.String(),
		Args:        []string{"-i", "delete", "user", user.String()},
	}
	Test.StandardTest(t)
}

// Default DELETE test for User
func Delete(t *testing.T, user proxmox.UserID) {
	Test := cliTest.Test{
		Contains: []string{user.String()},
		Args:     []string{"-i", "delete", "user", user.String()},
	}
	Test.StandardTest(t)
}

// Default SET test for User
func Set(t *testing.T, user proxmox.ConfigUser) {
	userID := user.User.String()
	user.User = proxmox.UserID{}
	Test := cliTest.Test{
		InputJson: user,
		Contains:  []string{userID},
		Args:      []string{"-i", "set", "user", userID},
	}
	Test.StandardTest(t)
}
