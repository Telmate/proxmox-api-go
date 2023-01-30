package storagesubtests

import (
	"encoding/json"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
)

func CloneJson(jsonStruct proxmox.ConfigStorage) *proxmox.ConfigStorage {
	s := &proxmox.ConfigStorage{}
	ori, _ := json.Marshal(jsonStruct)
	json.Unmarshal(ori, s)
	return s
}

func InlineMarshal(jsonStruct *proxmox.ConfigStorage) string {
	j, _ := json.Marshal(jsonStruct)
	return string(j)
}
