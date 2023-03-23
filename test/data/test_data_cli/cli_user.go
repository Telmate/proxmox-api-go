package test_data_cli

import (
	"fmt"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
)

func User_Empty_testData(testNumber uint) proxmox.ConfigUser {
	return proxmox.ConfigUser{
		User:   proxmox.UserID{Name: fmt.Sprintf("test-user%d", testNumber), Realm: "pve"},
		Enable: false,
		Expire: 0,
		Groups: &[]proxmox.GroupName{},
	}
}

func User_Full_testData(testNumber uint) proxmox.ConfigUser {
	return proxmox.ConfigUser{
		User:      proxmox.UserID{Name: fmt.Sprintf("test-user%d", testNumber), Realm: "pve"},
		Comment:   "this is a comment",
		Email:     "b.wayne@proxmox.com",
		Enable:    true,
		Expire:    253370811600,
		FirstName: "Bruce",
		Groups:    &[]proxmox.GroupName{proxmox.GroupName(fmt.Sprintf("user%d-group%d", testNumber, testNumber))},
		Keys:      "2fa key",
		LastName:  "Wayne",
	}
}
