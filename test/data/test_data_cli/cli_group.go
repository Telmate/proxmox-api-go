package test_data_cli

import (
	"fmt"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
)

func Group_Get_Empty_testData(testNumber uint) proxmox.ConfigGroup {
	return proxmox.ConfigGroup{
		Name:    proxmox.GroupName(fmt.Sprintf("group%d", testNumber)),
		Comment: "",
		Members: &[]proxmox.UserID{}}
}

func Group_Get_Full_testData(testNumber uint) proxmox.ConfigGroup {
	return proxmox.ConfigGroup{
		Name:    proxmox.GroupName(fmt.Sprintf("group%d", testNumber)),
		Comment: "comment",
		Members: &[]proxmox.UserID{
			{Name: "root", Realm: "pam"},
			{Name: fmt.Sprintf("group%d-user%d0", testNumber, testNumber), Realm: "pve"},
		},
	}
}
