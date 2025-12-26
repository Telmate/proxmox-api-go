package helpers

import (
	"strings"

	"github.com/Telmate/proxmox-api-go/proxmox"
)

// Converts an comma separated list of "username@realm" to a array of UserID objects
func ParseUserIDs(userIds string) (*[]proxmox.UserID, error) {
	if userIds == "" {
		return &[]proxmox.UserID{}, nil
	}
	tmpUserIds := strings.Split(userIds, ",")
	users := make([]proxmox.UserID, len(tmpUserIds))
	for i := range tmpUserIds {
		if err := users[i].Parse(tmpUserIds[i]); err != nil {
			return nil, err
		}
	}
	return &users, nil
}
