package test_data_cli

import (
	"fmt"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/proxmox"
)

func User_Empty_testData(testNumber uint) proxmox.ConfigUser {
	return proxmox.ConfigUser{
		User:   proxmox.UserID{Name: fmt.Sprintf("test-user%d", testNumber), Realm: "pve"},
		Enable: util.Pointer(false),
		Expire: util.Pointer(uint(0)),
		Groups: &[]proxmox.GroupName{},
	}
}

func User_Full_testData(testNumber uint) proxmox.ConfigUser {
	return proxmox.ConfigUser{
		User:      proxmox.UserID{Name: fmt.Sprintf("test-user%d", testNumber), Realm: "pve"},
		Comment:   util.Pointer("this is a comment"),
		Email:     util.Pointer("b.wayne@proxmox.com"),
		Enable:    util.Pointer(true),
		Expire:    util.Pointer(uint(253370811600)),
		FirstName: util.Pointer("Bruce"),
		Groups:    &[]proxmox.GroupName{proxmox.GroupName(fmt.Sprintf("user%d-group%d", testNumber, testNumber))},
		Keys:      util.Pointer("2fa key"),
		LastName:  util.Pointer("Wayne"),
	}
}
